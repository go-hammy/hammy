![Hammy Logo](https://gohammy.org/docs/hammy.svg)

# Hammy Project

Welcome to the Hammy Project! Hammy is a lightweight and efficient web server written in Go, designed to serve dynamic and static content with ease. This document will guide you through the structure of the project, how to get started quickly, and how to deploy it using Docker or directly on a server. For more detailed information, visit our [official website](https://gohammy.org) and check out the [documentation](https://gohammy.org/docs).

[![Docker Hub](https://img.shields.io/badge/Docker%20Hub-View%20Image-blue)](https://hub.docker.com/r/gohammy/hammy)

## Quickstart

Get started quickly via https://gohammy.org/getting-started

## Deployment Considerations

Hammy is optimally configured to operate behind a cloud-based reverse proxy, commonly utilized in Kubernetes environments. This configuration ensures that Hammy remains both fast and lightweight. While Hammy does not natively support SSL, it is capable of managing SSL termination when routed through a reverse proxy. The choice to expose Hammy on port 9090, rather than the standard port 80, underscores the importance of always running it behind a reverse proxy for enhanced security and performance.

## Conclusion

The Hammy Project is designed to be simple yet powerful, providing a robust solution for serving web content. Whether you choose to deploy it using Docker or directly on a server, Hammy is ready to deliver fast and reliable performance. Enjoy using Hammy, and feel free to explore and customize it to suit your needs. For further guidance, refer to our [documentation](https://gohammy.org/docs) and [official website](https://gohammy.org).
