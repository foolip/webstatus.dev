# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

locals {
  service_dir = "frontend"
  service_sha = sha1(join(
    "",
    [filesha1(join("", [path.cwd, "/../images/nodejs_service.Dockerfile"]))],
    [for f in fileset(path.cwd, "/../${local.service_dir}/**") : filesha1(f)],
    [for f in fileset(path.cwd, "/../lib/**") : filesha1(f)]
  ))
}

resource "docker_image" "frontend" {
  name = "${var.docker_repository_details.url}/frontend"
  build {
    context = "${path.cwd}/.."
    build_args = {
      service_dir : local.service_dir
      nginx_config_filename : "nginx.local.conf"
    }
    target     = "static"
    dockerfile = "images/nodejs_service.Dockerfile"
  }
  triggers = {
    dir_sha1 = local.service_sha
  }
}

resource "docker_registry_image" "frontend_remote_image" {
  name          = docker_image.frontend.name
  keep_remotely = true
  triggers = {
    dir_sha1 = local.service_sha
  }
}


locals {
  tagged_oauth2_proxy_image = "${var.docker_repository_details.url}/oauth2_proxy:${var.oauth2_proxy_tag}"
}

resource "docker_image" "oauth2_proxy" {
  name         = "quay.io/oauth2-proxy/oauth2-proxy:${var.oauth2_proxy_tag}"
  keep_locally = false # No need to store it locally long-term
}

resource "null_resource" "retag_oauth2_proxy" {
  depends_on = [docker_image.oauth2_proxy] # Ensure the image is pulled first

  provisioner "local-exec" {
    command = "docker tag ${docker_image.oauth2_proxy.name} ${local.tagged_oauth2_proxy_image}"
  }
}

resource "docker_registry_image" "oauth2_proxy_gcr" {
  name = local.tagged_oauth2_proxy_image

  depends_on = [null_resource.retag_oauth2_proxy] # Make sure the image is re-tagged
}


resource "google_service_account" "frontend" {
  account_id   = "frontend-${var.env_id}"
  provider     = google.public_project
  display_name = "Frontend service account for ${var.env_id}"
}

data "google_project" "host_project" {
}

data "google_project" "datastore_project" {
}



resource "google_cloud_run_v2_service" "service" {
  for_each     = var.region_to_subnet_info_map
  provider     = google.public_project
  launch_stage = "BETA"
  name         = "${var.env_id}-${each.key}-webstatus-frontend"
  location     = each.key
  ingress      = "INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER"

  template {
    containers {
      image = "${docker_image.frontend.name}@${docker_registry_image.frontend_remote_image.sha256_digest}"
      ports {
        container_port = 5555
      }
      env {
        name  = "API_URL"
        value = var.backend_api_host
      }
      env {
        name  = "GSI_CLIENT_ID"
        value = var.gsi_client_id
      }
      env {
        name  = "PROJECT_ID"
        value = data.google_project.datastore_project.number
      }
    }
    # # oauth2-proxy Container
    # containers {
    #   image = docker_registry_image.oauth2_proxy_gcr.name
    #   args = [
    #     "--provider=google",
    #     "--client-id=${var.gsi_client_id}",
    #     # "--upstream=http://127.0.0.1:5555",
    #     "--redirect-url=https://website-webstatus-dev.corp.goog/oauth2/callback",
    #     "--reverse-proxy=true",
    #     "--email-domain=google.com"
    #   ]

    #   env {
    #     name = "OAUTH2_PROXY_CLIENT_SECRET"
    #     value_source {
    #       secret_key_ref {
    #         secret  = "projects/845410142324/secrets/staging-oauth-client-secret"
    #         version = "latest"
    #       }
    #     }
    #   }

    #   env {
    #     name = "OAUTH2_PROXY_COOKIE_SECRET"
    #     value_source {
    #       secret_key_ref {
    #         secret  = "projects/845410142324/secrets/staging-oauth-cookie-secret"
    #         version = "latest"
    #       }
    #     }
    #   }
    # }
    vpc_access {
      network_interfaces {
        network    = "projects/${data.google_project.host_project.name}/global/networks/${var.vpc_name}"
        subnetwork = "projects/${data.google_project.host_project.name}/regions/${each.key}/subnetworks/${each.value.public}"
      }
      egress = "ALL_TRAFFIC"
    }
    service_account = google_service_account.frontend.email
  }
}

resource "google_cloud_run_service_iam_member" "public" {
  provider = google.public_project
  for_each = google_cloud_run_v2_service.service
  location = each.value.location
  service  = each.value.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_compute_region_network_endpoint_group" "neg" {
  provider = google.public_project
  for_each = google_cloud_run_v2_service.service

  name                  = "${var.env_id}-frontend-neg-${each.value.location}"
  network_endpoint_type = "SERVERLESS"
  region                = each.value.location

  cloud_run {
    service = each.value.name
  }
  depends_on = [
    google_cloud_run_v2_service.service
  ]
}

resource "google_compute_backend_service" "lb_backend" {
  provider = google.public_project
  name     = "${var.env_id}-frontend-service"
  dynamic "backend" {
    for_each = google_compute_region_network_endpoint_group.neg
    content {
      group = backend.value.id
    }
  }
}

resource "google_compute_url_map" "url_map" {
  provider = google.public_project
  name     = "${var.env_id}-frontend-url-map"

  default_service = google_compute_backend_service.lb_backend.id
}

resource "google_compute_global_forwarding_rule" "https" {
  provider    = google.public_project
  name        = "${var.env_id}-frontend-https-rule"
  ip_protocol = "TCP"
  port_range  = "443"
  ip_address  = google_compute_global_address.ub_ip_address.id
  target      = google_compute_target_https_proxy.lb_https_proxy.id
}

resource "google_compute_global_address" "ub_ip_address" {
  provider = google.public_project
  name     = "${var.env_id}-frontend-ip"
}

resource "google_compute_target_https_proxy" "lb_https_proxy" {
  provider = google.public_project
  name     = "${var.env_id}-frontend-https-proxy"
  url_map  = google_compute_url_map.url_map.id
  ssl_certificates = [
    "ub-self-sign" # Temporary for UB
  ]
}
