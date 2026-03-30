# Boni PAM - Enterprise Technical Blueprint

**Document Version:** 1.0  
**Date:** 2026-03-30  
**System Name:** Boni PAM (Privileged Access Management)

---

## 1. Project Overview

### System Name
Boni PAM

### Goals
- Centralize and secure privileged access to critical infrastructure.
- Enforce least privilege and Zero Trust principles across all access paths.
- Provide auditable, policy-driven, time-bound privileged sessions.
- Support enterprise scale with modular microservices and horizontal scalability.
- Reduce credential exposure by replacing direct credential sharing with controlled brokering.

### Key Features
- Enterprise authentication with SSO (OIDC/SAML) and MFA.
- Role-Based Access Control (RBAC) and extensible attribute-based policies.
- Fine-grained policy engine for command, asset, and context-aware access decisions.
- Asset inventory for servers, databases, clusters, and network appliances.
- Encrypted credential vault with automated rotation workflows.
- SSH Bastion / Access Gateway for controlled session brokering.
- Session monitoring, tamper-evident recording, and searchable metadata.
- Immutable audit logging with integrity verification.
- Just-in-Time (JIT) access with expiring grants and approvals.
- Access request and multi-stage approval workflow with SLA handling.

### Architecture Summary
Boni PAM is designed as a Go-based microservices platform with a React frontend. Core services communicate through REST APIs and asynchronous event messaging. PostgreSQL stores transactional data, Redis supports caching and ephemeral state, and object storage keeps encrypted session recordings. Security controls include mTLS service-to-service communication, KMS-backed encryption keys, and signed audit events.

---

## 2. Software Architecture

### High-Level Architecture Diagram (ASCII)

```text
                               +---------------------------+
                               |   Enterprise IdP (SSO)    |
                               |   OIDC / SAML / MFA       |
                               +------------+--------------+
                                            |
                                            v
+------------------+             +----------+----------+             +--------------------+
|   React Web UI   +-----------> |   API Gateway /     +-----------> |   Auth Service     |
|  Admin/Requester |   HTTPS     |   Edge Router        |             |  Tokens, Sessions  |
+---------+--------+             +-----+-----------+----+             +---------+----------+
          |                            |           |                            |
          |                            |           |                            v
          |                            |           |                     +------+------+
          |                            |           +-------------------->+ User & Role |
          |                            |                                 |  Management |
          |                            |                                 +------+------+
          |                            |                                        |
          |                            v                                        v
          |                   +--------+--------+                      +--------+--------+
          |                   |   Policy Engine  |<--------------------+ Approval Workflow|
          |                   +--------+--------+                      +--------+--------+
          |                            |                                        |
          |                            v                                        v
          |                   +--------+--------+                      +--------+--------+
          |                   |    JIT Access    |-------------------->+ Asset Mgmt      |
          |                   +--------+--------+                      +--------+--------+
          |                            |                                        |
          |                            v                                        v
          |                   +--------+--------+                      +--------+--------+
          +------------------>+  Access Gateway  +-------------------->+ Vault Service    |
             Session View      |   (SSH Proxy)    |   Ephemeral Creds   | Secrets/Rotation |
                               +--------+--------+                      +--------+--------+
                                        |
                                        v
                               +--------+--------+        Events        +------------------+
                               | Session Recorder +-------------------->+ Audit Logging    |
                               +--------+--------+                     | (Immutable Trail)|
                                        |                              +---------+--------+
                                        v                                        |
                             +----------+----------+                              v
                             | Encrypted Object    |                     +--------+--------+
                             | Storage (Recordings)|                     | SIEM / Analytics |
                             +---------------------+                     +------------------+
```

### Microservices Breakdown
- API Gateway
  - Unified API entry point, routing, rate limiting, request tracing, and coarse authorization.
- Auth Service
  - SSO integration, MFA orchestration, token issuance, session lifecycle.
- User & Role Management Service
  - Users, groups, role templates, permission assignments.
- Policy Engine Service
  - Evaluate RBAC/ABAC policies and return allow/deny + obligations.
- Asset Management Service
  - Inventory of managed assets and connection metadata.
- Vault Service
  - Store, retrieve, rotate credentials and keys.
- Access Gateway Service
  - SSH session proxying and controlled command channel.
- Session Recorder Service
  - Real-time stream ingestion, indexing, encrypted archival.
- Audit Logging Service
  - Immutable, signed audit events and query APIs.
- JIT Access Service
  - Time-bound entitlements and lease management.
- Approval Workflow Service
  - Access request state machine, approvers, escalations.

### Data Flow Explanation
1. User authenticates through SSO/MFA via Auth Service.
2. User requests privileged access to a target asset.
3. Approval Workflow evaluates request route and approvals.
4. JIT Access issues short-lived grant after approval.
5. Policy Engine evaluates current context (role, device, risk, time, asset tags).
6. Access Gateway validates grant and policy decision.
7. Gateway retrieves ephemeral credential/token from Vault Service.
8. Session starts through gateway; Session Recorder captures activity.
9. All actions generate signed audit events to Audit Logging.
10. Session end triggers credential revocation/lease cleanup and final audit closure.

---

## 3. Module Breakdown

### 3.1 Auth Service

**Description**  
Provides identity federation, MFA checks, token lifecycle, and user session state.

**Responsibilities**
- OIDC/SAML SSO integration.
- MFA challenge orchestration (TOTP/WebAuthn/push).
- Access token and refresh token issuance.
- Session revocation and idle timeout enforcement.

