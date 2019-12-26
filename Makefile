GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
GIT_DESCR = $(shell git describe --always) 
# build output folder
OUTPUTFOLDER = dist
# docker image
DOCKER_REGISTRY = noandrea
DOCKER_IMAGE = distill
DOCKER_TAG = $(shell git describe --always)
# build parameters
OS = linux
ARCH = amd64

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

default: build

workdir:
	mkdir -p dist

build: build-dist

build-dist: $(GOFILES)
	@echo build binary to $(OUTPUTFOLDER)
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -ldflags "-X main.Version=$(GIT_DESCR)" -o $(OUTPUTFOLDER)/distill .
	@echo copy resources
	cp -r README.md LICENSE configs $(OUTPUTFOLDER)
	@echo done

test: test-all

test-all:
	@echo running tests 
	go test $(GOPACKAGES) -race -coverprofile=coverage.txt -covermode=atomic
	@echo tests completed

bench: bench-all

bench-all:
	@go test -bench -v $(GOPACKAGES)

lint: lint-all

lint-all:
	@golint -set_exit_status $(GOPACKAGES)

clean:
	@echo remove $(OUTPUTFOLDER) folder
	@rm -rf dist
	@echo done

docker: docker-build

docker-build: build-dist
	@echo copy resources
	@cp configs/settings.docker.yaml $(OUTPUTFOLDER)/settings.yaml
	@echo build image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f ./build/docker/Dockerfile .
	@echo done

docker-push: docker-build
	@echo push image
	docker tag $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo done

docker-run: 
	@docker run -p 1804:1804 $(DOCKER_IMAGE) 

debug-start:
	@go run main.go -c examples/settings.yaml --debug start

gen-secret:
	@< /dev/urandom tr -dc _A-Z-a-z-0-9 | head -c40
