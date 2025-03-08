run:
	@go run ./cmd

build:
	@go build -o calef ./cmd

up:
	@docker compose up -d