**Key APIs (Examples)**
- `POST /api/v1/auth/sso/callback`
- `POST /api/v1/auth/mfa/verify`
- `POST /api/v1/auth/token/refresh`
- `POST /api/v1/auth/logout`

**Dependencies**
- Enterprise IdP
- Redis (session cache)
- User & Role Management
- Audit Logging

### 3.2 User & Role Management

**Description**  
Manages identity metadata, role catalogs, group mappings, and permissions.

**Responsibilities**
- User profile synchronization from IdP.
- Role CRUD and permission mapping.
- Group-to-role mapping and lifecycle.
- Account state controls (active, locked, disabled).

**Key APIs (Examples)**
- `GET /api/v1/users`
- `POST /api/v1/roles`
- `PUT /api/v1/roles/{roleId}/permissions`
- `POST /api/v1/groups/{groupId}/roles/{roleId}`

**Dependencies**
- PostgreSQL
- Auth Service
- Audit Logging

### 3.3 Policy Engine

**Description**  
Evaluates access policies and returns deterministic authorization decisions.

**Responsibilities**
- Evaluate RBAC and optional ABAC conditions.
- Support context attributes (time, source IP, device trust, risk score).
- Return obligations (record session, require approval level, block commands).
- Maintain policy versioning and rollback.

**Key APIs (Examples)**
- `POST /api/v1/policies/evaluate`
- `POST /api/v1/policies`
- `GET /api/v1/policies/{policyId}`
- `POST /api/v1/policies/{policyId}/publish`

**Dependencies**
- User & Role Management
- Asset Management
- JIT Access
- PostgreSQL

### 3.4 Asset Management

**Description**  
Maintains authoritative inventory of privileged assets and metadata.

**Responsibilities**
- Register/update servers, DBs, clusters, and devices.
- Maintain ownership, environment tags, criticality, and network zones.
- Health checks and connectivity validation.
- Asset grouping for policy assignment.

**Key APIs (Examples)**
- `POST /api/v1/assets`
- `GET /api/v1/assets/{assetId}`
- `PUT /api/v1/assets/{assetId}`
- `POST /api/v1/assets/{assetId}/test-connection`

**Dependencies**
- PostgreSQL
- Policy Engine
- Audit Logging

### 3.5 Vault Service

**Description**  
Stores and serves sensitive credentials using encryption and strict controls.

**Responsibilities**
- Secret storage and retrieval with envelope encryption.
- Dynamic secret issuance and lease expiration.
- Automated credential rotation and revocation.
- Secret access logging and policy enforcement.

**Key APIs (Examples)**
- `POST /api/v1/vault/secrets`
- `POST /api/v1/vault/secrets/{secretId}/issue`
- `POST /api/v1/vault/secrets/{secretId}/rotate`
- `POST /api/v1/vault/leases/{leaseId}/revoke`

**Dependencies**
- KMS/HSM
- PostgreSQL
- Audit Logging
- Access Gateway

### 3.6 Access Gateway (SSH Proxy)

**Description**  
Acts as controlled ingress for privileged SSH sessions, enforcing policy and recording.

**Responsibilities**
- SSH connection brokering without exposing target credentials.
- Policy checks at session start and per command (where applicable).
- Enforce session controls (idle timeout, command restrictions, file transfer policy).
- Stream telemetry to Session Recorder and Audit Logging.

**Key APIs (Examples)**
- `POST /api/v1/gateway/sessions/start`
- `POST /api/v1/gateway/sessions/{sessionId}/terminate`
- `GET /api/v1/gateway/sessions/{sessionId}/status`
- `POST /api/v1/gateway/commands/authorize`

**Dependencies**
- Policy Engine
- JIT Access
- Vault Service
- Session Recorder
- Audit Logging

### 3.7 Session Recorder

**Description**  
Captures, encrypts, stores, and indexes privileged session activities.

**Responsibilities**
- Stream ingestion from gateway.
- Record terminal I/O and metadata timeline.
- Encrypt recordings before persistence.
- Enable playback and search by session metadata.

**Key APIs (Examples)**
- `POST /api/v1/recordings/ingest`
- `GET /api/v1/recordings/{sessionId}`
- `GET /api/v1/recordings/{sessionId}/metadata`
- `POST /api/v1/recordings/{sessionId}/seal`

**Dependencies**
- Object Storage (S3-compatible)
- PostgreSQL (metadata)
- Audit Logging

### 3.8 Audit Logging

**Description**  
Provides immutable and signed audit records for all privileged actions.

**Responsibilities**
- Accept event streams from all services.
- Sign and chain events for tamper evidence.
- Provide search/filter/report APIs.
- Export to SIEM and compliance systems.

**Key APIs (Examples)**
- `POST /api/v1/audit/events`
- `GET /api/v1/audit/events`
- `GET /api/v1/audit/events/{eventId}/verify`
- `POST /api/v1/audit/exports`

**Dependencies**
- PostgreSQL
- Kafka/NATS (optional event bus)
- SIEM integration

### 3.9 JIT Access

**Description**  
Issues short-lived elevated access grants based on policy and approvals.

**Responsibilities**
- Grant creation with validity window.
- Lease renewal and expiration enforcement.
- Automatic deprovisioning on expiry/revocation.
- Exposure of active grants for enforcement.

**Key APIs (Examples)**
- `POST /api/v1/jit/grants`
- `GET /api/v1/jit/grants/{grantId}`
- `POST /api/v1/jit/grants/{grantId}/renew`
- `POST /api/v1/jit/grants/{grantId}/revoke`

**Dependencies**
- Approval Workflow
- Policy Engine
- Asset Management
- Audit Logging

### 3.10 Approval Workflow

**Description**  
Manages request lifecycle from submission to approval/rejection/escalation.

