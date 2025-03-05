
EXECUTABLE := aidns

ROOT := $(shell pwd)
DIST := $(ROOT)/release
coredns_origin := $(DIST)/coredns_origin
coredns_build := $(DIST)/coredns_build
binaries_path := $(DIST)/binaries
release_path := $(DIST)/release
release_upx_path := $(DIST)/release-upx

TARGETS ?= windows linux
ARCHS ?= amd64 arm64 386 arm
UPXLEVEL := 6

ifneq (!$(VERSION),)
	VERSION ?= v0.0.0
endif

ifneq ($(wildcard $(coredns_origin)),)
    $(info 已下载 CoreDNS 源代码)
else
    $(shell git clone --depth 1 git@github.com:coredns/coredns.git $(coredns_origin))
endif

build: build-aidns release-upx release-check-upx

build-aidns: build-before
	cd $(coredns_build) && go mod tidy && go generate && \
 		gox -os="$(TARGETS)" -arch="$(ARCHS)" -output="$(binaries_path)/aidns-$(VERSION)-{{.OS}}-{{.Arch}}"

build-before:
	rm -rf $(coredns_build) && rm -rf $(binaries_path) && mkdir -p $(binaries_path)
	cp -r $(coredns_origin) $(coredns_build)
	mkdir -p $(coredns_build)/plugin/aidns && cp *.go $(coredns_build)/plugin/aidns
	sed -i 's/k8s_external:k8s_external/aidns:aidns/g' $(coredns_build)/plugin.cfg
	sed -i '/kubernetes/d' $(coredns_build)/plugin.cfg

clear-all:
	rm -rf release && rm -rf aidns

release-upx:
	rm -rf $(release_upx_path) && mkdir -p $(release_upx_path);
	cd $(binaries_path); $(foreach file,$(wildcard $(binaries_path)/$(EXECUTABLE)-*),upx -$(UPXLEVEL) $(notdir $(file)) -o $(release_upx_path)/$(notdir $(file));)

release-check:
	cd $(release_path); $(foreach file,$(wildcard $(release_path)/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

release-check-upx:
	cd $(release_upx_path); $(foreach file,$(wildcard $(release_upx_path)/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

dev: build-before
	cd $(coredns_build) && go mod tidy && go generate && go build -o $(ROOT)/aidns