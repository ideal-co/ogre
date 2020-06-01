# # #                                             # # #
# To get started, run make help from the project root #
# # #                                             # # #
.PHONY: build build-alpine clean test help default

BIN_NAME=ogre

VERSION := $(shell grep "const Version " pkg/version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "ideal-co/ogre"
BIN_PATH :=

help:
	@echo 'Management commands for ogre:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make build-alpine    Compile optimized for alpine linux.'
	@echo '    make package         Build final docker image with just the go binary inside'
	@echo '    make tag             Tag image created by package with latest, git commit and version'
	@echo '    make test            Run tests on a compiled project.'
	@echo '    make push            Push tagged images to registry'
	@echo '    make clean           Clean the directory tree.'
	@echo

build:
	@go build ./cmd/ogred/
	@go build ./cmd/ogre/
	@cp ogre /usr/local/bin/
	@cp ogred /usr/local/bin/
	@rm ogre ogred

build-alpine:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags '-w -linkmode external -extldflags "-static" -X github.com/ideal-co/ogre/pkg/version/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/ideal-co/ogre/pkg/version/version.BuildDate=${BUILD_DATE}' -o bin/${BIN_NAME} ./cmd/ogre/
	go build -o ./bin/ ./cmd/ogred/

package:
	@echo "building image ${BIN_NAME} ${VERSION} $(GIT_COMMIT)"
	docker build --build-arg VERSION=${VERSION} --build-arg GIT_COMMIT=$(GIT_COMMIT) -t $(IMAGE_NAME):local .

tag: 
	@echo "Tagging: latest ${VERSION} $(GIT_COMMIT)"
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):$(GIT_COMMIT)
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):${VERSION}
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

push: tag
	@echo "Pushing docker image to registry: latest ${VERSION} $(GIT_COMMIT)"
	docker push $(IMAGE_NAME):$(GIT_COMMIT)
	docker push $(IMAGE_NAME):${VERSION}
	docker push $(IMAGE_NAME):latest

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}

test:
	go test -v ./...
