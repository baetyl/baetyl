package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path"
	"strings"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"gocv.io/x/gocv"
)

type content map[string]interface{}

func (c content) isDiscard() bool {
	if v, ok := c["imageDiscard"]; ok {
		if b, ok := v.(bool); ok && b {
			return true
		} else if s, ok := v.(string); ok && strings.EqualFold(s, "true") {
			return true
		}
	}
	return false
}

func (c content) location() string {
	if v, ok := c["imageLocation"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c content) topic() string {
	if v, ok := c["publishTopic"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c content) qos() uint32 {
	if v, ok := c["publishQOS"]; ok {
		if i, ok := v.(int); ok {
			return uint32(i)
		}
	}
	return 0
}

// Process the image processor
type Process struct {
	cfg   ProcessInfo
	cli   *mqtt.Dispatcher
	funcs *Functions

	size image.Point
	mean gocv.Scalar
}

// NewProcess creates a new process
func NewProcess(cfg ProcessInfo, cli *mqtt.Dispatcher, funcs *Functions) *Process {
	return &Process{
		cfg:   cfg,
		cli:   cli,
		funcs: funcs,
		size:  image.Pt(cfg.Before.Width, cfg.Before.Hight),
		mean:  gocv.NewScalar(cfg.Before.Mean.V1, cfg.Before.Mean.V2, cfg.Before.Mean.V3, cfg.Before.Mean.V4),
	}
}

// Before processes image before inference
func (p *Process) Before(img gocv.Mat) gocv.Mat {
	return gocv.BlobFromImage(img, p.cfg.Before.Scale, p.size, p.mean, p.cfg.Before.SwapRB, p.cfg.Before.Crop)
}

// After processes image after inference
func (p *Process) After(img gocv.Mat, results gocv.Mat, elapsedTime float64, captureTime time.Time) error {
	logger.Global.Debugln("type:", results.Type(), "total:", results.Total(), "size", results.Size())

	s := time.Now()
	out, err := p.funcs.Call(p.cfg.After.Function.Name, captureTime.UTC().UnixNano(), results.ToBytes())
	if err != nil {
		return err
	}
	logger.Global.Debugf("[After ][Call     ] elapsed time: %v", time.Since(s))

	if out == nil {
		return nil
	}

	s = time.Now()
	var cnt content
	logger.Global.Debugf(string(out))
	err = json.Unmarshal(out, &cnt)
	if err != nil {
		return err
	}
	logger.Global.Debugf("[After ][Unmarshal] elapsed time: %v", time.Since(s))

	discard := cnt.isDiscard()
	location := cnt.location()
	if !discard && location != "" {
		s = time.Now()
		if !gocv.IMWrite(location, img) {
			os.MkdirAll(path.Dir(location), 0755)
			if !gocv.IMWrite(location, img) {
				return fmt.Errorf("failed to save image: %s", location)
			}
		}
		logger.Global.Debugf("[After ][Write    ] elapsed time: %v", time.Since(s))
	}

	if p.cli == nil || cnt.topic() == "" {
		return nil
	}

	cnt["imageWidth"] = img.Cols()
	cnt["imageHight"] = img.Rows()
	cnt["imageCaptureTime"] = captureTime
	cnt["imageCaptureTime"] = captureTime
	cnt["imageInferenceTime"] = elapsedTime
	cnt["imageProcessTime"] = (time.Since(s)).Seconds() + elapsedTime
	if !discard && location == "" {
		s = time.Now()
		cnt["imageData"] = img.ToBytes()
		logger.Global.Debugf("[After ][ToBytes  ] elapsed time: %v", time.Since(s))
	}

	s = time.Now()
	msgData, err := json.Marshal(cnt)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	logger.Global.Debugf("[After ][Marshal  ] elapsed time: %v", time.Since(s))

	s = time.Now()
	err = p.cli.Publish(1, cnt.qos(), cnt.topic(), msgData, false, false)
	logger.Global.Debugf("[After ][Publish  ] elapsed time: %v", time.Since(s))
	return err
}
