package models

type TagCost struct {
	TagKey      string          `json:"tag_key"`
	TagValue    string          `json:"tag_value"`
	Cost        float64         `json:"cost"`
	Currency    string          `json:"currency"`
	Resources   []ResourceCost  `json:"resources"`
	CreatorName string          `json:"creator_name,omitempty"` // Add for aws:createdBy
}

type ResourceCost struct {
	ResourceType string  `json:"resource_type"`
	ResourceID   string  `json:"resource_id"`
	Cost         float64 `json:"cost"`
}