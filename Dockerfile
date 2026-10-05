FROM node:22-slim AS frontend-build

WORKDIR /frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend ./
RUN npm run build

FROM golang:1.26-alpine AS builder

ARG VERSION=dev

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
COPY --from=frontend-build /frontend/dist ./internal/web/dist

RUN CGO_ENABLED=0 go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /ytdl-triage ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /ytdl-triage /app/ytdl-triage

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["/app/ytdl-triage"]
