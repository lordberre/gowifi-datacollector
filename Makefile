BUILD := build
GOPATH := $(PWD)/$(BUILD)
GOCACHE := $(GOPATH)/.gocache
# GOBIN := $(GOPATH)/bin

TARGET_ARCH ?= arm
CMDS := $(shell find cmd/ -maxdepth 1 -mindepth 1 -type d | sed -E 's|cmd/|$(BUILD)/bin/|g')
PROJECT:= $(shell basename $(PWD))
# PROJECT_VERSION := $(shell git describe --always --dirty=-D --abbrev=7 | sed -E 's/^([^\-]*)$$/\1-0/;s/^([^\-]*-[^\-]*).*$$/\1/g')
PROJECT_VERSION := 0.0.1
GO_VERSION := v$(shell echo $(PROJECT_VERSION) | cut -d- -f1)
LDFLAGS=-ldflags "-X=main.version=$(PROJECT_VERSION) -X=main.Build=$(BUILD) -w"
all: $(CMDS)

$(BUILD)/bin/%: cmd/% go.mod
	cd $< && CGO_ENABLED=0 GOOS=linux GOARCH=$(TARGET_ARCH) go install -a -tags netgo $(LDFLAGS); \

go.mod:
	if [ -f $@ ]; then touch $@; else go mod init github.com/lordberre/gowifi-datacollector; fi

.EXPORT_ALL_VARIABLES:

.PRECIOUS: go.mod

strip: all
	strip $(BUILD)/bin/*

.PHONY strip:

fmt:
	gofmt -l -e -s -w cmd/

clean:
	chmod -R u+w build > /dev/null 2>&1 || true
	rm -rf build

