package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__main_gen(test *testing.T) {
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
			name:  "Time.Now",
			input: "{\"timestamp\": {{.Time.NowUnix}}}",
		},
		{
			name:  "Time.Now",
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