**Responsibilities**
- Rule-driven approval chains.
- Multi-approver quorum and conditional approvals.
- SLA timers, reminders, and escalations.
- Decision traceability and rationale capture.

**Key APIs (Examples)**
- `POST /api/v1/requests`
- `POST /api/v1/requests/{requestId}/approve`
- `POST /api/v1/requests/{requestId}/reject`
- `POST /api/v1/requests/{requestId}/escalate`

**Dependencies**
- User & Role Management
- Policy Engine
- JIT Access
- Notification Service (email/chat)
- Audit Logging

---

## 4. Database Design (PostgreSQL)

### Core Tables and Fields

#### `users`
- `id` UUID PK
- `external_id` VARCHAR(128) UNIQUE
- `email` VARCHAR(255) UNIQUE NOT NULL
- `display_name` VARCHAR(255)
- `status` VARCHAR(32) NOT NULL
- `last_login_at` TIMESTAMPTZ
- `created_at` TIMESTAMPTZ NOT NULL
- `updated_at` TIMESTAMPTZ NOT NULL

#### `roles`
- `id` UUID PK
- `name` VARCHAR(128) UNIQUE NOT NULL
- `description` TEXT
- `is_system` BOOLEAN DEFAULT FALSE
- `created_at` TIMESTAMPTZ NOT NULL
- `updated_at` TIMESTAMPTZ NOT NULL

#### `permissions`
- `id` UUID PK
- `resource` VARCHAR(128) NOT NULL
- `action` VARCHAR(64) NOT NULL
- `constraints` JSONB

#### `role_permissions`
- `role_id` UUID FK -> roles.id
- `permission_id` UUID FK -> permissions.id
- PK (`role_id`, `permission_id`)

#### `user_roles`
- `user_id` UUID FK -> users.id
- `role_id` UUID FK -> roles.id
- `assigned_by` UUID FK -> users.id
- `assigned_at` TIMESTAMPTZ NOT NULL
- PK (`user_id`, `role_id`)

#### `assets`
- `id` UUID PK
- `name` VARCHAR(255) NOT NULL
- `type` VARCHAR(64) NOT NULL
- `hostname` VARCHAR(255)
- `ip_address` INET
- `port` INT
- `environment` VARCHAR(64)
- `criticality` VARCHAR(32)
- `owner_team` VARCHAR(128)
- `tags` JSONB
- `created_at` TIMESTAMPTZ NOT NULL
- `updated_at` TIMESTAMPTZ NOT NULL

#### `asset_credentials`
- `id` UUID PK
- `asset_id` UUID FK -> assets.id
- `vault_secret_ref` VARCHAR(255) NOT NULL
- `credential_type` VARCHAR(64)
- `rotation_policy_id` UUID
- `last_rotated_at` TIMESTAMPTZ

#### `policies`
- `id` UUID PK
- `name` VARCHAR(128) UNIQUE NOT NULL
- `version` INT NOT NULL
- `status` VARCHAR(32) NOT NULL
- `definition` JSONB NOT NULL
- `created_by` UUID FK -> users.id
- `created_at` TIMESTAMPTZ NOT NULL
- `published_at` TIMESTAMPTZ

#### `access_requests`
- `id` UUID PK
- `requester_id` UUID FK -> users.id
- `asset_id` UUID FK -> assets.id
- `requested_role` VARCHAR(128)
- `reason` TEXT
- `start_time` TIMESTAMPTZ
- `end_time` TIMESTAMPTZ
- `status` VARCHAR(32) NOT NULL
- `created_at` TIMESTAMPTZ NOT NULL
- `updated_at` TIMESTAMPTZ NOT NULL

#### `approvals`
- `id` UUID PK
- `request_id` UUID FK -> access_requests.id
- `approver_id` UUID FK -> users.id
- `decision` VARCHAR(16) NOT NULL
- `comment` TEXT
- `decided_at` TIMESTAMPTZ

#### `jit_grants`
- `id` UUID PK
- `request_id` UUID FK -> access_requests.id
- `user_id` UUID FK -> users.id
- `asset_id` UUID FK -> assets.id
- `scope` JSONB
- `status` VARCHAR(32) NOT NULL
- `valid_from` TIMESTAMPTZ NOT NULL
- `valid_until` TIMESTAMPTZ NOT NULL
- `revoked_at` TIMESTAMPTZ

#### `sessions`
- `id` UUID PK
- `jit_grant_id` UUID FK -> jit_grants.id
- `user_id` UUID FK -> users.id
- `asset_id` UUID FK -> assets.id
- `gateway_node` VARCHAR(255)
- `status` VARCHAR(32)
- `started_at` TIMESTAMPTZ
- `ended_at` TIMESTAMPTZ
- `termination_reason` VARCHAR(255)

#### `session_recordings`
- `id` UUID PK
- `session_id` UUID FK -> sessions.id
- `storage_uri` TEXT NOT NULL
- `sha256_checksum` VARCHAR(64) NOT NULL
- `encryption_key_ref` VARCHAR(255) NOT NULL
- `size_bytes` BIGINT
- `sealed_at` TIMESTAMPTZ

#### `audit_events`
- `id` UUID PK
- `event_time` TIMESTAMPTZ NOT NULL
- `actor_id` UUID
- `service_name` VARCHAR(128) NOT NULL
- `event_type` VARCHAR(128) NOT NULL
- `entity_type` VARCHAR(128)
- `entity_id` UUID
- `payload` JSONB NOT NULL
- `prev_event_hash` VARCHAR(128)
- `event_hash` VARCHAR(128) NOT NULL
- `signature` TEXT NOT NULL

