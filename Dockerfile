FROM golang:1.20-alpine

COPY . /app

WORKDIR /app

ENTRYPOINT ["go", ""]
