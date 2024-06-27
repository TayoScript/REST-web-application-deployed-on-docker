FROM golang:1.19 as builder

WORKDIR /go/src/app/cmd

COPY ./go.sum /go/src/app/go.sum
COPY ./go.mod /go/src/app/go.mod
COPY ./handlers /go/src/app/handlers
COPY ./cmd /go/src/app/cmd
COPY ./renewable-share-energy.csv /go/src/app/renewable-share-energy.csv

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o ../server

EXPOSE 8080

CMD ["../server"]
