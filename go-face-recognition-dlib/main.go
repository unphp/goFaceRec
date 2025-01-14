package main

import (
	"fmt"
	"github.com/Kagami/go-face"
	"github.com/barnex/cuda5/cu"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
	"gopkg.in/eapache/queue.v1"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const dataDir = "images"
const modelsDir = "models"
const testDir = "testimages"

func CheckErr(err error) {
	if nil != err {
		panic(err)
	}
}

func singleObjShowRectangleWithName(imgPath string, name []string) {
	window := gocv.NewWindow("Hello")

	// method 1: opencv reRecgnizeFace and rectangle it
	img := gocv.IMRead(imgPath, gocv.IMReadColor)
	if img.Empty() {
		fmt.Println("Error reading image from: %v", imgPath)
		return
	}

	///////////////////////////////////////////////////////////////
	// color for the rect when faces detected
	//blue := color.RGBA{0, 0, 255, 0}
	blue := color.RGBA{255, 0, 0, 2}
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		fmt.Println("Error reading cascade file: data/haarcascade_frontalface_default.xml")
		return
	}
	// detect faces
	rects := classifier.DetectMultiScale(img)
	fmt.Printf("found %d faces\n", len(rects))
	// draw a rectangle around each face on the original image
	for i, r := range rects {
		gocv.Rectangle(&img, r, blue, 2)

		//size := gocv.GetTextSize(name[i], gocv.FontHersheyPlain, 1.2, 2)
		//pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
		pt := image.Pt(r.Min.X, r.Min.Y-20)
		gocv.PutText(&img, name[i], pt, gocv.FontHersheyPlain, 2, blue, 2)
	}
	/////////////////////////////////////////////////////////////////////

	for {
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func multiObjshowRectangleWithName(imgPath string, recfaces []face.Face, name []string) {
	window := gocv.NewWindow("Hello")

	fmt.Println("mj test : multiObjshowRectangleWithName name = ", name, "\n")

	// method 2 : use reced faces data before
	img := gocv.IMRead(imgPath, gocv.IMReadColor)
	if img.Empty() {
		fmt.Println("Error reading image from: %v", imgPath)
		return
	}

	blue := color.RGBA{255, 0, 0, 2}

	for i, r := range recfaces {
		gocv.Rectangle(&img, r.Rectangle, blue, 2)

		//size := gocv.GetTextSize(name[i], gocv.FontHersheyPlain, 0.5, 2)
		//pt := image.Pt(r.Rectangle.Min.X+(r.Rectangle.Min.X/2)-(size.X/2), r.Rectangle.Min.Y-20)
		pt := image.Pt(r.Rectangle.Min.X, r.Rectangle.Min.Y-20)
		gocv.PutText(&img, name[i], pt, gocv.FontHersheyPlain, 2, blue, 2)
	}

	for {
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

func main() {
	fmt.Println("Facial Recognition System v0.01", cu.DeviceGet(0).Name())

	if len(os.Args) < 2 {
		fmt.Println("How to run:\n\tgo run main.go 1/2/3(camera)\n")
		return
	}
	choseId, _ := strconv.Atoi(os.Args[1])

	// 1. init recognizer
	rec, err := face.NewRecognizer(modelsDir)
	if err != nil {
		fmt.Println("Cannot initialize recognizer")
	}
	fmt.Println("Recognizer Initialized")
	defer rec.Close()

	// 2. set samples to the recognizer------1
	/*
		avengersImage := filepath.Join(dataDir, "avengers-02.jpeg")
		faces , err := rec.RecognizeFile(avengersImage)
		if err != nil {
			log.Fatalf("Can't recognize: %v", err)
		}
		fmt.Println("Number of Faces in Image: ", len(faces))

		var samples []face.Descriptor
		var avengers []int32
		for i, f := range faces {
			samples = append(samples, f.Descriptor)
			// Each face is unique on that image so goes to its own category.
			avengers = append(avengers, int32(i))
		}
		// Name the categories, i.e. people on the image.
		labels := []string{
			"Dr Strange",
			"Tony Stark",
			"Bruce Banner",
			"Wong",
		}
		// Pass samples to the recognizer.
		rec.SetSamples(samples, avengers)
		fmt.Println("Pass samples to the recognizer OK,LET'S start test.")
	*/

	// 2. set samples to the recognizer------2
	var samples []face.Descriptor
	var avengers []int32
	var labels []string
	var count int32

	// iterate src face images from given dir
	faceImages, err := ioutil.ReadDir(dataDir)
	CheckErr(err)
	count = 0
	for _, faceImageInfo := range faceImages {
		//get a face in each image
		faceImagePath := filepath.Join(dataDir, faceImageInfo.Name())
		faces, err := rec.RecognizeFileCNN(faceImagePath)
		if err != nil {
			log.Fatalf("Can't recognize: %v", err)
		}
		if len(faces) > 1 {
			fmt.Printf("Gets %d faces in %s, please keep one face in image.\n", len(faces), faceImagePath)
			os.Exit(0)
		}
		//fmt.Println("Number of Faces in Image: ", len(faces),"and dir : ",faceImagePath)

		// Each face's Descriptor
		for _, f := range faces {
			samples = append(samples, f.Descriptor)
		}
		// Each face is unique on this image, so goes to its own category.
		avengers = append(avengers, int32(count))
		// Name the categories, i.e. people on the image.
		if strings.Contains(faceImageInfo.Name(), ".") {
			Name := (faceImageInfo.Name())[0:strings.Index(faceImageInfo.Name(), ".")]
			fmt.Println(Name)
			labels = append(labels, Name)
		}

		count++
	}

	// Pass samples to the recognizer.
	rec.SetSamples(samples, avengers)
	fmt.Println("mj test : labels = ", labels, "\n")

	////////////////////////////////////////////////////////////////////////
	// 3. Now let's try to classify some not yet known image.
	//test 1 : single-objective in single picture
	if choseId == 1 {
		fmt.Println("choseId == ", choseId)
		singleObjImgPath := filepath.Join(testDir, "mj1.jpg")
		//singleObjImgPath := filepath.Join(testDir, "ts.jpg")
		//singleObjImgPath := filepath.Join(testDir, "tony-stark.jpg")
		singleFace, err := rec.RecognizeSingleFileCNN(singleObjImgPath)
		if err != nil {
			log.Fatalf("Can't recognize: %v", err)
		}
		if singleFace == nil {
			log.Fatalf("Not a single face on the image")
			os.Exit(0)
		}

		var recName []string
		//singleFaceID := rec.Classify(singleFace.Descriptor)
		singleFaceID := rec.ClassifyThreshold(singleFace.Descriptor, 0.35)
		fmt.Println("mj retshld: ", singleFaceID)
		if singleFaceID < 0 {
			recName = append(recName, "unkown")
			fmt.Printf("Can't classify : %v\n", recName)
		} else {
			recName = append(recName, labels[singleFaceID])
			fmt.Printf("Classified : %v\n", recName)
		}
		// 4. Rectangle faces and relate name
		singleObjShowRectangleWithName(singleObjImgPath, recName)
	}
	////////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////////

	//test 2 : multi-objective in single picture
	if choseId == 2 {
		fmt.Println("choseId == ", choseId)
		multiObjImgPath := filepath.Join(testDir, "avengers-02.jpeg")
		//multiObjImgPath := filepath.Join(testDir, "lindan_chenlong.jpg")
		faces, err := rec.RecognizeFileCNN(multiObjImgPath)
		if err != nil {
			log.Fatalf("Can't recognize: %v", err)
		}
		if faces == nil {
			log.Fatalf("No faces on the image")
		}
		fmt.Println("Number of Faces in Image: ", len(faces))

		// rec each face in img
		var recMultiName []string
		for _, f := range faces {
			//faceID := rec.Classify(f.Descriptor)
			faceID := rec.ClassifyThreshold(f.Descriptor, 0.35)
			fmt.Println("mj retshld: ", faceID)
			if faceID < 0 {
				//recMultiName[i] = "unkown"
				//fmt.Printf("Can't classify : %s\n",recName)
				recMultiName = append(recMultiName, "unkown")
			} else {
				//recMultiName[i] = labels[faceID]
				//fmt.Printf("Classified : %s\n",recName)
				recMultiName = append(recMultiName, labels[faceID])
			}
		}
		fmt.Println("mj test : recMultiName = ", recMultiName, "\n")
		// 4. Rectangle faces and relate name
		multiObjshowRectangleWithName(multiObjImgPath, faces, recMultiName)
		//singleObjShowRectangleWithName(multiObjImgPath, recMultiName)
	}
	////////////////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////////////////

	//test 3 : Rec&Show multi-objective from camera stream
	if choseId == 3 {
		fmt.Println("choseId == ", choseId)
		// method 1	: capture frmae,rec frame,push frmae with one routine
		//cameraMultiObjShowRecFacesWithName1(rec, labels)

		// method 2 : capture frmae,rec&&push frmae with two routines
		cameraMultiObjShowRecFacesWithName2(rec, labels)
	}
	////////////////////////////////////////////////////////////////////////
}

func cameraMultiObjShowRecFacesWithName1(rec *face.Recognizer, labels []string) {
	// set to use a video capture device 0
	//deviceID := 0
	//deviceID := "rtmp://192.168.43.74:1935/live/movie"
	deviceID := "rtsp://admin:aA12345678@192.168.1.179:554/Streaming/Channels/101"

	// open webcam : OpenVideoCapture("2.mp4")/OpenVideoCapture("rtsp://192.168.1.123:1935/live/show/")
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()
	fmt.Println("open cam ok")

	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()
	fmt.Println("NewWindow ok")

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	//for ffmpeg push to rtmp server
	width := int(webcam.Get(gocv.VideoCaptureFrameWidth))
	height := int(webcam.Get(gocv.VideoCaptureFrameHeight))
	fps := int(webcam.Get(gocv.VideoCaptureFPS))

	cmdArgs := fmt.Sprintf("%s %s %s %d %s %s",
		"ffmpeg -y -an -f rawvideo -vcodec rawvideo -pix_fmt bgr24 -s",
		fmt.Sprintf("%dx%d", width, height),
		"-r",
		fps,
		"-i - -c:v libx264 -pix_fmt yuv420p -preset ultrafast -f flv",
		"rtmp://192.168.2.143:1935/live/movie",
	)
	list := strings.Split(cmdArgs, " ")
	cmd := exec.Command(list[0], list[1:]...)
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer cmdIn.Close()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("start reading camera device: %v\n", deviceID)
	cameraTmpImgPath := filepath.Join(testDir, "cameratmptest.jpg")
	for {
		// read a frame
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}
		fmt.Println("read frame ok")

		// get each face's name from lables[]
		///////////////////////////////////////////////////////////////////////////////
		gocv.IMWrite(cameraTmpImgPath, img)
		faces, err := rec.RecognizeFileCNN(cameraTmpImgPath)
		if err != nil {
			//log.Fatalf("Can't recognize: %v", err)
			fmt.Printf("Can't recognize...")
		}
		if faces == nil {
			//log.Fatalf("No faces on the image")
			fmt.Printf("No faces on the image")
		}
		fmt.Println("Number of Faces in Image: ", len(faces))

		// rec each face in img
		var recCamMultiName [30]string
		for i, f := range faces {
			faceID := rec.ClassifyThreshold(f.Descriptor, 0.35)
			fmt.Println("mj retshld: ", faceID)
			if faceID < 0 {
				recCamMultiName[i] = "unkown"
				//fmt.Printf("Can't classify : %s\n",recName)
			} else {
				recCamMultiName[i] = labels[faceID]
				//fmt.Printf("Classified : %s\n",recName)
			}
		}
		///////////////////////////////////////////////////////////////////////////////

		// set name and rectangle on img
		///////////////////////////////////////////////////////////////////////////////
		// draw a rectangle around each face on the original image
		for i, r := range faces {
			gocv.Rectangle(&img, r.Rectangle, blue, 2)

			//size := gocv.GetTextSize(recCamMultiName[i], gocv.FontHersheyPlain, 1.2, 2)
			//pt := image.Pt(r.Rectangle.Min.X+(r.Rectangle.Min.X/2)-(size.X/2), r.Rectangle.Min.Y-2)
			pt := image.Pt(r.Rectangle.Min.X, r.Rectangle.Min.Y-20)
			gocv.PutText(&img, recCamMultiName[i], pt, gocv.FontHersheyPlain, 2, blue, 2)
		}
		///////////////////////////////////////////////////////////////////////////////

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)

		// push to rtmp server by ffmpeg
		///////////////////////////////////////////////////////////////////////////////
		// push to rtmp server
		cnt, err := cmdIn.Write([]byte(img.ToBytes()))
		//cnt,err :=cmdIn.Write(img.ToBytes())
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(0)
		} else {
			fmt.Printf("send cnt=%d\n", cnt)
		}
		///////////////////////////////////////////////////////////////////////////////
	}
}

var wg sync.WaitGroup

const stopGetFrame int = 1
const stopRecFrame int = 2

func cameraMultiObjShowRecFacesWithName2(rec *face.Recognizer, labels []string) {
	frameQueue := queue.New() // queue for frame from camera

	quiteChan := make(chan int, 1) // quite channel for sync 2 routines
	defer close(quiteChan)

	// set src
	//deviceID := 0
	//deviceID := "rtmp://58.200.131.2:1935/livetv/hunantv"
	deviceID := "rtsp://admin:aA12345678@192.168.1.179:554/Streaming/Channels/101"

	// open webCam
	webCam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webCam.Close()
	fmt.Println("open webCam ok")

	// get frame from camera to frameQueue for rec,argsChan for ffmpeg push
	wg.Add(1)
	go getFrameFromCameraToQueue(webCam, frameQueue, quiteChan)
	//time.Sleep(1)

	// get frame from frameQueue to rec && tracke,then push it to rtmp server
	wg.Add(1)
	go recFaceAndMarkName(webCam, frameQueue, quiteChan, rec, labels)
	//time.Sleep(1)

	wg.Wait()
	fmt.Println("main exit...")
}

func getFrameFromCameraToQueue(webCam *gocv.VideoCapture, fQueue *queue.Queue, quiteChan chan int) {
	defer wg.Done()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

loopGetFrame:
	for {
		// select for listenning quite msg from rec/push frame routine
		select {
		case qMsg := <-quiteChan:
			if qMsg == 2 {
				fmt.Printf("========== 11111111.=========== \n")
				fmt.Printf("get routine: quite msg from rec frame routine.\n")
				break loopGetFrame
			}
		default:
			// do nothing
			//case <-time.After(3 * time.Millisecond):
			// wait 3 ms, do nothing while timeout
		}
		fmt.Printf("========== 00000000.=========== \n")
		if webCam.IsOpened() {
			// read frame from cam
			if ok := webCam.Read(&img); !ok {
				fmt.Printf("cannot read webCam device.\n")
				quiteChan <- stopGetFrame
				fmt.Printf("========== 2222222222.=========== \n")
				break
			}
			if img.Empty() {
				fmt.Printf("========== 333333333.=========== \n")
				continue
			}
			//fmt.Println("read frame ok")

			// resize 320*240
			//gocv.Resize(img,&dstImg,dstImg,320,240,gocv.InterpolationCubic)

			// put frame into fqueue
			fmt.Printf("========== 5555555555555.=========== \n")
			fQueue.Add(img)
		} else {
			fmt.Println("webCam has been closed!")
			fmt.Printf("========== 444444444.=========== \n")
			quiteChan <- stopGetFrame
			break
		}
	}
}

func recFaceAndMarkName(webCam *gocv.VideoCapture, fQueue *queue.Queue, quiteChan chan int, rec *face.Recognizer, labels []string) {
	defer wg.Done()

	window := gocv.NewWindow("Face Detect")
	defer window.Close()
	fmt.Println("NewWindow ok")

	// prepare image matrix
	recImg := gocv.NewMat()
	defer recImg.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}
	var recCamMultiName [30]string
	cameraTmpImgPath := filepath.Join(testDir, "cameratmptest.jpg")

	var trackers []gocv.Tracker
	var lastFaceCnt int
	var trackAll bool

	// for ffmpeg push to rtmp server
	// get webCam ops:width/height/fps
	width := int(webCam.Get(gocv.VideoCaptureFrameWidth))
	height := int(webCam.Get(gocv.VideoCaptureFrameHeight))
	fps := int(webCam.Get(gocv.VideoCaptureFPS))

	cmdArgs := fmt.Sprintf("%s %s %s %d %s %s",
		"ffmpeg -y -an -f rawvideo -vcodec rawvideo -pix_fmt bgr24 -s",
		fmt.Sprintf("%dx%d", width, height),
		"-r",
		fps,
		"-i - -c:v libx264 -pix_fmt yuv420p -preset ultrafast -f flv",
		"rtmp://192.168.2.143:1935/live/movie",
	)
	//cmdArgs := fmt.Sprintf("%s %s %s %d %s %s",
	//	"ffmpeg -y -an -f rawvideo -vcodec rawvideo -pix_fmt bgr24 -s",
	//	fmt.Sprintf("%dx%d", width, height),
	//	"-r",
	//	fps,
	//	"-i - -c:v libx264 -pix_fmt yuv420p -preset ultrafast -f flv",
	//	"rtmp://192.168.2.143:1935/live/movie",
	//)
	list := strings.Split(cmdArgs, " ")
	cmd := exec.Command(list[0], list[1:]...)
	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("%v\n", err)
		quiteChan <- stopRecFrame
		return
	}
	defer cmdIn.Close()
	if err := cmd.Start(); err != nil {
		fmt.Printf("%v\n", err)
		quiteChan <- stopRecFrame
		return
	}

