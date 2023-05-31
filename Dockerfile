FROM golang:latest AS build

WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0

RUN make build

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/app/bin/oci-instance-creator /oci-instance-creator

ENV KEY_PATH /keys/oci.key

ENTRYPOINT [ "/oci-instance-creator" ]