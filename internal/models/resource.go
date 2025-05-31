package models

type Resource struct {
	ResourceARN  string            `json:"resource_arn"`
	ResourceType string            `json:"resource_type"`
	Tags         map[string]string `json:"tags"`
}

type UnusedResource struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Reason       string `json:"reason"`
}