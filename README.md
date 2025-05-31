# LegalBot

LegalBot is a Telegram bot for assisting users with legal claims in Russia. It collects the user's problem, builds a "golden" prompt for OpenRouter, generates advice along with PDF and DOCX documents, and sends them back via Telegram.

## Features
- Commands: `/start`, `/help`, `/claim`, `/status`, `/delete`, `/lang`
- Input text up to 8000 characters
- Rate limit: 10 requests per minute per user
- Generates PDF and DOCX versions of claim letters and lawsuits
- [Data policy](DATA_POLICY.md) and `/delete` command for removing history

## Repository Structure
```
.
├─ cmd/
│  ├─ bot/      # Telegram webhook service
│  ├─ prompt/   # Prompt builder gRPC service
│  └─ worker/   # Task consumer
├─ internal/
│  ├─ telegram/    # Telegram SDK wrapper
│  ├─ openrouter/  # REST client for OpenRouter
│  ├─ prompt/      # Prompt templates
│  └─ db/          # Postgres repositories
├─ deploy/
│  ├─ docker-compose.yml
│  ├─ Dockerfile.bot
│  ├─ Dockerfile.worker
│  └─ Dockerfile.prompt
├─ Makefile
├─ SPEC.md        # Technical specification
```

## Running Locally
```
cp .env.example .env
# set TELEGRAM_SECRET_TOKEN to the value configured in your bot settings
docker compose up --build
```

`DOCS_BASE_URL` can be used to customize the base URL for document links.

## Linting
```bash
make lint
```
Uses `golangci-lint` with settings in `.golangci.yml`.

## Deployment
Automated deployment is handled by GitHub Actions. Secrets such as the server
IP address and SSH key are stored in HashiCorp Vault. The workflow reads these
values using `hashicorp/vault-action@v2` and then performs an SSH deployment via
`appleboy/ssh-action`.

Configure Vault with `SERVER_IP` and `SERVER_SSH_KEY` keys under the
`secret/data/legalbot` path. Provide `VAULT_ADDR`, `VAULT_ROLE_ID` and
`VAULT_SECRET_ID` as GitHub repository secrets so the workflow can fetch the
credentials during the deploy job.
