# Stage 1: Build frontend
FROM node:22-slim AS web

WORKDIR /app/web
COPY web/package.json web/package-lock.json ./

RUN npm ci

COPY web/ ./

RUN npm run build

# Stage 2: Build backend
FROM golang:1.25 AS server

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server


# Stage 3: Final image
FROM debian:trixie-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
WORKDIR /app

COPY --from=server /server ./server
COPY --from=web /app/web/build ./web/build

EXPOSE 8080
CMD ["./server"]
