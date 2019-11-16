module github.com/snowzach/doods

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/blendle/zapdriver v1.2.1
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/hybridgroup/mjpeg v0.0.0-20140228234708-4680f319790e
	github.com/lmittmann/ppm v1.0.0
	github.com/mattn/go-pointer v0.0.0-20190911064623-a0a44394634f
	github.com/mattn/go-tflite v0.0.0-20191105144949-8f71e638fc90
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/snowzach/certtools v1.0.2
	github.com/snowzach/mjpeg v0.0.1 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.5.0
	github.com/tensorflow/tensorflow v2.0.0+incompatible
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.12.0
	gocv.io/x/gocv v0.21.0
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8
	golang.org/x/net v0.0.0-20191105084925-a882066a44e0
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	golang.org/x/tools v0.0.0-20191105231337-689d0f08e67a // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6
	google.golang.org/grpc v1.25.0
	gopkg.in/yaml.v2 v2.2.5 // indirect
)

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422

go 1.13
