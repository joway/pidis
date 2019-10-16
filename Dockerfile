FROM golang:1.12
WORKDIR /go/src/github.com/joway/pikv/
COPY . .

RUN make

FROM golang:1.12-alpine

COPY --from=0 /go/src/github.com/joway/pikv/pikv /usr/bin/pikv

ENTRYPOINT ["pikv"]