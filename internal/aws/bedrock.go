package aws

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedBedrockResources identifies unused Bedrock models and knowledge bases.
func ListUnusedBedrockResources(cfg *config.Config, start, end time.Time, unusedForDays int) ([]models.UnusedResource, error) {
	// Initialize Bedrock client for custom models
	bedrockClient := bedrock.NewFromConfig(cfg.AWSConfig)
	// Initialize Bedrock Agent client for knowledge bases
	agentClient := bedrockagent.NewFromConfig(cfg.AWSConfig)

	// Initialize slice
	unusedResources := []models.UnusedResource{}

	// List custom models (fine-tuned models, paid)
	modelInput := &bedrock.ListCustomModelsInput{}
	modelResult, err := bedrockClient.ListCustomModels(context.TODO(), modelInput)
	if err != nil {
		log.Printf("Failed to list Bedrock custom models: %v", err)
		return unusedResources, err
	}

	// Check for unused custom models
	for _, model := range modelResult.ModelSummaries {
		isUnused, err := isBedrockModelUnused(cfg, aws.ToString(model.ModelArn), start, end)
		if err != nil {
			log.Printf("Failed to check usage for Bedrock model %s: %v", aws.ToString(model.ModelArn), err)
			continue
		}
		if isUnused {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "bedrock:custom-model",
				ResourceID:   aws.ToString(model.ModelArn),
				Reason:       "No inference calls for " + strconv.Itoa(unusedForDays) + " days",
			})
			log.Printf("Found unused Bedrock custom model: %s", aws.ToString(model.ModelArn))
		}
	}

	// List knowledge bases
	kbInput := &bedrockagent.ListKnowledgeBasesInput{}
	kbResult, err := agentClient.ListKnowledgeBases(context.TODO(), kbInput)
	if err != nil {
		log.Printf("Failed to list Bedrock knowledge bases: %v", err)
		return unusedResources, err
	}

	// Check for unused knowledge bases
	for _, kb := range kbResult.KnowledgeBaseSummaries {
		isUnused, err := isBedrockKBUnused(cfg, aws.ToString(kb.KnowledgeBaseId), start, end)
		if err != nil {
			log.Printf("Failed to check usage for Bedrock KB %s: %v", aws.ToString(kb.KnowledgeBaseId), err)
			continue
		}
		if isUnused {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "bedrock:knowledge-base",
				ResourceID:   aws.ToString(kb.KnowledgeBaseId),
				Reason:       "No queries for " + strconv.Itoa(unusedForDays) + " days",
			})
			log.Printf("Found unused Bedrock knowledge base: %s", aws.ToString(kb.KnowledgeBaseId))
		}
	}

	return unusedResources, nil
}

// isBedrockModelUnused checks if a Bedrock model has no inference calls.
func isBedrockModelUnused(cfg *config.Config, modelArn string, start, end time.Time) (bool, error) {
	client := cloudwatch.NewFromConfig(cfg.AWSConfig)

	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("invocations"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/Bedrock"),
						MetricName: aws.String("Invocations"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("ModelId"),
								Value: aws.String(modelArn),
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
		log.Printf("CloudWatch error for Bedrock model %s: %v", modelArn, err)
		return false, err
	}

	if len(result.MetricDataResults) == 0 || len(result.MetricDataResults[0].Values) == 0 {
		log.Printf("No invocation metrics for Bedrock model %s from %s to %s", modelArn, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return true, nil // No usage
	}

	for _, metricResult := range result.MetricDataResults {
		for i, value := range metricResult.Values {
			log.Printf("Bedrock model %s invocations at %s: %f", modelArn, metricResult.Timestamps[i].Format("2006-01-02 15:04:05"), value)
			if value > 0 {
				log.Printf("Bedrock model %s not unused (invocations: %f)", modelArn, value)
				return false, nil
			}
		}
	}

	log.Printf("Bedrock model %s is unused (no invocations) from %s to %s", modelArn, start.Format("2006-01-02"), end.Format("2006-01-02"))
	return true, nil
}

// isBedrockKBUnused checks if a Bedrock knowledge base has no queries.
func isBedrockKBUnused(cfg *config.Config, kbID string, start, end time.Time) (bool, error) {
	client := cloudwatch.NewFromConfig(cfg.AWSConfig)

	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("queries"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/Bedrock"),
						MetricName: aws.String("KnowledgeBaseQueries"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("KnowledgeBaseId"),
								Value: aws.String(kbID),
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
		log.Printf("CloudWatch error for Bedrock KB %s: %v", kbID, err)
		return false, err
	}

	if len(result.MetricDataResults) == 0 || len(result.MetricDataResults[0].Values) == 0 {
		log.Printf("No query metrics for Bedrock KB %s from %s to %s", kbID, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return true, nil // No usage
	}

	for _, metricResult := range result.MetricDataResults {
		for i, value := range metricResult.Values {
			log.Printf("Bedrock KB %s queries at %s: %f", kbID, metricResult.Timestamps[i].Format("2006-01-02 15:04:05"), value)
			if value > 0 {
				log.Printf("Bedrock KB %s not unused (queries: %f)", kbID, value)
				return false, nil
			}
		}
	}

	log.Printf("Bedrock KB %s is unused (no queries) from %s to %s", kbID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	return true, nil
}