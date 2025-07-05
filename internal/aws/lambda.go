package aws

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedLambdaResources identifies unused Lambda functions.
func ListUnusedLambdaResources(cfg *config.Config, start, end time.Time, unusedForDays int) ([]models.UnusedResource, error) {
	// Initialize Lambda client
	client := lambda.NewFromConfig(cfg.AWSConfig)

	// Initialize slice
	unusedResources := []models.UnusedResource{}

	// List Lambda functions
	input := &lambda.ListFunctionsInput{}
	result, err := client.ListFunctions(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to list Lambda functions: %v", err)
		return unusedResources, err
	}

	// Check for unused Lambda functions
	for _, function := range result.Functions {
		isUnused, err := isLambdaFunctionUnused(cfg, aws.ToString(function.FunctionArn), start, end)
		if err != nil {
			log.Printf("Failed to check usage for Lambda %s: %v", aws.ToString(function.FunctionArn), err)
			continue
		}
		if isUnused {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "lambda:function",
				ResourceID:   aws.ToString(function.FunctionArn),
				Reason:       "No invocations for " + strconv.Itoa(unusedForDays) + " days",
			})
			log.Printf("Found unused Lambda function: %s", aws.ToString(function.FunctionArn))
		}
	}

	return unusedResources, nil
}

// isLambdaFunctionUnused checks if a Lambda function has no invocations.
func isLambdaFunctionUnused(cfg *config.Config, functionArn string, start, end time.Time) (bool, error) {
	client := cloudwatch.NewFromConfig(cfg.AWSConfig)

	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("invocations"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/Lambda"),
						MetricName: aws.String("Invocations"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("FunctionName"),
								Value: aws.String(functionArn),
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
		log.Printf("CloudWatch error for Lambda %s: %v", functionArn, err)
		return false, err
	}

	if len(result.MetricDataResults) == 0 || len(result.MetricDataResults[0].Values) == 0 {
		log.Printf("No invocation metrics for Lambda %s from %s to %s", functionArn, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return true, nil // No usage
	}

	for _, metricResult := range result.MetricDataResults {
		for i, value := range metricResult.Values {
			log.Printf("Lambda %s invocations at %s: %f", functionArn, metricResult.Timestamps[i].Format("2006-01-02 15:04:05"), value)
			if value > 0 {
				log.Printf("Lambda %s not unused (invocations: %f)", functionArn, value)
				return false, nil
			}
		}
	}

	log.Printf("Lambda %s is unused (no invocations) from %s to %s", functionArn, start.Format("2006-01-02"), end.Format("2006-01-02"))
	return true, nil
}