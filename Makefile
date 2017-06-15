dependency:
	go get github.com/hashicorp/packer
	go get github.com/laher/goxc

build:
	go build -o $(GOPATH)/bin/packer-provisioner-mitamae -a
