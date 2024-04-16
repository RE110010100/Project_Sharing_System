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

### Setting Up Backend Services (on localhost)
To set up the backend services:
1. Clone the project repository:
   ```bash
   git clone git@github.com:RE110010100/Project_Sharing_System.git

2. For each backend service (fileservice, notificationservice, projectservice, searchservice), navigate to its directory and build :
   ```bash
   cd <service-directory>
   go run main.go


Setting Up the Frontend
To set up the frontend:

1. Navigate to the frontend directory:
   ```bash
   cd web/project_sharing_system_frontend

2. Install the necessary packages:
  ```bash
  npm install

3. Start the development server to run the frontend:
  ```bash
  npm start


Deployment on Minikube
Kubernetes Deployment

1. Ensure Minikube is active and configured to use Docker:

  ```bash
  minikube start
  eval $(minikube docker-env)

2. Build and tag Docker images for each backend service:

  ```bash
  docker build -t <service-name>:latest .

3. Deploy the services using Kubernetes:

  ```bash
  kubectl apply -f kubernetes/

4. The frontend can typically be accessed via the Minikube IP and the NodePort specified in the service's Kubernetes configuration:

  ```bash
  http://<minikube-ip>:<node-port>




