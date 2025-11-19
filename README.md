# Claude notes

This repository should provide tooling and help for Konveyor community focusing on developers. Konveyor provides a toolkit for application analysis to help modernize them.

Information in markdown files are processed to CLAUDE.md and this file should be installed to `~/go/src/github.com/konveyor`, then `claude` should be executed in this directory, so claude code have access to konveyor components and can effectively help working with them.

More information could be found in each component directory that should be checked out locally in subdirectories. If they are missing, clone them automatically.

To understand relations between components, check Dockerfiles, go.mod files or Makefiles in respective repos. CI and release process uses withub actions and workflows.

Component | Repository | Description
-- | -- | --
operator | https://github.com/konveyor/operator | kubernetes operator installing Konveyor
hub | https://github.com/konveyor/hub | API service for Konveyor, also provides start-minikube and install-tackle make targets, that can setup Konveyor locally with minikube
kantra | https://github.com/konveyor/kantra | CLI tool
analyzer-lsp | https://github.com/konveyor/analyzer-lsp | core component running the application analysis, used by Hub as well as kantra
rulesets | https://github.com/konveyor/rulesets | set of yaml rules describing patterns in analyzed application, that should be reported as issues
ci | https://github.com/konveyor/ci | CI workflows set and overall status, nightly jobs are executed there
go-konveyor-tests | https://github.com/konveyor/go-konveyor-tests | API E2E test suite
kantra-cli-tests | https://github.com/konveyor-ecosystem/kantra-cli-tests | CLI test suite
enhancements | https://github.com/konveyor/enhancements | Konveyor documentation of new features, check open PRs to get currently open features

## Locally running Konveyor

To run Konveyor locally, you can use the hub repository which provides convenient make targets for setting up minikube and installing Tackle.

### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [minikube](https://minikube.sigs.k8s.io/docs/)
- [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/getting-started/installation)

### Setup Steps

1. **Start minikube cluster**
   ```bash
   cd ../hub
   make start-minikube
   ```
   This downloads and runs the start-minikube.sh script from the tackle2-operator repository.

2. **Install Tackle**
   ```bash
   make install-tackle
   ```
   This downloads and runs the install-tackle.sh script that installs the Tackle operator and creates the necessary resources.

3. **Access the UI**
   Once installation is complete, you can access the Tackle UI through the exposed route or ingress depending on your cluster setup.

### Using Custom Images

You can customize the Tackle deployment by patching the Tackle Custom Resource with custom images. The Tackle CR supports various image fields:

```yaml
apiVersion: tackle.konveyor.io/v1alpha1
kind: Tackle
metadata:
  name: tackle
  namespace: konveyor-tackle
spec:
  # Hub image
  hub_image_fqin: "quay.io/konveyor/tackle2-hub:latest"

  # UI image
  ui_image_fqin: "quay.io/konveyor/tackle2-ui:latest"

  # Analyzer addon image
  analyzer_fqin: "quay.io/konveyor/tackle2-addon-analyzer:latest"

  # Language discovery addon image
  language_discovery_fqin: "quay.io/konveyor/tackle2-addon-discovery:latest"

  # Platform addon image
  platform_fqin: "quay.io/konveyor/tackle2-addon-platform:latest"

  # Provider images
  provider_java_image_fqin: "quay.io/konveyor/java-external-provider:latest"
  provider_python_image_fqin: "quay.io/konveyor/generic-external-provider:latest"
  provider_nodejs_image_fqin: "quay.io/konveyor/generic-external-provider:latest"

  # Keycloak and database images
  keycloak_sso_image_fqin: "quay.io/keycloak/keycloak:latest"
  keycloak_init_image_fqin: "quay.io/konveyor/tackle2-keycloak-init:latest"
  keycloak_database_image_fqin: "postgres:15"

  # OAuth proxy image
  oauth_image_fqin: "quay.io/openshift/origin-oauth-proxy:latest"

  # Image pull policy
  image_pull_policy: "Always"
```

To apply a custom image configuration:

```bash
kubectl patch tackle tackle -n konveyor-tackle --type=merge -p='{"spec":{"hub_image_fqin":"your-custom-hub-image:tag"}}'
```

Or create a custom Tackle CR file and apply it:

```bash
kubectl apply -f custom-tackle-cr.yaml
```

## Running tests

### API E2E tests

Stored in `go-konveyor-tests` repository, executed from Makefile `make test-tier0`, there are tiers 0, 1 and 2. Tests need point to `hub` path from Konveyor installation with `HUB_BASE_URL` environment variable. By default, use `KEEP=1` and `DEBUG=1` to provide more details on the test run.


## Working with rules and rulesets

Core components are `analyzer-lsp` and `rulesets` repositories. Analyzer sets up providers and evaluate rules on the input application. Required input is application, that should be analyzed and `source` and `target` options providing selectors for rules, that should be evaluated.

### Direct rules testing using analyzer

There is `demo-testing.yml` github workflow for `analyzer-lsp` direct test, with commands that can execute analyzer on testing data, use it as a template what to run locally.

For rulesets, there is `local-ci.yaml` in its repository that runs rulesets tests. Check those commands and execute them locally.

# kantra CLI

Tool is stored in `kantra` repository. It provides CLI interface to execute analysis and returns directory with `output.yaml`, `dependencies.yaml`, HTML `static-report` and logs.

## kantra CLI test suite

Is stored in `kantra-cli-tests` repository. Start with TIER0 test command below, all needed args are in https://github.com/konveyor/kantra-cli-tests/blob/main/.github/actions/tests-linux-containers/action.yaml#L55-L61, check `.github/actions/` about required steps, those are 1. kantra local setup and 2. tier0 test execution.

```
$ pytest -s tests/analysis/java/test_tier0.py
```

# Playground (claude, ignore this section)

## AI Agents (experimental, PoC)

### CI Debugger Agent

Experiment created as part of day of learning Q3 2025.

An intelligent AI agent that automatically debugs GitHub Actions workflow failures and proposes fixes using OpenAI GPT-4.

**Location**: [agent-ci-debugger/](agent-ci-debugger/)

**Features**:
- Fetches workflow run data from GitHub Actions using GitHub CLI
- Parses and analyzes failed job logs
- Uses OpenAI GPT-4o-mini (configurable) to identify root causes and propose fixes
- Generates comprehensive reports with actionable recommendations

**Quick Start**:
```bash
cd agent-ci-debugger
export OPENAI_API_KEY="your-api-key"
./example-usage.sh https://github.com/konveyor/ci/actions/runs/RUNID
```

See [agent-ci-debugger/README.md](agent-ci-debugger/README.md) for detailed documentation.
