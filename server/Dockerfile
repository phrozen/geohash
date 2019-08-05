# Multistep build, binary only deploy
FROM golang:1.12-alpine as builder
WORKDIR /go/src/app
COPY . .
RUN apk add --no-cache git
# Module support
ENV GO111MODULE=on
RUN go get
# Static bianry build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -o server .
# Install timezone data and certificate authorities files
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .
# -----------------------------------------------------------------------------
FROM scratch
# Copy binary
COPY --from=builder /go/src/app/server /server
# Copy zoneinfo and ca-certificates
ENV ZONEINFO /zoneinfo.zip
COPY --from=builder /zoneinfo.zip /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# export PORT
ENV PORT=3000
ENTRYPOINT [ "/server" ]
EXPOSE 3000
# -----------------------------------------------------------------------------
# docker build -t geohash-server -f Dockerfile .
# docker run -it -p 3000:3000 -v $(pwd)/geohash.db:/geohash.db geohash-server