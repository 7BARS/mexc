FROM golang:1.21.0 as builder
COPY go.mod go.sum /go/src/mexc/
WORKDIR /go/src/mexc
RUN go mod download
COPY . /go/src/mexc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/mexc mexc

FROM alpine
RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY --from=builder /go/src/mexc/build/mexc /usr/bin/mexc
EXPOSE 8080 8080
ENTRYPOINT ["/usr/bin/mexc"]