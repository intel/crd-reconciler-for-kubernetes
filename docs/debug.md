## Debug short tutorial

- **`make env-up`** bring up the environment

- **`export GODEBUGGER=dlv|gdb`** add your choice of GODEBUGGER to your shell profile
-- See references pages for [gdb](https://golang.org/doc/gdb) and [dlv](https://github.com/derekparker/delve)

- **`make debug`** attach to the controller process running in the docker container. 
You should see the debugger prompt 
```
Type 'help' for list of commands.
(dlv|gdb)
```

- **`b hooks.go:17`** set a break point in hooks.go

- **`c`** continue running the process

- you should break in the debugger once a new example custom resource is created.
```
Breakpoint 1, main.(*exampleHooks).Add (c=0xc420113510, obj=...)
    at /go/src/github.com/intel/crd-reconciler-for-kubernetes/cmd/example-controller/hooks.go:17
17              example := obj.(*crv1.Example)
```
