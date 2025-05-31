# devcost-api
Golang backend APIs for DevCost, a SaaS tool for AWS cost management and unused resource detection.

```plaintext
devcost-api/
├── cmd/
│   └── api/
│       └── main.go          # Updated to load AWS config
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── health.go    # Existing health check
│   │   │   └── costs.go     # New cost endpoint handler
│   │   └── routes.go        # Updated with /costs/projects
│   ├── aws/
│   │   └── costexplorer.go  # AWS Cost Explorer logic
│   ├── models/
│   │   └── cost.go          # Cost data model
│   ├── slack/               # Empty, for future
│   └── config/              # New: AWS and Redis config
│       └── config.go
├── pkg/
│   └── db/
│       └── redis.go         # Redis connection (optional)
├── .github/
│   └── workflows/
│       └── ci.yml           # Existing CI/CD
├── Dockerfile               # Existing
├── go.mod                   # Updated with new dependencies
├── README.md                # Existing
└── .env                     # New: Environment variables
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