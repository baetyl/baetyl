package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
)

func TestCleaner(t *testing.T) {
	prefix, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(prefix)
	target := prepareTest(t, prefix)
	type args struct {
		prefix  string
		target  string
		volumes map[string]baetyl.ComposeVolumeInfo
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "no volume",
			args: args{
				prefix:  prefix,
				target:  target,
				volumes: map[string]baetyl.ComposeVolumeInfo{},
			},
			want: []string{
				filepath.Join(target, "b"),
				filepath.Join(target, "c"),
				filepath.Join(target, "d"),
			},
			wantErr: false,
		},
		{
			name: "all volumes",
			args: args{
				prefix: prefix,
				target: target,
				volumes: map[string]baetyl.ComposeVolumeInfo{
					"no-exists": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": "no-exists",
						},
					},
					"b": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "b"),
						},
					},
					"cc1": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "c", "c1"),
						},
					},
					"dd1d1i": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "d", "d1", "d1i"),
						},
					},
				},
			},
			want: []string{
				filepath.Join(target, "c", "c2"),
			},
			wantErr: false,
		},
		{
			name: "target volumes",
			args: args{
				prefix: prefix,
				target: target,
				volumes: map[string]baetyl.ComposeVolumeInfo{
					target: baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": target,
						},
					},
					"c": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "c"),
						},
					},
					"dd1": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "d", "d1"),
						},
					},
				},
			},
			want: []string{
				filepath.Join(target, "b"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := list(tt.args.prefix, tt.args.target, tt.args.volumes)
			if (err != nil) != tt.wantErr {
				t.Errorf("list() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("list() = %v, want %v", got, tt.want)
			}
		})
	}

	// test cleaner
	log := newMockLogger()
	c := newCleaner(prefix, target, log)
	c.do("")
	assert.Equal(t, []string{"[Debugf]version () is ignored"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Equal(t, []string{"[Debugf]version (v1) is ignored"}, log.records)
	log.records = []string{}
	c.set("v1", prepareConfig(baetyl.ComposeVolumeInfo{
		DriverOpts: map[string]string{
			"device": prefix,
		},
	}).Volumes)
	c.do("v2")
	assert.Equal(t, []string{"[Debugf]version (v2) is ignored"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 4)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[3], "[Infof]end to clean, elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.FileExists(t, filepath.Join(target, "c", "c3"))
	assert.False(t, utils.DirExists(filepath.Join(target, "c", "c2")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 2)
	log.records = []string{}
	c.set("v1", prepareConfig(baetyl.ComposeVolumeInfo{
		DriverOpts: map[string]string{
			"device": target,
		},
	}).Volumes)
	c.do("v1")
	assert.Len(t, log.records, 4)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[3], "[Infof]end to clean, elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.False(t, utils.DirExists(filepath.Join(target, "b")))
	assert.False(t, utils.DirExists(filepath.Join(target, "c")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
}

func TestCleaner2(t *testing.T) {
	prefix := "var/db/baetyl"
	err := os.MkdirAll(prefix, 0777)
	assert.NoError(t, err)
	defer os.RemoveAll("var")
	target := prepareTest(t, prefix)
	type args struct {
		prefix  string
		target  string
		volumes map[string]baetyl.ComposeVolumeInfo
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "no volume",
			args: args{
				prefix:  prefix,
				target:  target,
				volumes: map[string]baetyl.ComposeVolumeInfo{},
			},
			want: []string{
				filepath.Join(target, "b"),
				filepath.Join(target, "c"),
				filepath.Join(target, "d"),
			},
			wantErr: false,
		},
		{
			name: "all volumes",
			args: args{
				prefix: prefix,
				target: target,
				volumes: map[string]baetyl.ComposeVolumeInfo{
					"no-exists": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": "no-exists",
						},
					},
					"b": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "b"),
						},
					},
					"cc1": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "c", "c1"),
						},
					},
					"dd1d1i": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "d", "d1", "d1i"),
						},
					},
				},
			},
			want: []string{
				filepath.Join(target, "c", "c2"),
			},
			wantErr: false,
		},
		{
			name: "target volumes",
			args: args{
				prefix: prefix,
				target: target,
				volumes: map[string]baetyl.ComposeVolumeInfo{
					prefix: baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": prefix,
						},
					},
					"c": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "c"),
						},
					},
					"dd1": baetyl.ComposeVolumeInfo{
						DriverOpts: map[string]string{
							"device": filepath.Join(prefix, "d", "d1"),
						},
					},
				},
			},
			want: []string{
				filepath.Join(target, "b"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := list(tt.args.prefix, tt.args.target, tt.args.volumes)
			if (err != nil) != tt.wantErr {
				t.Errorf("list() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("list() = %v, want %v", got, tt.want)
			}
		})
	}

	// test cleaner
	log := newMockLogger()
	c := newCleaner(prefix, target, log)
	c.do("")
	assert.Equal(t, []string{"[Debugf]version () is ignored"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Equal(t, []string{"[Debugf]version (v1) is ignored"}, log.records)
	log.records = []string{}
	c.set("v1", prepareConfig(baetyl.ComposeVolumeInfo{
		DriverOpts: map[string]string{
			"device": prefix,
		},
	}).Volumes)
	c.do("v2")
	assert.Equal(t, []string{"[Debugf]version (v2) is ignored"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 4)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[3], "[Infof]end to clean, elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.False(t, utils.DirExists(filepath.Join(target, "c", "c2")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 2)
	log.records = []string{}
	c.set("v1", prepareConfig(baetyl.ComposeVolumeInfo{}).Volumes)
	c.do("v1")
	assert.Len(t, log.records, 4)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[3], "[Infof]end to clean, elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.False(t, utils.DirExists(filepath.Join(target, "b")))
	assert.False(t, utils.DirExists(filepath.Join(target, "c")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
}

func prepareTest(t *testing.T, prefix string) string {
	target := filepath.Join(prefix, "volumes")
	err := os.MkdirAll(target, 0777)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(target, "a"), []byte{}, 0777)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(target, "b"), 0777)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(target, "b", "b1"), []byte{}, 0777)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(target, "c", "c1"), 0777)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(target, "c", "c2"), 0777)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(target, "c", "c3"), []byte{}, 0777)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(target, "c", "c1", "c1i"), []byte{}, 0777)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(target, "d", "d1", "d1i"), 0777)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(target, "d", "d1", "d1ii"), []byte{}, 0777)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(target, "d", "d1", "d1i", "d1i1"), []byte{}, 0777)
	assert.NoError(t, err)
	return target
}

