default: build

version = '0.0.1'

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
name := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

clean:
	rm -f bin/$(name)
	rm -f $(GOPATH)/bin/$(name)

build:
	CGOENABLED=0 go build -o bin/$(name) ./bin

install: build
	cp bin/$(name) $(GOPATH)/bin/

release:
	GOOS=linux GOARCH=amd64 CGOENABLED=0 go build -o bin/$(name) ./bin
	tar -cvzf bin/$(name)-$(version)-linux-amd64.tar.gz -C bin $(name)
	rm bin/$(name)
	GOOS=darwin GOARCH=amd64 CGOENABLED=0 go build -o bin/$(name) ./bin
	tar -cvzf bin/$(name)-$(version)-darwin-amd64.tar.gz -C bin $(name)
	rm bin/$(name)
	GOOS=windows GOARCH=amd64 CGOENABLED=0 go build -o bin/$(name).exe ./bin
	tar -cvzf bin/$(name)-$(version)-windows-amd64.tar.gz -C bin $(name).exe
	rm bin/$(name).exe

.PHONY: clean build install release
