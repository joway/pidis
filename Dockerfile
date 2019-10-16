FROM golang:1.12
WORKDIR /go/src/github.com/joway/pikv/
COPY . .

RUN make

FROM 804775010343.dkr.ecr.cn-north-1.amazonaws.com.cn/golang:1.12-alpine

COPY --from=0 /go/src/github.com/joway/pikv/pikv /usr/bin/pikv

ENTRYPOINT ["pikv"]
