# Boni PAM Backend Scaffold

This repository now contains a production-oriented Go microservices scaffold for Boni PAM.

## Implemented services

- `api-gateway` (`cmd/api-gateway`)
- `auth-service` (`cmd/auth-service`)

## Key features included

- Shared configuration loader in `pkg/config`
- Structured logger setup in `pkg/logger`
- Correlation ID middleware (`X-Correlation-ID`)
- Standard JSON response envelope
- Graceful shutdown and health endpoints
- Dockerfiles for both services

## Run locally

```bash
go mod tidy
chmod +x services/scripts/run-local.sh
./services/scripts/run-local.sh
```

Health checks:

- `GET http://localhost:8080/health`
- `GET http://localhost:8081/health`

## Initial API routes

Auth service (`:8081`):

- `POST /api/v1/auth/sso/callback`
- `POST /api/v1/auth/mfa/verify`
- `POST /api/v1/auth/token/refresh`
- `POST /api/v1/auth/logout`

Gateway (`:8080`):

- `POST /api/v1/gateway/sessions/start`
- `POST /api/v1/gateway/sessions/:sessionId/terminate`
- `GET /api/v1/gateway/sessions/:sessionId/status`

## Next suggested milestones

- Integrate real OIDC provider and JWT signing/verification.
- Add PostgreSQL repositories and migrations.
- Add policy evaluation and JIT grant validation in gateway flow.
- Add unit/integration tests and CI pipeline.
