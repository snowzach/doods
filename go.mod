module github.com/snowzach/doods

go 1.13

require (
	github.com/blendle/zapdriver v1.3.1
	github.com/go-chi/chi v1.5.0
	github.com/go-chi/cors v1.1.1
	github.com/go-chi/render v1.0.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/lmittmann/ppm v1.0.0
	github.com/mattn/go-pointer v0.0.1
	github.com/snowzach/certtools v1.0.2
	github.com/snowzach/mjpeg v0.0.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/tensorflow/tensorflow v2.0.3+incompatible
	go.uber.org/zap v1.16.0
	gocv.io/x/gocv v0.25.1-0.20201108120252-7f525fdbcb78
	golang.org/x/image v0.5.0
	golang.org/x/net v0.5.0
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f
	google.golang.org/grpc v1.52.0
	google.golang.org/grpc/examples v0.0.0-20230306234545-3292193519c3 // indirect
)

//replace github.com/tensorflow/tensorflow v2.3.1+incompatible => github.com/tensorflow/tensorflow v2.0.3+incompatible
