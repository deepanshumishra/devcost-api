package aws

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListResourcesByTags fetches AWS resources by tag key and value.
func ListResourcesByTags(cfg *config.Config, tagKey, tagValue string) ([]models.Resource, error) {
	// Initialize Resource Groups Tagging API client
	client := resourcegroupstaggingapi.NewFromConfig(cfg.AWSConfig)

	// Initialize slice
	resources := []models.Resource{}

	// Query resources by tag
	input := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []types.TagFilter{
			{
				Key:    aws.String(tagKey),
				Values: []string{tagValue},
			},
		},
	}

	var paginationToken *string
	for {
		input.PaginationToken = paginationToken
		result, err := client.GetResources(context.TODO(), input)
		if err != nil {
			log.Printf("Failed to get resources by tag %s=%s: %v", tagKey, tagValue, err)
			return resources, err
		}

		for _, resource := range result.ResourceTagMappingList {
			resources = append(resources, models.Resource{
				ResourceARN:  aws.ToString(resource.ResourceARN),
				ResourceType: getResourceTypeFromARN(aws.ToString(resource.ResourceARN)),
				Tags:         convertTags(resource.Tags),
			})
			log.Printf("Found resource %s with tag %s=%s", aws.ToString(resource.ResourceARN), tagKey, tagValue)
		}

		if result.PaginationToken == nil || *result.PaginationToken == "" {
			break
		}
		paginationToken = result.PaginationToken
	}

	return resources, nil
}

// getResourceTypeFromARN extracts the resource type from an ARN.
func getResourceTypeFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return "unknown"
	}
	service := parts[2]
	resource := parts[5]
	if service == "ec2" && strings.HasPrefix(resource, "instance/") {
		return "ec2:instance"
	}
	if service == "ec2" && strings.HasPrefix(resource, "volume/") {
		return "ebs:volume"
	}
	if service == "ec2" && strings.HasPrefix(resource, "eipalloc/") {
		return "ec2:elastic-ip"
	}
	if service == "rds" && strings.HasPrefix(resource, "db:") {
		return "rds:instance"
	}
	if service == "lambda" && strings.HasPrefix(resource, "function:") {
		return "lambda:function"
	}
	if service == "dynamodb" && strings.HasPrefix(resource, "table/") {
		return "dynamodb:table"
	}
	if service == "elasticloadbalancing" && strings.HasPrefix(resource, "loadbalancer/") {
		return "elasticloadbalancing:loadbalancer"
	}
	return service + ":" + strings.Split(resource, "/")[0]
}

// convertTags converts AWS tags to map.
func convertTags(tags []types.Tag) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return tagMap
}