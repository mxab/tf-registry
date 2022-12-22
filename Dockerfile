#syntax=docker/dockerfile:1.4
FROM golang:1.19-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN ls -al
RUN go mod download
COPY . .
RUN go build cmd/tfr/tfr.go

FROM alpine:3.14
COPY --from=builder /src/tfr /bin/tfr
ENTRYPOINT ["/bin/tfr"]
CMD [ "server" ]