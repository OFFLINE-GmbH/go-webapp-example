FROM golang:1.12-alpine as builder

MAINTAINER Tobias Kündig <tobias@offline.ch>

## Install build tools

ARG DOCKER_CLI_VERSION="18.06.3-ce"
ENV DOWNLOAD_URL="https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_CLI_VERSION.tgz"

# Base
RUN mkdir /app
RUN apk update && apk upgrade && \
        apk add --no-cache git nodejs yarn gcc g++ zip curl rsync openssh-client npm

# Vue
RUN yarn global add @vue/cli

# mage
RUN go get -u -d github.com/magefile/mage && cd $GOPATH/src/github.com/magefile/mage && go run bootstrap.go && cd -

WORKDIR /app

# Build Backend + Frontend
ADD ./ /app

ENV GOPATH=/app/vendor/go

RUN mage -v build:all

# --------------------- Build the resulting container in a separate stage

FROM alpine:3.10.1

# Backend
COPY --from=builder /app/output/go-webapp-example /app/go-webapp-example
# Frontend
COPY --from=builder /app/web/dist /app/static
# Required external files
ADD ./deployments/migrations /app/migrations
ADD ./deployments/seeds /app/seeds
ADD ./internal/locales /app/locales

WORKDIR "/app"

EXPOSE 80

ENTRYPOINT ["/app/go-webapp-example"]
