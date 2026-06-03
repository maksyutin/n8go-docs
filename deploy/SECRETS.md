# Secrets Reference

## GitHub Actions
Settings → Secrets and variables → Actions → New repository secret

### Universal (all providers)
| Secret | Description | Example |
|--------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | Bot token from @BotFather | `1234567890:AAF...` |
| `TELEGRAM_CHAT_ID` | Chat or channel ID | `-1001234567890` |

### Stage environment (docker-ssh / VPS)
| Secret | Description |
|--------|-------------|
| `STAGE_SSH_HOST` | IP or hostname of stage server |
| `STAGE_SSH_USER` | SSH user (e.g. `deploy`) |
| `STAGE_SSH_PORT` | SSH port (default `22`) |
| `STAGE_SSH_PRIVATE_KEY` | PEM private key (full content, including header/footer) |

### Prod environment (docker-ssh / VPS)
| Secret | Description |
|--------|-------------|
| `PROD_SSH_HOST` | IP or hostname of prod server |
| `PROD_SSH_USER` | SSH user |
| `PROD_SSH_PORT` | SSH port |
| `PROD_SSH_PRIVATE_KEY` | PEM private key |

### Yandex Cloud provider (additional)
| Secret | Description |
|--------|-------------|
| `YC_SERVICE_ACCOUNT_KEY_JSON` | Full JSON of YC SA key |
| `YC_CLOUD_ID` | Yandex Cloud ID |
| `YC_FOLDER_ID` | Folder ID |
| `YC_REGISTRY_ID` | Container Registry ID |

## GitHub Registry (GHCR)
`GITHUB_TOKEN` is auto-provided. No manual secret needed.
Ensure `packages: write` permission is set in the workflow.

## GitLab CI
Settings → CI/CD → Variables

| Variable | Protected | Masked |
|----------|-----------|--------|
| `STAGE_SSH_PRIVATE_KEY` | ✓ | ✓ |
| `STAGE_SSH_HOST` | ✓ | — |
| `STAGE_SSH_USER` | ✓ | — |
| `STAGE_SSH_PORT` | ✓ | — |
| `PROD_SSH_PRIVATE_KEY` | ✓ | ✓ |
| `PROD_SSH_HOST` | ✓ | — |
| `PROD_SSH_USER` | ✓ | — |
| `PROD_SSH_PORT` | ✓ | — |
| `TELEGRAM_BOT_TOKEN` | ✓ | ✓ |
| `TELEGRAM_CHAT_ID` | ✓ | — |

## GitVerse CI
Settings → CI/CD → Secrets (same names as GitHub Actions)

Additional GitVerse-specific:
| Secret | Description |
|--------|-------------|
| `REGISTRY_USER` | GitVerse registry username |
| `REGISTRY_TOKEN` | GitVerse registry token/password |

## GitFlic CI
Settings → CI/CD → Variables (same convention as GitLab)

Additional GitFlic-specific:
| Variable | Description |
|----------|-------------|
| `REGISTRY_USER` | GitFlic registry user |
| `REGISTRY_TOKEN` | GitFlic registry token |

## How to generate an SSH deploy key

```bash
# On your local machine
ssh-keygen -t ed25519 -C "ci-deploy" -f ~/.ssh/deploy_key -N ""

# Add the PUBLIC key to the server
ssh-copy-id -i ~/.ssh/deploy_key.pub deploy@your-server

# Copy the PRIVATE key content into the secret
cat ~/.ssh/deploy_key
```

## GitHub Environment protection (manual approval for prod)
1. Settings → Environments → prod → Add required reviewer
2. Add yourself or a team as required reviewer
3. The `deploy-prod` job will pause and wait for approval before running
