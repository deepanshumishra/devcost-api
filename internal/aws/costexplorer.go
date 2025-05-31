package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

func GetProjectCosts(cfg *config.Config) ([]models.ProjectCost, error) {
	// Initialize Cost Explorer client
	client := costexplorer.NewFromConfig(cfg.AWSConfig)

	// Set time range (last 7 days)
	end := time.Now()
	start := end.AddDate(0, 0, -7)

	// Query Cost Explorer
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
				Key:  aws.String("project"),
			},
		},
	}

	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	// Parse results
	var costs []models.ProjectCost
	for _, group := range result.ResultsByTime[0].Groups {
		project := group.Keys[0]
		amount := *group.Metrics["UnblendedCost"].Amount
		costs = append(costs, models.ProjectCost{
			Project:  project,
			Cost:     amount,
			Currency: *group.Metrics["UnblendedCost"].Unit,
		})
	}

	return costs, nil
}