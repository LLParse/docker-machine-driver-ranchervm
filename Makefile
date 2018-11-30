.PHONY: clean build install deps lint vet test dist-clean dist release tag-release

default: build

VERSION := $(shell cat VERSION)

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
name := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

clean:
	rm -f bin/$(name)
	rm -f $(GOPATH)/bin/$(name)

build:
	CGOENABLED=0 go build -o bin/$(name) ./bin

install: build
	cp bin/$(name) $(GOPATH)/bin/

deps:
	go get -v github.com/tcnksm/ghr
	go get -v github.com/golang/lint/golint

lint:
	@golint $$(go list ./... 2> /dev/null | grep -v /vendor/)

vet:
	@go vet $$(go list ./... 2> /dev/null | grep -v /vendor/)

test: lint vet
	@go test $$(go list ./... 2> /dev/null | grep -v /vendor/)

tag-release:
	git tag -f $(VERSION)
	git push -f origin master --tags

dist-clean:
	rm -rf release

dist: dist-clean
	mkdir -p release
	GOOS=linux GOARCH=amd64 CGOENABLED=0 go build -o release/$(name)-Linux-x86_64 ./bin
	GOOS=darwin GOARCH=amd64 CGOENABLED=0 go build -o release/$(name)-Darwin-x86_64 ./bin
	GOOS=windows GOARCH=amd64 CGOENABLED=0 go build -o release/$(name)-Windows-x86_64.exe ./bin
	for file in release/$(name)-*; do openssl dgst -md5 < $${file} > $${file}.md5; openssl dgst -sha256 < $${file} > $${file}.sha256; done

release:
	ghr -u llparse -r docker-machine-driver-ranchervm --replace $(VERSION) release/
