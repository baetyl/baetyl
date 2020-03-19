package main

import (
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"
)

const (
	templateActive = `
<!DOCTYPE html>
<head>
<meta charset="utf8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style type="text/css">
    div.active-title {font-family:Menlo; height: 50px; margin-top:12%; vertical-align: middle; text-align:center}
    div.active-input {font-family:SimHei; margin:0 auto; width: 30%; height: 150px; margin-top:40px}
    label.active-lable {float: left; font-size: 17px; margin-top: 5px}
    input.active-box{float:right; width: 200px; height: 18px; margin-top: 4px}
    input.active-submit{margin-top:25px; float:right; width: 100px; height: 30px; font-size: 30px}
</style>
<title>激活界面</title>
</head>
<body bgcolor="#f6f7fb">
<form method="post" action="update">
    <div class="active-title">
        <h1>设备激活</h1>
    </div>
    <div class="active-input">
        {{range $value := .Attributes }}
            <label class="active-lable">{{$value.Label}}：</label>
            <input class="active-box" name="{{$value.Name}}" placeholder="{{$value.Desc}}" value=""/>
            <br/><br/>
        {{ end }}
        <input class="active-submit" type="submit" value="激活"/>
    </div>
</form>
</body>
`
	templateFailed = `
<!DOCTYPE html>
<head>
    <meta charset="utf8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="refresh" content="3; url=/">
    <title>激活失败</title>
    <style type="text/css">
        div.active-title {font-family:Menlo; height: 50px; margin-top:20%; vertical-align: middle; text-align:center}
    </style>
</head>
<body bgcolor="#f6f7fb">
<div class="active-title">
    <h3>激活失败，请重新输入</h3>
</div>
</body>
`
	templateSuccess = `
<!DOCTYPE html>
<head>
    <meta charset="utf8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>激活成功</title>
    <style type="text/css">
        div.active-title {font-family:Menlo; height: 50px; margin-top:20%; vertical-align: middle; text-align:center}
    </style>
</head>
<body bgcolor="#f6f7fb">
<div class="active-title">
    <h3>激活成功</h3>
</div>
</body>
`
)

var (
	attrs = map[string]string{
		"batch":            "b1",
		"namespace":        "default",
		"fingerprintValue": "123",
	}
)

func initTemplate(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "pages")
	assert.Nil(t, err)
	activePath := path.Join(tmpDir, "active.html.template")
	err = ioutil.WriteFile(activePath, []byte(templateActive), 0755)
	assert.Nil(t, err)
	failedPath := path.Join(tmpDir, "filed.html.template")
	err = ioutil.WriteFile(failedPath, []byte(templateFailed), 0755)
	assert.Nil(t, err)
	successPath := path.Join(tmpDir, "success.html.template")
	err = ioutil.WriteFile(successPath, []byte(templateSuccess), 0755)
	return tmpDir
}

func TestServer(t *testing.T) {
	tmpDir := initTemplate(t)
	defer os.RemoveAll(tmpDir)
	a := initAgent(t)
	a.cfg.Server.Listen = "0.0.0.0:16868"
	a.cfg.Server.Pages = tmpDir
	a.cfg.Attributes = []config.Attribute{
		{Name: "batch"},
		{Name: "namespace"},
		{Name: "fingerprintValue"},
	}
	err := a.NewServer(a.cfg.Server, a.ctx.Log())
	assert.Nil(t, err)
	a.CloseServer()

	w := &httptest.ResponseRecorder{
		Code:    http.StatusOK,
		Body:    nil,
		Flushed: false,
	}
	a.handleView(w, nil)
	form := url.Values{}
	for k, v := range attrs {
		form.Set(k, v)
	}
	req := &http.Request{
		Method:   http.MethodPost,
		Form:     form,
		PostForm: form,
	}
	a.handleUpdate(w, req)
	assert.Equal(t, attrs, a.attrs)
}
