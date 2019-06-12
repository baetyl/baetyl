package main

import (
	"encoding/json"
	"fmt"
	"github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"syscall"
	"testing"
	"time"
)

type processConfigs struct {
	exec string
	pwd  string
	argv []string
	env  []string
}

func TestInstance_Python36(t *testing.T) {
	tmpPath := "tmp" + uuid.Generate().String()
	instanceName := "function-sayhi"
	functionName := "python36-sayhi"
	confTargetPath := path.Join(tmpPath, "etc", "openedge")
	codeTargetPath := path.Join(tmpPath, "var", "db", "openedge", instanceName)
	confSourcePath := path.Join("..", "example", "native", "var", "db", "openedge", "function-sayhi-conf")
	codeSourcePath := path.Join("..", "example", "native", "var", "db", "openedge", "function-sayhi-code")
	confFileName := "service.yml"
	codeFileName := "index.py"

	os.MkdirAll(confTargetPath, os.ModePerm)
	os.MkdirAll(codeTargetPath, os.ModePerm)
	utils.CopyFile(path.Join(confSourcePath, confFileName), path.Join(confTargetPath, confFileName))
	utils.CopyFile(path.Join(codeSourcePath, codeFileName), path.Join(codeTargetPath, codeFileName))
	defer os.RemoveAll(tmpPath)

	env := []string{}
	env = os.Environ()
	hostName := "127.0.0.1"
	port, err := utils.GetAvailablePort(hostName)
	assert.NoError(t, err)

	address := fmt.Sprintf("%s:%d", hostName, port)
	env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_ADDRESS", address))
	env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_NAME", instanceName))
	pwd, err := os.Getwd()
	assert.NoError(t, err)

	params := processConfigs{
		exec: path.Join(pwd, "..", "openedge-function-python", "openedge-function-python36.py"),
		env:  env,
		pwd: path.Join(pwd, tmpPath),
	}

	os.Chmod(params.exec, os.ModePerm)
	p, err := os.StartProcess(
		params.exec,
		params.argv,
		&os.ProcAttr{
			Dir: params.pwd,
			Env: params.env,
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		},
	)
	assert.NoError(t, err)

	fcc := openedge.FunctionClientConfig{}
	fcc.Address = address
	fcc.Message.Length.Max = 4194304
	fcc.Timeout = 30 * time.Second
	fcc.Backoff.Max = 30 * time.Second
	cli, err := openedge.NewFClient(fcc)
	assert.NoError(t, err)

	var msgId uint64 = 1234
	var msgQOS uint32 = 0
	var msgTopic string = "t"
	var msgPayload string = "OpenEdge Project"
	var msgFunctionName string = functionName
	var msgFunctionInvokeID string = uuid.Generate().String()
	msg := &openedge.FunctionMessage{
		ID:               msgId,
		QOS:              msgQOS,
		Topic:            msgTopic,
		Payload:          []byte(msgPayload),
		FunctionName:     msgFunctionName,
		FunctionInvokeID: msgFunctionInvokeID,
	}

	out, err := cli.Call(msg)
	assert.NoError(t, err)

	dataArr := map[string]interface{}{}
	err = json.Unmarshal([]byte(out.Payload), &dataArr)
	assert.NoError(t, err)
	assert.Equal(t, dataArr["messageTopic"], msgTopic)
	assert.Equal(t, dataArr["functionName"], msgFunctionName)
	assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
	assert.Equal(t, dataArr["bytes"], msgPayload)
	assert.Equal(t, dataArr["Say"], "Hello OpenEdge")

	msgPayload = "{\"Project\":\"OpenEdge\"}"
	msg = &openedge.FunctionMessage{
		ID:               msgId,
		QOS:              msgQOS,
		Topic:            msgTopic,
		Payload:          []byte(msgPayload),
		FunctionName:     msgFunctionName,
		FunctionInvokeID: msgFunctionInvokeID,
	}

	out, err = cli.Call(msg)
	assert.NoError(t, err)

	dataArr = map[string]interface{}{}
	err = json.Unmarshal([]byte(out.Payload), &dataArr)
	assert.NoError(t, err)
	assert.Equal(t, dataArr["messageTopic"], msgTopic)
	assert.Equal(t, dataArr["functionName"], msgFunctionName)
	assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
	assert.Equal(t, dataArr["Project"], "OpenEdge")
	assert.Equal(t, dataArr["Say"], "Hello OpenEdge")

	err = p.Signal(syscall.SIGTERM)
	assert.NoError(t, err)
}


