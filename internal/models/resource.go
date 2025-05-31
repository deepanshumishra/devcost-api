package models

type Resource struct {
	ResourceARN  string            `json:"resource_arn"`
	ResourceType string            `json:"resource_type"`
	Tags         map[string]string `json:"tags"`
}