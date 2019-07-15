FROM golang:1.12

WORKDIR /go/src/app
COPY . .

RUN go install -v .

ENTRYPOINT [ "app" ]
