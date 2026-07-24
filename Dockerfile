# syntax=docker/dockerfile:1

# ---- Stage 1: build the Vue3 frontend (views/) ----
FROM node:22-alpine AS views-build
WORKDIR /app/views
COPY views/package.json views/package-lock.json* ./
RUN npm ci
COPY views/ .
RUN npm run build

# ---- Stage 2: build the Go backend (server/), embedding the frontend build ----
FROM golang:1.23-alpine AS server-build
WORKDIR /app/server
COPY server/go.mod ./
COPY server/go.sum* ./
RUN go mod download 2>/dev/null || true
COPY server/ .
# Populate the go:embed source directory with the built frontend assets
# before compiling, so the binary ships the real SPA instead of the placeholder.
RUN rm -rf internal/web/dist/*
COPY --from=views-build /app/views/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Stage 3: minimal runtime image ----
FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=server-build /out/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
