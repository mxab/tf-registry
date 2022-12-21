#syntax=docker/dockerfile:1.4
FROM golang:1.19-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN ls -al
RUN go mod download
COPY . .
RUN go build -o /bin/tf-registry

FROM alpine:3.14
COPY --from=builder /bin/tf-registry /bin/tf-registry
ENTRYPOINT ["/bin/tf-registry"]
CMD [ "server" ]