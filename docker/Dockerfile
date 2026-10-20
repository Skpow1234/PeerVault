# syntax=docker/dockerfile:1

FROM golang:1.24.6-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /shugo ./cmd/shugo

FROM alpine:3.20
WORKDIR /
COPY --from=build /shugo /usr/local/bin/shugo
EXPOSE 3000 5000 7000
ENTRYPOINT ["/usr/local/bin/shugo"]

