package main

import (
	"bytes"
	"math/rand"
	"text/template"
	"time"
)

type _data struct {
	Time *_time
	Rand *_rand
}

type _rand struct {
	*rand.Rand
}

func (r *_rand) Float64n(n float64) float64 {
	return n * r.Float64()
}

type _time struct{}

func (t *_time) Now() time.Time {
	return time.Now()
}

func (t *_time) NowUnix() int64 {
	return time.Now().Unix()
}

func (t *_time) NowUnixNano() int64 {
	return time.Now().UnixNano()
}

type _template struct {
	*template.Template
	data *_data
}

func newTemplate(t string) (*_template, error) {
	rand.Seed(time.Now().UnixNano())
	temp := template.New("dummy")
	temp, err := temp.Parse(t)
	if err != nil {
		return nil, err
	}
	return &_template{
		Template: temp,
		data: &_data{
			Time: new(_time),
			Rand: &_rand{Rand: rand.New(rand.NewSource(time.Now().UnixNano()))},
		},
	}, nil
}

func (t *_template) gen() ([]byte, error) {
	kk := new(bytes.Buffer)
	err := t.Execute(kk, t.data)
	if err != nil {
		return nil, err
	}
	return kk.Bytes(), nil
}
