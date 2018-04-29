GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
OUTPUTFOLDER = 'dist'
DOCKERIMAGE = 'welance/ilij'

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

default: build

workdir:
	mkdir -p dist

build: build-dist

build-dist: $(GOFILES)
	@echo build binary to $(OUTPUTFOLDER)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(OUTPUTFOLDER)/ilij .
	@echo copy resources
	cp README.md LICENSE configs/ilij.conf.sample.yaml $(OUTPUTFOLDER)
	@echo done

test: test-all

test-all:
	@go test -v $(GOPACKAGES)

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
	@cp configs/ilij.conf.docker.yaml $(OUTPUTFOLDER)/ilij.conf.yaml
	@echo build image
	docker build -t $(DOCKERIMAGE) -f ./build/docker/Dockerfile .
	@echo done

docker-push: docker-build
	@echo push image
	docker push $(DOCKERIMAGE)
	@echo done

docker-run: 
	@docker run -p 1804:1804 $(DOCKERIMAGE) 

debug-start:
	@go run main.go -c configs/ilij.conf.sample.yaml --debug start
