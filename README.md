# s3fs-go

A simple S3-compatible filesystem API server written in Go that stores files on the local filesystem.

## Overview

s3fs-go provides a basic S3-compatible REST API that maps bucket/key operations to a local filesystem. It supports the core S3 operations: PUT (upload), GET (download), and DELETE.

## Features

- S3-compatible REST API endpoints
- Local filesystem storage backend
- Path traversal protection
- Docker support with multi-stage builds
- Configurable storage root directory

## API Endpoints

- `PUT /<bucket>/<key>` - Upload a file
- `GET /<bucket>/<key>` - Download a file  
- `DELETE /<bucket>/<key>` - Delete a file

## Usage

### Command Line

```bash
go run main.go <storage-root-path>
```

Example:
```bash
go run main.go ./storage
```

### Docker

Build and run with Docker:

```bash
docker build -t s3fs-go .
docker run -p 8080:8080 -v $(pwd)/storage:/app/storage s3fs-go
```

### Docker Compose

```bash
docker-compose up
```

The service will be available on `http://localhost:8081`.

## Examples

Upload a file:
```bash
curl -X PUT "http://localhost:8080/my-bucket/path/to/file.txt" \
     --data-binary @local-file.txt
```

Download a file:
```bash
curl "http://localhost:8080/my-bucket/path/to/file.txt" \
     -o downloaded-file.txt
```

Delete a file:
```bash
curl -X DELETE "http://localhost:8080/my-bucket/path/to/file.txt"
```

## Requirements

- Go 1.20 or later
- Docker (optional)

## Security

The server includes path traversal protection to prevent access outside the configured storage root directory.