FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o executor .

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add docker-cli

COPY --from=builder /app/executor .

CMD ["./executor"]
