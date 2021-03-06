#
# Copyright (c) 2018 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: EPL-2.0
#

FROM crd-reconciler-for-kubernetes-dep

ADD . /go/src/github.com/intel/crd-reconciler-for-kubernetes
WORKDIR /go/src/github.com/intel/crd-reconciler-for-kubernetes
RUN go get github.com/kubernetes/gengo/examples/deepcopy-gen
RUN make code-generation
RUN make test
CMD /bin/bash
