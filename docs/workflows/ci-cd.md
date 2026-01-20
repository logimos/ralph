# CI/CD Integration

Running Ralph in continuous integration and deployment pipelines.

## Environment Detection

Ralph automatically detects CI environments:

| Environment | Detection |
|-------------|-----------|
| GitHub Actions | `GITHUB_ACTIONS` |
| GitLab CI | `GITLAB_CI` |
| Jenkins | `JENKINS_URL` |
| CircleCI | `CIRCLECI` |
| Travis CI | `TRAVIS` |
| Azure DevOps | `TF_BUILD` |
| Generic CI | `CI` |

## Automatic Adaptations

In CI environments, Ralph automatically:

- Enables verbose output for better logs
- Increases timeouts for slower runners
- Disables colors for cleaner logs
- Adjusts parallel hints based on resources

## GitHub Actions

```yaml
name: Ralph Development

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  develop:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          
      - name: Install Ralph
        run: go install github.com/start-it/ralph@latest
        
      - name: Run Development Iterations
        run: |
          ralph -iterations 5 -json-output
        env:
          CURSOR_API_KEY: ${{ secrets.CURSOR_API_KEY }}
          
      - name: Check Progress
        run: ralph -list-all
```

## GitLab CI

```yaml
stages:
  - develop
  - test

develop:
  stage: develop
  image: golang:1.21
  script:
    - go install github.com/start-it/ralph@latest
    - ralph -iterations 5 -verbose
  artifacts:
    paths:
      - plan.json
      - progress.txt
    expire_in: 1 week
```

## Jenkins Pipeline

```groovy
pipeline {
    agent any
    
    stages {
        stage('Setup') {
            steps {
                sh 'go install github.com/start-it/ralph@latest'
            }
        }
        
        stage('Develop') {
            steps {
                sh 'ralph -iterations 5 -verbose'
            }
        }
        
        stage('Verify') {
            steps {
                sh 'ralph -list-tested'
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'plan.json,progress.txt'
        }
    }
}
```

## Configuration for CI

Create a CI-optimized config:

```yaml
# .ralph.yaml
agent: cursor-agent
build_system: go
verbose: true
json_output: true

# Recovery settings
max_retries: 5
auto_replan: true
replan_threshold: 5

# Scope control
scope_limit: 3
deadline: "30m"
```

## Best Practices

### 1. Use JSON Output

```bash
ralph -iterations 5 -json-output
```

Parse output for automated decisions:

```bash
ralph -iterations 5 -json-output | jq '.type'
```

### 2. Set Appropriate Timeouts

CI runners may be slower:

```yaml
deadline: "1h"
scope_limit: 5
```

### 3. Archive Artifacts

Always save plan and progress files:

```yaml
# GitHub Actions
- uses: actions/upload-artifact@v4
  with:
    name: ralph-outputs
    path: |
      plan.json
      progress.txt
```

### 4. Use Secrets for API Keys

Never hardcode API keys:

```yaml
env:
  CURSOR_API_KEY: ${{ secrets.CURSOR_API_KEY }}
```

### 5. Enable Auto-Replan

For long runs, enable replanning:

```bash
ralph -iterations 20 -auto-replan
```

## Validating in CI

Run validations after iterations:

```yaml
- name: Run Validations
  run: ralph -validate
```

## Conditional Workflows

Only run on plan changes:

```yaml
on:
  push:
    paths:
      - 'plan.json'
      - '.ralph.yaml'
```

## Matrix Builds

Test with different configurations:

```yaml
strategy:
  matrix:
    build-system: [go, npm, cargo]
    
steps:
  - run: ralph -iterations 3 -build-system ${{ matrix.build-system }}
```
