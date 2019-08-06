# Example Go application using go-darknet

This is an example Go application which uses go-darknet.

## Install

```shell
go get github.com/gyonluks/go-darknet
go install github.com/gyonluks/go-darknet/example

# Alternatively
go build github.com/gyonluks/go-darknet/example
```

An executable named `example` should be in your `$GOPATH/bin`, if using
`go install`; otherwise it will be in your current working directory (`$PWD`),
if using `go build`.

## Run

```shell
$GOPATH/bin/example
```

Please ensure that `libdarknet.so` is in your `$LD_LIBRARY_PATH`.

## Notes

Note that the bounding boxes' values are ratios. To get the actual values, use
the ratios and multiply with either the image's width or height, depending on
which ratio is used.
