FROM golang:alpine as builder

RUN apk update && apk add git
COPY . $GOPATH/src/github.com/AlbinoDrought/creamy-gateway-picker
WORKDIR $GOPATH/src/github.com/AlbinoDrought/creamy-gateway-picker

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN go get -d -v && go build -a -installsuffix cgo -o /go/bin/creamy-gateway-picker

FROM scratch

COPY --from=builder /go/bin/creamy-gateway-picker /go/bin/creamy-gateway-picker

ENTRYPOINT ["/go/bin/creamy-gateway-picker"]
