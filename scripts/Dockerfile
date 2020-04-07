FROM golang:1.13 AS builder
WORKDIR /server
COPY . .

ARG GIT_COMMIT
ARG PROJECT
ARG BUILD_TIME

RUN CGO_ENABLED=0 GOFLAGS=-mod=vendor go build \
-ldflags "-X ${PROJECT}/version.Commit=${GIT_COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
./cmd/server

RUN CGO_ENABLED=0 GOFLAGS=-mod=vendor go build \
-ldflags "-X ${PROJECT}/version.Commit=${GIT_COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
./cmd/indexer

FROM alpine

RUN apk add --no-cache bash

COPY --from=builder /server/server /server
COPY --from=builder /server/indexer /indexer
COPY --from=builder /server/wait-for-it.sh /wait-for-it.sh

CMD ["/server"]