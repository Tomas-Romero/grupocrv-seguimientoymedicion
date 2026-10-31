.DEFAULT_GOAL := help
SHELL := /bin/bash

GO          ?= go
TEMPL       ?= templ
MIN_DOMAIN  ?= 85
MIN_TOTAL   ?= 70

# Version del binario standalone de Tailwind. Mantenerla igual que ARG TAILWIND_VERSION del Dockerfile.
TAILWIND_VERSION ?= v4.3.1

.PHONY: help up down dev build run gen css test cover bdd lint fmt check migrate seed trace clean tools tailwind

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

tools: ## Instala las herramientas de desarrollo (correr una sola vez)
	$(GO) install github.com/a-h/templ/cmd/templ@v0.2.793
	$(GO) install github.com/pressly/goose/v3/cmd/goose@v3.26.0
	$(GO) install github.com/air-verse/air@v1.61.7
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	$(MAKE) tailwind

up: ## Levanta Postgres, aplica migraciones y carga datos de ejemplo
	docker compose up -d db
	@echo "Esperando a que Postgres este listo..."
	@# -h 127.0.0.1: sin el host, pg_isready usa el socket, y el servidor temporal de la primera inicializacion ya contesta por ahi
	@# antes de que el definitivo acepte conexiones por red (DEF-008).
	@until docker compose exec -T db pg_isready -h 127.0.0.1 -U metrics -d metrics >/dev/null 2>&1; do sleep 1; done
	$(MAKE) migrate
	$(MAKE) seed

down: ## Baja los contenedores (los datos sobreviven en el volumen)
	docker compose down

gen: ## Genera el codigo Go de las plantillas templ
	$(TEMPL) generate

tailwind: ## Descarga el binario standalone de Tailwind a ./bin si falta o cambio la version
	@mkdir -p bin
	@if [ "$$(cat bin/.tailwind-version 2>/dev/null)" = "$(TAILWIND_VERSION)" ]; then exit 0; fi; \
	os=$$(uname -s); arch=$$(uname -m); destino=bin/tailwindcss; \
	case "$$os-$$arch" in \
	  Linux-x86_64) asset=tailwindcss-linux-x64 ;; \
	  Linux-aarch64|Linux-arm64) asset=tailwindcss-linux-arm64 ;; \
	  Darwin-arm64) asset=tailwindcss-macos-arm64 ;; \
	  Darwin-x86_64) asset=tailwindcss-macos-x64 ;; \
	  MINGW*|MSYS*|CYGWIN*) asset=tailwindcss-windows-x64.exe; destino=bin/tailwindcss.exe ;; \
	  *) echo "Sistema no soportado: $$os-$$arch"; exit 1 ;; \
	esac; \
	echo "Descargando Tailwind $(TAILWIND_VERSION) ($$asset)..."; \
	rm -f bin/tailwindcss bin/tailwindcss.exe; \
	curl -fsSL -o "$$destino" "https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/$$asset"; \
	chmod +x "$$destino"; \
	echo "$(TAILWIND_VERSION)" > bin/.tailwind-version

css: tailwind ## Compila Tailwind (descarga el binario si falta)
	./bin/tailwindcss -i web/static/css/input.css -o web/static/css/app.css --minify

dev: css ## Desarrollo con recarga automatica
	$(TEMPL) generate --watch --proxy=http://localhost:8080 --cmd="$(GO) run ./cmd/server"

build: gen ## Compila el binario en ./bin/server
	$(GO) build -o bin/server ./cmd/server

run: build ## Compila y ejecuta
	./bin/server

migrate: ## Aplica las migraciones pendientes, aunque lleguen fuera de orden (T-012)
	goose -allow-missing -dir migraciones postgres "$${DATABASE_URL}" up

seed: ## Carga datos de ejemplo
	psql -v ON_ERROR_STOP=1 -d "$${DATABASE_URL}" -f migraciones/datos/ejemplo.sql

test: gen ## Corre todos los tests unitarios
	$(GO) test ./... -race -count=1

cover: gen ## Corre tests con cobertura y verifica los umbrales
	bash ./scripts/coverage.sh

bdd: gen ## Corre los escenarios BDD (godog)
	@if ls features/*.feature >/dev/null 2>&1; then $(GO) test ./features/... -race -count=1; else echo "Todavia no hay escenarios BDD (igual que en CI)."; fi

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
