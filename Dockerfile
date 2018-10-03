# https://github.com/alextanhongpin/go-docker-multi-stage-build
FROM golang:latest as builder

WORKDIR /go/bin/

COPY ./src .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o claimcheck .


FROM alpine:3.8
RUN apk --no-cache add ca-certificates

WORKDIR /go/bin/
COPY --from=builder /go/bin/claimcheck .

EXPOSE 1323
ENTRYPOINT ["/go/bin/claimcheck"]