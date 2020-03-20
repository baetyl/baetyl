package baetyl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	engine := new(mockEngine)
	f := func(is InfoStats, opts Options) (Engine, error) {
		return engine, nil
	}
	_, err := New("unknown", nil, Options{})
	assert.Error(t, err)
	assert.Equal(t, "no such engine", err.Error())
	fac := Factories()
	fac["test"] = f
	e, err := New("test", nil, Options{})
	assert.NoError(t, err)
	assert.Equal(t, engine, e)
}

type mockEngine struct{}

func (*mockEngine) Name() string {
	return ""
}

func (*mockEngine) Recover() {
	return
}

func (*mockEngine) Prepare(ComposeAppConfig) {
	return
}

func (*mockEngine) SetInstanceStats(serviceName, instanceName string, partialStats PartialStats, persist bool) {
	return
}

func (*mockEngine) DelInstanceStats(serviceName, instanceName string, persist bool) {
	return
}

func (*mockEngine) DelServiceStats(serviceName string, persist bool) {
	return
}

func (*mockEngine) Run(string, ComposeService, map[string]ComposeVolume) (Service, error) {
	return nil, nil
}

func (*mockEngine) Close() error {
	return nil
}
