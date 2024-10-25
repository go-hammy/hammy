# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# # Install PHP
# RUN apk add --no-cache php

# Copy the go.mod and go.sum files to the working directory
COPY go.mod ./

# Download the Go module dependencies
RUN go mod download
 
# Copy the rest of the application code to the working directory, excluding the content directory
COPY . /app
RUN rm -rf /app/content /app/cache /app/docker-compose.yaml /app/kubernetes.yaml /app/readme.md /app/examples


# Build the Go application
RUN go build -o main .

# Create the directory for mounting content
RUN mkdir -p /var/www/html && mkdir -p /var/log/hammy && mkdir -p /var/cache/hammy || true

# Expose the port that the application will run on
EXPOSE 9090

# Command to run the executable
CMD ["./main"]
