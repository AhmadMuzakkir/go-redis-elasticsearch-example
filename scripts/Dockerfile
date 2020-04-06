FROM golang:1.13 AS builder
WORKDIR /server
COPY . .

RUN CGO_ENABLED=0 GOFLAGS=-mod=vendor go build ./cmd/server
RUN CGO_ENABLED=0 GOFLAGS=-mod=vendor go build ./cmd/indexer

FROM alpine

RUN apk add --no-cache bash

COPY --from=builder /server/server /server
COPY --from=builder /server/indexer /indexer
COPY --from=builder /server/wait-for-it.sh /wait-for-it.sh

CMD ["/server"]