# Copyright 2022 Iguazio
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
FROM golang:alpine as builder

ARG GO_LINK_FLAGS="-s -w"

ARG GOOS=linux

ARG GOARCH=amd64

WORKDIR /app

COPY . .

RUN apk add -U --no-cache ca-certificates

RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
    -a \
    -installsuffix cgo \
    -ldflags="${GO_LINK_FLAGS}"  \
    -o registrycredshandler cmd/registrycredshandler/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/registrycredshandler /usr/bin/
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["registrycredshandler"]
