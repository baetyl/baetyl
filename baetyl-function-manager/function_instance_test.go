package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_FunctionInstance(t *testing.T) {
	tests := []struct {
		name         string
		exec         string
		functionName string
		workPath     string
		filePath     []string
	}{
		{
			name:         "test python3 runtime",
			exec:         "python3",
			functionName: "python36-sayhi",
			workPath:     "testrun/python",
			filePath:     []string{"..", "..", "..", "baetyl-function-python", "baetyl-function-python36.py"},
		},
		{
			name:         "test node8 runtime",
			exec:         "node",
			functionName: "node85-sayhi",
			workPath:     "testrun/node",
			filePath:     []string{"..", "..", "..", "baetyl-function-node", "baetyl-function-node85.js"},
		},
		{
			name:         "test python2.7 runtime",
			exec:         "python2.7",
			functionName: "python27-sayhi",
			workPath:     "testrun/python",
			filePath:     []string{"..", "..", "..", "baetyl-function-python", "baetyl-function-python27.py"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec, err := exec.LookPath(tt.exec)
			if err != nil {
				t.Skip("need " + tt.exec)
			}

			instanceName := "function-sayhi"
			functionName := tt.functionName
			err = os.Chdir(tt.workPath)
			assert.NoError(t, err)
			defer os.Chdir("../..")

			hostName := "127.0.0.1"
			port, err := utils.GetAvailablePort(hostName)
			assert.NoError(t, err)
			address := fmt.Sprintf("%s:%d", hostName, port)

			env := os.Environ()
			env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_ADDRESS", address))
			env = append(env, fmt.Sprintf("%s=%s", "OPENEDGE_SERVICE_INSTANCE_NAME", instanceName))
			filePath := path.Join(tt.filePath...)

			p, err := os.StartProcess(
				exec,
				[]string{"python3", filePath},
				&os.ProcAttr{
					Env: env,
					Files: []*os.File{
						os.Stdin,
						os.Stdout,
						os.Stderr,
					},
				},
			)
			assert.NoError(t, err)

			fcc := baetyl.FunctionClientConfig{}
			err = utils.UnmarshalJSON([]byte("{\"address\":\""+address+"\"}"), &fcc)
			assert.NoError(t, err)
			cli, err := baetyl.NewFClient(fcc)
			assert.NoError(t, err)

			// round 1: test binary payload
			msgID := uint64(1234)
			msgQOS := uint32(0)
			msgTopic := "t"
			msgPayload := "Baetyl Project"
			msgFunctionName := functionName
			msgFunctionInvokeID := uuid.Generate().String()
			msgTimestamp := time.Now().Unix()
			msg := &baetyl.FunctionMessage{
				ID:               msgID,
				QOS:              msgQOS,
				Topic:            msgTopic,
				Payload:          []byte(msgPayload),
				FunctionName:     msgFunctionName,
				FunctionInvokeID: msgFunctionInvokeID,
				Timestamp:        msgTimestamp,
			}

			out, err := cli.Call(msg)
			assert.NoError(t, err)

			dataArr := map[string]interface{}{}
			err = json.Unmarshal(out.Payload, &dataArr)
			assert.NoError(t, err)
			assert.Equal(t, len(dataArr), 7)
			assert.Equal(t, uint32(dataArr["messageQOS"].(float64)), msgQOS)
			assert.Equal(t, dataArr["messageTopic"], msgTopic)
			assert.Equal(t, dataArr["functionName"], msgFunctionName)
			assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
			assert.Equal(t, int64(dataArr["messageTimestamp"].(float64)), msgTimestamp)
			assert.Equal(t, dataArr["bytes"], msgPayload)
			assert.Equal(t, dataArr["Say"], "Hello Baetyl")

			// round 2: test json payload
			msg.Payload = []byte("{\"Project\":\"Baetyl\"}")
			out, err = cli.Call(msg)
			assert.NoError(t, err)

			dataArr = map[string]interface{}{}
			err = json.Unmarshal(out.Payload, &dataArr)
			assert.NoError(t, err)
			assert.Equal(t, len(dataArr), 7)
			assert.Equal(t, uint32(dataArr["messageQOS"].(float64)), msgQOS)
			assert.Equal(t, dataArr["messageTopic"], msgTopic)
			assert.Equal(t, dataArr["functionName"], msgFunctionName)
			assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
			assert.Equal(t, int64(dataArr["messageTimestamp"].(float64)), msgTimestamp)
			assert.Equal(t, dataArr["Project"], "Baetyl")
			assert.Equal(t, dataArr["Say"], "Hello Baetyl")

			// round 3: test empty payload
			msg.Payload = []byte("")
			out, err = cli.Call(msg)
			assert.NoError(t, err)

			dataArr = map[string]interface{}{}
			err = json.Unmarshal(out.Payload, &dataArr)
			assert.NoError(t, err)
			assert.Equal(t, len(dataArr), 6)
			assert.Equal(t, uint32(dataArr["messageQOS"].(float64)), msgQOS)
			assert.Equal(t, dataArr["messageTopic"], msgTopic)
			assert.Equal(t, dataArr["functionName"], msgFunctionName)
			assert.Equal(t, dataArr["functionInvokeID"], msgFunctionInvokeID)
			assert.Equal(t, int64(dataArr["messageTimestamp"].(float64)), msgTimestamp)
			assert.Equal(t, dataArr["Say"], "Hello Baetyl")

			// round 4: test error
			msg.Payload = []byte("{\"err\":\"Baetyl\"}")
			out, err = cli.Call(msg)
			assert.Error(t, err)

			// round 5: function not exist
			msg.FunctionName = "xxx"
			out, err = cli.Call(msg)
			assert.Error(t, err)

			err = p.Signal(syscall.SIGTERM)
			assert.NoError(t, err)
			p.Wait()

			b, err := ioutil.ReadFile(path.Join("var", "log", "baetyl", "service.log"))
			assert.NoError(t, err)
			logInfo := string(b)
			assert.True(t, strings.Contains(logInfo, "service starting"))
			assert.True(t, strings.Contains(logInfo, "service closed"))
			defer os.RemoveAll(path.Join("var", "log"))
		})
	}
}