type mackLogger struct {
	records []string
	data    map[string]interface{}
	err     error
}

func newMockLogger() *mackLogger {
	return &mackLogger{
		data:    map[string]interface{}{},
		records: []string{},
	}
}

func (l *mackLogger) WithField(key string, value interface{}) logger.Logger {
	l.data[key] = value
	return l
}
func (l *mackLogger) WithError(err error) logger.Logger {
	l.err = err
	return l
}
func (l *mackLogger) Debugf(format string, args ...interface{}) {
	l.records = append(l.records, "[Debugf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Infof(format string, args ...interface{}) {
	l.records = append(l.records, "[Infof]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Warnf(format string, args ...interface{}) {
	l.records = append(l.records, "[Warnf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Errorf(format string, args ...interface{}) {
	l.records = append(l.records, "[Errorf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Fatalf(format string, args ...interface{}) {
	l.records = append(l.records, "[Fatalf]"+fmt.Sprintf(format, args...))
}
func (l *mackLogger) Debugln(args ...interface{}) {
	l.records = append(l.records, "[Debugln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Infoln(args ...interface{}) {
	l.records = append(l.records, "[Infoln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Warnln(args ...interface{}) {
	l.records = append(l.records, "[Warnln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Errorln(args ...interface{}) {
	l.records = append(l.records, "[Errorln]"+fmt.Sprintln(args...))
}
func (l *mackLogger) Fatalln(args ...interface{}) {
	l.records = append(l.records, "[Fatalln]"+fmt.Sprintln(args...))
}

func prepareConfig(v baetyl.ComposeVolumeInfo) *baetyl.ComposeAppConfig {
	return &baetyl.ComposeAppConfig{
		Version: "v1",
		Volumes: map[string]baetyl.ComposeVolumeInfo{
			"no-exists": baetyl.ComposeVolumeInfo{
				DriverOpts: map[string]string{
					"device": "no-exists",
				},
			},
			"b": baetyl.ComposeVolumeInfo{
				DriverOpts: map[string]string{
					"device": filepath.Join(v.DriverOpts["device"], "b"),
				},
			},
			"cd1": baetyl.ComposeVolumeInfo{
				DriverOpts: map[string]string{
					"device": filepath.Join(v.DriverOpts["device"], "c", "c1"),
				},
			},
		},
	}
}
