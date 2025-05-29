# devcost-api
Golang backend APIs for DevCost, a SaaS tool for AWS cost management and unused resource detection.

```plaintext
devcost-backend/
├── cmd/
│   └── api/
│       └── main.go          # Entry point for the API server
├── internal/
│   ├── api/                # API handlers and routes
│   │   ├── handlers/       # HTTP handlers (e.g., cost, resources, slack)
│   │   └── routes.go       # API route definitions
│   ├── aws/                # AWS SDK logic (Cost Explorer, CloudWatch, EC2)
│   ├── slack/              # Slack integration logic
│   └── models/             # Data models (e.g., User, Project, Resource)
├── pkg/
│   └── db/                 # Database connection and queries (PostgreSQL, Redis)
├── .gitignore              # Git ignore file (from Go template)
├── go.mod                  # Go module file
├── go.sum                  # Dependency checksums
├── README.md               # Project documentation
└── .github/
    └── workflows/
        └── ci.yml          # GitHub Actions CI/CD pipeline
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