### Relationship Summary
- One `user` to many `user_roles`, `access_requests`, `jit_grants`, `sessions`.
- One `role` to many `permissions` via `role_permissions`.
- One `asset` to many `access_requests`, `jit_grants`, `sessions`, `asset_credentials`.
- One `access_request` to many `approvals`, and zero/one `jit_grant`.
- One `session` to one/many `session_recordings` (chunking allowed).
- All operational entities emit events into `audit_events`.

---

## 5. Security Design

### Zero Trust Model
- Never trust network location; verify every request.
- Continuous authorization checks at request, grant, and command boundaries.
- Strong identity binding between user, device posture, and session.
- Short-lived credentials and contextual policy evaluation.

### Encryption Strategy
- TLS 1.3 for all client and service communications.
- mTLS for service-to-service traffic inside the cluster.
- AES-256-GCM for data at rest in DB fields and object storage artifacts.
- Envelope encryption using KMS/HSM-backed data keys.
- Rotating signing keys for audit integrity signatures.

### Secrets Handling
- No static credentials in source control or environment defaults.
- Runtime secret injection via Vault/KMS integration.
- Secret lease TTL and automatic revocation on session end.
- Dual control for sensitive secret retrieval and break-glass operations.

### Access Control Model
- Layered authorization: RBAC baseline + ABAC conditions + JIT grants.
- Approval workflow as prerequisite for high-risk operations.
- Policy obligations enforce recording, command restrictions, and session limits.
- Emergency access with mandatory post-incident review and strict audit.

### Additional Security Controls
- Rate limiting and adaptive lockouts for authentication paths.
- Device/IP risk scoring integrated into policy decisions.
- Tamper-evident logs with hash chaining and digital signatures.
- Secure SDLC: SAST, DAST, dependency and container image scanning.

---

## 6. Backend Structure (Go Project)

```text
boni-pam/
├── cmd/
│   ├── api-gateway/
│   ├── auth-service/
│   ├── user-role-service/
│   ├── policy-service/
│   ├── asset-service/
│   ├── vault-service/
│   ├── access-gateway/
│   ├── session-recorder/
│   ├── audit-service/
│   ├── jit-service/
│   └── approval-service/
├── internal/
│   ├── app/
│   ├── domain/
│   ├── repository/
│   ├── service/
│   ├── transport/
│   ├── middleware/
│   ├── security/
│   └── observability/
├── pkg/
│   ├── logger/
│   ├── config/
│   ├── errors/
│   ├── crypto/
│   ├── authz/
│   ├── tracing/
│   └── utils/
├── services/
│   ├── docker/
│   ├── helm/
│   ├── k8s/
│   ├── migrations/
│   └── scripts/
└── go.work
```

### Folder Explanations
- `/cmd`
  - Entrypoints (`main.go`) for each microservice, bootstrapping configs, dependencies, server startup.
- `/internal`
  - Private application code per Go module boundary; contains core business logic, adapters, middleware, and domain models.
- `/pkg`
  - Reusable shared libraries safe for cross-service usage (logging, crypto wrappers, config loaders, error contracts).
- `/services`
  - Deployment and operations artifacts: Dockerfiles, Helm charts, Kubernetes manifests, DB migrations, and utility scripts.

---

## 7. Backend Task List (50 Tasks)

- **TASK-001: Implement OIDC Login Flow**  
  Description: Integrate OIDC authorization code flow with enterprise IdP and callback handling.  
  Module: Auth Service
  Status: Done

- **TASK-002: Add SAML SSO Adapter**  
  Description: Build SAML assertion consumer endpoint and map attributes to internal user model.  
  Module: Auth Service
  Status: Done

- **TASK-003: Implement MFA Challenge API**  
  Description: Add endpoints for TOTP/WebAuthn challenge generation and verification.  
  Module: Auth Service
  Status: Done

- **TASK-004: Build JWT/Refresh Token Service**  
  Description: Generate signed access tokens and rotate refresh tokens securely.  
  Module: Auth Service
  Status: Done

- **TASK-005: Session Revocation Endpoint**  
  Description: Create API to revoke all or specific sessions and invalidate refresh tokens.  
  Module: Auth Service
  Status: Done

- **TASK-006: User Sync from IdP**  
  Description: Implement periodic and login-triggered synchronization of user profiles and groups.  
  Module: User & Role Management
  Status: Done

- **TASK-007: User CRUD with Soft Delete**  
  Description: Create admin APIs for user lifecycle with soft-delete and recovery controls.  
  Module: User & Role Management
  Status: Done

- **TASK-008: Role CRUD APIs**  
  Description: Build endpoints for role create/read/update/delete with validation.  
  Module: User & Role Management
  Status: Done

- **TASK-009: Permission Catalog Service**  
  Description: Define permission model and APIs for resource-action combinations.  
  Module: User & Role Management
  Status: Done

- **TASK-010: Group-to-Role Mapping**  
  Description: Support assigning roles to external groups and reconcile membership changes.  
  Module: User & Role Management
  Status: Done

- **TASK-011: Policy Definition Schema**  
  Description: Create JSON schema for policy documents with versioned validation.  
  Module: Policy Engine
  Status: Done

- **TASK-012: Policy CRUD Endpoints**  
  Description: Implement create/update/list/get APIs for policy artifacts.  
  Module: Policy Engine
  Status: Done

- **TASK-013: Policy Evaluation Runtime**  
  Description: Build deterministic evaluator returning allow/deny and obligations.  
  Module: Policy Engine
  Status: Done

