# devcost-api
Golang backend APIs for DevCost, a SaaS tool for AWS cost management and unused resource detection.

```plaintext
devcost-api
│   ├── cmd
│   │   └── api
│   │       └── main.go
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── api
│   │   │   ├── handlers
│   │   │   │   ├── health_test.go
│   │   │   │   ├── health.go
│   │   │   │   ├── resources_test.go
│   │   │   │   ├── resources.go
│   │   │   │   └── users.go
│   │   │   └── routes.go
│   │   ├── aws
│   │   │   ├── bedrock.go
│   │   │   ├── cloudwatch.go
│   │   │   ├── cost.go
│   │   │   ├── dynamodb.go
│   │   │   ├── ec2.go
│   │   │   ├── elb.go
│   │   │   ├── iam.go
│   │   │   ├── lambda.go
│   │   │   ├── rds.go
│   │   │   ├── resourcegroupstagging.go
│   │   │   ├── resources.go
│   │   │   └── secretsmanager.go
│   │   ├── config
│   │   │   └── config.go
│   │   ├── models
│   │   │   ├── cost.go
│   │   │   ├── resource.go
│   │   │   └── user.go
│   │   └── slack
│   ├── pkg
│   │   └── db
│   │       └── redis.go
│   └── README.md
├── Dockerfile
└── UI
    └── web.html
```


## Setup
1. Install Go 1.21+: `go version`
2. Clone repo: `git clone https://github.com/yourusername/devcost-backend.git`
3. Install dependencies: `go mod tidy`
4. Run: `go run cmd/api/main.go`

## Features
- Per-project cost dashboards (AWS Cost Explorer)
- Unused resource detection (idle EC2 instances, unattached EBS volumes)
- Slack daily summaries

## Run inside docker
 ✗ docker run --rm -p 8080:8080 \
  -e AWS_REGION=ap-south-1 \
  -e AWS_ACCESS_KEY_ID=<access_key> \
  -e AWS_SECRET_ACCESS_KEY=<secret_key> \
  deepanshu1411/devcost-api:latest