loopRecFrame:
	for {
		// select for listenning quite msg from get/push frame routine
		select {
		case qMsg := <-quiteChan:
			if qMsg == 1 {
				fmt.Printf("rec routine: quite msg from get frame routine.\n")
				break loopRecFrame
			}
		default:
			// do nothing
			//case <-time.After(3 * time.Millisecond):
			// wait 3 ms, do nothing while timeout
		}

		if fQueue.Length() > 0 {
			queueImg := fQueue.Get(0)
			switch qImg := queueImg.(type) {
			case gocv.Mat:
				fmt.Printf("========== loopRecFrame. 001=========== \n")
				recImg = qImg
			default:
				continue
			}

			//_, err = cmdIn.Write(recImg.ToBytes())
			//if err != nil {
			//	fmt.Printf("push error 00 : %v\n", err)
			//	//quiteChan <- stopRecFrame
			//	//break
			//}

			////////////////////////////////////////////////////////////////////
			// rec each face in recImg
			gocv.IMWrite(cameraTmpImgPath, recImg)
			faces, err := rec.RecognizeFileCNN(cameraTmpImgPath)
			//faces, err := rec.RecognizeCNN([]byte(recImg.ToBytes()))
			if err != nil {
				fmt.Printf("Can't recognize...\n")
			}
			if faces == nil {
				//fmt.Printf("No faces on the image\n")
			}
			//fmt.Printf("lastFaceCnt=%d,len(faces)=%d\n",lastFaceCnt,len(faces))

			if len(trackers) == 0 || lastFaceCnt != len(faces) {
				// rec name from faces && init tracker for each face.
				// clear condition
				trackers = trackers[0:0]

				if len(faces) > 0 {
					// get each face's name from lables[] and init tracker...
					for i, f := range faces {
						faceID := rec.ClassifyThreshold(f.Descriptor, 0.30)
						if faceID < 0 {
							recCamMultiName[i] = "unkown"
						} else {
							recCamMultiName[i] = labels[faceID]
						}
						fmt.Printf("========== loopRecFrame. trackers=========== \n")
						tracker := contrib.NewTrackerKCF()
						//tracker := contrib.NewTrackerCSRT()
						tracker.Init(recImg, f.Rectangle)
						trackers = append(trackers, tracker)
					}
				}

				lastFaceCnt = len(faces)
			} else if len(trackers) != 0 && lastFaceCnt == len(faces) {
				// track frame
				trackAll = true

				for i, t := range trackers {
					r, ok := t.Update(recImg)
					if ok {
						gocv.Rectangle(&recImg, r, blue, 3)
						pt := image.Pt(r.Min.X, r.Min.Y-20)
						gocv.PutText(&recImg, recCamMultiName[i], pt, gocv.FontHersheyPlain, 2, blue, 3)

						// use faces point to rectangle
						//gocv.Rectangle(&recImg, faces[i].Rectangle, blue, 3)
						//pt := image.Pt(faces[i].Rectangle.Min.X, faces[i].Rectangle.Min.Y-20)
						//gocv.PutText(&recImg, recCamMultiName[i], pt, gocv.FontHersheyPlain, 2, blue, 2)
					} else {
						fmt.Printf("track %s failed,reInit tracker again.\n", recCamMultiName[i])
						trackAll = false
						break
					}
				}

				// tracking someone failed,reRecognize to track
				if !trackAll {
					for _, t := range trackers {
						t.Close()
					}

					trackers = trackers[0:0]
				}
			}
			window.IMShow(recImg)
			window.WaitKey(1)
			////////////////////////////////////////////////////////////////////
			fmt.Printf("========== loopRecFrame. push to server =========== \n")
			//push to rtmp server
			//_, err = cmdIn.Write([]byte(recImg.ToBytes()))
			//if err != nil {
			//	fmt.Printf("push error: %v\n", err)
			//	//quiteChan <- stopRecFrame
			//	//break
			//}
		}
	}
}
