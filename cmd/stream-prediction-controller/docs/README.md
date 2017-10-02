Stream Prediction Controller
============================

## Job States

<img src="https://raw.githubusercontent.com/NervanaSystems/kube-controllers-go/dpitt/states-doc/cmd/stream-prediction-controller/docs/states.png?token=ABI5xNHGyhuu6GhqHxRGZ6tmhGmAPrrOks5Z29rqwA%3D%3D" width=700>

A Stream Prediction job's state space consists of 4 states:

```go
const (
	Deploying states.State = iota
	Deployed
	Completed
	Error
)
```

* `Deploying` - In this states, a job has been created, but its sub-resources are pending.
* `Deployed` - This is the _ready_ state for a stream prediction job. In this state, it is ready to respond to queries.
* `Completed` - A `Completed` job has been undeployed.  `Completed` is a terminal state.
* `Error` - A job is in an `Error` state if an error has caused it to no longer be available to respond to queries.
