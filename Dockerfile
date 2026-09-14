# --- build ---
FROM golang:1.23-alpine AS build
WORKDIR /src

RUN apk add --no-cache git
RUN go install github.com/a-h/templ/cmd/templ@v0.2.793

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN templ generate
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime ---
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /src/migrations /app/migrations
COPY --from=build /src/web/static /app/web/static
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
