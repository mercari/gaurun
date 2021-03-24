FROM golang:1.15 as build-env
ADD . /opt/gaurun
RUN cd /opt/gaurun/cmd/gaurun &&\
    go build -tags=internal -a -ldflags="-s -w -linkmode external -extldflags -static" -v  &&\
    cd /opt/gaurun/cmd/gaurun_recover &&\
    go build -tags=internal -a -ldflags="-s -w -linkmode external -extldflags -static" -v

FROM alpine:3.12.4
RUN apk add --no-cache ca-certificates tzdata &&\
    addgroup -g 1000 -S gaurun &&\
    adduser -u 1000 -S gaurun -G gaurun
USER gaurun
WORKDIR /app
COPY --from=build-env /opt/gaurun/cmd/gaurun/gaurun ./
COPY --from=build-env /opt/gaurun/cmd/gaurun_recover/gaurun_recover ./
EXPOSE 1056