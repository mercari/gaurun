VERSION=0.9.1
TARGETS_NOVENDOR=$(shell glide novendor)
BUILD_FLAG=

all: bin/gaurun bin/gaurun_recover

build-cross: cmd/gaurun/gaurun.go cmd/gaurun_recover/gaurun_recover.go gaurun/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/linux/amd64/gaurun-${VERSION}/gaurun cmd/gaurun/gaurun.go
	GOOS=linux GOARCH=amd64 go build -o bin/linux/amd64/gaurun-${VERSION}/gaurun_recover  cmd/gaurun_recover/gaurun_recover.go
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/gaurun-${VERSION}/gaurun cmd/gaurun/gaurun.go
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin/amd64/gaurun-${VERSION}/gaurun_recover cmd/gaurun_recover/gaurun_recover.go

dist: build-cross
	cd bin/linux/amd64 && tar zcvf gaurun-linux-amd64-${VERSION}.tar.gz gaurun-${VERSION}
	cd bin/darwin/amd64 && tar zcvf gaurun-darwin-amd64-${VERSION}.tar.gz gaurun-${VERSION}

bin/gaurun: cmd/gaurun/gaurun.go gaurun/*.go
	${BUILD_FLAG} go build -o bin/gaurun cmd/gaurun/gaurun.go

bin/gaurun_recover: cmd/gaurun_recover/gaurun_recover.go gaurun/*.go
	${BUILD_FLAG} go build -o bin/gaurun_recover cmd/gaurun_recover/gaurun_recover.go

bin/gaurun_client: samples/client.go
	${BUILD_FLAG} go build -o bin/gaurun_client samples/client.go

bundle:
	glide install

docker:
	docker build -t mercari/gaurun .

fmt:
	@echo $(TARGETS_NOVENDOR) | xargs go fmt

check:
	go test -v $(TARGETS_NOVENDOR)

clean:
	rm -rf bin/*
