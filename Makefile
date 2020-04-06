.PHONY: protoc-docker protoc

protoc:
	protoc --gogofaster_out=plugins=grpc:./proto -I=./proto proto/messages.proto

protoc-docker:
	docker run --rm -v `pwd`:/src znly/protoc:0.4.0 --gogofaster_out=plugins=grpc:. -I=. src/proto/messages.proto