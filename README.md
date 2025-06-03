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

# Build and Push Container Images

## Manual Build

```bash
docker build --tag ghcr.io/oglimmer/s3fs-go:latest .
docker push ghcr.io/oglimmer/s3fs-go:latest
```

## Using GitHub Actions

This repository includes a GitHub Actions workflow that automatically builds and pushes the Docker image to GitHub Container Registry (ghcr.io) when you:

1. Push to the main/master branch
2. Create a release tag (v1.0.0, v1.2.3, etc.)

The workflow file is located at `.github/workflows/docker-build-push.yml`.

### Image Tags

The following tags will be created:
- Branch name (e.g., `main`, `develop`)
- Git SHA (short format)
- Semantic version tags when pushing a tag (e.g., `v1.0.0`, `1.0`)

### Using the Container Image

```bash
docker pull ghcr.io/oglimmer/s3fs-go:latest
docker run -p 8080:8080 -v $(pwd)/storage:/data ghcr.io/oglimmer/s3fs-go:latest
```
