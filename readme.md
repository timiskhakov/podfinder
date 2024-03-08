# podfinder

## Requirements

- Go 1.22
- [golangci-lint](https://golangci-lint.run)

## Running Tests

```shell
mockgen -source=./app/itunes/httpclient.go -destination=./app/itunes/mock/mock_httpclient.go -package=mock
go test -v ./...
```

## Running Linter

```shell
golangci-lint run
```

## Running Application

Locally:
```shell
go run ./app
```

In a Docker container:

```shell
docker-compose up --build
```