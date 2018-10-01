FROM crd-reconciler-for-kubernetes

WORKDIR /go/src/github.com/intel/crd-reconciler-for-kubernetes/cmd/example-controller
# -gcflags: disable compiler optimizations and inlining
RUN make code-generation
RUN go install -gcflags "-N -l" github.com/intel/crd-reconciler-for-kubernetes/cmd/example-controller
RUN make test
CMD /go/bin/example-controller
