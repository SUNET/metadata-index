FROM golang:latest as build
RUN mkdir -p /go/src/github.com/sunet/metadata-index
ADD . /go/src/github.com/sunet/metadata-index/
WORKDIR /go/src/github.com/sunet/metadata-index
RUN make
RUN env GOBIN=/usr/bin go install ./cmd/mix
RUN mkdir -p /etc/datasets

# Now copy it into our base image.
FROM gcr.io/distroless/base:debug
COPY --from=build /usr/bin/mix /usr/bin/mix
VOLUME /etc/datasets
EXPOSE 3000
CMD ["/usr/bin/mix","-index","/etc/datasets/.bleve"]
