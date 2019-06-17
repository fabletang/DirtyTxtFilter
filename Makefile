# The binary to build (just the basename).
VERSION := 1.0.1
BIN := DirtyTxtFilter
PORT:= 8080
BUILD_PARA:= '-tags=jsoniter'
# Where to push the docker image.
# docker 仓库
REGISTRY ?= 192.168.76.172:5000
# BASEIMAGE ?= gcr.io/distroless/static
BASEIMAGE ?= fabletang/golang-run:alpine-3.9
# BUILD_IMAGE ?= golang:1.12-alpine
BUILD_IMAGE ?= fabletang/golang-compile:1.12.6

### ==================================================================
### These variables should not need tweaking.
###
# This version-strategy uses git tags to set the version string
#VERSION := $(shell git describe --tags --always --dirty)
#$(subst ' ','',$(shell git rev-list HEAD --count))
VERSION :=$(VERSION)-$(subst ' ','',$(shell git rev-list HEAD --count))
ENV ?=dev
# 转换小写
bin := $(shell echo $(BIN) | tr A-Z a-z )
# directories which hold app source (not vendored)
SRC_DIRS := src   #src pkg
#MAIN_DIR := src/github.com/fabletang/DirtyTxtFilter/
MAIN_DIR := src/$(shell go list -m)
#SRC_MAIN := /src/github.com/fabletang/DirtyTxtFilter/main.go
SRC_MAIN :=  $(MAIN_DIR)/main.go
RES_DIR := $(MAIN_DIR)/res/

# ALL_PLATFORMS := linux/amd64 linux/arm linux/arm64 linux/ppc64le linux/s390x
# ALL_PLATFORMS := linux/amd64 darwin/amd64
ALL_PLATFORMS := linux/amd64

# Used internally.  Users should pass GOOS and/or GOARCH.
OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))


# IMAGE := $(REGISTRY)/$(BIN)
IMAGE := $(REGISTRY)/$(ENV)/$(bin)
# IMAGE := $(REGISTRY_H):$(REGISTRY_P)/$(BIN)
TAG := $(VERSION)_$(OS)_$(ARCH)


var:
	@echo ENV:$(ENV) BIN:$(BIN) VERSION:$(VERSION) OS:$(OS) ARCH:$(ARCH)
	@echo MAIN_DIR:$(MAIN_DIR) SRC_MAIN:$(SRC_MAIN)
	@echo docker image:$(IMAGE):$(TAG)
# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
#
help: 
	@echo ============================使用指南===================================
	@echo "make run"  :自动检测当前系统环境,在bin目录生成二进制以及copy res/资源文件
	@echo "make docker"  :本地打包linux/amd64 docker镜像
	@echo "make pushdocker"  :本地打包linux/amd64 docker镜像,并且推送到docker仓库
	@echo ============================使用指南===================================

all: build
docker: container-linux_amd64
pushdocker: clean push-linux_amd64	
#run: build-$(OS)_$(ARCH) 
#	cd bin/$(OS)_$(ARCH) && ./$(BIN)
run: clean 
	mkdir -p bin/$(OS)_$(ARCH)/res/ &&                   \
	cp -r $(RES_DIR) bin/$(OS)_$(ARCH)/res/ &&           \
	go build -o bin/$(OS)_$(ARCH)/$(BIN) $(SRC_MAIN) &&  \
	cd bin/$(OS)_$(ARCH) && ./$(BIN)
	
# /bin/sh $(OUTBIN) 

# For the following OS/ARCH expansions, we transform OS/ARCH into OS_ARCH
# because make pattern rules don't match with embedded '/' characters.

build-%:
	@$(MAKE) build                        \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

container-%:
	@$(MAKE) container                    \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

push-%:
	@$(MAKE) push                         \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

all-build: $(addprefix build-, $(subst /,_, $(ALL_PLATFORMS)))

all-container: $(addprefix container-, $(subst /,_, $(ALL_PLATFORMS)))

all-push: $(addprefix push-, $(subst /,_, $(ALL_PLATFORMS)))

build: bin/$(OS)_$(ARCH)/$(BIN)
# build: bin/$(OS)_$(ARCH)/$(bin)

# Directories that we need created to build/test.
BUILD_DIRS := bin/$(OS)_$(ARCH)     \
              .go/bin/$(OS)_$(ARCH) \
              .go/cache

# The following structure defeats Go's (intentional) behavior to always touch
# result files, even if they have not changed.  This will still run `go` but
# will not trigger further work if nothing has actually changed.
OUTBIN = bin/$(OS)_$(ARCH)/$(BIN)
# OUTBIN = bin/$(OS)_$(ARCH)/$(bin)

