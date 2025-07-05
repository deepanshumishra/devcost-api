package aws

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedDynamoDBResources identifies unused DynamoDB tables.
func ListUnusedDynamoDBResources(cfg *config.Config, start, end time.Time, unusedForDays int) ([]models.UnusedResource, error) {
	// Initialize DynamoDB client
	client := dynamodb.NewFromConfig(cfg.AWSConfig)

	// Initialize slice
	unusedResources := []models.UnusedResource{}

	// List DynamoDB tables
	input := &dynamodb.ListTablesInput{}
	result, err := client.ListTables(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to list DynamoDB tables: %v", err)
		return unusedResources, err
	}

	// Check for unused tables
	for _, tableName := range result.TableNames {
		isUnused, err := isDynamoDBTableUnused(cfg, tableName, start, end)
		if err != nil {
			log.Printf("Failed to check usage for DynamoDB table %s: %v", tableName, err)
			continue
		}
		if isUnused {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "dynamodb:table",
				ResourceID:   tableName,
				Reason:       "No reads or writes for " + strconv.Itoa(unusedForDays) + " days",
			})
			log.Printf("Found unused DynamoDB table: %s", tableName)
		}
	}

	return unusedResources, nil
}

// isDynamoDBTableUnused checks if a DynamoDB table has no reads or writes.
func isDynamoDBTableUnused(cfg *config.Config, tableName string, start, end time.Time) (bool, error) {
	client := cloudwatch.NewFromConfig(cfg.AWSConfig)

	// Check ReadCapacityUnits and WriteCapacityUnits
	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("reads"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/DynamoDB"),
						MetricName: aws.String("ConsumedReadCapacityUnits"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("TableName"),
								Value: aws.String(tableName),
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Sum"),
				},
			},
			{
				Id: aws.String("writes"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/DynamoDB"),
						MetricName: aws.String("ConsumedWriteCapacityUnits"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("TableName"),
								Value: aws.String(tableName),
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Sum"),
				},
			},
		},
		StartTime: &start,
		EndTime:   &end,
	}

	result, err := client.GetMetricData(context.TODO(), input)
	if err != nil {
		log.Printf("CloudWatch error for DynamoDB table %s: %v", tableName, err)
		return false, err
	}

	for _, metricResult := range result.MetricDataResults {
		if len(metricResult.Values) == 0 {
			log.Printf("No metrics for DynamoDB table %s from %s to %s", tableName, start.Format("2006-01-02"), end.Format("2006-01-02"))
			continue
		}
		for i, value := range metricResult.Values {
			log.Printf("DynamoDB table %s %s at %s: %f", tableName, *metricResult.Id, metricResult.Timestamps[i].Format("2006-01-02 15:04:05"), value)
			if value > 0 {
				log.Printf("DynamoDB table %s not unused (%s: %f)", tableName, *metricResult.Id, value)
				return false, nil
			}
		}
	}

	log.Printf("DynamoDB table %s is unused (no reads or writes) from %s to %s", tableName, start.Format("2006-01-02"), end.Format("2006-01-02"))
	return true, nil
}