- **TASK-014: Policy Version Publish Workflow**  
  Description: Add draft, publish, rollback mechanics for policy deployment.  
  Module: Policy Engine
  Status: Done

- **TASK-015: Context Attribute Resolver**  
  Description: Resolve runtime attributes (device, IP, time, risk) for ABAC checks.  
  Module: Policy Engine
  Status: Done

- **TASK-016: Asset Registration API**  
  Description: Implement asset onboarding endpoint with connection metadata validation.  
  Module: Asset Management
  Status: Done

- **TASK-017: Asset Tagging and Grouping**  
  Description: Support environment, owner, and criticality tags for policy targeting.  
  Module: Asset Management
  Status: Done

- **TASK-018: Asset Connectivity Tester**  
  Description: Build secure test connection workflow for SSH and database assets.  
  Module: Asset Management
  Status: Done

- **TASK-019: Asset Ownership Workflow**  
  Description: Implement owner assignment and transfer with approvals and audit events.  
  Module: Asset Management
  Status: Done

- **TASK-020: Asset Import Pipeline**  
  Description: Add CSV/API bulk asset import with deduplication and validation.  
  Module: Asset Management
  Status: Done

- **TASK-021: Secret Storage API**  
  Description: Create encrypted secret write/read APIs with envelope encryption.  
  Module: Vault Service

- **TASK-022: Dynamic Credential Issuance**  
  Description: Generate short-lived credentials for supported targets with lease metadata.  
  Module: Vault Service

- **TASK-023: Credential Rotation Scheduler**  
  Description: Build periodic rotation jobs and policy-driven rotation windows.  
  Module: Vault Service

- **TASK-024: Lease Revocation Mechanism**  
  Description: Revoke active credential leases on session termination or policy trigger.  
  Module: Vault Service

- **TASK-025: KMS Integration Adapter**  
  Description: Integrate external KMS/HSM for key generation and envelope key wrapping.  
  Module: Vault Service

- **TASK-026: SSH Proxy Core**  
  Description: Implement SSH bastion core for inbound user channels and outbound target sessions.  
  Module: Access Gateway

- **TASK-027: Gateway Session Authorization**  
  Description: Enforce JIT grant and policy checks before session establishment.  
  Module: Access Gateway

- **TASK-028: Command Restriction Filter**  
  Description: Add real-time command allow/deny enforcement with policy obligations.  
  Module: Access Gateway

- **TASK-029: Idle Timeout and Kill Switch**  
  Description: Terminate stale sessions and provide admin emergency termination endpoint.  
  Module: Access Gateway

- **TASK-030: File Transfer Governance**  
  Description: Restrict SCP/SFTP actions based on policy and sensitive paths.  
  Module: Access Gateway

- **TASK-031: Session Stream Ingestion API**  
  Description: Build high-throughput endpoint for terminal stream fragments from gateway nodes.  
  Module: Session Recorder

- **TASK-032: Session Metadata Indexer**  
  Description: Store searchable session metadata (user, asset, command markers, timestamps).  
  Module: Session Recorder

- **TASK-033: Recording Encryption Pipeline**  
  Description: Encrypt stream chunks before persisting to object storage.  
  Module: Session Recorder

- **TASK-034: Playback API with Access Control**  
  Description: Provide secure playback URLs and frame streaming for authorized auditors.  
  Module: Session Recorder

- **TASK-035: Recording Integrity Verification**  
  Description: Generate and validate checksums/signatures for tamper detection.  
  Module: Session Recorder

- **TASK-036: Unified Audit Event Contract**  
  Description: Define common audit event schema and service-side SDK for emission.  
  Module: Audit Logging

- **TASK-037: Immutable Audit Store**  
  Description: Implement append-only event persistence with retention policies.  
  Module: Audit Logging

- **TASK-038: Hash Chain Signatures**  
  Description: Chain event hashes and sign each event to ensure tamper evidence.  
  Module: Audit Logging

- **TASK-039: Audit Search API**  
  Description: Build query endpoints with filters by actor, asset, event type, and date range.  
  Module: Audit Logging

- **TASK-040: SIEM Export Connector**  
  Description: Stream audit events to external SIEM via webhook/syslog/Kafka adapters.  
  Module: Audit Logging

- **TASK-041: JIT Grant Issuance API**  
  Description: Create grants with constrained scope, validity window, and policy obligations.  
  Module: JIT Access

- **TASK-042: Grant Renewal Workflow**  
  Description: Implement conditional extension flow with re-approval requirements.  
  Module: JIT Access

- **TASK-043: Automatic Grant Expiry Worker**  
  Description: Revoke and clean up grants when TTL expires or policy state changes.  
  Module: JIT Access

- **TASK-044: Active Grant Lookup API**  
  Description: Provide low-latency API for gateway to validate active entitlements.  
  Module: JIT Access

- **TASK-045: Access Request API**  
  Description: Implement request submission endpoint with reason, timeframe, and asset scope.  
  Module: Approval Workflow

- **TASK-046: Multi-Step Approval Engine**  
  Description: Build state machine for quorum/sequence-based approvals and conditions.  
  Module: Approval Workflow

- **TASK-047: Escalation and SLA Timers**  
  Description: Trigger reminders and escalate pending approvals after SLA breaches.  
  Module: Approval Workflow

- **TASK-048: Notification Service Integration**  
  Description: Integrate email/chat notifications for request and decision events.  
  Module: Approval Workflow

- **TASK-049: End-to-End Correlation IDs**  
  Description: Add distributed tracing and correlation IDs across all request flows.  
  Module: Cross-Cutting Platform

