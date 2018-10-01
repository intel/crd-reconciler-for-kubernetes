FROM crd-reconciler-for-kubernetes-dep

ADD . /go/src/github.com/intel/crd-reconciler-for-kubernetes
WORKDIR /go/src/github.com/intel/crd-reconciler-for-kubernetes
RUN go get github.com/kubernetes/gengo/examples/deepcopy-gen
RUN make code-generation
RUN make test
CMD /bin/bash
