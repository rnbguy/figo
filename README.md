Figo
===

_I moved on to Rust-lang. Checkout my similar rust-project: [*rope*](https://github.com/rnbguy/rope)._
---

A simple, Zeroconf-based, peer to peer file transfer utility where you and your friend are in same local network and can talk to each other.

This app is inspired by [zget](https://github.com/nils-werner/zget) but written in [Golang](https://golang.org).

Installation
---
* Make sure
	- `golang-1.8` is installed and `go` executable is available in system path.
	- `$GOPATH/bin` is included system path. 
	- `git` is installed.
* Run `go get github.com/rnbdev/figo/...` to install `figo`. Mind the trailing `/...`.
	- It will download `figo` in `$GOPATH` and build and copy the executable in `$GOPATH/bin`.

Usage
---
* `figos` to send file.
* `figor` to receive file.

Suppose you want to share a video with Jay. Run
```
$ figos sunday_picnic.mp4
```

Then tell Jay to run,
```
$ figor sunday_picnic.mp4
```
