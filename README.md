# protobuilder

### Geting Started

First install [`protoc`](https://github.com/golang/protobuf#installation).

Set your `$GOPATH` to ensure access to the `protoc` executable:

```
export PATH=$PATH:$GOPATH/bin
```


### Build protobufs

Time to build your protobufs:

```
$ go run cmd/build.go protos
```

The program will run `protoc` on our example proto files, compiling them to our desired protocol buffers in a generated directory `protobuf/..`

Just insert your own `.proto` files into the `protos` directory.