# syntax=docker/dockerfile:1

# ---------- build stage ----------
FROM golang:1.25 AS build
WORKDIR /src
# Cache deps first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Static, stripped binary. CGO off so it runs on distroless/static.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---------- runtime stage ----------
# distroless/static:nonroot = no shell, no package manager, runs as UID 65532.
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
