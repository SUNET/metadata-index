FROM golang:1.12-alpine
MAINTAINER Leif Johansson <leifj@sunet.se>
RUN apk add --update --no-cache git
WORKDIR /go/src/metadata-index
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
RUN mkdir -p /etc/datasets
VOLUME /etc/datasets
EXPOSE 3000
CMD ["metadata-index","-index","/etc/datasets/.bleve"]
