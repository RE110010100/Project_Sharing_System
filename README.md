# Project Sharing System

## Introduction
The Project Sharing System is a comprehensive solution designed to enhance collaboration and project management within organizations. This system is composed of multiple services including file management, notifications, project handling, and search functionalities, coupled with a robust frontend interface.

## Project Structure
The system is structured into several services and a frontend, detailed as follows:
- **`fileservice`**: Manages the storage and retrieval of project files.
- **`notificationservice`**: Handles the distribution of notifications to users.
- **`projectservice`**: Responsible for various project-related operations.
- **`searchservice`**: Provides robust search capabilities within the project environment.
- **`web`**:
  - **`project_sharing_system_frontend`**: The frontend part of the system built using ReactJS. It is located under the `web` directory.
    - **`src`**:
      - **`components`**: Contains React components used across the frontend application.
- **`kubernetes`**: Contains all necessary Kubernetes YAML files for deploying the system on Minikube.

## Prerequisites
To run and deploy the Project Sharing System, you will need:
- [Go](https://golang.org/doc/install): For backend service development.
- [Node.js and npm](https://nodejs.org/en/download/): For frontend development.
- [Minikube](https://minikube.sigs.k8s.io/docs/start/): For local deployment using Kubernetes.
- [Docker](https://docs.docker.com/get-docker/): For creating containers for the services.

## Installation

### Setting Up Backend Services
To set up the backend services:
1. Clone the project repository:
   ```bash
   git clone 