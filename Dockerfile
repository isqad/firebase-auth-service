FROM golang:1.17.1-buster AS compile

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download
COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/service cmd/service/main.go

FROM alpine:latest

WORKDIR /api

RUN apk add tini

ENTRYPOINT ["tini", "--"]

COPY --from=compile /build/service /api/

EXPOSE 50053

CMD /api/service --listen-address :50053
