package config

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type Config struct {
	AWSConfig aws.Config
}

func NewConfig() (*Config, error) {
	// Load AWS config with custom profile and region
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile("devcost-api-user"), // Use custom profile
		config.WithRegion(os.Getenv("AWS_REGION")),         // Load region from .env
	)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AWSConfig: awsCfg,
	}

	return cfg, nil
}