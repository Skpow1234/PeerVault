# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /peervault ./cmd/peervault

FROM alpine:3.20
WORKDIR /
COPY --from=build /peervault /usr/local/bin/peervault
EXPOSE 3000 5000 7000
ENTRYPOINT ["/usr/local/bin/peervault"]

