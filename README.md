# Gator
A RSS feed aggregator.

## Requirements

- PostgreSQL 17.6
- Go 1.25

## Install

Install with Go:
```go
go install github.com/harryyu/gator
```

Create config file at ~/.gatorconfig.json:
```json
{
    "db_url": "<postgres_db_url>?sslmode=disable"
}
```

## Usage
Try running a few commands:
```bash
go run . register johndoe
go run . addfeed "Hacker News RSS" "https://hnrss.org/newest"
go run . agg 30s
```
(In another shell)
```bash
go run browse 10
```

