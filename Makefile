.DEFAULT_GOAL := help
SHELL := /bin/bash

GO          ?= go
TEMPL       ?= templ
MIN_DOMAIN  ?= 85
MIN_TOTAL   ?= 70

.PHONY: help up down dev build run gen css test cover bdd lint fmt check migrate seed trace clean tools

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

tools: ## Instala las herramientas de desarrollo (correr una sola vez)
	$(GO) install github.com/a-h/templ/cmd/templ@v0.2.793
	$(GO) install github.com/pressly/goose/v3/cmd/goose@latest
	$(GO) install github.com/air-verse/air@latest

up: ## Levanta Postgres, aplica migraciones y carga datos de ejemplo
	docker compose up -d db
	@echo "Esperando a que Postgres este listo..."
	@until docker compose exec -T db pg_isready -U metrics -d metrics >/dev/null 2>&1; do sleep 1; done
	$(MAKE) migrate
	$(MAKE) seed

down: ## Baja los contenedores (los datos sobreviven en el volumen)
	docker compose down

gen: ## Genera el codigo Go de las plantillas templ
	$(TEMPL) generate

css: ## Compila Tailwind
	./bin/tailwindcss -i web/static/css/input.css -o web/static/css/app.css --minify

dev: ## Desarrollo con recarga automatica
	$(TEMPL) generate --watch --proxy=http://localhost:8080 --cmd="$(GO) run ./cmd/server"

build: gen ## Compila el binario en ./bin/server
	$(GO) build -o bin/server ./cmd/server

run: build ## Compila y ejecuta
	./bin/server

migrate: ## Aplica las migraciones pendientes
	goose -dir migrations postgres "$${DATABASE_URL}" up

seed: ## Carga datos de ejemplo
	psql "$${DATABASE_URL}" -f scripts/seed.sql

test: gen ## Corre todos los tests unitarios
	$(GO) test ./... -race -count=1

cover: gen ## Corre tests con cobertura y verifica los umbrales
	./scripts/coverage.sh

bdd: gen ## Corre los escenarios BDD (godog)
	$(GO) test ./features/... -race -count=1

lint: ## Analisis estatico
	$(GO) vet ./...
	golangci-lint run

fmt: ## Formatea el codigo
	$(GO) fmt ./...
	$(TEMPL) fmt .

check: lint cover bdd ## Todo lo que corre CI. Ejecutar ANTES de abrir un PR.
	@echo ""
	@echo "  Todo en verde. Ya podes abrir el PR."

trace: ## Regenera la matriz de trazabilidad
	$(GO) run ./cmd/trace > docs/trazabilidad.md

clean: ## Borra artefactos generados
	rm -rf bin/ coverage.out domain.out coverage.html
	find . -name '*_templ.go' -delete
