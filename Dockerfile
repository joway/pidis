FROM golang:1.13
WORKDIR /go/src/github.com/joway/pidis/
COPY . .
RUN make

FROM alpine
COPY --from=0 /go/src/github.com/joway/pidis/bin/* /usr/bin/
ENTRYPOINT ["pidis"]
