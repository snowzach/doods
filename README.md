# DOODS
Dedicated Open Object Detection Service - Yes, it's a backronym...

DOODS is a GRPC/REST service that detects objects in images. It's designed to be very easy to use, run as a container and available remotely.

## API
The API uses gRPC to communicate but it has a REST gateway built in for ease of use. It supports both a single call RPC and a streaming interface.
It supports very basic pre-shared key authentication if you wish to protect it. It also supports TLS encryption but is disabled by default.
It uses the content-type header to automatically determine if you are connecting in REST mode or GRPC mode. It listens on port 8080 by default.

### GRPC Endpoints
The protobuf API definitations are in the `odrpc/odrpc.proto` file. There are 3 endpoints. 

- GetDetector - Get the list of configured detectors.
- Detect -  Detect objects in an image - Data should be passed as raw bytes in GRPC.
- DetectStream - Detect objects in a stream of images

### REST/JSON
The services are available via rest API at these endpoints
* `GET /version` - Get the version
* `GET /detectors` - Get the list of configured detectors
* `POST /detect` - Detect objects in an image

For `POST /detect` it expects JSON in the following format.
```
{
	"detector_name": "default",
	"data": "<base64 encoded image information>",
	"detect": {
		"*": 50
	}
}
```

The result is returned as:
```
{
    "id": "test",
    "detections": [
        {
            "top": 0,
            "left": 0.05,
            "bottom": .8552,
            "right": 0.9441,
            "label": "person",
            "confidence": 87.890625
        }
    ]
}
```
This will perform a detection using the detector called default. (If omitted, it will use one called default if it exists)
The `data`, when using the REST interface is base64 encoded image data. DOODS can decode png, bmp and jpg. 
The `detect` object allows you to specify the list of objects to detect as defined in the labels file. You can give a min percentage match.
You can also use "*" which will match anything with a minimum percentage.

Example 1-Liner to call the API using curl with image data: 
```
echo "{\"detector_name\":\"default\", \"detect\":{\"*\":60}, \"data\":\"`cat grace_hopper.png|base64 -w0`\"}" > /tmp/postdata.json && curl -d@/tmp/postdata.json -H "Content-Type: application/json" -X POST http://localhost:8080/detect
```

## Detectors
You should optimally pass image data in the requested size for the detector. If not, it will be automatically resized.
It can read BMP, PNG and JPG as well as PPM. For detectors that do not specify a size (inception) you do not need to resize

### TFLite
If you pass PPM image data in the right dimensions, it can be fed directly into tensorflow lite. This skips a couple steps for speed. 
You can also specify `hwAccel: true` in the config and it will enable Coral EdgeTPU hardware acceleration. 
You must also provide it an appropriate EdgeTPU model file. There are none included with the base image.

## Compiling
This is designed as a go module aware program and thus requires go 1.12 or better. It also relies heavily on CGO. The easiest way to compile it
is to use the Dockerfile which will build a functioning docker image. It's a little large but it includes 2 models.

## Configuration
The configuration can be specified in a number of ways. By default you can create a json file and call it with the -c option
you can also specify environment variables that align with the config file values.

Example:
```json
{
	"logger": {
        "level": "debug"
	}
}
```
Can be set via an environment variable:
```
LOGGER_LEVEL=debug
```

### Options:
| Setting                   | Description                                         | Default      |
| ------------------------- | --------------------------------------------------- | ------------ |
| logger.level              | The default logging level                           | "info"       |
| logger.encoding           | Logging format (console or json)                    | "console"    |
| logger.color              | Enable color in console mode                        | true         |
| logger.disable_caller     | Hide the caller source file and line number         | false        |
| logger.disable_stacktrace | Hide a stacktrace on debug logs                     | true         |
| ---                       | ---                                                 | ---          |
| server.host               | The host address to listen on (blank=all addresses) | ""           |
| server.port               | The port number to listen on                        | 8080         |
| server.tls                | Enable https/tls                                    | false        |
| server.devcert            | Generate a development cert                         | false        |
| server.certfile           | The HTTPS/TLS server certificate                    | "server.crt" |
| server.keyfile            | The HTTPS/TLS server key file                       | "server.key" |
| server.log_requests       | Log API requests                                    | true         |
| server.profiler_enabled   | Enable the profiler                                 | false        |
| server.profiler_path      | Where should the profiler be available              | "/debug"     |
| ---                       | ---                                                 | ---          |
| pidfile                   | Write a pidfile (only if specified)                 | ""           |
| profiler.enabled          | Enable the debug pprof interface                    | "false"      |
| profiler.host             | The profiler host address to listen on              | ""           |
| profiler.port             | The profiler port to listen on                      | "6060"       |
| ---                       | ---                                                 | ---          |
| doods.auth_key            | A pre-shared auth key. Disabled if blank            | ""           |
| doods.detectors           | The detector configurations                         | <see below>  |

