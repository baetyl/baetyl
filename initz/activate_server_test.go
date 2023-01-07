package initz

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
	mc "github.com/baetyl/baetyl/v2/mock"
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
	tmpDir := t.TempDir()
	activePath := path.Join(tmpDir, "active.html.template")
	err := os.WriteFile(activePath, []byte(templateActive), 0755)
	assert.Nil(t, err)
	failedPath := path.Join(tmpDir, "filed.html.template")
	err = os.WriteFile(failedPath, []byte(templateFailed), 0755)
	assert.Nil(t, err)
	successPath := path.Join(tmpDir, "success.html.template")
	err = os.WriteFile(successPath, []byte(templateSuccess), 0755)
	return tmpDir
}

func TestActivate_Server(t *testing.T) {
	initTemplate(t)
	c := &config.Config{}
	c.Init.Active.Collector.Server.Listen = "www.baidu.com"
	c.Init.Active.Collector.Attributes = []config.Attribute{
		{Name: "batch"},
		{Name: "namespace"},
		{Name: "fingerprintValue"},
	}

	active, err := NewActivate(c)
	assert.Error(t, err)

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	ami := mc.NewMockAMI(mockCtl)
	active = genActivate(t, c, ami)

	active.Start()

	w := &httptest.ResponseRecorder{
		Code:    http.StatusOK,
		Body:    nil,
		Flushed: false,
	}
	active.handleView(w, nil)
	form := url.Values{}
	for k, v := range attrs {
		form.Set(k, v)
	}

	req := &http.Request{
		Method:   http.MethodPost,
		Form:     form,
		PostForm: form,
	}
	active.handleUpdate(w, req)
	assert.Equal(t, attrs, active.attrs)

	req = &http.Request{
		Method: http.MethodGet,
	}
	active.handleUpdate(w, req)
	req = &http.Request{
		Method: http.MethodPost,
		Form:   nil,
	}
	active.handleUpdate(w, req)
}
