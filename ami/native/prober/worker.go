package prober

import (
	"math/rand"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/kardianos/service"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/clock"
)

// Type of probe (liveness, readiness or startup)
type probeType int

const (
	liveness probeType = iota
	startup
)

type probeKey struct {
	Name      string
	Version   string
	ProbeType probeType
}

type worker struct {
	// Channel for stopping the probe.
	stopCh chan struct{}
	// Describes the probe configuration (read-only)
	spec      *v1.Probe
	probeType probeType
	// The process to probe
	svc          service.Service
	app          *specV1.Application
	probeManager *manager
	log          *log.Logger
	startedAt    time.Time
	// The last probe result for this worker.
	lastResult ProbeResult
	// How many times in a row the probe has returned the same result.
	resultRun int
}

func newWorker(m *manager, svc service.Service, probeType probeType, app *specV1.Application) *worker {
	return &worker{
		stopCh:       make(chan struct{}, 1),        // Buffer so stop() can be non-blocking.
		spec:         app.Services[0].LivenessProbe, // native only support one service
		app:          app,
		svc:          svc,
		probeManager: m,
		probeType:    probeType,
		log:          m.log,
		startedAt:    clock.RealClock{}.Now(),
	}
}

func (w *worker) run() {
	probeTickerPeriod := time.Duration(w.spec.PeriodSeconds) * time.Second

	// If baetyl-core restarted the probes could be started in rapid succession.
	// Let the worker wait for a random portion of tickerPeriod before probing.
	// Do it only if the baetyl-core has started recently.
	if probeTickerPeriod > time.Since(w.probeManager.start) {
		time.Sleep(time.Duration(rand.Float64() * float64(probeTickerPeriod)))
	}
	probeTicker := time.NewTicker(probeTickerPeriod)
	defer func() {
		// Clean up.
		probeTicker.Stop()
		key := probeKey{Name: w.app.Name, Version: w.app.Version, ProbeType: w.probeType}
		w.probeManager.removeWorker(key)
	}()
probeLoop:
	for w.doProbe() {
		// Wait for next probe tick.
		select {
		case <-w.stopCh:
			break probeLoop
		case <-probeTicker.C:
		}
	}
}

// stop stops the probe worker. The worker handles removes itself from its manager.
// It is safe to call stop multiple times.
func (w *worker) stop() {
	select {
	case w.stopCh <- struct{}{}:
	default: // Non-blocking.
	}
}

// doProbe probes the container once
func (w *worker) doProbe() (keepGoing bool) {
	defer func() { recover() }() // Actually eat panics (HandleCrash takes care of logging)
	defer runtime.HandleCrash(func(_ interface{}) { keepGoing = true })

	key := &probeKey{Name: w.app.Name, Version: w.app.Version, ProbeType: w.probeType}
	status, ok := w.svc.Status()
	if ok != nil {
		w.log.Debug("No status for process", log.Any("key", key))
		return true
	}
	if status == service.StatusStopped {
		w.log.Debug("Process is terminated, exiting probe worker", log.Any("key", key))
		return false
	}
	// Stop probing for liveness until process has started.
	if w.probeType == liveness && status == service.StatusUnknown {
		w.log.Debug("No status for process", log.Any("key", key))
		return true
	}
	// Stop probing for startup once process has started.
	// we keep it running to make sure it will work for restarted process.
	if w.probeType == startup && status == service.StatusRunning {
		return true
	}
	// Probe disabled for InitialDelaySeconds.
	if int32(time.Since(w.startedAt).Seconds()) < w.spec.InitialDelaySeconds {
		return true
	}

	result, err := w.probeManager.prober.probe(key, w.spec)
	if err != nil {
		// Prober error, throw away the result.
		return true
	}
	if w.lastResult == result {
		w.resultRun++
	} else {
		w.lastResult = result
		w.resultRun = 1
	}
	if (result == Failure && w.resultRun < int(w.spec.FailureThreshold)) ||
		(result == Success && w.resultRun < int(w.spec.SuccessThreshold)) {
		// Success or failure is below threshold - leave the probe state unchanged.
		return true
	}
	if result == Failure {
		// The process fails a check, it will need to be restarted.
		// Stop probing and restart the process.
		w.log.Warn("Process failed probe, restarting", log.Any("key", key))
		err = w.svc.Restart()
		if err != nil {
			w.log.Error("Failed to restart process", log.Any("key", key), log.Error(err))
		}
		w.resultRun = 0
	}
	return true
}
