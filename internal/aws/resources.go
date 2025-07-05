package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedResources fetches all unused paid AWS resources.
func ListUnusedResources(ctx context.Context, cfg *config.Config, start, end time.Time, unusedForDays int) ([]models.UnusedResource, error) {
	// Initialize to avoid nil response
	allResources := []models.UnusedResource{}
	var errors []error

	// Get unused EC2 resources (instances, EBS volumes, Elastic IPs)
	ec2Resources, err := ListUnusedEC2Resources(cfg, start, end)
	if err != nil {
		log.Printf("Failed to list EC2 resources: %v", err)
		errors = append(errors, err)
	} else {
		allResources = append(allResources, ec2Resources...)
	}

	// Get unused RDS resources
	rdsResources, err := ListUnusedRDSResources(cfg, start, end, unusedForDays)
	if err != nil {
		log.Printf("Failed to list RDS resources: %v", err)
		errors = append(errors, err)
	} else {
		allResources = append(allResources, rdsResources...)
	}

	// Get unused Bedrock resources
	bedrockResources, err := ListUnusedBedrockResources(cfg, start, end, unusedForDays)
	if err != nil {
		log.Printf("Failed to list Bedrock resources: %v", err)
		errors = append(errors, err)
	} else {
		allResources = append(allResources, bedrockResources...)
	}

	// Get unused Lambda resources
	lambdaResources, err := ListUnusedLambdaResources(cfg, start, end, unusedForDays)
	if err != nil {
		log.Printf("Failed to list Lambda resources: %v", err)
		errors = append(errors, err)
	} else {
		allResources = append(allResources, lambdaResources...)
	}

	// Get unused DynamoDB resources
	dynamoDBResources, err := ListUnusedDynamoDBResources(cfg, start, end, unusedForDays)
	if err != nil {
		log.Printf("Failed to list DynamoDB resources: %v", err)
		errors = append(errors, err)
	} else {
		allResources = append(allResources, dynamoDBResources...)
	}

	// Get unused Secrets Manager resources
	secretResources, err := ListUnusedSecrets(ctx, cfg, unusedForDays)
	if err != nil {
		log.Printf("Failed to list Secrets Manager resources: %v", err)
		errors = append(errors, err)
	} else {
		allResources = append(allResources, secretResources...)
	}

	// Get unused S3 buckets
	s3Client := s3.NewFromConfig(cfg.AWSConfig)
	cloudwatchClient := cloudwatch.NewFromConfig(cfg.AWSConfig)
	s3Resp, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Printf("Failed to list S3 buckets: %v", err)
		errors = append(errors, err)
	} else {
		for _, bucket := range s3Resp.Buckets {
			metrics, err := cloudwatchClient.GetMetricData(ctx, &cloudwatch.GetMetricDataInput{
				MetricDataQueries: []types.MetricDataQuery{
					{
						Id: aws.String("requests"),
						MetricStat: &types.MetricStat{
							Metric: &types.Metric{
								Namespace:  aws.String("AWS/S3"),
								MetricName: aws.String("AllRequests"),
								Dimensions: []types.Dimension{
									{Name: aws.String("BucketName"), Value: bucket.Name},
									{Name: aws.String("FilterId"), Value: aws.String("EntireBucket")},
								},
							},
							Period: aws.Int32(86400),
							Stat:   aws.String("Sum"),
						},
					},
				},
				StartTime: aws.Time(start),
				EndTime:   aws.Time(end),
			})
			if err != nil {
				log.Printf("Failed to get metrics for S3 bucket %s: %v", *bucket.Name, err)
				continue
			}
			if len(metrics.MetricDataResults) == 0 || len(metrics.MetricDataResults[0].Values) == 0 {
				allResources = append(allResources, models.UnusedResource{
					ResourceType: "s3:bucket",
					ResourceID:   *bucket.Name,
					Reason:       "No requests for " + fmt.Sprintf("%d days", unusedForDays),
				})
			}
		}
	}

	// If no resources and errors occurred, return an error
	if len(allResources) == 0 && len(errors) > 0 {
		log.Printf("No unused resources found, with %d errors", len(errors))
		return nil, fmt.Errorf("failed to list unused resources: %v errors occurred", len(errors))
	}

	log.Printf("Returning %d unused paid resources", len(allResources))
	return allResources, nil
}