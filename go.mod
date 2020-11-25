module github.com/snowzach/doods

go 1.13

require (
	github.com/blendle/zapdriver v1.3.1
	github.com/go-chi/chi v1.5.0
	github.com/go-chi/cors v1.1.1
	github.com/go-chi/render v1.0.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/lmittmann/ppm v1.0.0
	github.com/mattn/go-pointer v0.0.1
	github.com/snowzach/certtools v1.0.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/tensorflow/tensorflow v2.3.1+incompatible
	go.uber.org/zap v1.16.0
	gocv.io/x/gocv v0.25.0
	golang.org/x/image v0.0.0-20200927104501-e162460cd6b5
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	google.golang.org/genproto v0.0.0-20201119123407-9b1e624d6bc4
	google.golang.org/grpc v1.33.2
)

replace github.com/tensorflow/tensorflow v2.3.1+incompatible => github.com/tensorflow/tensorflow v2.0.3+incompatible
