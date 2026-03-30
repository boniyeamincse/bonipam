# [B-PAM] Boni Privileged Access Management

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-19-61DAFB.svg)](https://react.dev)
[![Status](https://img.shields.io/badge/Status-Beta-orange.svg)]()

**B-PAM** (Boni PAM) is a next-generation, open-source Privileged Access Management (PAM) platform designed for modern, cloud-native infrastructure. Built on **Zero Trust** principles, B-PAM centralizes and secures privileged access to servers, databases, and clusters through policy-driven, time-bound sessions.

---

## 🚀 Key Features

### 🔐 Identity & Access
- **Enterprise Auth**: Seamless integration with OIDC and SAML (GitHub, Google, Okta).
- **Adaptive MFA**: Multi-factor authentication via TOTP and WebAuthn.
- **Unified RBAC**: Granular role-based access control with group-to-role mappings.

### 🛡️ Policy Engine
- **ABAC Runtime**: Attribute-Based Access Control considering IP, device trust, time, and risk.
- **JIT Grants**: Just-in-Time access requests with lifecycle management.
- **Versioned Policies**: Immutable policy history with instant rollback capabilities.

### 📦 Asset & Secrets Management
- **Asset Inventory**: Centralized management of servers, DBs, and Kubernetes clusters.
- **Credential Vault**: Encrypted secret storage with automated rotation and lease revocation.
- **KMS Integration**: Envelope encryption using external KMS/HSM providers.

### 🖥️ Secure Connectivity
- **Access Gateway**: Controlled SSH bastion core for secure session brokering.
- **Live Recording**: (Coming Soon) Real-time session capture and audit-ready playbacks.
- **Tamper-Evident Logs**: Signed, hash-chained audit trails for compliance.

---

## 🏗️ Architecture

B-PAM is built using a modular microservices architecture:
- **Backend**: High-performance Go services (Auth, Gateway, Policy, Vault).
- **Frontend**: Premium, animated React 19 dashboard with Framer Motion.
- **Infrastructure**: PostgreSQL for persistence, Redis for caching.

---

## 🛠️ Getting Started

### 1. Prerequisites
- [Go 1.23+](https://golang.org/doc/install)
- [Node.js 20+](https://nodejs.org/en/download/)
- [PostgreSQL](https://www.postgresql.org/download/)

### 2. Installation

#### Clone the Repository
```bash
git clone https://github.com/boniyeamincse/bonipam.git
cd bonipam
```

#### Start Backend Services
```bash
bash backend/services/scripts/run-local.sh
```

#### Start Frontend Dashboard
```bash
cd frontend
npm install
npm run dev
```

### 3. Test Credentials
The dev environment is pre-configured with:
- **Email**: `admin@bonipam.local`
- **Password**: `admin123`
- **Portal**: [http://localhost:5173](http://localhost:5173)

---

## 🤝 Contributing

We love contributors! Please read our [CONTRIBUTING.md](./CONTRIBUTING.md) to get started.
- Found a bug? [Open an issue](https://github.com/boniyeamincse/bonipam/issues).
- Have a feature idea? Join the discussion!

## 🛡️ Security

For reporting security vulnerabilities, please refer to our [SECURITY.md](./SECURITY.md).

---

## 📄 License

Distributed under the Apache License 2.0. See `LICENSE` for more information.

---
*Developed by [Boni Yeamin](https://github.com/boniyeamincse) and the B-PAM Community.*
