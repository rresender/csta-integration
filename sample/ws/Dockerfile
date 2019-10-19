FROM golang
RUN export GOPATH=$HOME/go
RUN export GOROOT=/usr/local/go
RUN go get github.com/gorilla/mux
RUN go get github.com/go-redis/redis
RUN mkdir -p $GOPATH/src/github.com/rresender/csta-integration/sample/callinfows
COPY callinfows/ $GOPATH/src/github.com/rresender/csta-integration/sample/callinfows
RUN mkdir -p $GOPATH/src/github.com/rresender/csta-integration/sample/common
COPY common/ $GOPATH/src/github.com/rresender/csta-integration/sample/common
WORKDIR $GOPATH/src/github.com/rresender/csta-integration/sample/ws
RUN go build -o ws
CMD ["./ws"]

EXPOSE 7070
