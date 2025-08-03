# Stage 1: Build the Go application
FROM golang:1.18-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /main .

# Stage 2: Create the final image
FROM alpine:latest
WORKDIR /root/
COPY --from=build /main .
CMD ["./main"]
