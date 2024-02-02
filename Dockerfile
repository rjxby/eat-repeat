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
    if [ "$DRONE" = "true" ]; then \
        DRONE_COMMIT_SHORT=$(echo $DRONE_COMMIT | cut -c 1-7) ; \
        version=${DRONE_TAG}${DRONE_BRANCH}-${DRONE_COMMIT_SHORT}-$(date +%Y%m%d-%H:%M:%S) ; \
    else \
        echo "runs outside of drone" && version="unknown" ; \
    fi && \
    echo "version=$version" && \
    go build --tags "embed fts5" -o service -ldflags "-X main.revision=${version} -s -w"

# Final stage
FROM alpine:3.18.4

WORKDIR /srv

COPY --from=build-backend /backend/service .

ARG RUN_MIGRATION
ENV RUN_MIGRATION=$RUN_MIGRATION

ARG PDF_READER_ENDPOINT
ENV PDF_READER_ENDPOINT=$PDF_READER_ENDPOINT

EXPOSE 8080

CMD ["/srv/service"]
