GOPATH=$(shell pwd)

all: buildroot main plugins

buildroot:
	mkdir -p build/plugins/{action,connection}

plugins: buildroot
	GOPATH=$(GOPATH) go build -buildmode=plugin -o build/plugins/action/normal.so src/ansible/plugins/action/main/normal.go
	GOPATH=$(GOPATH) go build -buildmode=plugin -o build/plugins/action/debug.so src/ansible/plugins/action/main/debug.go
	GOPATH=$(GOPATH) go build -buildmode=plugin -o build/plugins/connection/local.so src/ansible/plugins/connection/main/local.go
	GOPATH=$(GOPATH) go build -buildmode=plugin -o build/plugins/connection/ssh.so src/ansible/plugins/connection/main/ssh.go

main: buildroot
	GOPATH=$(GOPATH) go build -o build/ansible ansible.go

clean:
	rm -rf build
