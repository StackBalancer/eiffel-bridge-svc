# EiffelBridge

[![Go Report Card](https://goreportcard.com/badge/github.com/example/eiffel-bridge-svc)](https://goreportcard.com/report/github.com/example/eiffel-bridge-svc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A proof-of-concept adapter service that bridges GitLab webhooks to Eiffel events, enabling integration between GitLab CI/CD and Eiffel-based event systems.

> **⚠️ Proof of Concept**: This is a minimal implementation for demonstration purposes. See [TODO.md](TODO.md) for production readiness requirements.

This project provides a simple adapter service that simulates the GitLab → Eiffel integration bridge by receiving and processing **GitLab webhook-like** events locally.
Instead of connecting to a real GitLab instance, a lightweight built-in webhook emulator is implemented in Golang.
It accepts HTTP POST requests that mimic GitLab push events (via curl and JSON files), converts them into corresponding **Eiffel events**, and publishes them to **RabbitMQ** for downstream processing.

The purpose of adapter service to migrate CI/CD systems (e.g. Jenkins + Gerrit + Eiffel) into a **GitLab + Eiffel + RabbitMQ** architecture.

---

## 🔄 Architecture

```mermaid
flowchart LR
    subgraph GitLab
        A[Developer Commit/Push]
        B[GitLab Webhook JSON]
    end

    subgraph Adapter[GitLab → Eiffel Adapter]
        C[Webhook Endpoint /webhook]
        D[Translate GitLab JSON → Eiffel Event]
    end

    subgraph Broker[RabbitMQ]
        E[(Queue: eiffel.events)]
    end

    subgraph Consumers
        F[Consumer 1: Log Eiffel Events]
        G[Consumer 2: Trigger CI Job]
        H[Consumer 3: Dashboard / Monitoring]
    end

    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    E --> G
    E --> H
```

---

## Components

- **GitLab webhook emulator** → generates webhook events on commits, merges, etc.
- **Adapter** → receives webhook JSON, converts it to Eiffel event JSON with error handling and logging
- **RabbitMQ** → acts as the central event bus (queue: eiffel.events)
- **Consumers** → tools that subscribe to Eiffel events (cli, logging, CI triggers, dashboards)

### Supported Events

- **Push Hook** → `EiffelSourceChangeCreatedEvent`
- **Pipeline Hook** → `EiffelActivityTriggeredEvent` + `EiffelActivityFinishedEvent`

### Endpoints

- `POST /webhook` - GitLab webhook receiver
- `GET /health` - Health check endpoint

---

## Deploy

### Prerequisites

- Docker or Podman
- Python 3.8+
- curl, tar, sudo available

Project includes a deployment script that sets up all components — Minikube, RabbitMQ and the EiffelBridge service on your local machine.

Run the deploy.py script from project root to deploy stack:
```bash
$ ./deploy.py
```

Script will:

- Build your EiffelBridge Docker image (if needed)
- Package & install EiffelBridge Helm chart
- Expose services (via NodePort or port-forward)
- Print final curl + rabbitmqadmin commands for testing

### Manual Testing

Once deployed, test the webhook endpoint:

```bash
# Test push event (JSON)
curl -X POST http://localhost:8080/webhook \
     -H 'Content-Type: application/json' \
     -H 'X-Gitlab-Event: Push Hook' \
     --data-binary @test-webhooks/push.json

# Test push event (form-encoded, like real GitLab)
curl -X POST http://localhost:8080/webhook \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     -H 'X-Gitlab-Event: Push Hook' \
     --data-urlencode "payload@test-webhooks/push.json"

# Test pipeline event
curl -X POST http://localhost:8080/webhook \
     -H 'Content-Type: application/json' \
     -H 'X-Gitlab-Event: Pipeline Hook' \
     --data-binary @test-webhooks/pipeline-success.json

# Health check
curl http://localhost:8080/health
```

### Monitoring

- **RabbitMQ Management UI**: http://localhost:15672 (guest/guest)
- **Service Health**: http://localhost:8080/health

---

## Tear down
```bash
minikube delete
```

## Development

### Local Development

```bash
# Install dependencies
go mod download

# Run locally (requires RabbitMQ)
export RABBIT_URL="amqp://guest:guest@localhost:5672/"
go run cmd/adapter/main.go
```

### Project Structure

```
eiffel-bridge-svc/
├── cmd/adapter/          # Main application
├── internal/
│   ├── eiffel/           # Eiffel event structures
│   ├── gitlab/           # GitLab webhook handlers
│   └── publisher/        # RabbitMQ publisher
├── charts/eiffel-bridge/ # Helm chart
├── test-webhooks/        # Sample payloads
└── k8s/                  # Kubernetes manifests
```

## Contributing

This is a proof-of-concept project. For production use, please review [TODO.md](TODO.md) for required improvements.

## License

MIT License - see [LICENSE](LICENSE) file for details.
