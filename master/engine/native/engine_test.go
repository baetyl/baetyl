package native

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mount(t *testing.T) {
	epwd, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(epwd)
	srcdir := path.Join(epwd, "srcdir")
	srcfile := path.Join(epwd, "srcfile")
	srcnone := path.Join(epwd, "srcnone")
	srcerr := path.Join(epwd, "srcfile/srcnone")
	err = os.MkdirAll(srcdir, 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(srcfile, []byte(""), 0700)
	assert.NoError(t, err)
	dstdir := path.Join(epwd, "dstdir")
	dstfile := path.Join(epwd, "dstfile")
	dstnone := path.Join(epwd, "dstnone")
	dsterr := path.Join(epwd, "dstfile/dstnone")

	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "file",
			args: args{
				src: srcfile,
				dst: dstfile,
			},
		},
		{
			name: "dir1",
			args: args{
				src: srcdir,
				dst: dstdir,
			},
		},
		{
			name: "dir2",
			args: args{
				src: srcnone,
				dst: dstnone,
			},
		},
		{
			name: "err1",
			args: args{
				src: srcerr,
				dst: dsterr,
			},
			wantErr: true,
		},
		{
			name: "err2",
			args: args{
				src: srcfile,
				dst: srcfile,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mount(tt.args.src, tt.args.dst)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("mount() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				wfile := tt.args.dst
				rfile := tt.args.src
				if strings.HasPrefix(tt.name, "dir") {
					wfile = path.Join(wfile, "tmpfile")
					rfile = path.Join(rfile, "tmpfile")
				}
				err = ioutil.WriteFile(wfile, []byte(epwd), 0700)
				assert.NoError(t, err)
				out, err := ioutil.ReadFile(rfile)
				assert.NoError(t, err)
				assert.Equal(t, []byte(epwd), out)
			}
		})
	}
}
