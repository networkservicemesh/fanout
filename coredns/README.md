## How to build Coredns image with fanout plugin?
At this moment exist two ways how to build Coredns image with fanout plugin.

#### Build via Coredns soruce code
You can build Coredns image via source code of Corends for this case you need to run:
```bash
$ cd $GOPATH
$ git clone https://github.com/coredns/coredns
$ cd coredns
$ echo fanout:github.com/networkservicemesh/fanout >> plugin.cfg
$ make
```

#### Build via custom `main.go` file
As alternative you can create your own `main.go` file and build your own Coredns binary. Take a look at [Official example](https://coredns.io/2017/07/25/compile-time-enabling-or-disabling-plugins/). After that you will need also to create your own `Dockerfile`.
You can also use files prepared for networkservicemesh. For this you need to run:
```bash
$ go build -o coredns/coredns coredns/main.go
$ docker build coredns/. -t "${ORG}/coredns:${TAG}"
```