# Contributing to BONI PAM (B-PAM)

Thank you for your interest in contributing to B-PAM! We welcome contributions from the community to help make B-PAM the most secure and accessible open-source PAM platform.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. (Placeholder for link to CoC).

## How Can I Contribute?

### Reporting Bugs
- Use the **GitHub Issues** tracker to report bugs.
- Include a clear title, description, steps to reproduce, and expected result.
- Attach screenshots or logs if applicable.

### Suggesting Enhancements
- Open a **GitHub Issue** with the "enhancement" label.
- Explain why this feature is needed and how it fits into the B-PAM vision.

### Pull Requests
1. **Fork** the repository.
2. **Clone** your fork: `git clone https://github.com/YOUR_USERNAME/bonipam.git`
3. **Create a branch**: `git checkout -b feature/your-feature-name`
4. **Make your changes**: Ensure code follows Go/React best practices.
5. **Run tests**: `go test ./...` in backend and `npm test` in frontend.
6. **Commit**: Use descriptive commit messages.
7. **Push**: `git push origin feature/your-feature-name`
8. **Open a PR**: Submit your PR to the `main` branch with a detailed description.

## Development Setup

### Backend (Go)
- **Pre-requisites**: Go 1.23+, PostgreSQL, Redis.
- **Run**: `bash backend/services/scripts/run-local.sh`

### Frontend (React)
- **Pre-requisites**: Node.js 20+, npm.
- **Run**: `cd frontend && npm install && npm run dev`

## Community
Join our [Discord/Slack] (Placeholder) to discuss development and get help!

---
*B-PAM is built with ❤️ for the security community.*
