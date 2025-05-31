package aws

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedEC2Resources identifies unused EC2 instances, EBS volumes, and Elastic IPs.
func ListUnusedEC2Resources(cfg *config.Config, start, end time.Time) ([]models.UnusedResource, error) {
	// Initialize EC2 client
	client := ec2.NewFromConfig(cfg.AWSConfig)

	// Initialize slice to avoid nil
	unusedResources := []models.UnusedResource{}

	// List EC2 instances
	ec2Input := &ec2.DescribeInstancesInput{}
	ec2Result, err := client.DescribeInstances(context.TODO(), ec2Input)
	if err != nil {
		log.Printf("Failed to describe EC2 instances: %v", err)
		return unusedResources, err
	}

	// Check for idle EC2 instances
	for _, reservation := range ec2Result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State.Name != types.InstanceStateNameRunning {
				continue
			}
			isIdle, err := isInstanceIdle(cfg, aws.ToString(instance.InstanceId), start, end)
			if err != nil {
				log.Printf("Failed to check idle status for instance %s: %v", aws.ToString(instance.InstanceId), err)
				continue
			}
			if isIdle {
				unusedResources = append(unusedResources, models.UnusedResource{
					ResourceType: "ec2:instance",
					ResourceID:   aws.ToString(instance.InstanceId),
					Reason:       "CPU utilization <20% for 7 days",
				})
				log.Printf("Found unused EC2 instance: %s", aws.ToString(instance.InstanceId))
			}
		}
	}

	// List EBS volumes
	volumeInput := &ec2.DescribeVolumesInput{}
	volumeResult, err := client.DescribeVolumes(context.TODO(), volumeInput)
	if err != nil {
		log.Printf("Failed to describe EBS volumes: %v", err)
		return unusedResources, err
	}

	// Check for unattached EBS volumes
	for _, volume := range volumeResult.Volumes {
		if len(volume.Attachments) == 0 {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "ebs:volume",
				ResourceID:   aws.ToString(volume.VolumeId),
				Reason:       "Unattached",
			})
			log.Printf("Found unused EBS volume: %s", aws.ToString(volume.VolumeId))
		}
	}

	// List Elastic IPs
	eipInput := &ec2.DescribeAddressesInput{}
	eipResult, err := client.DescribeAddresses(context.TODO(), eipInput)
	if err != nil {
		log.Printf("Failed to describe Elastic IPs: %v", err)
		return unusedResources, err
	}

	// Check for unassociated Elastic IPs
	for _, address := range eipResult.Addresses {
		if address.AssociationId == nil {
			unusedResources = append(unusedResources, models.UnusedResource{
				ResourceType: "ec2:elastic-ip",
				ResourceID:   aws.ToString(address.AllocationId),
				Reason:       "Not associated with any resource",
			})
			log.Printf("Found unused Elastic IP: %s", aws.ToString(address.AllocationId))
		}
	}

	return unusedResources, nil
}