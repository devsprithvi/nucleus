Infrastructure & Deployment Architecture

1. Executive Summary

This document outlines the high-level architecture for our platform. We utilize KubeVela as the Application Control Plane and Terraform as the Infrastructure Engine.

This architecture is designed to be Serverless-like for the developer: they define what they need (e.g., "Database", "WebService"), and the platform handles the complexity of how it is provisioned and maintained.

2. Core Components

2.1 The Control Plane (KubeVela)

Role: Acts as the "Manager." It interprets Application.yaml files and orchestrates the deployment.

Responsibility: It manages the lifecycle of the application, including rollouts, rollbacks, and policy enforcement.

2.2 The Provisioning Engine (Terraform Controller)

Role: Acts as the "Worker." It is an add-on running inside the Kubernetes cluster.

Behavior: When KubeVela requests infrastructure (like an RDS database or S3 bucket), it delegates the task to the Terraform Controller.

State Management:

No Manual State Files: We do not manage .tfstate files locally or in S3 buckets manually.

Kubernetes Backend: The Controller automatically encrypts and stores the Terraform state inside Kubernetes Secrets.

Self-Healing: If the infrastructure drifts (e.g., someone deletes a security group manually), the Controller detects the change and auto-corrects it to match the configuration.

3. The Deployment Workflow (CI/CD)

We follow a strict "Artifact-First" deployment model. The configuration repository is the source of truth, but the CI pipeline drives the updates.

3.1 Day 0: The "Placeholder" Strategy

Before the first line of code is ever built, the application must be deployable and "Green" (Healthy).

The Problem: We cannot deploy an image that doesn't exist yet.

The Solution: We use a Placeholder Image in our initial Application.yaml.

Recommended Placeholder: nginx:alpine

Why? It is extremely lightweight (5MB), pulls instantly, and starts a web server on Port 80, satisfying Kubernetes liveness probes immediately.

Initial Configuration:

apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
name: order-system
spec:
components: - name: order-api
type: webservice
properties:
image: "nginx:alpine" # <--- Placeholder: Keeps the app "Healthy"
ports: - port: 80

3.2 Day 1: The Build & Replace Pipeline

Once the developer pushes code, the CI pipeline takes over.

Identify: The build system (Nx) identifies which Target (Artifact) changed (e.g., api or worker).

Build & Push: The CI builds the Docker image and pushes it to the registry.

Result: myregistry.com/order-api:v1-sha123

Update (The Handshake): The CI pipeline executes a KubeVela command to replace the placeholder with the real artifact.

The Logic:

"I have successfully built the artifact. Now, update the live system to use this specific version."

# Example CI Command

vela app update order-system --component order-api --image [myregistry.com/order-api:v1-sha123](https://myregistry.com/order-api:v1-sha123)

4. Component Mapping Strategy

To maintain a clean architecture in our monorepo, we adhere to the 1-to-1 Mapping Rule.

Rule: One Build Target = One KubeVela Component.

Anti-Pattern: We do not map "One Project" to "One Component" if that project produces multiple images.

Nx Project

Build Target (Artifact)

KubeVela Component Name

order-service

api

order-api

order-service

worker

order-worker

frontend

production

frontend-main

This ensures that updating the worker image never accidentally restarts the api service.

5. Summary of Benefits

Zero-Touch State: Developers never run terraform init. The cluster manages its own state securely.

Always Deployable: Thanks to the Placeholder Strategy, the application definition is valid from the very first commit.

Decoupled Registry: The project.json (build config) does not know about the registry. The CI pipeline injects the destination at runtime, allowing for total portability.
