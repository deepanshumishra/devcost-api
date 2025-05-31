package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
	"fmt"
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
			return nil, err
		}

		for _, secret := range page.SecretList {
			// Skip secrets marked for deletion
			if secret.DeletedDate != nil {
				continue
			}

			// Get secret details
			describeInput := &secretsmanager.DescribeSecretInput{
				SecretId: secret.ARN,
			}
			secretDetails, err := client.DescribeSecret(ctx, describeInput)
			if err != nil {
				return nil, err
			}

			// Check if secret is unused
			if secretDetails.LastAccessedDate == nil || time.Since(*secretDetails.LastAccessedDate).Hours()/24 > float64(unusedForDays) {
				unusedSecrets = append(unusedSecrets, models.UnusedResource{
					ResourceType: "secretsmanager:secret",
					ResourceID:   aws.ToString(secret.ARN),
					Reason:       "Not accessed in " + fmt.Sprint(unusedForDays) + " days",
				})
			}
		}
	}

	return unusedSecrets, nil
}