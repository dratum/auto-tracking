.PHONY: up down run build logs ps web-dev web-build web-install docker-build help

COMPOSE_FILE := deployments/docker-compose.yml

## Docker Compose — инфраструктура (TimescaleDB + MongoDB + App)

up:                ## Поднять контейнеры
	docker compose -f $(COMPOSE_FILE) up -d

down:              ## Остановить контейнеры
	docker compose -f $(COMPOSE_FILE) down

logs:              ## Логи контейнеров (follow)
	docker compose -f $(COMPOSE_FILE) logs -f

ps:                ## Статус контейнеров
	docker compose -f $(COMPOSE_FILE) ps

## Docker

docker-build:      ## Собрать Docker-образ приложения
	DOCKER_BUILDKIT=0 docker build -t autotrack-app:latest -f Dockerfile .

## Локальный запуск сервера

run:               ## Запустить сервер локально
	go run ./cmd/server

build:             ## Собрать бинарник
	go build -o bin/server ./cmd/server

## Frontend

web-install:       ## Установить зависимости фронтенда
	cd web && npm install

web-dev:           ## Запустить фронтенд dev-сервер
	cd web && npm run dev

web-build:         ## Собрать фронтенд для продакшена
	cd web && npm run build

help:              ## Показать справку
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'
