# Estágio de Compilação
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instala dependências do sistema
RUN apk add --no-cache git

# Copia arquivos de dependência e baixa os pacotes
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Compila o binário do chaincode
RUN CGO_ENABLED=0 GOOS=linux go build -o chaincode kitnet.go

# Estágio Final
FROM alpine:3.18

WORKDIR /app

# Copia o binário compilado
COPY --from=builder /app/chaincode .

# Expõe a porta padrão se necessário (Chaincode as a Service usa portas específicas)
# EXPOSE 9999

# Comando para iniciar o chaincode
CMD ["./chaincode"]
