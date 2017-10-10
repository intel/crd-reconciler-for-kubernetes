FROM kube-controllers-go:v0.1.0

WORKDIR /go/src/github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller
# -gcflags: disable compiler optimizations and inlining
RUN go install -gcflags "-N -l" github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller
RUN make test
CMD /go/bin/stream-prediction-controller
