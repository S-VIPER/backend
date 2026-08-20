# backend

## Description

Key tools: 
openapi-gen
testcontainers



## Container run

```bash
docker compose up
```


## Run on host

```bash
cd gin-api
```

```bash
go mod download
```

```bash
go generate ./...
```

```bash
go build -o main cmd/main.go
```

```bash
export MONGODB_URI="mongodb://root:example@mongodb:27017" 
```

```bash
./main
```

