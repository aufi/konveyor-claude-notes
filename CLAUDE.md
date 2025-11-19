# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is claude-notes, a documentation and tooling repository for the Konveyor community. Konveyor is a toolkit for application analysis and modernization, helping developers migrate applications between platforms and frameworks.

This repository should be cloned to `~/go/src/github.com/konveyor/` alongside other Konveyor component repositories. When working with Konveyor components, execute `claude` from this directory to provide Claude Code with comprehensive context about the entire Konveyor ecosystem.

## Konveyor Ecosystem Architecture

The Konveyor project consists of multiple interconnected components in sibling directories:

### Core Components
- **hub** - REST API service and task management system, built in Go with Gin/GORM
- **operator** - Kubernetes operator for installing and managing Konveyor (Ansible-based)
- **kantra** - CLI tool for offline application analysis (Go)
- **analyzer-lsp** - Language Server Protocol implementation for code analysis engine (Go)
- **go-konveyor-tests** - API E2E integration test suite for Hub (Go)
- **kantra-cli-tests** - CLI test suite for kantra (Python/pytest)
- **rulesets** - YAML rule definitions for application analysis patterns
- **ci** - CI workflows and overall status tracking, nightly jobs execution
- **enhancements** - Documentation of new features (check open PRs for in-progress features)

### Component Relationships
- Hub provides the API and task orchestration layer
- Operator deploys and manages Hub + UI on Kubernetes clusters
- Kantra uses analyzer-lsp for offline analysis without Hub
- Hub uses analyzer-lsp via Kubernetes pod tasks for online analysis
- Both Hub and Kantra consume rulesets for analysis rules
- go-konveyor-tests validates the entire stack integration
- kantra-cli-tests validates kantra CLI functionality

**Understanding Dependencies:**
To understand relationships between components, examine:
- `Dockerfile` - Image build dependencies and runtime requirements
- `go.mod` - Go module dependencies
- `Makefile` - Build targets and inter-component references
- `.github/workflows/` - CI processes and release workflows

## Local Development Setup

### Prerequisites
- kubectl, minikube, docker/podman installed
- Go 1.21+ for component development

### Quick Setup Commands
```bash
# Start local Konveyor environment
cd ../hub
make start-minikube    # Sets up minikube with OLM
make install-tackle    # Installs Konveyor operator and components

# Access UI at minikube IP (get with `minikube ip`)
```

### Running Tests

#### API E2E Tests (go-konveyor-tests)
```bash
cd ../go-konveyor-tests
export HUB_BASE_URL="http://$(minikube ip)/hub"
make test-tier0    # Core functionality tests
make test-tier1    # Standard feature tests
make test-tier2    # Advanced feature tests
# Use KEEP=1 and DEBUG=1 for more details
KEEP=1 DEBUG=1 make test-tier0
```

#### Kantra CLI Tests (kantra-cli-tests)
Tests use pytest and require kantra binary setup. Check `.github/actions/` for required steps.
```bash
cd ../kantra-cli-tests
# Run TIER0 tests
pytest -s tests/analysis/java/test_tier0.py
```

## Development Workflows

### Working with Hub API
Hub development requires understanding the task-based analysis system:
- Applications are analyzed via Kubernetes pod tasks
- Tasks execute analyzer-lsp containers with rule engines
- Results are stored in SQLite database with GORM models
- API provides REST endpoints for managing applications, analyses, and tasks

### Component Build Commands
```bash
# Hub
cd ../hub && make hub           # Build hub binary
cd ../hub && make run           # Run locally for development
cd ../hub && make test          # Run unit tests
cd ../hub && make test-api      # Run API tests
cd ../hub && make fmt           # Format code
cd ../hub && make vet           # Run go vet

# Kantra CLI
cd ../kantra && go build        # Build kantra binary
cd ../kantra && go test ./...   # Run tests

# Analyzer LSP
cd ../analyzer-lsp && make      # Build analyzer binary

# Operator
cd ../operator && make bundle   # Build operator bundle
cd ../operator && make docker-build    # Build operator image
cd ../operator && make install # Install via Helm
cd ../operator && make uninstall # Uninstall via Helm
```

