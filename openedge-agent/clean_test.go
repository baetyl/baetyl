package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/stretchr/testify/assert"
)

func Test_cleaner(t *testing.T) {
	target, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(target)
	prepareTest(t, target)
	type args struct {
		target  string
		volumes []openedge.VolumeInfo
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
				target:  target,
				volumes: []openedge.VolumeInfo{},
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
				target: target,
				volumes: []openedge.VolumeInfo{
					openedge.VolumeInfo{
						Path: "no-exists",
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "b"),
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "c", "c1"),
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "d", "d1", "d1i"),
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
				target: target,
				volumes: []openedge.VolumeInfo{
					openedge.VolumeInfo{
						Path: target,
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "c"),
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "d", "d1"),
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
			got, err := list(tt.args.target, tt.args.volumes)
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
	c := newCleaner(target, log)
	c.do("")
	assert.Equal(t, []string{"[Debugf]report version is empty"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Equal(t, []string{"[Debugf]last app config is not cached"}, log.records)
	log.records = []string{}
	c.reset(prepareConfig, openedge.VolumeInfo{Path: target})
	c.do("v2")
	assert.Equal(t, []string{"[Debugf]report version is not matched"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 0)
	c.do("v1")
	assert.Len(t, log.records, 0)
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.DirExists(t, filepath.Join(target, "c", "c2"))
	assert.FileExists(t, filepath.Join(target, "c", "c3"))
	assert.FileExists(t, filepath.Join(target, "d", "d1", "d1i", "d1i1"))
	c.do("v1")
	assert.Len(t, log.records, 2)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[1], "[Infof]end to clean,  elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.FileExists(t, filepath.Join(target, "c", "c3"))
	assert.False(t, utils.DirExists(filepath.Join(target, "c", "c2")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 0)

	appcfg, _, _ := c.reset(prepareConfig, openedge.VolumeInfo{Path: target})
	appcfg.Volumes = []openedge.VolumeInfo{}
	c.do("v1")
	assert.Len(t, log.records, 0)
	c.do("v1")
	assert.Len(t, log.records, 0)
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.FileExists(t, filepath.Join(target, "c", "c3"))
	assert.False(t, utils.DirExists(filepath.Join(target, "c", "c2")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
	c.do("v1")
	assert.Len(t, log.records, 2)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[1], "[Infof]end to clean,  elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.False(t, utils.DirExists(filepath.Join(target, "b")))
	assert.False(t, utils.DirExists(filepath.Join(target, "c")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
}

func Test_cleaner2(t *testing.T) {
	target := "var/db/openedge"
	err := os.MkdirAll(target, 0777)
	assert.NoError(t, err)
	defer os.RemoveAll("var")
	prepareTest(t, target)
	type args struct {
		target  string
		volumes []openedge.VolumeInfo
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
				target:  target,
				volumes: []openedge.VolumeInfo{},
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
				target: target,
				volumes: []openedge.VolumeInfo{
					openedge.VolumeInfo{
						Path: "no-exists",
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "b"),
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "c", "c1"),
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "d", "d1", "d1i"),
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
				target: target,
				volumes: []openedge.VolumeInfo{
					openedge.VolumeInfo{
						Path: target,
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "c"),
					},
					openedge.VolumeInfo{
						Path: filepath.Join(target, "d", "d1"),
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
			got, err := list(tt.args.target, tt.args.volumes)
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
	c := newCleaner(target, log)
	c.do("")
	assert.Equal(t, []string{"[Debugf]report version is empty"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Equal(t, []string{"[Debugf]last app config is not cached"}, log.records)
	log.records = []string{}
	c.reset(prepareConfig, openedge.VolumeInfo{Path: target})
	c.do("v2")
	assert.Equal(t, []string{"[Debugf]report version is not matched"}, log.records)
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 0)
	c.do("v1")
	assert.Len(t, log.records, 0)
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.DirExists(t, filepath.Join(target, "c", "c2"))
	assert.FileExists(t, filepath.Join(target, "d", "d1", "d1i", "d1i1"))
	c.do("v1")
	assert.Len(t, log.records, 2)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[1], "[Infof]end to clean,  elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.False(t, utils.DirExists(filepath.Join(target, "c", "c2")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
	log.records = []string{}
	c.do("v1")
	assert.Len(t, log.records, 0)

	appcfg, _, _ := c.reset(prepareConfig, openedge.VolumeInfo{Path: target})
	appcfg.Volumes = []openedge.VolumeInfo{}
	c.do("v1")
	assert.Len(t, log.records, 0)
	c.do("v1")
	assert.Len(t, log.records, 0)
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.FileExists(t, filepath.Join(target, "b", "b1"))
	assert.FileExists(t, filepath.Join(target, "c", "c1", "c1i"))
	assert.False(t, utils.DirExists(filepath.Join(target, "c", "c2")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
	c.do("v1")
	assert.Len(t, log.records, 2)
	assert.Equal(t, "[Infof]start to clean '"+target+"'", log.records[0])
	assert.Contains(t, log.records[1], "[Infof]end to clean,  elapsed time:")
	assert.FileExists(t, filepath.Join(target, "a"))
	assert.False(t, utils.DirExists(filepath.Join(target, "b")))
	assert.False(t, utils.DirExists(filepath.Join(target, "c")))
	assert.False(t, utils.DirExists(filepath.Join(target, "d")))
}

func prepareTest(t *testing.T, target string) {
	err := ioutil.WriteFile(filepath.Join(target, "a"), []byte{}, 0777)
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

func prepareConfig(v openedge.VolumeInfo) (*openedge.AppConfig, string, error) {
	return &openedge.AppConfig{
		Version: "v1",
		Volumes: []openedge.VolumeInfo{
			openedge.VolumeInfo{
				Path: "no-exists",
			},
			openedge.VolumeInfo{
				Path: filepath.Join(v.Path, "b"),
			},
			openedge.VolumeInfo{
				Path: filepath.Join(v.Path, "c", "c1"),
			},
		},
	}, "", nil
}
