FROM golang:1.23.3 as builder

WORKDIR /app
COPY . /app

RUN go mod tidy

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/main.go

RUN ls -la /app

# Compile the binary
RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://proxy.golang.org go build -o app cmd/main.go

# Stage 2: Final Image
FROM alpine:latest

USER root
RUN apk update && apk add tzdata
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy the binary and other necessary files
COPY --from=builder /app/app .
COPY --from=builder /app/.env .
COPY --from=builder /app/docs .

ENTRYPOINT ["./app"]
