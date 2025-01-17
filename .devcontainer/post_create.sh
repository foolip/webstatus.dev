#!/bin/bash
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

# Clean up minikube just in case to ensure a fresh cluster.
make minikube-delete

# Install oapi-codegen
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.15.0

# Install wrench CLI
go install github.com/cloudspannerecosystem/wrench@v1.7.0

# Install repo-wide npm tools
npm i --workspaces=false

# Generate files
make gen -B
