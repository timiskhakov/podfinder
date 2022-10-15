# podfinder

## Running Tests

```shell
mockgen -source=./app/itunes/httpclient.go -destination=./app/itunes/mock/mock_httpclient.go -package=mock
go test -v ./...
```

## Running Locally

Run the following command:
```shell
go run ./app
```

## Running in Docker

Run the following command:
```shell
docker-compose up --build
```