.PHONY: protoc-docker protoc build-docker run-docker clean-docker

protoc:
	protoc --gogofaster_out=plugins=grpc:./proto -I=./proto proto/messages.proto

protoc-docker:
	docker run --rm -v `pwd`:/src znly/protoc:0.4.0 --gogofaster_out=plugins=grpc:. -I=. src/proto/messages.proto

# Build the docker image
build-docker:
	./scripts/build.docker.sh

# Build the docker image, then run the docker compose
run-docker: build-docker
	docker-compose down
	docker-compose up --build

# Cleanup the docker compose
clean-docker:
	docker-compose down --volumes