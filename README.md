# EiffelBridge

[![Go Report Card](https://goreportcard.com/badge/github.com/example/eiffel-bridge-svc)](https://goreportcard.com/report/github.com/example/eiffel-bridge-svc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An adapter service that bridges GitLab webhooks to Eiffel events, enabling integration between GitLab CI/CD and Eiffel-based event systems.

It receives HTTP POST requests from GitLab, converts them into corresponding **Eiffel events**, and publishes them to **RabbitMQ** for downstream processing. The purpose of the adapter is to support migration of CI/CD systems (e.g. Jenkins + Gerrit + Eiffel) into a **GitLab + Eiffel + RabbitMQ** architecture.

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

- **Adapter** → receives webhook JSON, converts it to Eiffel events with error handling and logging
- **RabbitMQ** → acts as the central event bus (queue: eiffel.events)
- **Consumers** → tools that subscribe to Eiffel events (CLI, logging, CI triggers, dashboards)

### Supported Events

- **Push Hook** → `EiffelSourceChangeCreatedEvent`
- **Pipeline Hook** → `EiffelActivityTriggeredEvent` + `EiffelActivityFinishedEvent`

### Endpoints

- `POST /webhook` — GitLab webhook receiver
- `GET /health` — Health check endpoint

---

## Deploy

### Prerequisites

- Docker or Podman
- Python 3.8+
- curl, tar, sudo available

Run the deploy script from the project root to set up all components (Minikube, RabbitMQ, and the EiffelBridge service):

```bash
./deploy.py
```

The script will:

- Build the EiffelBridge Docker image
- Package and install the EiffelBridge Helm chart
- Expose services via NodePort or port-forward
- Print curl and rabbitmqadmin commands for testing

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

## Tear Down

```bash
minikube delete
```

## Development

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

## License

MIT License — see [LICENSE](LICENSE) for details.
