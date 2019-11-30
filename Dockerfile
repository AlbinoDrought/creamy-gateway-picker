FROM golang:alpine as builder

RUN apk update && apk add git
COPY . $GOPATH/src/github.com/AlbinoDrought/creamy-gateway-override
WORKDIR $GOPATH/src/github.com/AlbinoDrought/creamy-gateway-override

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN go get -d -v && go build -a -installsuffix cgo -o /go/bin/creamy-gateway-override

FROM scratch

COPY --from=builder /go/bin/creamy-gateway-override /go/bin/creamy-gateway-override

ENTRYPOINT ["/go/bin/creamy-gateway-override"]
