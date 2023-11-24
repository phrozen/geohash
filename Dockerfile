# Multistep build, binary only deploy 🐳
# We establish a separate stage for building the app.
FROM golang:1.21-alpine as build

WORKDIR /app 

# We create a caching layer by copying our mod file and downloading dependencies 🧭
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Statically compile our binary for final image
RUN CGO_ENABLED=0 go build \
  -ldflags '-s -w -extldflags "-static"' \
  -tags musl,netgo,osusergo  \
  -o geohash-server ./server

# -----------------------------------------------------------------------------
# se the Google Distroless image for a minimal container. Can use scratch
FROM gcr.io/distroless/static

WORKDIR /
# Copy binary from our previously build image
COPY --from=build /app/geohash-server geohash-server
# use a non-root user for the execution
USER nonroot:nonroot
# export PORT
ENV PORT=3000
EXPOSE 3000
ENTRYPOINT [ "/geohash-server" ]

# -----------------------------------------------------------------------------
# docker build -t geohash-server -f Dockerfile .
# docker run -it -p 3000:3000 -v $(pwd)/geohash.db:/geohash.db geohash-server