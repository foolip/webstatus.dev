// Package repo_downloader provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package repo_downloader

// Defines values for ArchiveTypes.
const (
	TAR ArchiveTypes = "TAR"
)

// ArchiveTypes defines model for ArchiveTypes.
type ArchiveTypes string

// BasicErrorModel defines model for BasicErrorModel.
type BasicErrorModel struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// FileFilter defines model for FileFilter.
type FileFilter struct {
	Prefix *string `json:"prefix,omitempty"`
	Suffix *string `json:"suffix,omitempty"`
}

// GCSUploadReport defines model for GCSUploadReport.
type GCSUploadReport struct {
	Bucket     string    `json:"bucket"`
	Filenames  *[]string `json:"filenames,omitempty"`
	RepoPrefix string    `json:"repo_prefix"`
}

// TarInput defines model for TarInput.
type TarInput struct {
	// TarStripComponents The --strip-components flag for tar
	TarStripComponents *int         `json:"tar_strip_components,omitempty"`
	Type               ArchiveTypes `json:"type"`
}

// UploadDestinationReport defines model for UploadDestinationReport.
type UploadDestinationReport struct {
	Gcs *GCSUploadReport `json:"gcs,omitempty"`
}

// PostV1GithubComOwnerNameJSONBody defines parameters for PostV1GithubComOwnerName.
type PostV1GithubComOwnerNameJSONBody struct {
	Archive     TarInput      `json:"archive"`
	FileFilters *[]FileFilter `json:"file_filters,omitempty"`
}

// PostV1GithubComOwnerNameJSONRequestBody defines body for PostV1GithubComOwnerName for application/json ContentType.
type PostV1GithubComOwnerNameJSONRequestBody PostV1GithubComOwnerNameJSONBody