func TestInstance_Python27(t *testing.T) {
	tmpPath := "tmp" + uuid.Generate().String()
	instanceName := "function-sayhi"
	functionName := "python27-sayhi"
	confTargetPath := path.Join(tmpPath, "etc", "openedge")
	codeTargetPath := path.Join(tmpPath, "var", "db", "openedge", instanceName)
	confSourcePath := path.Join("..", "example", "native", "var", "db", "openedge", "function-sayhi-conf")
	codeSourcePath := path.Join("..", "example", "native", "var", "db", "openedge", "function-sayhi-code")
	confFileName := "service.yml"
	codeFileName := "index.py"

	os.MkdirAll(confTargetPath, os.ModePerm)
	os.MkdirAll(codeTargetPath, os.ModePerm)
	utils.CopyFile(path.Join(confSourcePath, confFileName), path.Join(confTargetPath, confFileName))
	utils.CopyFile(path.Join(codeSourcePath, codeFileName), path.Join(codeTargetPath, codeFileName))
	defer os.RemoveAll(tmpPath)

	env := []string{}
	env = os.Environ()
	hostName := "127.0.0.1"
	port, err := utils.GetAvailablePort(hostName)
	assert.NoError(t, err)

	address := fmt.Sprintf("%s:%d", hostName, port)
	env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_ADDRESS", address))
	env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_NAME", instanceName))
	pwd, err := os.Getwd()
	assert.NoError(t, err)

	params := processConfigs{
		exec: path.Join(pwd, "..", "openedge-function-python", "openedge-function-python27.py"),
		env:  env,
		pwd: path.Join(pwd, tmpPath),
	}

	os.Chmod(params.exec, os.ModePerm)
	p, err := os.StartProcess(
		params.exec,
		params.argv,
		&os.ProcAttr{
			Dir: params.pwd,
			Env: params.env,
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		},
	)
	assert.NoError(t, err)

	fcc := openedge.FunctionClientConfig{}
	fcc.Address = address
	fcc.Message.Length.Max = 4194304
	fcc.Timeout = 30 * time.Second
	fcc.Backoff.Max = 30 * time.Second
	cli, err := openedge.NewFClient(fcc)
	assert.NoError(t, err)

	var msgId uint64 = 1234
	var msgQOS uint32 = 0
	var msgTopic string = "t"
	var msgPayload string = "OpenEdge Project"
	var msgFunctionName string = functionName
	var msgFunctionInvokeID string = uuid.Generate().String()
	msg := &openedge.FunctionMessage{
		ID:               msgId,
		QOS:              msgQOS,
		Topic:            msgTopic,
		Payload:          []byte(msgPayload),
		FunctionName:     msgFunctionName,
		FunctionInvokeID: msgFunctionInvokeID,
	}

	out, err := cli.Call(msg)
	assert.NoError(t, err)

	dataArr := map[string]interface{}{}
	err = json.Unmarshal([]byte(out.Payload), &dataArr)
	assert.NoError(t, err)
	assert.Equal(t, dataArr["messageTopic"], msgTopic)
	assert.Equal(t, dataArr["functionName"], msgFunctionName)
	assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
	assert.Equal(t, dataArr["bytes"], msgPayload)
	assert.Equal(t, dataArr["Say"], "Hello OpenEdge")

	msgPayload = "{\"Project\":\"OpenEdge\"}"
	msg = &openedge.FunctionMessage{
		ID:               msgId,
		QOS:              msgQOS,
		Topic:            msgTopic,
		Payload:          []byte(msgPayload),
		FunctionName:     msgFunctionName,
		FunctionInvokeID: msgFunctionInvokeID,
	}

	out, err = cli.Call(msg)
	assert.NoError(t, err)

	dataArr = map[string]interface{}{}
	err = json.Unmarshal([]byte(out.Payload), &dataArr)
	assert.NoError(t, err)
	assert.Equal(t, dataArr["messageTopic"], msgTopic)
	assert.Equal(t, dataArr["functionName"], msgFunctionName)
	assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
	assert.Equal(t, dataArr["Project"], "OpenEdge")
	assert.Equal(t, dataArr["Say"], "Hello OpenEdge")

	err = p.Signal(syscall.SIGTERM)
	assert.NoError(t, err)
}


