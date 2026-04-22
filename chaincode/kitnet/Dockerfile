# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.20-alpine AS build

COPY . /go/src/github.com/chaincode/kitnet
WORKDIR /go/src/github.com/chaincode/kitnet

RUN go mod vendor
RUN go build -o chaincode github.com/chaincode/kitnet

FROM alpine:latest
COPY --from=build /go/src/github.com/chaincode/kitnet/chaincode /usr/local/bin/chaincode
CMD ["chaincode"]
