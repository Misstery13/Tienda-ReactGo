# ===== Etapa de compilación =====
FROM golang:alpine AS build

WORKDIR /src

# Las dependencias van antes que el código para aprovechar la caché de capas.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Binario estático: la imagen final no necesita bibliotecas de C.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api .

# ===== Imagen final =====
FROM alpine:latest

# Certificados raíz por si el backend llama a servicios externos por HTTPS.
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app

COPY --from=build /out/api /usr/local/bin/api

USER app
EXPOSE 3000

CMD ["api"]