### Testing Rules and Rulesets
Direct testing of analyzer and rulesets without full Hub deployment:

```bash
# Analyzer LSP testing
cd ../analyzer-lsp
# Check demo-testing.yml GitHub workflow for example commands

# Rulesets testing
cd ../rulesets
# Check local-ci.yaml for test commands to run locally
```

### Testing Custom Images
Patch the Tackle CR to use custom development images:
```bash
kubectl patch tackle tackle -n konveyor-tackle --type=merge \
  -p='{"spec":{"hub_image_fqin":"your-custom-hub:tag"}}'
```

### Environment Configuration
Key environment variables for development:
- `HUB_BASE_URL` - Points tests to Hub API endpoint (default: http://localhost:8080)
- `DB_PATH` - SQLite database location for Hub
- `NAMESPACE` - Kubernetes namespace (default: konveyor-tackle)
- `AUTH_REQUIRED=false` - Disable auth for local development

## Architecture Patterns

### Task Execution Model
- Hub orchestrates analysis as Kubernetes pod tasks
- Each task runs analyzer-lsp with specific rulesets
- Tasks have priority-based scheduling with preemption
- Lifecycle: Created → Ready → Running → Succeeded/Failed

### Database Layer
- Hub uses SQLite with GORM ORM (built with json1 tag)
- Migration system in hub/migration/
- Connection pooling (default 1 connection)
- JSON support via json1 SQLite extension

### Authentication & Authorization
- Keycloak integration for SSO
- JWT token handling in hub/auth/
- Can be disabled for local development

## Test Organization

Tests are organized by tiers in go-konveyor-tests:
- **Tier 0**: Core functionality that must never fail
- **Tier 1**: Standard features for most users
- **Tier 2**: Advanced features and edge cases
- **Tier 3**: Tests requiring credentials/private resources

Each test tier validates different aspects of the analysis pipeline and API functionality.

## Common Development Commands

### Hub Development
```bash
cd ../hub
make generate       # Generate code and manifests
make docs          # Build all documentation
make debug         # Build with debug symbols
make addon         # Build sample addon
```

### Operator Development
```bash
cd ../operator
make run           # Run operator locally
make bundle-build  # Build bundle image
make catalog-build # Build catalog image
make start-minikube install-tackle  # Full local setup
```

### Testing Integration
```bash
cd ../go-konveyor-tests
make setup         # Setup local minikube with tackle
make clean         # Clean local minikube
make test-all      # Run all test tiers
make update-hub    # Update to latest hub dependency
```

### Image Management
```bash
# Build custom images
cd ../hub && make docker-build IMG=your-registry/hub:tag
cd ../operator && make docker-build IMG=your-registry/operator:tag

# Push images
cd ../hub && make docker-push IMG=your-registry/hub:tag
cd ../operator && make docker-push IMG=your-registry/operator:tag
```

## Experimental Tools

### CI Debugger Agent (agent-ci-debugger/)
AI-powered tool that automatically analyzes GitHub Actions workflow failures and proposes fixes using OpenAI GPT-4.

**Quick Start:**
```bash
cd agent-ci-debugger
export OPENAI_API_KEY="your-api-key"
go build -o github-workflow-debugger github-workflow-debugger.go

# Analyze a failed workflow
./github-workflow-debugger https://github.com/konveyor/ci/actions/runs/RUNID

# Or use the convenience script
./example-usage.sh https://github.com/konveyor/ci/actions/runs/RUNID
```

**Features:**
- Fetches workflow data via GitHub CLI (`gh`)
- Extracts error patterns, timeouts, and failed tests
- Uses GPT-4o-mini (default) or GPT-4o for analysis
- Generates markdown reports with root cause and fix proposals
- Handles large logs with intelligent filtering

**Configuration:**
```bash
# Use better quality model (higher cost)
export OPENAI_MODEL="gpt-4o"

# Default (cost-effective)
export OPENAI_MODEL="gpt-4o-mini"
```

See `agent-ci-debugger/README.md` for detailed documentation.