default: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s %s\n\033[0m", $$1, $$2}'

# general
APP_NAME=hostdb-collector-vcenter
APP_VER=0.1
CIRCLE_WORKFLOW_ID?=""
WORK_DIR=$(shell pwd)

# git
ifeq ($(CIRCLECI), true)
GIT_BRANCH_DIRTY=$(CIRCLE_BRANCH)
else
GIT_BRANCH_DIRTY=$(shell git rev-parse --abbrev-ref HEAD)
endif
GIT_BRANCH=$(shell echo "$(GIT_BRANCH_DIRTY)" | sed s/[[:punct:]]/_/g | tr '[:upper:]' '[:lower:]')
GIT_COMMIT_NUM=$(shell git rev-list --count HEAD)
GIT_COMMIT_SHA=$(shell git rev-list -1 HEAD | cut -b 1-7)

# version
ifeq ($(GIT_BRANCH), master)
TAG="latest"
VERSION=$(APP_VER).$(GIT_COMMIT_NUM)
else
TAG=$(GIT_BRANCH)
VERSION=$(APP_VER).$(GIT_COMMIT_NUM).$(GIT_BRANCH)
endif
export TAG
export VERSION

# container
CONTAINER_REPO=registry.pdxfixit.com
CONTAINER_IMAGE_NAME=$(CONTAINER_REPO)/$(APP_NAME)

.PHONY: all
all: get test build container_build container_push

.PHONY: get
get: ## get dependencies
	go get -t -v

.PHONY: test
test: ## run the golang tests
	go test -v --failfast

.PHONY: build
build: ## build the linux/amd64 binary with c lib bindings for use in a scratch container
	go get github.com/mitchellh/gox
	env CGO_ENABLED=0 gox -osarch="linux/amd64" -tags netgo -output $(APP_NAME)

.PHONY: container_build
container_build: ## create container image
	docker build -t $(APP_NAME) \
		--label "name=$(APP_NAME)" \
		--label "vendor=PDXfixIT, LLC" \
		--label "version=$(VERSION)" \
		--label "maintainer=Ben Sandberg <info@pdxfixit.com>" \
		.

.PHONY: container_push
container_push: ## push hostdb container to registry
	docker tag $(APP_NAME) $(CONTAINER_IMAGE_NAME):$(VERSION)
	docker tag $(CONTAINER_IMAGE_NAME):$(VERSION) $(CONTAINER_IMAGE_NAME):$(TAG)

	@echo $(CONTAINER_PASS) | docker login -u $(CONTAINER_USER) --password-stdin $(CONTAINER_REPO)
	docker push $(CONTAINER_IMAGE_NAME):$(VERSION)
	docker push $(CONTAINER_IMAGE_NAME):$(TAG)

.PHONY: container_stop
container_stop: ## stop the hostdb container
	if [ "$$(docker ps -a -q -f 'name=$(APP_NAME)')" ]; then docker stop -t0 $(APP_NAME); fi

.PHONY: container_run
container_run: ## run the hostdb container
	docker run -it --rm --name $(APP_NAME) $(APP_NAME)

.PHONY: clean
clean: ## clean up any artifacts
	rm -f $(WORK_DIR)/$(APP_NAME)
	if [ "$$(docker ps -a -q -f 'name=$(APP_NAME)')" ]; then docker stop -t0 $(APP_NAME); fi
