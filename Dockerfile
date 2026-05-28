# --- STAGE 1: Build the Go Binary ---
FROM golang:1.24-alpine AS builder

# Git ve SSL sertifikaları için gerekli araçları yükleyelim
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Bağımlılıkları kopyalayalım ve indirelim
COPY go.mod ./
RUN go mod download

# Kaynak kodları kopyalayalım
COPY . .

# Binary'i optimize ve statik olarak derleyelim (CGO devre dışı)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o gsm-core main.go

# --- STAGE 2: Lightweight Runtime Environment ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Builder stage'den derlenmiş binary'i alalım
COPY --from=builder /app/gsm-core .

# Frontend klasörünü kopyalayalım (Statik servis için)
COPY --from=builder /app/frontend ./frontend

# Uygulama portunu dışarı açalım
EXPOSE 8080

# Uygulamayı çalıştıralım
CMD ["./gsm-core"]
