package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type Config struct {
	AWSConfig aws.Config
}

func NewConfig() (*Config, error) {
	var awsCfg aws.Config
	var err error

	// Try loading from environment variables first
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	region := os.Getenv("AWS_REGION")

	if accessKeyID != "" && secretAccessKey != "" {
		// Use static credentials from env vars
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(
				aws.NewCredentialsCache(
					credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken),
				),
			),
		)
	} else {
		// Fallback to shared config profile or IAM role
		awsCfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile("devcost-api-user"),
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	cfg := &Config{
		AWSConfig: awsCfg,
	}

	return cfg, nil
}