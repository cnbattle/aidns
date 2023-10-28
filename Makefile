
coredns_origin := tmp/coredns
coredns_build := tmp/coredns_build

ifneq (,$(wildcard $(coredns_origin)))
    $(info $(coredns_origin) exists1)
else
    $(shell git clone --depth 1 https://github.com/coredns/coredns.git tmp/coredns)
endif

build: clear build-aidns

build-aidns:
	cp -r $(coredns_origin) $(coredns_build)
	mkdir $(coredns_build)/plugin/aidns && cp *.go $(coredns_build)/plugin/aidns
	echo "aidns:aidns" >> $(coredns_build)/plugin.cfg
	cd $(coredns_build) && go mod tidy && go generate && go build -o aidns
	cp $(coredns_build)/aidns aidns && chmod +x aidns

clear:
	rm -rf $(coredns_build) && rm -rf aidns

clear-all:
	rm -rf tmp && rm -rf aidns