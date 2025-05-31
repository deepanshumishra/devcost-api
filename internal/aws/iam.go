package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/deepanshumishra/devcost-api/internal/config"
)

// ListIAMUsernames fetches all IAM usernames from the AWS account.
func ListIAMUsernames(cfg *config.Config) ([]string, error) {
	// Initialize IAM client
	client := iam.NewFromConfig(cfg.AWSConfig)

	var usernames []string
	var marker *string

	for {
		// Call ListUsers API
		input := &iam.ListUsersInput{Marker: marker}
		result, err := client.ListUsers(context.TODO(), input)
		if err != nil {
			return nil, err
		}

		// Append usernames to the list
		for _, user := range result.Users {
			usernames = append(usernames, *user.UserName)
		}

		// Check for pagination
		if !result.IsTruncated {
			break
		}
		marker = result.Marker
	}

	return usernames, nil
}