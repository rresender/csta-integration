FROM golang
RUN export GOPATH=$HOME/go
RUN export GOROOT=/usr/local/go
RUN go get github.com/gorilla/mux
RUN go get github.com/streadway/amqp
RUN go get github.com/garyburd/redigo/redis
RUN go get github.com/basgys/goxml2json
RUN mkdir -p $GOPATH/src/github.com/rresender/csta-integration/cti
COPY . $GOPATH/src/github.com/rresender/csta-integration/cti
WORKDIR $GOPATH/src/github.com/rresender/csta-integration/cti
RUN go build -o cti_monitoring
CMD ["./cti_monitoring"]

EXPOSE 7700
