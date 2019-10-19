FROM golang
RUN export GOPATH=$HOME/go
RUN export GOROOT=/usr/local/go
RUN go get github.com/gorilla/mux
RUN go get github.com/go-redis/redis
RUN go get github.com/streadway/amqp
RUN go get github.com/tidwall/gjson
RUN mkdir -p $GOPATH/src/github.com/rresender/csta-integration/sample/common/callinfoconsumer/
COPY callinfoconsumer/ $GOPATH/rresender/csta-integration/sample/callinfoconsumer/
RUN mkdir -p $GOPATH/src/github.com/rresender/csta-integration/sample/common
COPY common/ $GOPATH/src/github.com/rresender/csta-integration/sample/common
WORKDIR $GOPATH/src/github.com/rresender/csta-integration/sample/callinfoconsumer
RUN go build -o consumer
CMD ["./consumer"]

EXPOSE 7070