# This will build the binary under ./.go and update the real binary iff needed.
# docker run --rm -it -v "$PWD":/goapp/src/myapp -w /goapp/src/myapp -e GOOS=linux -e GOARCH=amd64 fabletang/golang-compile:1.11.5 go build
#	     -v $$(pwd):/src                                         \
#	-u $$(id -u):$$(id -g)                                  \
#   -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
#    -v $$(pwd)/.go/cache:/.cache                            \
	
	   
	     
$(OUTBIN): .go/$(OUTBIN).stamp
	@true
	    
.PHONY: .go/$(OUTBIN).stamp
.go/$(OUTBIN).stamp: $(BUILD_DIRS)
	@echo "making $(OUTBIN)"
	@docker run                                         \
	    -it                                             \
	    --rm                                            \
	    -v $$(pwd):/goapp/src/myapp                     \
	    -w /goapp/src/myapp                             \
	    --env HTTP_PROXY=$(HTTP_PROXY)                  \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                \
	    -e GOOS=$(OS)     			            \
	    -e GOARCH=$(ARCH)                               \
	    -e GOFLAGS="-mod=vendor" 	            	    \
	    -e GO111MODULE=on                               \
	    -e CGO_ENABLED=0                                \
	    $(BUILD_IMAGE)                                  \
	    go build                                        \
	    $(BUILD_PARA)                                   \
	    -o /goapp/src/myapp/.go/$(OUTBIN)               \
	    $(SRC_MAIN) 

	@if ! cmp -s .go/$(OUTBIN) $(OUTBIN); then        \
	    cp -r $(RES_DIR) bin/$(OS)_$(ARCH)/res/;           \
	    mv .go/$(OUTBIN) $(OUTBIN);                   \
	    date >$@;                                     \
	fi

# Example: make shell CMD="-c 'date > datefile'"
shell: $(BUILD_DIRS)
	@echo "launching a shell in the containerized build environment"
	@docker run                                                 \
	    -ti                                                     \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/sh $(CMD)
#	     -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \


# Used to track state in hidden files.
# DOTFILE_IMAGE = $(subst /,_,$(IMAGE))-$(TAG)
# DOTFILE_IMAGE = $(subst /,_,$(bin))-$(TAG)
DOTFILE_IMAGE = $(subst /,_,$(BIN))-$(TAG)
# DOTFILE_IMAGE = $(TAG)

container: .container-$(DOTFILE_IMAGE) say_container_name

.container-$(DOTFILE_IMAGE): bin/$(OS)_$(ARCH)/$(BIN) Dockerfile.in
	@echo "===== build docker ===="
	@sed                                 \
 	    -e 's|{ARG_BIN}|$(BIN)|g'        \
	    -e 's|{ARG_ARCH}|$(ARCH)|g'      \
	    -e 's|{ARG_OS}|$(OS)|g'          \
	    -e 's|{PORT}|$(PORT)|g'          \
 	    -e 's|{ARG_FROM}|$(BASEIMAGE)|g' \
	    Dockerfile.in > .dockerfile-$(OS)_$(ARCH)

        ifeq ($(OS),linux)  
	@docker build -t $(IMAGE):$(TAG) -f .dockerfile-$(OS)_$(ARCH) . 
	@docker images -q $(IMAGE):$(TAG) > $@                          
        else 
		@echo "os not linux cannot build docker!"
        endif

say_container_name:
	@echo "container: $(IMAGE):$(TAG)"

push: .push-$(DOTFILE_IMAGE) say_push_name
.push-$(DOTFILE_IMAGE): .container-$(DOTFILE_IMAGE)
        ifeq ($(OS),linux)  
	@docker push $(IMAGE):$(TAG)
        endif

say_push_name:
	ifeq ($(OS),linux)  
	@echo "pushed: $(IMAGE):$(TAG)"
	endif

manifest-list: push
	platforms=$$(echo $(ALL_PLATFORMS) | sed 's/ /,/g');  \
	manifest-tool                                         \
	    --username=oauth2accesstoken                      \
	    --password=$$(gcloud auth print-access-token)     \
	    push from-args                                    \
	    --platforms "$$platforms"                         \
	    --template $(REGISTRY)/$(BIN):$(VERSION)__OS_ARCH \
	    --target $(REGISTRY)/$(BIN):$(VERSION)

version:
	@echo $(VERSION)

test: $(BUILD_DIRS)
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/goapp/src/myapp                     \
	    -w /goapp/src/myapp                             \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/sh -c "                                            \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        ./build/test.sh $(SRC_DIRS)                         \
	    "

$(BUILD_DIRS):
	@mkdir -p $@

clean: container-clean bin-clean

container-clean:
	rm -rf .container-* .dockerfile-* .push-*

bin-clean:
	rm -rf .go bin
