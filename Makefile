
all: bin/gaurun bin/gaurun_recover

bin/gaurun: gaurun.go gaurun/*.go
	gom build -o bin/gaurun gaurun.go

bin/gaurun_recover: recover.go gaurun/*.go
	gom build -o bin/gaurun_recover recover.go

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
