# syntax=docker/dockerfile:1

# Build
FROM golang:1.18-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY *.go ./
ADD ./crawlers ./crawlers
COPY *.proto ./

RUN go build -o /crawler


# Deploy
FROM alpine:3.14

ARG USER=nonroot
ENV HOLE /home/$USER

RUN adduser -D $USER \
        && mkdir -p /etc/sudoers.d \
        && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
        && chmod 0440 /etc/sudoers.d/$USER

WORKDIR /

COPY --from=build /crawler /crawler

USER nonroot:nonroot

ENTRYPOINT [ "/crawler" ]

