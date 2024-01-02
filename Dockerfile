# Frontend build stage
FROM node:20.8.1-alpine AS build-frontend

WORKDIR /frontend

COPY frontend/package*.json ./
COPY frontend/icon/* ./icon/

RUN npm install && npm run build

# Backend build stage
FROM golang:1.21.3-alpine AS build-backend

WORKDIR /backend

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add --no-cache gcc musl-dev && go env -w CGO_ENABLED=1

ARG DRONE
ARG DRONE_TAG
ARG DRONE_COMMIT
ARG DRONE_BRANCH

# if DRONE presented use DRONE_* git env to make version
RUN \
    if [ -z "$DRONE" ] ; then echo "runs outside of drone" && version="$(/script/git-rev.sh)" ; \
    else version=${DRONE_TAG}${DRONE_BRANCH}-${DRONE_COMMIT:0:7}-$(date +%Y%m%d-%H:%M:%S) ; fi && \
    echo "version=$version" && \
    go build -tags embed -o service -ldflags "-X main.revision=${version} -s -w"

# Final stage
FROM alpine:3.18.4

WORKDIR /srv

COPY --from=build-backend /backend/service .

ARG RUN_MIGRATION
ENV RUN_MIGRATION=$RUN_MIGRATION

EXPOSE 8080

CMD ["/srv/service"]
