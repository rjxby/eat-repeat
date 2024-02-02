# Project: Eat Repeat

## Overview

Eat Repeat is a pet project featuring a full-stack web application. It consists of a backend API written in Go and a frontend using htmx and Go templates. The project revolves around managing recipes, scheduling them for the week, and providing an interface to view the selected recipes for the current week.

## Features

### Backend (Go)

#### Docker Setup

The backend includes a Docker setup for containerization. The Dockerfile is divided into build stages for frontend and backend.

- **Frontend Build Stage (Node.js):**
  - Uses Node.js to build the frontend assets.
  - Copies necessary files and runs npm install and npm run build.

- **Backend Build Stage (Go):**
  - Utilizes Go for the backend.
  - Handles dependencies using Go modules.
  - Incorporates Alpine Linux as the base image.
  - Enables CGO for the Go build.
  - Supports versioning using Git information and Drone CI/CD environment variables.
  - Includes a mechanism to run migrations if the `RUN_MIGRATION` environment variable is set to `true`.
  - Supports an `.env` file for settings like `RUN_MIGRATION`, `PDF_READER_ENDPOINT`, and `WORKER_TIMEOUT_IN_SECONDS`.

- **Final Stage (Alpine):**
  - Creates a lightweight Alpine-based image for production.
  - Exposes port 8080 for communication.
  - Sets environment variables, including `RUN_MIGRATION`, which controls whether migrations should be run, and other configurable settings.
  - Defines the entry point command for the service.

### Frontend (HTML, htmx, Node.js)

- **Two Main Pages:**
  - **Recipes:** Displays a list of recipes.
  - **Current Week's Selected Recipes:** Shows recipes scheduled for the current week.

- **Scheduling Recipes:**
  - Allows users to schedule recipes for a specific week.

- **Docker Container:**
  - The frontend build is containerized using Node.js.

## CI/CD Pipeline (Drone)

The project has a CI/CD pipeline set up using Drone. The pipeline triggers on pushes to the main branch and feature branches, as well as pull requests. The key steps include:

- **Build and Deploy:**
  - Uses Docker plugins to build and deploy the application.
  - Retrieves environment variables such as `DRONE`, `DRONE_TAG`, `DRONE_COMMIT`, `DRONE_BRANCH`, and configurable settings from secrets.
  - Tags Docker images based on the branch name.
  - Pushes the images to the specified Docker registry.

## How to Run

To run the project locally, follow these steps:

1. Clone the repository.
2. Navigate to the project root.
3. Build the Docker image:

    ```bash
    docker build -t eat-repeat .
    ```

4. Run the Docker container:

    ```bash
    docker run -p 8080:8080 -e RUN_MIGRATION=true -e PDF_READER_ENDPOINT=http://example.com -e WORKER_TIMEOUT_IN_SECONDS=30 eat-repeat
    ```

5. Access the application in your browser at [http://localhost:8080](http://localhost:8080).

Feel free to explore and enhance the project as needed. If you have any questions or suggestions, please let me know!
