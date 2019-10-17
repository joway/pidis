FROM golang:1.12
WORKDIR /go/src/github.com/joway/pikv/
COPY . .
RUN make

FROM alpine
COPY --from=0 /go/src/github.com/joway/pikv/bin/* /usr/bin/
ENTRYPOINT ["pikv"]
