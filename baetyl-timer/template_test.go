package main

import (
	"fmt"
	"testing"

	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
)

func Test__template_gen(test *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "Time.Now",
			input: "{\"timestamp\": {{.Time.Now}}}",
		},
		{
			name:  "Time.NowUnix",
			input: "{\"timestamp\": {{.Time.NowUnix}}}",
		},
		{
			name:  "Time.NowUnixNano",
			input: "{\"timestamp\": {{.Time.NowUnixNano}}}",
		},
		{
			name:  "Rand.Int",
			input: "{\"Rand.Int\": {{.Rand.Int}}}",
		},
		{
			name:  "Rand.Int63",
			input: "{\"Rand.Int63\": {{.Rand.Int63}}}",
		},
		{
			name:  "Rand.Intn",
			input: "{\"Rand.Intn\": {{.Rand.Intn 10}}}",
		},
		{
			name:  "Rand.Float64",
			input: "{\"Rand.Float64\": {{.Rand.Float64}}}",
		},
		{
			name:  "Rand.Float64n",
			input: "{\"Rand.Float64n\": {{.Rand.Float64n 60}}}",
		},
		{
			name:  "anyString",
			input: "{\"anyString\": \"inputString\"}",
		},
		{
			name:  "allInOne",
			input: "{\"datetime\": {{.Time.Now}},\"timestamp\": {{.Time.NowUnix}},\"timestampNano\": {{.Time.NowUnixNano}},\"random1\": {{.Rand.Int}},\"random2\": {{.Rand.Int63}},\"random3\": {{.Rand.Intn 10}},\"random4\": {{.Rand.Float64}},\"random5\": {{.Rand.Float64n 60}},\"anyString\": \"inputString\"}",
		},
	}
	for _, t := range tests {
		test.Run(t.name, func(testOne *testing.T) {
			_template, err := newTemplate(t.input)
			assert.NoError(testOne, err)
			pld, err := _template.gen()
			assert.NoError(testOne, err)
			fmt.Println(string(pld))
			pld, err = _template.gen()
			assert.NoError(testOne, err)
			fmt.Println(string(pld))
		})
	}
}

func Test__template_gen_old_config(t *testing.T) {
	cfgData := `
publish:
  topic: t
  payload:
    id: 1
`
	var cfg config
	err := utils.UnmarshalYAML([]byte(cfgData), &cfg)
	assert.NoError(t, err)
	_template, err := newTemplate(string(cfg.Publish.Payload))
	assert.NoError(t, err)
	pld, err := _template.gen()
	assert.NoError(t, err)
	fmt.Println(string(pld))
	pld, err = _template.gen()
	assert.NoError(t, err)
	fmt.Println(string(pld))
}
