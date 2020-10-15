FROM golang:1.13 as build-env
ADD . /opt/gaurun
RUN cd /opt/gaurun/cmd/gaurun &&\
    go build -tags=internal -a -ldflags="-s -w -linkmode external -extldflags -static" -v  &&\
    cd /opt/gaurun/cmd/gaurun_recover &&\
    go build -tags=internal -a -ldflags="-s -w -linkmode external -extldflags -static" -v

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata &&\
    addgroup -g 1000 -S gaurun &&\
    adduser -u 1000 -S gaurun -G gaurun
COPY --from=build-env /opt/gaurun/cmd/gaurun/gaurun /app/
COPY --from=build-env /opt/gaurun/cmd/gaurun_recover/gaurun_recover /app/
COPY ./conf/gaurun.toml /app/conf/gaurun.toml
USER gaurun
WORKDIR /app
EXPOSE 1056
CMD ["./gaurun", "-c", "conf/gaurun.toml"]
