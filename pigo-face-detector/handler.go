package function

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/esimov/pigo/core"
	"github.com/fogleman/gg"
)

type FaceDetector struct {
	cascadeFile  string
	minSize      int
	maxSize      int
	shiftFactor  float64
	scaleFactor  float64
	iouThreshold float64
}

type DetectionResult struct {
	Faces       []image.Rectangle
	ImageBase64 string
}

var dc *gg.Context

// Handle a serverless request
func Handle(req []byte) string {
	var data []byte

	if val, exists := os.LookupEnv("input_mode"); exists && val == "url" {
		inputUrl := strings.TrimSpace(string(req))

		res, err := http.Get(inputUrl)
		if err != nil {
			return fmt.Sprintf("Unable to download image file from URI: %s, status %d", inputUrl, res.Status)
		}
		defer res.Body.Close()

		data, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Sprintf("Unable to read response body", err)
		}
	} else {
		var decodeError error
		data, decodeError = base64.StdEncoding.DecodeString(string(req))
		if decodeError != nil {
			data = req
		}

		contentType := http.DetectContentType(req)
		if contentType != "image/jpeg" && contentType != "image/png" {
			return fmt.Sprintf("Only jpeg or png images, either raw uncompressed bytes or base64 encoded are acceptable inputs, you uploaded: %s", contentType)
		}
	}
	tmpfile, err := ioutil.TempFile("/tmp", "image")
	if err != nil {
		log.Fatal("Unable to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = io.Copy(tmpfile, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Sprintf("Unable to copy the source URI to the destionation file")
	}

	var output string
	query, err := url.ParseQuery(os.Getenv("Http_Query"))
	if err == nil {
		output = query.Get("output")
	}

	if val, exists := os.LookupEnv("output_mode"); exists {
		output = val
	}

	fd := NewFaceDetector("./data/facefinder", 20, 2000, 0.1, 1.1, 0.2)
	faces, err := fd.DetectFaces(tmpfile.Name())

	if err != nil {
		return fmt.Sprintf("Error on face detection:, %v", err)
	}

	var resp DetectionResult
	var rects []image.Rectangle
	var image []byte

	if output == "image" || output == "json_image" {
		var err error
		rects, image, err = fd.DrawFaces(faces, false)
		if err != nil {
			return fmt.Sprintf("Error creating image output: %s", err)
		}

		resp = DetectionResult{
			Faces:       rects,
			ImageBase64: base64.StdEncoding.EncodeToString(image),
		}
	}
	if output == "image" {
		return string(image)
	}

	j, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf("Error encoding output: %s", err)
	}

	// Return face rectangle coordinates
	return string(j)
}

func NewFaceDetector(cf string, minSize, maxSize int, shf, scf, iou float64) *FaceDetector {
	return &FaceDetector{
		cascadeFile:  cf,
		minSize:      minSize,
		maxSize:      maxSize,
		shiftFactor:  shf,
		scaleFactor:  scf,
		iouThreshold: iou,
	}
}

func (fd *FaceDetector) DetectFaces(source string) ([]pigo.Detection, error) {
	src, err := pigo.GetImage(source)
	if err != nil {
		return nil, err
	}

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     fd.minSize,
		MaxSize:     fd.maxSize,
		ShiftFactor: fd.shiftFactor,
		ScaleFactor: fd.scaleFactor,
	}
	imgParams := pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}

	cascadeFile, err := ioutil.ReadFile(fd.cascadeFile)
	if err != nil {
		return nil, err
	}

	pigo := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err := pigo.Unpack(cascadeFile)
	if err != nil {
		return nil, err
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	faces := classifier.RunCascade(imgParams, cParams)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = classifier.ClusterDetections(faces, fd.iouThreshold)

	return faces, nil
}

func (fd *FaceDetector) DrawFaces(faces []pigo.Detection, isCircle bool) ([]image.Rectangle, []byte, error) {
	var qThresh float32 = 5.0
	var rects []image.Rectangle

	for _, face := range faces {
		if face.Q > qThresh {
			if isCircle {
				dc.DrawArc(
					float64(face.Col),
					float64(face.Row),
					float64(face.Scale/2),
					0,
					2*math.Pi,
				)
			} else {
				dc.DrawRectangle(
					float64(face.Col-face.Scale/2),
					float64(face.Row-face.Scale/2),
					float64(face.Scale),
					float64(face.Scale),
				)
			}
			rects = append(rects, image.Rect(
				face.Col-face.Scale/2,
				face.Row-face.Scale/2,
				face.Scale,
				face.Scale,
			))
			dc.SetLineWidth(3.0)
			dc.SetStrokeStyle(gg.NewSolidPattern(color.RGBA{R: 255, G: 0, B: 0, A: 255}))
			dc.Stroke()
		}
	}

	img := dc.Image()

	filename := fmt.Sprintf("/tmp/%d.jpg", time.Now().UnixNano())

	output, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, nil, err
	}
	defer os.Remove(filename)

	jpeg.Encode(output, img, &jpeg.Options{Quality: 100})

	rf, err := ioutil.ReadFile(filename)
	return rects, rf, err
}