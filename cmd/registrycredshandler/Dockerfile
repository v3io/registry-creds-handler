FROM golang:alpine as builder

ARG VERSION_LABEL=unstable

ARG GO_LINK_FLAGS="-s -w"

ARG GOOS=linux

ARG GOARCH=amd64

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -a -installsuffix cgo -ldflags="${GO_LINK_FLAGS}"  -o registrycredshandler cmd/registrycredshandler/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/registrycredshandler /usr/bin/

ENTRYPOINT ["registrycredshandler"]
