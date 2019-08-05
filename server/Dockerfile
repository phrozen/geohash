# Multistep build, binary only deploy
FROM golang:1.12-alpine as builder
WORKDIR /go/src/app
COPY . .
RUN apk add --no-cache git
RUN go get
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -o server .
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .
# -----------------------------------------------------------------------------
FROM scratch
COPY --from=builder /go/src/app/server /server
ENV ZONEINFO /zoneinfo.zip
COPY --from=builder /zoneinfo.zip /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV PORT=3000
ENTRYPOINT [ "/server" ]
EXPOSE 3000
# -----------------------------------------------------------------------------
# docker build -t geohash-server -f Dockerfile .
# docker run -it -p 3000:3000 geohash-server