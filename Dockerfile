FROM golang:1.24.3-alpine3.20
WORKDIR /app

COPY . .

EXPOSE 5000
CMD ["go", "run", "server.go"]