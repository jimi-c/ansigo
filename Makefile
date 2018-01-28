all:
	go build ansible.go
	go build -buildmode=plugin -o plugins/action/normal.so src/ansible/plugins/action/main/normal.go
	go build -buildmode=plugin -o plugins/connection/local.so src/ansible/plugins/connection/main/local.go
	go build -buildmode=plugin -o plugins/connection/ssh.so src/ansible/plugins/connection/main/ssh.go
