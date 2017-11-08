FROM golang

ADD . /go/src/pullr-krul
WORKDIR /go/src/pullr-krul
ARG PULLR_TAG

ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ARG DEV_PULLR_SNS_ARN

RUN go generate
RUN go build -v
ENTRYPOINT ["/go/src/pullr-krul/pullr-krul"]
