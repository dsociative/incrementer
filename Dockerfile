FROM golang:alpine as builder
WORKDIR $GOPATH/src/github.com/dsociative/incrementer
ADD . .
RUN go install -v ./

FROM alpine
COPY --from=builder /go/bin/incrementer /usr/local/bin
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/incrementer"]
