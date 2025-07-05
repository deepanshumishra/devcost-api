package aws

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// GetTagCosts fetches costs by a specified tag key, aggregated over the date range.
func GetTagCosts(cfg *config.Config, tagKey string, start, end time.Time) ([]models.TagCost, error) {
	// Initialize clients
	client := costexplorer.NewFromConfig(cfg.AWSConfig)
	iamClient := iam.NewFromConfig(cfg.AWSConfig)

	// Validate tag is active in cost allocation tags
	listTagsOutput, err := client.ListCostAllocationTags(context.TODO(), &costexplorer.ListCostAllocationTagsInput{})
	if err != nil {
		log.Printf("Failed to list cost allocation tags: %v", err)
		return nil, fmt.Errorf("failed to validate tag '%s': %v", tagKey, err)
	}

	isActive := false
	for _, tag := range listTagsOutput.CostAllocationTags {
		if aws.ToString(tag.TagKey) == tagKey {
			isActive = true
			log.Printf("Tag '%s' is active in cost allocation tags (Type: %s, Status: %s)", tagKey, tag.Type, tag.Status)
			break
		}
	}
	if !isActive {
		log.Printf("Tag '%s' is not active in cost allocation tags", tagKey)
		return nil, fmt.Errorf("tag '%s' is not active in cost allocation tags", tagKey)
	}

	// Initialize map to aggregate costs by tag value
	tagCostMap := make(map[string]struct {
		TotalCost   float64
		Resources   map[string]float64
		CreatorName string
	})

	// Query Cost Explorer for costs by tag and service
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start.Format("2006-01-02")),
			End:   aws.String(end.Format("2006-01-02")),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeTag,
				Key:  aws.String(tagKey),
			},
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to get cost and usage for tag %s: %v", tagKey, err)
		return nil, fmt.Errorf("failed to fetch costs for tag '%s': %v", tagKey, err)
	}

	// Log raw response
	log.Printf("Cost Explorer response for tag %s: %+v", tagKey, result)

	// Resolve IAM names for aws:createdBy
	getCreatorName := func(tagValue string) string {
		if tagKey != "aws:createdBy" {
			return ""
		}
		parts := strings.Split(tagValue, ":")
		if len(parts) < 2 {
			return tagValue
		}
		principalType, principalID := parts[0], parts[1]
		switch principalType {
		case "IAMUser":
			users, err := iamClient.ListUsers(context.TODO(), &iam.ListUsersInput{})
			if err != nil {
				log.Printf("Failed to list IAM users for %s: %v", tagValue, err)
				return tagValue
			}
			for _, user := range users.Users {
				if aws.ToString(user.UserId) == principalID {
					return aws.ToString(user.UserName)
				}
			}
		case "AssumedRole":
			roles, err := iamClient.ListRoles(context.TODO(), &iam.ListRolesInput{})
			if err != nil {
				log.Printf("Failed to list IAM roles for %s: %v", tagValue, err)
				return tagValue
			}
			for _, role := range roles.Roles {
				if aws.ToString(role.RoleId) == principalID {
					return aws.ToString(role.RoleName)
				}
			}
		case "Root":
			return "Root Account"
		}
		return tagValue
	}

	// Aggregate costs
	for _, group := range result.ResultsByTime {
		for _, metric := range group.Groups {
			tagValue := metric.Keys[0]
			serviceName := metric.Keys[1]
			// Clean tag value
			if strings.HasPrefix(tagValue, tagKey+"$") {
				log.Printf("Cleaning tag value %s to %s for tag %s", tagValue, strings.TrimPrefix(tagValue, tagKey+"$"), tagKey)
				tagValue = strings.TrimPrefix(tagValue, tagKey+"$")
			}
			if tagValue == "" || tagValue == tagKey+"$" {
				log.Printf("Skipping untagged or empty value for tag %s=%s, service %s", tagKey, tagValue, serviceName)
				continue
			}
			costAmount, err := strconv.ParseFloat(aws.ToString(metric.Metrics["UnblendedCost"].Amount), 64)
			if err != nil {
				log.Printf("Failed to parse cost for tag %s=%s, service %s: %v", tagKey, tagValue, serviceName, err)
				continue
			}
			data, exists := tagCostMap[tagValue]
			if !exists {
				data = struct {
					TotalCost   float64
					Resources   map[string]float64
					CreatorName string
				}{Resources: make(map[string]float64)}
			}
			data.TotalCost += costAmount
			data.Resources[serviceName] += costAmount
			data.CreatorName = getCreatorName(tagValue)
			tagCostMap[tagValue] = data
			log.Printf("Aggregated cost for tag %s=%s, service %s on %s: %f", tagKey, tagValue, serviceName, *group.TimePeriod.Start, costAmount)
		}
	}

	// Convert to slice
	costs := []models.TagCost{}
	for tagValue, data := range tagCostMap {
		resources := []models.ResourceCost{}
		for serviceName, cost := range data.Resources {
			resources = append(resources, models.ResourceCost{
				ResourceType: serviceName,
				ResourceID:   "",
				Cost:         cost,
			})
		}
		costs = append(costs, models.TagCost{
			TagKey:      tagKey,
			TagValue:    tagValue,
			Cost:        data.TotalCost,
			Currency:    "USD",
			Resources:   resources,
			CreatorName: data.CreatorName,
		})
		log.Printf("Total cost for tag %s=%s: %f USD, Creator: %s", tagKey, tagValue, data.TotalCost, data.CreatorName)
	}

	if len(costs) == 0 {
		log.Printf("No costs found for tag %s from %s to %s", tagKey, start.Format("2006-01-02"), end.Format("2006-01-02"))
	}

	return costs, nil
}