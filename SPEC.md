Техническое задание

КОНТЕКСТ
• Назначение: Telegram-бот принимает юридическую проблему пользователя, собирает «золотой» промпт, вызывает OpenRouter и возвращает советы + PDF / DOCX претензий и исков.
• SLA: не менее 95 % ответов за ≤ 10 с, 99 % без ошибок.
• Языки: русский.
ВЫСОКО-УРОВНЕВАЯ АРХИТЕКТУРА (ASCII-схема)
User → Telegram API → Bot Service (Go Gin) ─┬─→ Prompt Builder gRPC
│
├─→ Redis (rate-limit)
│
├─→ RabbitMQ → Task Queue → Worker Pool (Go)
│ │
│ ├─→ OpenRouter REST
│ └─→ Postgres
└─→ wkhtmltopdf side-car (PDF)
▲
└───── Prometheus + Grafana + Loki (обзор).

ПОТОК ЗАПРОСА
Пользователь вводит /claim. Bot Service валидирует данные и кладёт задачу claim.create в RabbitMQ.
Worker забирает задачу, вызывает Prompt Builder.Build(), собирает промпт.
Выполняет POST https://openrouter.ai/v1/chat/completions.
Результат (JSON) сохраняется в Postgres, создаётся PDF через wkhtmltopdf.
Bot Service отправляет пользователю разъяснение + файл.
ФУНКЦИОНАЛЬНЫЕ ТРЕБОВАНИЯ
IDТребованиеКритерий приёмки
F-01Команды /start, /help, /claim, /statusОтвет ≤ 200 мс
F-02Приём текста до 8 000 символовБот режет/склеивает без потери данных
F-03Генерация PDF и DOCX
F-05Ограничение 10 запроса в минуту на пользователя11-й запрос отклонён Redis-bucket’ом
НЕФУНКЦИОНАЛЬНЫЕ ТРЕБОВАНИЯ
Производительность: P95 задержка Bot Service ≤ 400 мс
Масштабирование: Worker Pool горизонтально до 100 запросов/с
Безопасность: TLS 1.3, проверка X-Telegram-Bot-Api-Secret-Token, секреты в Vault
Контейнеры: multi-stage, итоговый образ < 300 MB
Покрытие тестами: ≥ 90 % unit, ≥ 70 % интеграционных

СТРУКТУРА РЕПОЗИТОРИЯ
.
├─ cmd/
│ ├─ bot/ (веб-хуки Telegram)
│ ├─ prompt/ (gRPC-сервис промптов)
│ └─ worker/ (консьюмер задач)
├─ internal/
│ ├─ telegram/ (SDK-обёртка)
│ ├─ openrouter/ (REST-клиент)
│ ├─ prompt/ (шаблоны)
│ └─ db/ (Postgres-репозитории)
├─ deploy/
│ ├─ docker-compose.yml
│ ├─ Dockerfile.bot
│ ├─ Dockerfile.worker
│ └─ Dockerfile.prompt
├─ .github/workflows/deploy.yml
├─ Makefile
├─ README.md
└─ SPEC.md (этот файл)

ШАБЛОН «ЗОЛОТОГО» ПРОМПТА (пример)
SYSTEM:
You are a licensed Russian attorney with 15+ years practice…
CONTEXT:
– Jurisdiction: Russian Federation
– Date: {{ .Date }}
– Law excerpts: {{ .Laws }}
USER_QUESTION:
{{ .UserText }}
TASKS:
Qualify the issue.
Advise step-by-step actions.
Draft claim letter (Markdown).
Draft lawsuit (Markdown).
STYLE: Russian, formal, references to articles.
OUTPUT_FORMAT: JSON with keys [advice_md, claim_md, lawsuit_md]

GITHUB ACTIONS — АВТОМАТИЧЕСКОЕ ОБНОВЛЕНИЕ
Файл .github/workflows/deploy.yml:

name: CI/CD
on:
push:
branches: [ main ]

jobs:
build-and-deploy:
runs-on: ubuntu-latest
env:
REGISTRY: ghcr.io
IMAGE_NAME: ${{ github.repository }}-bot
steps:
- uses: actions/checkout@v4
- uses: actions/setup-go@v5
with:
go-version: '1.22'
  - name: Run tests  
    run: make test  

  - name: Build Docker image  
    run: docker build -f deploy/Dockerfile.bot -t $REGISTRY/$IMAGE_NAME:${{ github.sha }} .  

  - name: Push to GHCR  
    uses: docker/login-action@v3  
    with:  
      registry: ${{ env.REGISTRY }}  
      username: ${{ github.actor }}  
      password: ${{ secrets.GITHUB_TOKEN }}  

  - run: docker push $REGISTRY/$IMAGE_NAME:${{ github.sha }}  

  - name: Deploy via SSH & docker compose  
    uses: appleboy/ssh-action@v1.0.0  
    with:  
      host: ${{ secrets.SERVER_IP }}  
      username: deploy  
      key: ${{ secrets.SERVER_SSH_KEY }}  
      script: |  
        cd /opt/legalbot  
        docker pull $REGISTRY/$IMAGE_NAME:${{ github.sha }}  
        docker compose down  
        IMAGE_TAG=${{ github.sha }} docker compose up -d

ВЫДЕРЖКА docker-compose.yml
version: "3.9"
services:
bot:
image: ghcr.io/owner/legalbot-bot:${IMAGE_TAG:-latest}
env_file:
- .env
depends_on:
- postgres
- rabbitmq
restart: always

prompt:
image: ghcr.io/owner/legalbot-prompt:${IMAGE_TAG:-latest}
env_file: [.env]
restart: always

worker:
image: ghcr.io/owner/legalbot-worker:${IMAGE_TAG:-latest}
env_file: [.env]
depends_on: [rabbitmq, postgres]
restart: always

postgres:
image: postgres:16
environment:
POSTGRES_PASSWORD: legalbot
volumes: [db_data:/var/lib/postgresql/data]

rabbitmq:
image: rabbitmq:3-management

grafana:
image: grafana/grafana

volumes:
db_data:

MAKEFILE (основные цели)
test:
 go test ./... -race -count=1

lint:
 golangci-lint run

docker-build:
 docker build -f deploy/Dockerfile.bot -t legalbot/bot:local .

README — ЗАПУСК ЛОКАЛЬНО
git clone https://github.com/owner/legalbot.git
cd legalbot
cp .env.example .env
docker compose up --build
