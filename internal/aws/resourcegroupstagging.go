package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListResourcesByTags fetches AWS resources with the specified tag key and value.
func ListResourcesByTags(cfg *config.Config, tagKey, tagValue string) ([]models.Resource, error) {
	// Initialize Resource Groups Tagging API client
	client := resourcegroupstaggingapi.NewFromConfig(cfg.AWSConfig)

	var resources []models.Resource
	input := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []types.TagFilter{
			{
				Key:    &tagKey,
				Values: []string{tagValue},
			},
		},
	}

	for {
		// Call GetResources API
		result, err := client.GetResources(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		// Append resources to the list
		for _, rm := range result.ResourceTagMappingList {
			tags := make(map[string]string)
			for _, tag := range rm.Tags {
				tags[*tag.Key] = *tag.Value
			}
			resources = append(resources, models.Resource{
				ResourceARN:  *rm.ResourceARN,
				ResourceType: getResourceType(*rm.ResourceARN),
				Tags:         tags,
			})
		}

		// Check for pagination
		if result.PaginationToken == nil || *result.PaginationToken == "" {
			break
		}
		input.PaginationToken = result.PaginationToken
	}

	return resources, nil
}

// getResourceType extracts the resource type from the ARN (e.g., "ec2:instance").
func getResourceType(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 6 {
		resourceParts := strings.Split(parts[5], "/")
		if len(resourceParts) >= 2 {
			return parts[2] + ":" + resourceParts[0]
		}
		return parts[2] + ":" + resourceParts[0]
	}
	return "unknown"
}