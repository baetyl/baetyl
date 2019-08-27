package main

import (
	"fmt"
	"io"
	"math"

	"github.com/baetyl/baetyl/logger"
	"gocv.io/x/gocv"
)

// Capture the video capture interfaces
type Capture interface {
	// Get parameter with property (=key).
	Get(prop gocv.VideoCaptureProperties) float64
	// Read reads the next frame from the VideoCapture to the Mat passed in
	// as the param. It returns false if the VideoCapture cannot read frame.
	Read(m *gocv.Mat) bool
	// Grab skips a specific number of frames.
	Grab(skip int)
	// Close
	io.Closer
}

// Video the video for inference
type Video struct {
	cfg VideoInfo
	cap Capture

	fps   float64
	width float64
	hight float64
	skip  int
}

// NewVideo creates a new video source
func NewVideo(cfg VideoInfo) (*Video, error) {
	v := &Video{cfg: cfg}
	return v, v.reopen()
}

// Read reads next image
func (v *Video) Read(m *gocv.Mat) error {
	if v.skip > 0 {
		v.cap.Grab(v.skip)
	}
	if !v.cap.Read(m) {
		v.reopen()
		return fmt.Errorf("failed to read image")
	}
	return nil
}

// Close closes the video source
func (v *Video) Close() error {
	if v.cap != nil {
		return v.cap.Close()
	}
	return nil
}

func (v *Video) reopen() error {
	cap, err := gocv.OpenVideoCapture(v.cfg.URL)
	if err != nil {
		return err
	}
	v.Close()
	v.cap = cap
	v.fps = cap.Get(gocv.VideoCaptureFPS)
	v.width = cap.Get(gocv.VideoCaptureFrameWidth)
	v.hight = cap.Get(gocv.VideoCaptureFrameHeight)
	if v.cfg.Limit.FPS > 0 {
		v.skip = int(math.Ceil((v.fps / v.cfg.Limit.FPS))) - 1
	}
	logger.Global.Infof("fps: %f, width: %f, hight: %f, skip: %d", v.fps, v.width, v.hight, v.skip)
	return nil
}
