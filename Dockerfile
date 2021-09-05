FROM golang:1.17 as build-dev

WORKDIR /biscuit

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o ~/biscuit *.go

FROM alpine:3.14.2

COPY --from=build-dev /root/biscuit /usr/local/bin/.

RUN apk update && apk add --no-cache \
  ca-certificates \
  bash \
  openssl

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

ENTRYPOINT ["biscuit"]
