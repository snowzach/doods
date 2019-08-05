package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"

	"google.golang.org/grpc"

	"github.com/snowzach/doods/odrpc"

	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
)

var (
	deviceID int
	err      error
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("How to run:\n\trtspdetector [source url] [host:port] [doods server] [detector]")
		return
	}

	// parse args
	source := os.Args[1]
	host := os.Args[2]
	server := os.Args[3]
	detector := os.Args[4]

	// open webcam
	webcam, err = gocv.OpenVideoCapture(source)
	if err != nil {
		fmt.Printf("Error opening capture device: %v: %v\n", source, err)
		return
	}
	defer webcam.Close()

	// create the mjpeg stream
	stream = mjpeg.NewStream()

	// start capturing
	go mjpegCapture(server, detector)

	fmt.Println("Capturing. Point your browser to " + host)

	// start http server
	http.Handle("/", stream)
	log.Fatal(http.ListenAndServe(host, nil))
}

func mjpegCapture(server string, detector string) {

	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}

	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial(server, dialOptions...)
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	// gRPC version Client
	client := odrpc.NewOdrpcClient(conn)
	detectStream, err := client.DetectStream(context.Background())
	if err != nil {
		log.Fatalf("Could not stream: %v", err)
	}

	img := gocv.NewMat()
	defer img.Close()
	detectImg := gocv.NewMat()
	defer detectImg.Close()

	// color for the rect when faces detected
	green := color.RGBA{0, 255, 0, 0}
	var rs = make([]image.Rectangle, 0)
	var labels = make([]string, 0)
	var m sync.Mutex
	var detectorReady bool = true

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		request := &odrpc.DetectRequest{
			DetectorName: detector,
			Detect: map[string]float32{
				"*": 50, //
			},
		}

		gocv.Resize(img.Region(image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 1080, Y: 1080}}), &detectImg, image.Point{X: 300, Y: 300}, 0.0, 0.0, gocv.InterpolationNearestNeighbor)
		request.Data, err = gocv.IMEncode(".bmp", detectImg)
		if err != nil {
			continue
		}

		m.Lock()
		if detectorReady {
			detectorReady = false
			if err := detectStream.Send(request); err != nil {
				log.Fatalf("could not stream send %v", err)
			}
			go func() {
				response, err := detectStream.Recv()
				if err == io.EOF {
					log.Fatalf("can not receive %v", err)
				}
				if err != nil {
					log.Fatalf("can not receive %v", err)
				}
				log.Printf("Processed: %v", response)

				m.Lock()
				detections := len(response.Detections)
				rs = make([]image.Rectangle, detections, detections)
				labels = make([]string, detections, detections)
				for x := 0; x < detections; x++ {
					rs[x] = image.Rectangle{
						Min: image.Point{X: int(float32(response.Detections[x].X1) * 3.6), Y: int(float32(response.Detections[x].Y1)*3.6) * 4},
						Max: image.Point{X: int(float32(response.Detections[x].X2) * 3.6), Y: int(float32(response.Detections[x].Y2)*3.6) * 4},
					}
					labels[x] = response.Detections[x].Label
				}
				detectorReady = true
				m.Unlock()
			}()
		}
		m.Unlock()

		for x := 0; x < len(rs); x++ {
			gocv.Rectangle(&img, rs[x], green, 2)
			size := gocv.GetTextSize(labels[x], gocv.FontHersheyPlain, 1.2, 1)
			pt := image.Pt(rs[x].Min.X+(rs[x].Min.X/2)-(size.X/2), rs[x].Min.Y-2)
			label := labels[x]
			gocv.PutText(&img, label, pt, gocv.FontHersheyPlain, 1.2, green, 1)
		}

		// re-encode with boxes
		request.Data, err = gocv.IMEncode(".jpg", img)
		if err != nil {
			continue
		}

		// buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(request.Data)

	}
}