### TLS/HTTPS
You can enable https by setting the config option server.tls = true and pointing it to your keyfile and certfile.
To create a self-signed cert: `openssl req -new -newkey rsa:2048 -days 3650 -nodes -x509 -keyout server.key -out server.crt`
You will need to mount these in the container and adjust the config to find them. 

### Detector Config
Detector config must be done with a configuration file. The default config includes one Tensorflow Lite mobilenet detector and the Tensorflow Inception model.
This is the default config with the exception of the threads and concurrent are tuned a bit for the architecture they are running on.
```
doods:
  detectors:
    - name: default
      type: tflite
      modelFile: models/coco_ssd_mobilenet_v1_1.0_quant.tflite
      labelFile: models/coco_labels0.txt
      numThreads: 4
      numConcurrent: 4
      hwAccel: false
      timeout: 2m
    - name: tensorflow
      type: tensorflow
      modelFile: models/faster_rcnn_inception_v2_coco_2018_01_28.pb
      labelFile: models/coco_labels1.txt
      numThreads: 4
      numConcurrent: 4
      hwAccel: false
      timeout: 2m
```
The default models are downloaded from google: coco_ssd_mobilenet_v1_1.0_quant_2018_06_29 and faster_rcnn_inception_v2_coco_2018_01_28.pb

The `numThreads` option is the number of threads that will be available for compatible operations in a model
The `numConcurrent` option sets the number of models that will be able to run at the same time. This should be 1 unless you have a beefy machine.
The `hwAccel` option is used to specify that a hardware device should be used. The only device supported is the edgetpu currently
If `timeout` is set than a detector (namely an edgetpu) that hangs for longer than the timeout will cause doods to error and exit. Generally this error is not recoverable and Doods needs to be restarted.

### Detector Types Supported
 * tflite - Tensorflow lite models - Supports Coral EdgeTPU if hwAccel: true and appropriate model is used
 * tensorflow - Tensorflow 

EdgeTPU models can be downloaded from here: https://coral.ai/models/ (Use the Object Detection Models)

## Examples - Clients
See the examples directory for sample clients

## Docker
To run the container in docker you need to map port 8080. If you want to update the models, you need to map model files and a config to use them. 
`docker run -it -p 8080:8080 snowzach/doods:latest`

There is a script called `fetch_models.sh` that you can download and run to create a models directory and download several models and outputs an `example.yaml` config file.
You could then run: `docker run -it -v ./models:/opt/doods/models -v ./example.yaml:/opt/doods/config.yaml -p 8080:8080 snowzach/doods:latest`

### Coral EdgeTPU
If you want to run it in docker using the Coral EdgeTPU, you need to pass the device to the container with: `--device /dev/bus/usb`
Example: `docker run -it --device /dev/bus/usb -p 8080:8080 snowzach/doods:latest`

## Misc
Special thanks to https://github.com/mattn/go-tflite as I would have never been able to figure out all the CGO stuff. I really wanted to write this in Go but I'm not good enough at C++/CGO to do it. Most of the tflite code is taken from that repo and customized for this tool.

And special thanks to @lhelontra, @marianopeck and @PINTO0309 for help in building tensorflow and binaries for bazel on the arm. 

## Docker Images
There are several published Docker images that you can use

* latest - This is a multi-arch image that points to the arm32 image, arm64 and noavx image
* noavx - 64 bit x86 image that should be a highly compatible with any cpu. 
* arm64 - Arm 64 bit image
* arm32 - Arm 32 bit/arm7 image optimized for the Raspberry Pi
* amd64 - 64 bit x86 image with all the fancy cpu features like avx and sse4.2
* cuda - Support for NVidia GPUs

## CUDA Support
There is now NVidia GPU support with an docker image tagged cuda, to run:
`docker run -it --gpus all -p 8080:8080 snowzach/doods:cuda`
For whatever reason, it can take a good 60-80 seconds before the model finishes loading.

## Compiling
You can compile it yourself using the plain `Dockerfile` which should pick the optimal CPU flags for your architecture. 
Make the `snowzach/doods:local` image with this command:
```
$ make libedgetpu
$ make docker
```
You only need to make libedgetpu once, it will download and compile it for all architectures. I hope to streamline that process into the main dockerfile at some point

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=QG353JUXA6BFW&source=url)
