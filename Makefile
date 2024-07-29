# Copyright (c) Abstract Machines

PROGRAM = callhome
MG_DOCKER_IMAGE_NAME_PREFIX ?= magistrala
SOURCES = $(wildcard *.go) cmd/main.go
CGO_ENABLED ?= 0
GOARCH ?= amd64
VERSION ?= $(shell git describe --abbrev=0 --tags 2>/dev/null || echo "0.13.0")
COMMIT ?= $(shell git rev-parse HEAD)
TIME ?= $(shell date +%F_%T)
DOMAIN ?= deployments.magistrala.abstractmachines.fr

all: $(PROGRAM)

.PHONY: all clean $(PROGRAM)

define make_docker
	docker build \
		--no-cache \
		--build-arg SVC=$(PROGRAM) \
		--build-arg GOARCH=$(GOARCH) \
		--build-arg GOARM=$(GOARM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg TIME=$(TIME) \
		--tag=$(MG_DOCKER_IMAGE_NAME_PREFIX)/$(PROGRAM) \
		-f docker/Dockerfile .
endef

define make_dev_cert
	sudo openssl req -x509 -out ./docker/certbot/conf/live/$(DOMAIN)/fullchain.pem \
	-keyout ./docker/certbot/conf/live/$(DOMAIN)/privkey.pem \
	-newkey rsa:2048 -nodes -sha256 \
	-subj '/CN=localhost'
endef

$(PROGRAM): $(SOURCES)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
	go build -ldflags "-s -w \
	-X 'github.com/andychao217/magistrala.BuildTime=$(TIME)' \
	-X 'github.com/andychao217/magistrala.Version=$(VERSION)' \
	-X 'github.com/andychao217/magistrala.Commit=$(COMMIT)'" \
	-o ./build/$(PROGRAM)-$(PROGRAM) cmd/main.go

clean:
	rm -rf $(PROGRAM)

docker-image:
	$(call make_docker)
dev-cert:
	$(call make_dev_cert)

run:
	docker compose -f ./docker/docker-compose.yml up

test:
	go test -v --race -covermode=atomic -coverprofile cover.out ./...
