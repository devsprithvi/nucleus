# Service Registry

This directory acts as the internal application hub for discovering all available services within our organization.

## Purpose
The registry is used to track and manage the lifecycle of services across the Nucleus ecosystem. It provides a single point of reference to find out what services exist and how they are configured.

## Service Registration
All services must be registered within this registry. We use a standardized `component.yaml` file (or similar configuration) to represent each service. This file contains machine-readable metadata about the service, including its domain, name, status, and capabilities.

## Usage
By maintaining this registry, developers and automated systems can:
- Discover existing service capabilities.
- Understand service dependencies.
- Access documentation and integration details for any organizational service.
