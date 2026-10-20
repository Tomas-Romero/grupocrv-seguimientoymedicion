# --- build ---
FROM golang:1.23-alpine AS build
WORKDIR /src

# libstdc++ y libgcc los necesita el binario de Tailwind (variante musl).
RUN apk add --no-cache git curl libstdc++ libgcc
RUN go install github.com/a-h/templ/cmd/templ@v0.2.793

# Mantener la version igual que TAILWIND_VERSION del Makefile.
ARG TAILWIND_VERSION=v4.3.1
RUN set -eu; \
    case "$(uname -m)" in \
      x86_64) asset=tailwindcss-linux-x64-musl ;; \
      aarch64) asset=tailwindcss-linux-arm64-musl ;; \
      *) echo "arquitectura no soportada: $(uname -m)"; exit 1 ;; \
    esac; \
    curl -fsSL -o /usr/local/bin/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/${asset}"; \
    chmod +x /usr/local/bin/tailwindcss

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN templ generate
RUN tailwindcss -i web/static/css/input.css -o web/static/css/app.css --minify
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime ---
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /src/web/static /app/web/static
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
