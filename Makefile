# Variables
GOCMD = go
GOBUILD_ENVS = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOTEST_SUDO_PREFIX = sudo --preserve-env=HOME --preserve-env=GOPATH
GOGET = $(GOCMD) get
BINARY_NAME = kubecop
GOFILES = $(shell find . -type f -name '*.go')

# Take image name from env variable or use default
IMAGE_NAME ?= kubecop:latest


$(BINARY_NAME): $(GOFILES) go.mod go.sum Makefile
	CGO_ENABLED=0 go build -o $(BINARY_NAME) cmd/main.go

test:
	$(GOTEST) -v ./... -coverprofile=coverage.out

install: $(BINARY_NAME)
	./scripts/install-in-pod.sh $(BINARY_NAME)

open-shell:
	./scripts/open-shell-in-pod.sh

close-shell:
	cat cop_pids.txt | xargs kill -15

deploy-dev-pod:
	kubectl apply -f etc/app-profile.crd.yaml
	kubectl apply -f dev/devpod.yaml

build: $(BINARY_NAME)

build-image: $(GOFILES) go.mod go.sum Makefile
	docker build -t $(IMAGE_NAME) -f ./Containerfile .

build-image-and-push: build-image
	docker push $(IMAGE_NAME)

clean:
	rm -f $(BINARY_NAME)

validate-crd:
	./scripts/validate-crd.sh

all: $(BINARY_NAME)

.PHONY: clean all install deploy-dev-pod test open-shell build validate-crd 