FROM golang:1.11 as builder

WORKDIR /go/src/github.com/mercari/gaurun

RUN apt update && apt install make
RUN go get -u -v github.com/Masterminds/glide

COPY . ./
RUN make bundle && make BUILD_FLAG="GOOS=linux GOARCH=amd64 CGO_ENABLED=0"

FROM alpine

RUN apk add --update --no-cache ca-certificates tzdata

COPY --from=builder /go/src/github.com/mercari/gaurun/bin /bin

ENTRYPOINT ["/bin/gaurun"]
