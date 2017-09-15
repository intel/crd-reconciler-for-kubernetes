FROM kube-controllers-go:v0.1.0

WORKDIR /go/src/github.com/NervanaSystems/kube-controllers-go/cmd/example-controller
RUN go install github.com/NervanaSystems/kube-controllers-go/cmd/example-controller
RUN make test
CMD /go/bin/example-controller
