# go-garminconnect

Go port of the Garmin Connect API client

## Overview
This project is a Go port of the Python Garmin Connect API wrapper. It provides programmatic access to Garmin Connect data through a structured Go API.

## Getting Started

### Prerequisites
- Go 1.21+
- Docker

### Installation
```sh
git clone https://github.com/sstent/go-garminconnect
cd go-garminconnect
```

### Building and Running
```sh
# Build and run with Docker
cd docker
docker compose up -d --build

# Run tests
go test ./...
```

### Development
See [PORTING_PLAN.md](PORTING_PLAN.md) for implementation progress and [JUNIOR_ENGINEER_GUIDE.md](JUNIOR_ENGINEER_GUIDE.md) for contribution guidelines.

## Project Structure
```
├── cmd/         - Main application
├── internal/    - Internal packages
│   ├── api/     - API endpoint implementations
│   └── auth/    - Authentication handling
├── docker/      - Docker configuration
└── tests/       - Test files
```

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
