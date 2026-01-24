# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-01-24

### Added
- Support GitLab's form-encoded webhook format
- Health check endpoint at `/health` for container readiness probes
- Comprehensive error handling for webhook payload validation
- Enhanced logging throughout the application with structured messages
- Safe type assertions for GitLab webhook payloads

### Changed
- Improved error messages in webhook handlers
- Better logging format with file names and line numbers
- More robust JSON marshaling error handling in RabbitMQ publisher

### Fixed
- Potential panic from unsafe type assertions in webhook handlers
- Missing error handling for JSON marshaling operations

## [0.1.0] - 2025-10-05

### Added
- Initial EiffelBridge service implementation
- GitLab webhook receiver with support for Push Hook and Pipeline Hook events
- Eiffel event generation (SourceChangeCreated, ActivityTriggered, ActivityFinished)
- RabbitMQ publisher for event distribution
- Docker containerization with multi-stage build
- Kubernetes deployment manifests
- Helm chart for easy deployment
- Automated deployment script with Minikube support
- Sample webhook payloads for testing
- Comprehensive README with architecture diagrams

### Infrastructure
- Go 1.23 based implementation
- RabbitMQ integration for event messaging
- Kubernetes-ready deployment configuration
- Support for both Docker and Podman container runtimes