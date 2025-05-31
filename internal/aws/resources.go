package aws

import (
	"log"
	"time"

	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedResources fetches all unused paid AWS resources.
func ListUnusedResources(cfg *config.Config, start, end time.Time, unusedForDays int) ([]models.UnusedResource, error) {
	// Initialize to avoid nil response
	allResources := []models.UnusedResource{}

	// Get unused EC2 resources (instances, EBS volumes, Elastic IPs)
	ec2Resources, err := ListUnusedEC2Resources(cfg, start, end)
	if err != nil {
		log.Printf("Failed to list EC2 resources: %v", err)
	}
	allResources = append(allResources, ec2Resources...)

	// Get unused RDS resources
	rdsResources, err := ListUnusedRDSResources(cfg, start, end, unusedForDays)
	if err != nil {
		log.Printf("Failed to list RDS resources: %v", err)
	}
	allResources = append(allResources, rdsResources...)

	// Get unused Bedrock resources
	bedrockResources, err := ListUnusedBedrockResources(cfg, start, end, unusedForDays)
	if err != nil {
		log.Printf("Failed to list Bedrock resources: %v", err)
	}
	allResources = append(allResources, bedrockResources...)

	log.Printf("Returning %d unused paid resources", len(allResources))
	return allResources, nil
}