package main

import (
	"fmt"

	"gocv.io/x/gocv"
)

// Infer video image inference
type Infer struct {
	cfg InferInfo
	net gocv.Net

	freq float64
}

// NewInfer creates a new inference
func NewInfer(cfg InferInfo) (*Infer, error) {
	net := gocv.ReadNet(cfg.Model, cfg.Config)
	if net.Empty() {
		return nil, fmt.Errorf("failed to read net from : %s %s", cfg.Model, cfg.Config)
	}
	backend := gocv.ParseNetBackend(cfg.Backend)
	target := gocv.ParseNetTarget(cfg.Device)
	net.SetPreferableBackend(gocv.NetBackendType(backend))
	net.SetPreferableTarget(gocv.NetTargetType(target))
	return &Infer{
		cfg:  cfg,
		net:  net,
		freq: gocv.GetTickFrequency(),
	}, nil
}

// Run runs inference on an image
func (n *Infer) Run(blob gocv.Mat) gocv.Mat {
	n.net.SetInput(blob, "")
	return n.net.Forward("")
}

// GetElapsedTime returns the elapsed time (seconds) for the last inference
func (n *Infer) GetElapsedTime() float64 {
	return n.net.GetPerfProfile() / n.freq
}

// Close closes this inference
func (n *Infer) Close() error {
	return n.net.Close()
}
