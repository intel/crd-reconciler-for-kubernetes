FROM kube-controllers-go-dep

ADD . /go/src/github.com/NervanaSystems/kube-controllers-go
WORKDIR /go/src/github.com/NervanaSystems/kube-controllers-go
RUN make test
CMD /bin/bash
