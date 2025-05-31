package aws

import (
	"context"
	"log"
	"time"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedRDSResources identifies unused RDS instances.
func ListUnusedRDSResources(cfg *config.Config, start, end time.Time, unusedForDays int) ([]models.UnusedResource, error) {
	// Initialize RDS client
	client := rds.NewFromConfig(cfg.AWSConfig)

	// Initialize slice
	unusedResources := []models.UnusedResource{}

	// List RDS instances
	input := &rds.DescribeDBInstancesInput{}
	result, err := client.DescribeDBInstances(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to describe RDS instances: %v", err)
		return unusedResources, err
	}

	// Check for stopped or idle RDS instances
	for _, db := range result.DBInstances {
		// Stopped instances
		if aws.ToString(db.DBInstanceStatus) == "stopped" {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "rds:instance",
				ResourceID:   aws.ToString(db.DBInstanceIdentifier),
				Reason:       "Stopped",
			})
			log.Printf("Found unused RDS instance (stopped): %s", aws.ToString(db.DBInstanceIdentifier))
			continue
		}

		// Only check running instances
		if aws.ToString(db.DBInstanceStatus) != "available" {
			continue
		}

		// Check CPU utilization
		isIdle, err := isRDSInstanceIdle(cfg, aws.ToString(db.DBInstanceIdentifier), start, end)
		if err != nil {
			log.Printf("Failed to check idle status for RDS %s: %v", aws.ToString(db.DBInstanceIdentifier), err)
			continue
		}
		if isIdle {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "rds:instance",
				Reason:       "CPU utilization <5% for " + strconv.Itoa(unusedForDays) + " days",
			})
			log.Printf("Found unused RDS instance (idle): %s", aws.ToString(db.DBInstanceIdentifier))
		}
	}

	return unusedResources, nil
}

// isRDSInstanceIdle checks if an RDS instance has CPU utilization <5%.
func isRDSInstanceIdle(cfg *config.Config, dbInstanceID string, start, end time.Time) (bool, error) {
	client := cloudwatch.NewFromConfig(cfg.AWSConfig)

	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("cpu"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  aws.String("AWS/RDS"),
						MetricName: aws.String("CPUUtilization"),
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("DBInstanceIdentifier"),
								Value: aws.String(dbInstanceID),
							},
						},
					},
					Period: aws.Int32(3600),
					Stat:   aws.String("Average"),
				},
			},
		},
		StartTime: &start,
		EndTime:   &end,
	}

	result, err := client.GetMetricData(context.TODO(), input)
	if err != nil {
		log.Printf("CloudWatch error for RDS %s: %v", dbInstanceID, err)
		return false, err
	}

	if len(result.MetricDataResults) == 0 || len(result.MetricDataResults[0].Values) == 0 {
		log.Printf("No CPU metrics for RDS %s from %s to %s", dbInstanceID, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return false, nil
	}

	if len(result.MetricDataResults[0].Values) < 3 {
		log.Printf("Insufficient CPU metrics (%d points) for RDS %s", len(result.MetricDataResults[0].Values), dbInstanceID)
		return false, nil
	}

	for _, metricResult := range result.MetricDataResults {
		for i, value := range metricResult.Values {
			log.Printf("RDS %s CPU at %s: %f%%", dbInstanceID, metricResult.Timestamps[i].Format("2006-01-02 15:04:05"), value)
			if value >= 20.0 {
				log.Printf("RDS %s not idle (CPU %f%% >= 5%%)", dbInstanceID, value)
				return false, nil
			}
		}
	}

	log.Printf("RDS %s is idle (<5%% CPU) from %s to %s", dbInstanceID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	return true, nil
}