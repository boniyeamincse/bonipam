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

### Backend
```bash
# From project root
bash backend/services/scripts/run-local.sh
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

## Test Credentials

For development and testing, the `auth-service` is pre-configured with the following in-memory user:

- **Email**: `admin@bonipam.local`
- **Password**: `admin123` (Note: Backend currently accepts any password for this user in dev mode)
- **Role**: `SuperAdmin`

## Health Checks

- **API Gateway**: `GET http://localhost:8080/health`
- **Auth Service**: `GET http://localhost:8081/health`

## Initial API routes

Auth service (`:8081`):

- `POST /api/v1/auth/sso/callback`
- `POST /api/v1/auth/mfa/challenge`
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
