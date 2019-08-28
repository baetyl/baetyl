package main

import (
	"time"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"gocv.io/x/gocv"
)

func main() {
	baetyl.Run(func(ctx baetyl.Context) error {
		var cfg Config
		// load custom config
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		ctx.Log().Infoln(cfg)
		// create a hub client
		cli, err := ctx.NewHubClient("", nil)
		if err != nil {
			return err
		}
		cli.Start(nil)
		defer cli.Close()
		//  create inference
		infer, err := NewInfer(cfg.Infer)
		if err != nil {
			return err
		}
		defer infer.Close()
		// create video
		video, err := NewVideo(cfg.Video)
		if err != nil {
			return err
		}
		defer video.Close()
		// create function clients
		funcs, err := NewFunctions(cfg.Functions)
		if err != nil {
			return err
		}
		// create process
		proc := NewProcess(cfg.Process, cli, funcs)

		var s time.Time
		img := gocv.NewMat()
		defer img.Close()
		quit := ctx.WaitChan()
		for {
			select {
			case <-quit:
				ctx.Log().Infof("quit")
				return nil
			default:
			}
			s = time.Now()
			err := video.Read(&img)
			if err != nil {
				ctx.Log().Errorf(err.Error())
				time.Sleep(time.Second) // TODO: configured
				continue
			}
			if img.Empty() {
				continue
			}
			ctx.Log().Debugf("[Read  ] elapsed time: %v", time.Since(s))

			t := time.Now()
			blob := proc.Before(img)
			ctx.Log().Debugf("[Before] elapsed time: %v", time.Since(t))

			s = time.Now()
			prob := infer.Run(blob)
			ctx.Log().Debugf("[Infer ] elapsed time: %v", time.Since(s))

			s = time.Now()
			err = proc.After(img, prob, infer.GetElapsedTime(), t)
			if err != nil {
				ctx.Log().Errorf(err.Error())
			}
			ctx.Log().Debugf("[After ] elapsed time: %v", time.Since(s))

			prob.Close()
			blob.Close()
		}
	})
}
