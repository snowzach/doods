module github.com/snowzach/doods

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/blendle/zapdriver v1.1.6
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.9.6
	github.com/hybridgroup/mjpeg v0.0.0-20140228234708-4680f319790e
	github.com/lmittmann/ppm v0.0.0-20190816103856-2887a48f2203
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-pointer v0.0.0-20180825124634-49522c3f3791
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/snowzach/certtools v1.0.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/tensorflow/tensorflow v1.14.0
	go.uber.org/zap v1.10.0
	gocv.io/x/gocv v0.20.0
	golang.org/x/image v0.0.0-20190802002840-cff245a6509b
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	google.golang.org/genproto v0.0.0-20190817000702-55e96fffbd48
	google.golang.org/grpc v1.23.0
)

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422

go 1.13
