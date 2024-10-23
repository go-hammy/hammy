# Hammy Project

Welcome to the Hammy Project! Hammy is a lightweight and efficient web server written in Go, designed to serve dynamic and static content with ease. This document will guide you through the structure of the project, how to get started quickly, and how to deploy it using Docker or directly on a server.

## Project Structure

The Hammy Project is organized into several key components:

- **serverPlugin**: This is the core of the Hammy server. It handles HTTP requests, serves files, and manages server operations. The serverPlugin is responsible for routing requests, executing PHP files, and serving HTML content.

- **cachePlugin**: This module is responsible for caching responses to improve performance. Cached responses are stored and served when the same request is made, reducing the need to repeatedly read files from disk.

- **content**: This directory is intended to be used with Docker or Docker Compose. It contains the static and dynamic content that the server will serve. When using Docker, this directory is mounted to `/var/www/html` inside the container.

- **serverPlugin/pages**: This folder contains custom error pages like `hammy-404.html` and `hammy-500.html`, which are served when the server encounters 404 or 500 errors, respectively. These are served when there is no 500.html or php in the original /var/www/html.

## Quickstart

### Using Docker

To quickly get started with Hammy using Docker, follow these steps:

1. **Build the Docker Image**:

   ```bash
   docker-compose build
   ```

2. **Run the Docker Container**:
   ```bash
   docker-compose up
   ```

This will start the Hammy server inside a Docker container, serving content from the `content` directory.

### Running on a Server

If you prefer to run Hammy directly on a server, ensure that `/var/www/html` is populated with your content. Then, follow these steps:

1. **Build the Go Application**:

   ```bash
   go build -o hammy
   ```

2. **Run the Server**:
   ```bash
   ./hammy
   ```

This will start the Hammy server on port 8080, ready to serve content from `/var/www/html`.

## Deployment Considerations

Hammy is designed to be run behind a cloud-based reverse proxy, such as those used in Kubernetes environments. This setup helps maintain its speed and lightweight nature. Hammy does not support SSL by default, but it can handle SSL termination when forwarded through a reverse proxy.

## Conclusion

The Hammy Project is designed to be simple yet powerful, providing a robust solution for serving web content. Whether you choose to deploy it using Docker or directly on a server, Hammy is ready to deliver fast and reliable performance. Enjoy using Hammy, and feel free to explore and customize it to suit your needs!
