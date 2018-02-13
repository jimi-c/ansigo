all: buildroot main plugins

buildroot:
	mkdir -p build/plugins/{action,connection,strategy}

plugins: buildroot
	go build -buildmode=plugin -o build/plugins/action/normal.so ansible/plugins/action/main/normal.go
	go build -buildmode=plugin -o build/plugins/action/debug.so ansible/plugins/action/main/debug.go
	go build -buildmode=plugin -o build/plugins/connection/local.so ansible/plugins/connection/main/local.go
	go build -buildmode=plugin -o build/plugins/connection/ssh.so ansible/plugins/connection/main/ssh.go
	go build -buildmode=plugin -o build/plugins/strategy/linear.so ansible/plugins/strategy/main/linear.go

main: buildroot
	go build -o build/ansible ansible.go

clean:
	rm -rf build
