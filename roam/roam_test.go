// Package roam 大文件同步模块，用于集群环境部署下大文件在集群中的传播
package roam

import (
	_ "github.com/baetyl/baetyl/v2/ami/native"
)

//func TestNewRoam(t *testing.T) {
//	// bad case
//	cfg := config.Config{}
//	r1, err := NewRoam(cfg)
//	assert.Error(t, err)
//	assert.Nil(t, r1)
//
//	// good case
//	err = os.Setenv(context.KeyRunMode, context.RunModeNative)
//	assert.NoError(t, err)
//	cfg.AMI.Native.PortsRange.Start = 40000
//	cfg.AMI.Native.PortsRange.End = 50000
//	r2, err := NewRoam(cfg)
//	assert.NoError(t, err)
//	assert.NotNil(t, r2)
//}
//
//func TestRoamImpl_RoamObject(t *testing.T) {
//	wg := sync.WaitGroup{}
//
//	dir := os.TempDir()
//	file, err := ioutil.TempFile("", "")
//	assert.NoError(t, err)
//	txt := "object"
//	_, err = file.WriteString(txt)
//	assert.NoError(t, err)
//	md5 := "md5"
//	unpack := "unpack"
//	// mock http server
//	ms := httptest.NewUnstartedServer(gohttp.HandlerFunc(func(writer gohttp.ResponseWriter, request *gohttp.Request) {
//		receUnpack := request.Header.Get(specv1.HeaderKeyObjectUnpack)
//		receMD5 := request.Header.Get(specv1.HeaderKeyObjectMD5)
//		receDir := request.Header.Get(specv1.HeaderKeyObjectDir)
//		assert.Equal(t, unpack, receUnpack)
//		assert.Equal(t, md5, receMD5)
//		assert.Equal(t, dir, receDir)
//
//		reader, er := request.MultipartReader()
//		assert.NoError(t, er)
//		assert.NotNil(t, reader)
//
//		dst, er := ioutil.TempFile("", "")
//		assert.NoError(t, er)
//
//		defer os.Remove(dst.Name())
//		for {
//			var p *multipart.Part
//			p, er = reader.NextPart()
//			if er == io.EOF {
//				break
//			}
//			assert.NoError(t, er)
//			if p.FileName() == "" {
//				_, er = ioutil.ReadAll(p)
//				assert.NoError(t, er)
//			} else {
//				// begin download file
//				er = dst.Truncate(0)
//				assert.NoError(t, er)
//
//				_, er = io.Copy(dst, p)
//				assert.NoError(t, er)
//			}
//		}
//		res, er := ioutil.ReadFile(dst.Name())
//		assert.NoError(t, er)
//		assert.Equal(t, txt, string(res))
//
//		writer.WriteHeader(gohttp.StatusOK)
//		wg.Done()
//	}))
//	ms.Start()
//	defer ms.Close()
//	fmt.Println(ms.URL)
//
//	// gen config
//	cfg := config.Config{}
//	cfg.Roam.Path = "/v1/object"
//	cfg.Roam.Timeout = 10 * time.Minute
//	cfg.Roam.InsecureSkipVerify = true
//
//	// mock ami
//	ctl := gomock.NewController(t)
//	defer ctl.Finish()
//	am := mc.NewMockAMI(ctl)
//
//	ops, err := cfg.Roam.ToClientOptions()
//	assert.NoError(t, err)
//
//	r := &roamImpl{
//		cfg:  cfg,
//		mode: "kube",
//		ami:  am,
//		cli:  http.NewClient(ops),
//		wg:   &sync.WaitGroup{},
//		mx:   &sync.Mutex{},
//		log:  log.L().With(log.Any("test", "roam")),
//	}
//
//	// bad case
//	am.EXPECT().StatsApps(context.EdgeSystemNamespace()).Return(nil, os.ErrInvalid).Times(1)
//	err = r.RoamObject(dir, file.Name(), md5, unpack)
//	assert.Error(t, err)
//
//	// good case with no md5
//	stats := []specv1.AppStats{
//		{
//			AppInfo: specv1.AppInfo{
//				Name:    "baetyl-agent-bnmdq",
//				Version: "18795958761",
//			},
//			DeployType: specv1.WorkloadDaemonSet,
//			InstanceStats: map[string]specv1.InstanceStats{
//				"baetyl-agent-faguang": {
//					Name:     "baetyl-agent-faguang",
//					IP:       strings.TrimPrefix(ms.URL, "http://"),
//					NodeName: "zb01",
//				},
//				"baetyl-agent-dahuang": {
//					Name:     "baetyl-agent-dahuang",
//					IP:       strings.TrimPrefix(ms.URL, "http://"),
//					NodeName: "zb02",
//				},
//			},
//		},
//		{
//			AppInfo: specv1.AppInfo{
//				Name:    "baetyl-init-ewnu834",
//				Version: "12389743279",
//			},
//		},
//	}
//	wg.Add(2)
//	am.EXPECT().StatsApps(context.EdgeSystemNamespace()).Return(stats, nil).Times(1)
//	err = r.RoamObject(dir, file.Name(), md5, unpack)
//	assert.NoError(t, err)
//	wg.Wait()
//}