- **TASK-050: Security and Performance Test Harness**  
  Description: Build load tests, authz fuzz tests, and penetration baseline test suites.  
  Module: Cross-Cutting Platform

---

## 8. Frontend Task List (50 Tasks)

- **FE-001: Build Login Page**  
  Description: Create responsive login UI with SSO options and clear validation errors.  
  Page: Auth - Login

- **FE-002: Build MFA Verification Screen**  
  Description: Implement MFA challenge UI supporting TOTP and WebAuthn prompts.  
  Page: Auth - MFA

- **FE-003: Session Timeout Re-Auth Modal**  
  Description: Add modal to handle token expiry and secure reauthentication flow.  
  Page: Global Shell

- **FE-004: Logout and Session Device List UI**  
  Description: Display active sessions and allow user-initiated revocation.  
  Page: Profile - Security

- **FE-005: Unauthorized Access Screen**  
  Description: Provide clear denied-access UI with request-access CTA.  
  Page: Error - 403

- **FE-006: Users List Page**  
  Description: Create searchable and filterable user table with status badges.  
  Page: Admin - Users

- **FE-007: User Detail Drawer**  
  Description: Add detail panel for user roles, group mappings, and recent actions.  
  Page: Admin - Users

- **FE-008: Role Management Page**  
  Description: Implement role listing, creation, editing, and deletion workflows.  
  Page: Admin - Roles

- **FE-009: Permission Matrix Component**  
  Description: Build matrix UI for assigning permissions to roles by resource/action.  
  Page: Admin - Roles

- **FE-010: Group Mapping UI**  
  Description: Add interface to map external groups to internal roles.  
  Page: Admin - Identity

- **FE-011: Policy List Page**  
  Description: Create policy inventory with status, version, and owner columns.  
  Page: Policies - List

- **FE-012: Policy Editor (JSON + Form Mode)**  
  Description: Build dual-mode policy editor with schema validation.  
  Page: Policies - Editor

- **FE-013: Policy Simulation Panel**  
  Description: Add test harness UI for evaluating sample access scenarios.  
  Page: Policies - Simulator

- **FE-014: Policy Version Timeline**  
  Description: Show policy history and rollback actions.  
  Page: Policies - Detail

- **FE-015: Policy Publish Confirmation Flow**  
  Description: Add impact warning and approval dialog before publishing policy changes.  
  Page: Policies - Detail

- **FE-016: Asset Inventory Page**  
  Description: Implement searchable asset table with environment and criticality filters.  
  Page: Assets - Inventory

- **FE-017: Asset Registration Form**  
  Description: Create asset onboarding form with protocol and connectivity fields.  
  Page: Assets - New

- **FE-018: Asset Detail Page**  
  Description: Display metadata, credentials status, and related sessions.  
  Page: Assets - Detail

- **FE-019: Asset Bulk Import Wizard**  
  Description: Build CSV upload, preview, validation, and import confirmation flow.  
  Page: Assets - Import

- **FE-020: Asset Connectivity Test UI**  
  Description: Add one-click connection test and diagnostics output view.  
  Page: Assets - Detail

- **FE-021: Vault Secret Catalog Page**  
  Description: Show secret records without exposing values; include lease and rotation info.  
  Page: Vault - Secrets

- **FE-022: Secret Creation Wizard**  
  Description: Build secure input form for secret metadata and rotation policies.  
  Page: Vault - New Secret

- **FE-023: Rotation Policy Config UI**  
  Description: Create schedule and policy configuration interface for rotation rules.  
  Page: Vault - Policies

- **FE-024: Lease Activity Timeline**  
  Description: Display active and expired leases with revoke action controls.  
  Page: Vault - Leases

- **FE-025: Secret Access Audit Panel**  
  Description: Present audit trail for each secret reference and service usage.  
  Page: Vault - Detail

- **FE-026: Access Request Form Page**  
  Description: Build guided form for selecting asset, reason, scope, and timeframe.  
  Page: Requests - New

- **FE-027: My Requests Dashboard**  
  Description: Display requester submissions with status chips and SLA timers.  
  Page: Requests - My Requests

- **FE-028: Approval Inbox Page**  
  Description: Create approver queue with priority sorting and quick decision actions.  
  Page: Approvals - Inbox

- **FE-029: Request Decision Modal**  
  Description: Capture approval/rejection rationale and optional constraints.  
  Page: Approvals - Inbox

- **FE-030: Escalation Indicator Widget**  
  Description: Highlight requests nearing SLA breach with visual urgency markers.  
  Page: Approvals - Inbox

- **FE-031: JIT Grants List Page**  
  Description: Show active, pending, and expired grants with scope summary.  
  Page: JIT - Grants

- **FE-032: Grant Detail View**  
  Description: Display grant timeline, issuance policy, and associated request details.  
  Page: JIT - Detail

- **FE-033: Grant Renewal Action UI**  
  Description: Implement renewal request flow with additional justification input.  
  Page: JIT - Detail

- **FE-034: Grant Revocation Flow**  
  Description: Add immediate revoke action with confirmation and reason capture.  
  Page: JIT - Detail

- **FE-035: Live Session Monitor Dashboard**  
  Description: Build real-time view of active privileged sessions and risk indicators.  
  Page: Sessions - Live

- **FE-036: Session Detail Timeline**  
  Description: Show chronological events, commands, and policy decisions during session.  
  Page: Sessions - Detail

- **FE-037: Session Termination Control**  
  Description: Add emergency terminate action with RBAC gating and justification.  
  Page: Sessions - Detail

- **FE-038: Session Playback UI**  
  Description: Implement terminal replay player with seek, speed, and marker navigation.  
  Page: Sessions - Playback

