module github.com/snowzach/doods

require (
	github.com/blendle/zapdriver v1.3.1
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/lmittmann/ppm v1.0.0
	github.com/mattn/go-pointer v0.0.0-20190911064623-a0a44394634f
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/snowzach/certtools v1.0.2
	github.com/snowzach/mjpeg v0.0.1
	github.com/spf13/afero v1.3.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0
	github.com/tensorflow/tensorflow v2.0.0+incompatible
	go.uber.org/zap v1.15.0
	gocv.io/x/gocv v0.21.0
	golang.org/x/image v0.0.0-20200618115811-c13761719519
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200623045635-ff88973b1e4e // indirect
	google.golang.org/genproto v0.0.0-20200623002339-fbb79eadd5eb
	google.golang.org/grpc v1.30.0
	gopkg.in/ini.v1 v1.57.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422

go 1.13
