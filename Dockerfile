# ---------- build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# สำหรับ https call ต่าง ๆ (สำคัญ)
RUN apk add --no-cache ca-certificates

# go modules
COPY go.mod go.sum ./
RUN go mod download

# source
COPY . .

# build binary จาก cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app ./cmd/api

# ---------- runtime stage ----------
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/app /app/app

EXPOSE 8080
USER nonroot:nonroot

CMD ["/app/app"]