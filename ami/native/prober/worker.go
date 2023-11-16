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

	"github.com/baetyl/baetyl/v2/utils"
)

type worker struct {
	// Channel for stopping the probe.
	stopCh chan struct{}
	// Describes the probe configuration (read-only)
	spec *v1.Probe
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

func newWorker(m *manager, svc service.Service, app *specV1.Application) *worker {
	return &worker{
		stopCh:       make(chan struct{}, 1),        // Buffer so stop() can be non-blocking.
		spec:         app.Services[0].LivenessProbe, // native only support one service
		app:          app,
		svc:          svc,
		probeManager: m,
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
		key := utils.MakeKey(specV1.KindApplication, w.app.Name, w.app.Version)
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

	key := utils.MakeKey(specV1.KindApplication, w.app.Name, w.app.Version)
	status, ok := w.svc.Status()
	if ok != nil || status == service.StatusUnknown {
		w.log.Debug("No status for process", log.Any("app", key))
		return true
	}
	if status == service.StatusStopped {
		w.log.Debug("Process is terminated, exiting probe worker", log.Any("app", key))
		return false
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
		// The process fails a liveness check, it will need to be restarted.
		// Stop probing and restart the process.
		w.log.Warn("Process failed liveness probe, restarting", log.Any("app", key))
		err = w.svc.Restart()
		if err != nil {
			w.log.Error("Failed to restart process", log.Any("app", key), log.Error(err))
		}
		w.resultRun = 0
	}
	return true
}
