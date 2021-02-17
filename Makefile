VERSION=0.13.1

all: bin/gaurun bin/gaurun_recover

build-cross: cmd/gaurun/gaurun.go cmd/gaurun_recover/gaurun_recover.go gaurun/*.go buford/**/*.go gcm/*.go
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/linux/amd64/gaurun-${VERSION}/gaurun cmd/gaurun/gaurun.go
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/linux/amd64/gaurun-${VERSION}/gaurun_recover  cmd/gaurun_recover/gaurun_recover.go
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/gaurun-${VERSION}/gaurun cmd/gaurun/gaurun.go
	GO111MODULE=on GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/gaurun-${VERSION}/gaurun_recover cmd/gaurun_recover/gaurun_recover.go

dist: build-cross
	cd bin/linux/amd64 && tar zcvf gaurun-linux-amd64-${VERSION}.tar.gz gaurun-${VERSION}
	cd bin/darwin/amd64 && tar zcvf gaurun-darwin-amd64-${VERSION}.tar.gz gaurun-${VERSION}

bin/gaurun: cmd/gaurun/gaurun.go gaurun/*.go buford/**/*.go gcm/*.go
	GO111MODULE=on go build -o bin/gaurun cmd/gaurun/gaurun.go

bin/gaurun_recover: cmd/gaurun_recover/gaurun_recover.go gaurun/*.go buford/**/*.go gcm/*.go
	GO111MODULE=on go build -o bin/gaurun_recover cmd/gaurun_recover/gaurun_recover.go

bin/gaurun_client: samples/client.go
	GO111MODULE=on go build -o bin/gaurun_client samples/client.go

fmt:
	go fmt ./...

check:
	GO111MODULE=on go test -v ./...

clean:
	rm -rf bin/*
