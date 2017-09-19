FROM kube-controllers-go:v0.1.0

WORKDIR /go/src/github.com/NervanaSystems/kube-controllers-go/cmd/example-controller
RUN mv ../../vendor .
RUN go install -a github.com/NervanaSystems/kube-controllers-go/cmd/example-controller
RUN make test
CMD /go/bin/example-controller
