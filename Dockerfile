# Implement multi-stage build
# 1. Build stage
FROM golang:1.19-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# 2. Run stage
FROM alpine:3.16
WORKDIR /app
# add binary executable file
COPY --from=builder /app/main .
# add config file
COPY app.env .

EXPOSE 8080
CMD [ "/app/main" ]