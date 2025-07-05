package aws

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedSecrets identifies secrets not accessed in the specified number of days.
func ListUnusedSecrets(ctx context.Context, cfg *config.Config, unusedForDays int) ([]models.UnusedResource, error) {
	client := secretsmanager.NewFromConfig(cfg.AWSConfig)
	var unusedSecrets []models.UnusedResource

	input := &secretsmanager.ListSecretsInput{}
	paginator := secretsmanager.NewListSecretsPaginator(client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("Failed to list secrets: %v", err)
			return nil, fmt.Errorf("failed to list secrets: %v", err)
		}

		for _, secret := range page.SecretList {
			// Skip secrets marked for deletion
			if secret.DeletedDate != nil {
				log.Printf("Skipping deleted secret %s", aws.ToString(secret.ARN))
				continue
			}

			// Check if secret is unused
			if secret.LastAccessedDate == nil || time.Since(*secret.LastAccessedDate).Hours()/24 > float64(unusedForDays) {
				unusedSecrets = append(unusedSecrets, models.UnusedResource{
					ResourceType: "secretsmanager:secret",
					ResourceID:   aws.ToString(secret.ARN),
					Reason:       "Not accessed in " + strconv.Itoa(unusedForDays) + " days",
				})
			}
		}
	}

	if len(unusedSecrets) == 0 {
		log.Printf("No unused secrets found for %d days", unusedForDays)
	}

	return unusedSecrets, nil
}