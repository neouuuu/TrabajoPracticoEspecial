# Compilación

# Una imagen de alpine con Go pre-instalado para compilar la aplicación.
FROM golang:1.26-alpine AS builder

# Donde se van a guardar los archivos.
WORKDIR /app

# Instalación de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Se copia el codigo fuente.
COPY . .

# Se compila el main.go en un binario llamado server. 
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Ejecución
FROM alpine:latest

WORKDIR /app

# Nos quedamos solo con el binario compilado.
COPY --from=builder /app/server .

# Escuchando en el puerto 8080.
EXPOSE 8080

# Se ejecuta el binario por defecto cuando se inicia el contenedor.
CMD ["./server"]