GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
GIT_DESCR = $(shell git describe)
# build output folder
OUTPUTFOLDER = dist
# docker image
DOCKERIMAGE = welance/distill
TAG = $(shell git describe)
# build paramters
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
	@go test -v $(GOPACKAGES) -coverprofile .testCoverage.txt

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
	docker build -t $(DOCKERIMAGE):$(TAG) -f ./build/docker/Dockerfile .
	@echo done

docker-push: docker-build
	@echo push image
	docker push $(DOCKERIMAGE):$(TAG)
	@echo done

docker-run: 
	@docker run -p 1804:1804 $(DOCKERIMAGE) 

debug-start:
	@go run main.go -c configs/settings.sample.yaml --debug start
