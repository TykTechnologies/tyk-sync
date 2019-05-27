FROM golang:1.12.5

RUN go get -u -v github.com/TykTechnologies/tyk-sync

RUN go install -v github.com/TykTechnologies/tyk-sync

WORKDIR $GOPATH/src/github.com/TykTechnologies/tyk-sync