FROM golang

ADD . /go/src/pullr-krul
WORKDIR /go/src/pullr-krul
RUN go build -v
ENTRYPOINT ["/go/src/pullr-krul/pullr-krul"]