func TestInstance_Node85(t *testing.T) {
	tmpPath := "tmp" + uuid.Generate().String()
	instanceName := "function-sayhi"
	functionName := "node85-sayhi"
	confTargetPath := path.Join(tmpPath, "etc", "openedge")
	codeTargetPath := path.Join(tmpPath, "var", "db", "openedge", instanceName)
	confSourcePath := path.Join("..", "example", "native", "var", "db", "openedge", "function-sayjs-conf")
	codeSourcePath := path.Join("..", "example", "native", "var", "db", "openedge", "function-sayjs-code")
	confFileName := "service.yml"
	codeFileName := "index.js"

	os.MkdirAll(confTargetPath, os.ModePerm)
	os.MkdirAll(codeTargetPath, os.ModePerm)
	utils.CopyFile(path.Join(confSourcePath, confFileName), path.Join(confTargetPath, confFileName))
	utils.CopyFile(path.Join(codeSourcePath, codeFileName), path.Join(codeTargetPath, codeFileName))
	defer os.RemoveAll(tmpPath)

	env := []string{}
	env = os.Environ()
	hostName := "127.0.0.1"
	port, err := utils.GetAvailablePort(hostName)
	assert.NoError(t, err)

	address := fmt.Sprintf("%s:%d", hostName, port)
	env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_ADDRESS", address))
	env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_NAME", instanceName))
	pwd, err := os.Getwd()
	assert.NoError(t, err)

	params := processConfigs{
		exec: path.Join(pwd, "..", "openedge-function-node", "openedge-function-node85.js"),
		env:  env,
		pwd: path.Join(pwd, tmpPath),
	}

	os.Chmod(params.exec, os.ModePerm)
	p, err := os.StartProcess(
		params.exec,
		params.argv,
		&os.ProcAttr{
			Dir: params.pwd,
			Env: params.env,
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		},
	)
	assert.NoError(t, err)

	fcc := openedge.FunctionClientConfig{}
	fcc.Address = address
	fcc.Message.Length.Max = 4194304
	fcc.Timeout = 30 * time.Second
	fcc.Backoff.Max = 30 * time.Second
	cli, err := openedge.NewFClient(fcc)
	assert.NoError(t, err)

	var msgId uint64 = 1234
	var msgQOS uint32 = 0
	var msgTopic string = "t"
	var msgPayload string = "OpenEdge Project"
	var msgFunctionName string = functionName
	var msgFunctionInvokeID string = uuid.Generate().String()
	msg := &openedge.FunctionMessage{
		ID:               msgId,
		QOS:              msgQOS,
		Topic:            msgTopic,
		Payload:          []byte(msgPayload),
		FunctionName:     msgFunctionName,
		FunctionInvokeID: msgFunctionInvokeID,
	}
	out, err := cli.Call(msg)
	assert.NoError(t, err)

	dataArr := map[string]interface{}{}
	err = json.Unmarshal([]byte(out.Payload), &dataArr)
	assert.NoError(t, err)
	assert.Equal(t, dataArr["messageTopic"], msgTopic)
	assert.Equal(t, dataArr["functionName"], msgFunctionName)
	assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
	assert.Equal(t, dataArr["bytes"], msgPayload)
	assert.Equal(t, dataArr["Say"], "Hello OpenEdge")

	msgPayload = "{\"Project\":\"OpenEdge\"}"
	msg = &openedge.FunctionMessage{
		ID:               msgId,
		QOS:              msgQOS,
		Topic:            msgTopic,
		Payload:          []byte(msgPayload),
		FunctionName:     msgFunctionName,
		FunctionInvokeID: msgFunctionInvokeID,
	}
	out, err = cli.Call(msg)
	assert.NoError(t, err)

	dataArr = map[string]interface{}{}
	err = json.Unmarshal([]byte(out.Payload), &dataArr)
	assert.NoError(t, err)
	assert.Equal(t, dataArr["messageTopic"], msgTopic)
	assert.Equal(t, dataArr["functionName"], msgFunctionName)
	assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
	assert.Equal(t, dataArr["Project"], "OpenEdge")
	assert.Equal(t, dataArr["Say"], "Hello OpenEdge")

	err = p.Signal(syscall.SIGTERM)
	assert.NoError(t, err)
}
