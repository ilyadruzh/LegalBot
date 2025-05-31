test:
	go test ./... -race -count=1

lint:
	golangci-lint run

docker-build:
	docker build -f deploy/Dockerfile.bot -t legalbot/bot:local .
