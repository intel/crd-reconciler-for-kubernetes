FROM kube-controllers-go:v0.1.0

WORKDIR /go/src/github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller
RUN mv ../../vendor .
RUN go install github.com/NervanaSystems/kube-controllers-go/cmd/stream-prediction-controller
RUN make test
CMD /go/bin/stream-prediction-controller