- **FE-039: Command Search in Recordings**  
  Description: Add indexed command search and jump-to-time in playback.  
  Page: Sessions - Playback

- **FE-040: Audit Explorer Page**  
  Description: Build filterable event log explorer with advanced query controls.  
  Page: Audit - Explorer

- **FE-041: Audit Event Detail Drawer**  
  Description: Show full event payload, signature status, and related entities.  
  Page: Audit - Explorer

- **FE-042: Compliance Report Builder UI**  
  Description: Create report presets and export scheduling controls.  
  Page: Audit - Reports

- **FE-043: Security Posture Dashboard**  
  Description: Present metrics on approvals, JIT usage, failed access, and policy violations.  
  Page: Dashboard - Security

- **FE-044: Notification Center Component**  
  Description: Implement in-app notifications for approvals, expiries, and policy alerts.  
  Page: Global Shell

- **FE-045: Global Search Bar**  
  Description: Add universal search across users, assets, requests, sessions, and events.  
  Page: Global Shell

- **FE-046: Theme and Accessibility Baseline**  
  Description: Implement WCAG-compliant color contrast, keyboard navigation, and ARIA patterns.  
  Page: Global

- **FE-047: API Error Boundary Framework**  
  Description: Standardize API error handling, retries, and user-friendly failure states.  
  Page: Global

- **FE-048: Frontend Authorization Guards**  
  Description: Add route/component guards based on role and permission claims.  
  Page: Global Routing

- **FE-049: Observability Hooks for UX Telemetry**  
  Description: Capture client-side performance and action traces with correlation IDs.  
  Page: Global

- **FE-050: End-to-End UI Test Suite**  
  Description: Create Cypress/Playwright coverage for auth, requests, approvals, and session flows.  
  Page: QA - Test Suite

---

## 9. API Design

### API Principles
- Base path: `/api/v1`
- Auth: `Authorization: Bearer <token>`
- Idempotency for sensitive POST operations using `Idempotency-Key` header.
- Correlation tracing via `X-Correlation-ID`.
- Standard response envelope:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {
    "request_id": "req_01J..."
  }
}
```

### REST Endpoints by Module

#### Auth Service
- `POST /auth/sso/callback`
- `POST /auth/mfa/verify`
- `POST /auth/token/refresh`
- `POST /auth/logout`

**Example: MFA Verify**

Request:
```json
{
  "challenge_id": "mfa_ch_123",
  "method": "totp",
  "code": "123456"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOi...",
    "refresh_token": "rfr_abc...",
    "expires_in": 900
  },
  "error": null,
  "meta": {
    "request_id": "req_auth_001"
  }
}
```

#### User & Role Management
- `GET /users`
- `GET /users/{userId}`
- `POST /roles`
- `PUT /roles/{roleId}/permissions`

#### Policy Engine
- `POST /policies/evaluate`
- `POST /policies`
- `GET /policies/{policyId}`
- `POST /policies/{policyId}/publish`

**Example: Policy Evaluate**

Request:
```json
{
  "subject": {
    "user_id": "u-1",
    "roles": ["db_admin"]
  },
  "resource": {
    "asset_id": "a-1",
    "type": "server",
    "tags": ["prod"]
  },
  "action": "ssh.connect",
  "context": {
    "ip": "10.1.2.3",
    "time": "2026-03-30T11:00:00Z",
    "risk_score": 18
  }
}
```

Response:
```json
{
  "success": true,
  "data": {
    "decision": "allow",
    "obligations": [
      "record_session",
      "max_duration_60m"
    ]
  },
  "error": null,
  "meta": {
    "request_id": "req_pol_003"
  }
}
```

#### Asset Management
- `POST /assets`
- `GET /assets`
- `GET /assets/{assetId}`
- `POST /assets/{assetId}/test-connection`

#### Vault Service
- `POST /vault/secrets`
- `POST /vault/secrets/{secretId}/issue`
- `POST /vault/secrets/{secretId}/rotate`
- `POST /vault/leases/{leaseId}/revoke`

#### Access Gateway
- `POST /gateway/sessions/start`
- `POST /gateway/sessions/{sessionId}/terminate`
- `GET /gateway/sessions/{sessionId}/status`

**Example: Start Session**

Request:
```json
{
  "user_id": "u-1",
  "asset_id": "a-1",
  "jit_grant_id": "g-1",
  "protocol": "ssh"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "session_id": "s-123",
    "gateway_host": "gw-01.bonipam.internal",
    "connect_command": "ssh -p 3022 s-123@gw-01.bonipam.internal"
  },
  "error": null,
  "meta": {
    "request_id": "req_gw_101"
  }
}
```

#### Session Recorder
- `POST /recordings/ingest`
- `GET /recordings/{sessionId}`
- `GET /recordings/{sessionId}/metadata`
- `POST /recordings/{sessionId}/seal`

#### Audit Logging
- `POST /audit/events`
- `GET /audit/events`
- `GET /audit/events/{eventId}/verify`
- `POST /audit/exports`

#### JIT Access
- `POST /jit/grants`
- `GET /jit/grants/{grantId}`
- `POST /jit/grants/{grantId}/renew`
- `POST /jit/grants/{grantId}/revoke`

#### Approval Workflow
- `POST /requests`
- `GET /requests/{requestId}`
- `POST /requests/{requestId}/approve`
- `POST /requests/{requestId}/reject`

**Example: Create Access Request**

Request:
```json
{
  "requester_id": "u-1",
  "asset_id": "a-1",
  "requested_role": "db_admin",
  "reason": "Emergency patching",
  "start_time": "2026-03-30T12:00:00Z",
  "end_time": "2026-03-30T14:00:00Z"
}
```

Response:
```json
{
  "success": true,
  "data": {
    "request_id": "r-001",
    "status": "pending_approval",
    "next_approvers": ["manager-ops", "security-oncall"]
  },
  "error": null,
  "meta": {
    "request_id": "req_wrk_013"
  }
}
```

---

## 10. Deployment Guide

### Docker Setup

#### Example `docker-compose.yml` (Local Development)
```yaml
version: "3.9"
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: bonipam
      POSTGRES_USER: boni
      POSTGRES_PASSWORD: boni_dev_password
    ports:
      - "5432:5432"

  redis:
    image: redis:7
    ports:
      - "6379:6379"

  api-gateway:
    build: ./services/docker/api-gateway
    environment:
      APP_ENV: development
      DB_DSN: postgres://boni:boni_dev_password@postgres:5432/bonipam?sslmode=disable
      REDIS_ADDR: redis:6379
    depends_on:
      - postgres
      - redis
    ports:
      - "8080:8080"
