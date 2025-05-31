package aws

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/deepanshumishra/devcost-api/internal/config"
)

// isInstanceIdle checks if an EC2 instance has CPU utilization <20% for the given time range.
func isInstanceIdle(cfg *config.Config, instanceID string, start, end time.Time) (bool, error) {
	// Initialize CloudWatch client
	client := cloudwatch.NewFromConfig(cfg.AWSConfig)

	// Query CPU utilization
	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("cpu"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/EC2"),
						MetricName: aws.String("CPUUtilization"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("InstanceId"),
								Value: aws.String(instanceID),
							},
						},
					},
					Period: aws.Int32(3600), // 1-hour period
					Stat:   aws.String("Average"),
				},
			},
		},
		StartTime: &start,
		EndTime:   &end,
	}

	result, err := client.GetMetricData(context.TODO(), input)
	if err != nil {
		log.Printf("CloudWatch error for instance %s: %v", instanceID, err)
		return false, err
	}

	// Log metrics for debugging
	if len(result.MetricDataResults) == 0 || len(result.MetricDataResults[0].Values) == 0 {
		log.Printf("No CPU metrics for instance %s from %s to %s", instanceID, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return false, nil // No data, assume not idle
	}

	// Require at least 3 data points
	dataPoints := len(result.MetricDataResults[0].Values)
	if dataPoints < 3 {
		log.Printf("Insufficient CPU metrics (%d points) for instance %s from %s to %s", dataPoints, instanceID, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return false, nil
	}

	// Check if all CPU utilization values are <20%
	for _, metricResult := range result.MetricDataResults {
		for i, value := range metricResult.Values {
			log.Printf("Instance %s CPU at %s: %f%%", instanceID, metricResult.Timestamps[i].Format("2006-01-02 15:04:05"), value)
			if value >= 20.0 {
				log.Printf("Instance %s not idle (CPU %f%% >= 20%%)", instanceID, value)
				return false, nil
			}
		}
	}

	log.Printf("Instance %s is idle (<20%% CPU) from %s to %s with %d data points", instanceID, start.Format("2006-01-02"), end.Format("2006-01-02"), dataPoints)
	return true, nil
}