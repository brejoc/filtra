FROM golang:alpine

WORKDIR /go/src/filtra
COPY . .

RUN go install -v .

ENTRYPOINT [ "/go/bin/filtra" ]
CMD [ "-config=/go/etc/config.toml" ]