```

### Environment Variables

#### Global
- `APP_ENV` (`development|staging|production`)
- `LOG_LEVEL` (`debug|info|warn|error`)
- `HTTP_PORT`
- `GRPC_PORT` (optional)

#### Database/Cache
- `DB_DSN`
- `DB_MAX_OPEN_CONNS`
- `DB_MAX_IDLE_CONNS`
- `REDIS_ADDR`
- `REDIS_PASSWORD`

#### Security
- `JWT_SIGNING_KEY_REF`
- `KMS_ENDPOINT`
- `KMS_KEY_ID`
- `MTLS_CERT_PATH`
- `MTLS_KEY_PATH`
- `MTLS_CA_PATH`

#### Integration
- `OIDC_ISSUER_URL`
- `OIDC_CLIENT_ID`
- `OIDC_CLIENT_SECRET`
- `SAML_METADATA_URL`
- `OBJECT_STORAGE_ENDPOINT`
- `OBJECT_STORAGE_BUCKET`

### Basic Deployment Steps
1. Build service containers.
2. Run DB migrations for each service schema.
3. Provision KMS keys and secret references.
4. Deploy backend services (Docker or Kubernetes).
5. Configure ingress, TLS certificates, and DNS.
6. Deploy React frontend with API base URL and SSO configs.
7. Run smoke tests for auth, request, approval, and session lifecycle.
8. Enable observability and alerting (metrics, logs, traces).

### Kubernetes (Optional Baseline)
- Namespace separation: `bonipam-core`, `bonipam-data`, `bonipam-observability`.
- Use Helm charts per service.
- HPA enabled for stateless services.
- StatefulSets for PostgreSQL/Redis if self-hosted.
- NetworkPolicies to restrict east-west traffic.
- PodSecurity standards and runtime seccomp profiles.

---

## 11. Frontend Dashboard & Menu Design

### Global Layout (The Shell)
The application uses a persistent **Global Shell** to provide consistent navigation and context.
- **Sidebar**: Collapsible navigation with grouped menu items. Includes brand logo and user profile summary at the bottom.
- **Topbar**: Breadcrumbs for location awareness, Global Search bar, Notification Center (bell icon), and System Health indicator.
- **Main Content Area**: Responsive container with smooth transitions between views.

### Navigation Menu Structure

#### 1. Overview
- **Dashboard**: Real-time metrics and shortcut widgets.

#### 2. Identity & Access
- **Users**: List, Edit, Sync from IdP.
- **Roles**: RBAC management and permission matrix.
- **Groups**: External IdP group-to-role mappings.

#### 3. Asset Management
- **Inventory**: Categorized list of servers, databases, and clusters.
- **Register Asset**: Onboarding wizard for new infrastructure.

#### 4. Governance
- **Policies**: Versioned policy editor (JSON + Form mode) and simulators.
- **Vault**: Secrets management, rotation policies, and lease tracking.

#### 5. Access Workflows
- **Requests**: Submit and track access requests.
- **Approvals**: Inbox for pending authorization decisions.
- **JIT Grants**: Active time-bound entitlements.

#### 6. Monitoring & Audit
- **Sessions**: Live session monitor and emergency termination.
- **Playback**: Searchable session recordings player.
- **Audit Logs**: Immutable event explorer and compliance reporter.

### Dashboard Widgets (Ideas)
- **Active Sessions**: Counter with "View All" link to live monitor.
- **Pending Approvals**: Badge with count and "Action Needed" list.
- **Expiring Grants**: Warning widget for JIT access ending within 30m.
- **Security Posture Score**: Gauge reflecting MFA usage and policy violations.
- **Recent Activity**: Stream of the last 5 relevant audit events for the user.

---

## Bonus: Future Improvements

### Future Enhancements
- Add database and cloud provider native connectors for dynamic account brokering.
- Introduce risk-adaptive authentication using UEBA signals.
- Add privileged command anomaly detection with ML-assisted alerting.
- Integrate hardware-backed WebAuthn enterprise passkeys.

### Scaling Strategy
- Partition services by domain and scale horizontally with autoscaling.
- Introduce event bus (Kafka/NATS) for decoupled workflows.
- Use read replicas for audit and reporting workloads.
- Apply caching for policy decision artifacts and asset metadata.

### Security Hardening Tips
- Enforce FIPS-compliant crypto modules where required.
- Implement strict CSP and secure headers on frontend delivery.
- Periodically rotate all service credentials and encryption keys.
- Run red-team simulations and incident response game days quarterly.
- Enable immutable backups and cross-region disaster recovery drills.
