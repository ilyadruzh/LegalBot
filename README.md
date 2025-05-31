# LegalBot

LegalBot is a Telegram bot for assisting users with legal claims in Russia. It collects the user's problem, builds a "golden" prompt for OpenRouter, generates advice along with PDF and DOCX documents, and sends them back via Telegram.

## Features
- Commands: `/start`, `/help`, `/claim`, `/status`, `/delete`
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
docker compose up --build
```
