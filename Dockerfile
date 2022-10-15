# Implement multi-stage build
# 1. Build stage
FROM golang:1.19-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
# install curl in the builder stage image
RUN apk add curl
# Run curl command to download and extract the migrate binary
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

# 2. Run stage
FROM alpine:3.16
WORKDIR /app
# add binary executable file
COPY --from=builder /app/main .
# copy from builder the downloaded migrate binary to the final image
COPY --from=builder /app/migrate ./migrate
# add config file
COPY app.env .
# copy start.sh file into the docker image
COPY start.sh .
# copy wait-for.sh file into the docker image
COPY wait-for.sh .
# Copy all migration SQL files from db/migration folder to the image in the migration folder under the current working directory
COPY db/migration ./migration

EXPOSE 8080
CMD [ "/app/main" ]
# Change the way to start app: run db migration before running the main binary
# -> specify the /app/start.sh file as the main entry point of the docker image using command ENTRYPOINT
# when CMD instruction is used together with ENTRYPOINT, it acts as additional parameters passed into the entry point script
ENTRYPOINT [ "/app/start.sh" ]


