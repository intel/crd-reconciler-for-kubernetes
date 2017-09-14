FROM golang:1.8.3

RUN mkdir -p /go/src/github.com/NervanaSystems
ADD . /go/src/github.com/NervanaSystems/kube-controllers-go
WORKDIR /go/src/github.com/NervanaSystems/kube-controllers-go
RUN make test
CMD /bin/bash
