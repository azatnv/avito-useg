FROM golang:1.20-alpine

WORKDIR /app

COPY . /app
COPY .env.db /app/.env.db

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /useg-app

CMD ["/useg-app"]
