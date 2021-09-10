FROM golang

MAINTAINER Nick Lubyshev <lubyshev@gmail.com>

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN go build -o app main.go

EXPOSE $APP_SERVER_PORT

CMD ["go","run","./main.go"]
