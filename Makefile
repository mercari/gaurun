VERSION=0.4.2

all: bin/gaurun bin/gaurun_recover

build-cross: cmd/gaurun/gaurun.go cmd/gaurun_recover/gaurun_recover.go gaurun/*.go
	GOOS=linux GOARCH=amd64 gom build -o bin/linux/amd64/gaurun-${VERSION}/gaurun cmd/gaurun/gaurun.go
	GOOS=linux GOARCH=amd64 gom build -o bin/linux/amd64/gaurun-${VERSION}/gaurun_recover  cmd/gaurun_recover/gaurun_recover.go
	GOOS=darwin GOARCH=amd64 gom build -o bin/darwin/amd64/gaurun-${VERSION}/gaurun cmd/gaurun/gaurun.go
	GOOS=darwin GOARCH=amd64 gom build -o bin/darwin/amd64/gaurun-${VERSION}/gaurun_recover cmd/gaurun_recover/gaurun_recover.go

dist: build-cross
	cd bin/linux/amd64 && tar zcvf gaurun-linux-amd64-${VERSION}.tar.gz gaurun-${VERSION}
	cd bin/darwin/amd64 && tar zcvf gaurun-darwin-amd64-${VERSION}.tar.gz gaurun-${VERSION}

bin/gaurun: cmd/gaurun/gaurun.go gaurun/*.go
	gom build -o bin/gaurun cmd/gaurun/gaurun.go

bin/gaurun_recover: cmd/gaurun_recover/gaurun_recover.go gaurun/*.go
	gom build -o bin/gaurun_recover cmd/gaurun_recover/gaurun_recover.go

bin/gaurun_client: samples/client.go
	gom build -o bin/gaurun_client samples/client.go

gom:
	go get -u github.com/mattn/gom

bundle:
	gom install

fmt:
	go fmt ./...

check:
	cd gaurun; gom test

clean:
	rm -rf bin/*
