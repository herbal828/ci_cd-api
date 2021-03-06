# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:1.12

# Add Maintainer Info
LABEL maintainer="Hernan Balmes <herbal828@gmail.com>"

# Copy go mod and sum files
COPY go.mod go.sum /

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY api .

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD go run /app

# Set the Current Working Directory inside the container
WORKDIR /app