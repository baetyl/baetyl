package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/baetyl/baetyl/utils"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/assert"
	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

func TestDefaultsValidator(t *testing.T) {
	type mm struct {
		Min *int `yaml:"Min" default:"1" validate:"min=1"`
		Max int  `yaml:"Max" default:"1" validate:"min=1"`
	}
	type dummy struct {
		T string `yaml:"T" default:"1m" validate:"regexp=^[1-9][0-9]{0\\,5}[smh]?$"`
		M []mm   `yaml:"M"`
	}
	d := new(dummy)
	defaults.Set(d)
	assert.Equal(t, "1m", d.T)
	assert.Nil(t, d.M)
	err := validator.Validate(d)
	assert.NoError(t, err)
	d.T = "1"
	err = validator.Validate(d)
	assert.NoError(t, err)
	d.T = "abc"
	err = validator.Validate(d)
	assert.Error(t, err)

	err = yaml.Unmarshal([]byte(`
T: 60
M:
  - Min: 0
  - Max: 0
  - Min: 2
    Max: 2
`), d)
	assert.NoError(t, err)
	assert.Len(t, d.M, 3)
	for i := range d.M {
		defaults.Set(&d.M[i])
	}
	assert.Equal(t, 1, *d.M[0].Min)
	assert.Equal(t, 1, d.M[0].Max)
	assert.Equal(t, 1, *d.M[1].Min)
	assert.Equal(t, 1, d.M[1].Max)
	assert.Equal(t, 2, *d.M[2].Min)
	assert.Equal(t, 2, d.M[2].Max)

	v := 0
	d.M[0].Min = &v
	err = validator.Validate(d)
	assert.EqualError(t, err, "M[0].Min: less than min")
}

func TestConfig(t *testing.T) {
	var c Config
	err := utils.LoadYAML("../../example/native/var/db/baetyl/localhub-conf/service.yml", &c)
	assert.NoError(t, err)

	assert.Equal(t, "var/db/baetyl/data", c.Storage.Dir)

	assert.Equal(t, 10000, c.Message.Ingress.Qos0.Buffer.Size)
	assert.Equal(t, 100, c.Message.Ingress.Qos1.Buffer.Size)
	assert.Equal(t, 50, c.Message.Ingress.Qos1.Batch.Max)
	assert.Equal(t, time.Hour*48, c.Message.Ingress.Qos1.Cleanup.Retention)
	assert.Equal(t, time.Minute, c.Message.Ingress.Qos1.Cleanup.Interval)

	assert.Equal(t, 10000, c.Message.Egress.Qos0.Buffer.Size)
	assert.Equal(t, 100, c.Message.Egress.Qos1.Buffer.Size)
	assert.Equal(t, 50, c.Message.Egress.Qos1.Batch.Max)
	assert.Equal(t, time.Second*20, c.Message.Egress.Qos1.Retry.Interval)
	assert.Equal(t, 10000, c.Message.Offset.Buffer.Size)
	assert.Equal(t, 100, c.Message.Offset.Batch.Max)

	assert.Equal(t, time.Minute, c.Metrics.Report.Interval)

	assert.Equal(t, int64(32768), c.Message.Length.Max)
}

func TestConfigSubscription(t *testing.T) {
	c, err := New([]byte(`
id: id
name: name
subscriptions:
- source:
    topic: abc
    qos: 0
  target:
    topic: x/abc
    qos: 1
- source:
    topic: abc
    qos: 1
  target:
    topic: y/abc
    qos: 1
`))
	assert.NoError(t, err)
	assert.Len(t, c.Subscriptions, 2)
	assert.Equal(t, "abc", c.Subscriptions[0].Source.Topic)
	assert.Equal(t, byte(0), c.Subscriptions[0].Source.QOS)
	assert.Equal(t, "x/abc", c.Subscriptions[0].Target.Topic)
	assert.Equal(t, byte(1), c.Subscriptions[0].Target.QOS)
	assert.Equal(t, "abc", c.Subscriptions[1].Source.Topic)
	assert.Equal(t, byte(1), c.Subscriptions[1].Source.QOS)
	assert.Equal(t, "y/abc", c.Subscriptions[1].Target.Topic)
	assert.Equal(t, byte(1), c.Subscriptions[1].Target.QOS)
}

func TestPrincipalsValidate(t *testing.T) {
	// round 1: regular principals config validate
	principals := []Principal{{
		Username: "test",
		Password: "hahaha",
		Permissions: []Permission{
			{"pub", []string{"test", "benchmark", "#", "+", "test/+", "test/#"}},
			{"sub", []string{"test", "benchmark", "#", "+", "test/+", "test/#"}},
		}}, {
		Username: "temp",
		Password: "3f29e1b2b05f8371595dc761fed8e8b37544b38d56dfce81a551b46c82f2f56b",
		Permissions: []Permission{
			{"pub", []string{"test", "benchmark", "a/+/b", "+/a/+", "+/a/#"}},
			{"sub", []string{"test", "benchmark", "a/+/b", "+/a/+", "+/a/#"}},
		}}}
	err := principalsValidate(principals, "")
	assert.NoError(t, err)

	// round 2: duplicate username validate
	principals = principals[:len(principals)-1]
	principals = append(principals, Principal{
		Username: "test",
		Password: "3f29e1b2b05f8371595dc761fed8e8b37544b38d56dfce81a551b46c82f2f56b",
		Permissions: []Permission{
			{"pub", []string{"test", "benchmark"}},
		}})
	err = principalsValidate(principals, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("username (test) duplicate"), err.Error())

	// round 3: invalid publish topic validate
	principals = principals[:len(principals)-1]
	principals = append(principals, Principal{
		Username: "hello",
		Password: "3f29e1b2b05f8371595dc761fed8e8b37544b38d56dfce81a551b46c82f2f56b",
		Permissions: []Permission{
			{"pub", []string{"test/a+/b", "benchmark"}},
		}})
	err = principalsValidate(principals, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("pub topic(test/a+/b) invalid"), err.Error())

	// round 4: invalid subscribe topic validate
	principals = principals[:len(principals)-1]
	principals = append(principals, Principal{
		Username: "hello",
		Password: "3f29e1b2b05f8371595dc761fed8e8b37544b38d56dfce81a551b46c82f2f56b",
		Permissions: []Permission{
			{"pub", []string{"test", "benchmark"}},
			{"sub", []string{"test", "test/#/temp"}},
		}})
	err = principalsValidate(principals, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("sub topic(test/#/temp) invalid"), err.Error())
}

func TestSubscriptionsValidate(t *testing.T) {
	// round 1: normal config validate
	subscriptions := make([]Subscription, 4)
	subscriptions[0].Source.Topic = "test/#"
	subscriptions[0].Target.Topic = "remote/iothub/test"
	subscriptions[1].Source.Topic = "test"
	subscriptions[1].Target.Topic = "point"
	subscriptions[2].Source.Topic = "remote/awshub/aws_test"
	subscriptions[2].Target.Topic = "demo"
	subscriptions[3].Source.Topic = "remote/azurehub/azure_test"
	subscriptions[3].Target.Topic = "point"
	err := subscriptionsValidate(subscriptions, "")
	assert.NoError(t, err)

	// round 6: source topic invalid
	subscriptions = subscriptions[:len(subscriptions)-1]
	var substopic Subscription
	substopic.Source.Topic = "test/#/a"
	substopic.Target.Topic = "test"
	subscriptions = append(subscriptions, substopic)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, "[{Topic:test/#/a QOS:0}] source topic invalid", err.Error())

	// round 7: target topic invalid
	subscriptions = subscriptions[:len(subscriptions)-1]
	var subttopic Subscription
	subttopic.Source.Topic = "test1"
	subttopic.Target.Topic = "test/+"
	subscriptions = append(subscriptions, subttopic)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, "[{Topic:test/+ QOS:0}] target topic invalid", err.Error())

	// round 8: duplicate source and target config(equal, all normal)
	subscriptions = subscriptions[:len(subscriptions)-2]
	var subdup1, subdup2 Subscription
	subdup1.Source.Topic = "test1"
	subdup1.Target.Topic = "test2"
	subdup2.Source.Topic = "test1"
	subdup2.Target.Topic = "test2"
	subscriptions = append(subscriptions, subdup1, subdup2)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("duplicate source and target config"), err.Error())

	// round 9: duplicate source and target config(equal, one normal, the other one wildcard)
	subscriptions = subscriptions[:len(subscriptions)-2]
	subdup1.Source.Topic = "test/#"
	subdup1.Target.Topic = "point"
	subdup2.Source.Topic = "test/#"
	subdup2.Target.Topic = "point"
	subscriptions = append(subscriptions, subdup1, subdup2)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("duplicate source and target config"), err.Error())

	// round 10: duplicate source and target config(equal, one normal, the other one function)
	subscriptions = subscriptions[:len(subscriptions)-2]
	subdup1.Source.Topic = "test1"
	subdup1.Target.Topic = "point"
	subdup2.Source.Topic = "test1"
	subdup2.Target.Topic = "point"
	subscriptions = append(subscriptions, subdup1, subdup2)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("duplicate source and target config"), err.Error())

	// round 11: duplicate source and target config(equal, one wildcard, the other one remote)
	subscriptions = subscriptions[:len(subscriptions)-2]
	subdup1.Source.Topic = "test/#"
	subdup1.Target.Topic = "remote/iothub/test"
	subdup2.Source.Topic = "test/#"
	subdup2.Target.Topic = "remote/iothub/test"
	subscriptions = append(subscriptions, subdup1, subdup2)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("duplicate source and target config"), err.Error())

	// round 12: duplicate source and target config(equal, one remote, the other one function)
	subscriptions = subscriptions[:len(subscriptions)-2]
	subdup1.Source.Topic = "remote/awshub/aws_test"
	subdup1.Target.Topic = "point"
	subdup2.Source.Topic = "remote/awshub/aws_test"
	subdup2.Target.Topic = "point"
	subscriptions = append(subscriptions, subdup1, subdup2)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("duplicate source and target config"), err.Error())

	// round 13: duplicate source and target config(equal, all remote)
	subscriptions = subscriptions[:len(subscriptions)-2]
	subdup1.Source.Topic = "remote/azurehub/azure_test"
	subdup1.Target.Topic = "remote/huaweihub/huawei_test"
	subdup2.Source.Topic = "remote/azurehub/azure_test"
	subdup2.Target.Topic = "remote/huaweihub/huawei_test"
	subscriptions = append(subscriptions, subdup1, subdup2)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("duplicate source and target config"), err.Error())

	subscriptions[0].Source.Topic = "test/#"
	subscriptions[0].Target.Topic = "remote/iothub/test"
	subscriptions[1].Source.Topic = "test"
	subscriptions[1].Target.Topic = "point"

	// // round 14: contains source and target config(one normal, the other one wildcard)
	// subscriptions = subscriptions[:len(subscriptions)-1]
	// subdup1.Source.Topic = "test"
	// subdup1.Target.Topic = "point"
	// subdup2.Source.Topic = "test/#"
	// subdup2.Target.Topic = "point"
	// subscriptions = append(subscriptions, subdup1, subdup2)
	// err = subscriptionsValidate(subscriptions, "")
	// assert.NoError(t, err)

	// round 15: intersection(test/a/b/from) source and target config(all wildcard)
	subscriptions = subscriptions[:len(subscriptions)-2]
	subdup1.Source.Topic = "test/+/+/from"
	subdup1.Target.Topic = "point"
	subdup2.Source.Topic = "test/a/b/#"
	subdup2.Target.Topic = "point"
	subscriptions = append(subscriptions, subdup1, subdup2)
	err = subscriptionsValidate(subscriptions, "")
	assert.NoError(t, err)

	// round 16: transmit source and target config
	subscriptions = subscriptions[:len(subscriptions)-3]
	var subdup3 Subscription
	subdup1.Source.Topic = "test/a"
	subdup1.Target.Topic = "test/b"
	subdup2.Source.Topic = "test/b"
	subdup2.Target.Topic = "test/c"
	subdup3.Source.Topic = "test/a"
	subdup3.Target.Topic = "test/c"
	subscriptions = append(subscriptions, subdup1, subdup2, subdup3)
	err = subscriptionsValidate(subscriptions, "")
	assert.NoError(t, err)

	// round 17: cycle in one source and target config(self-cycle, normal and normal)
	subscriptions = make([]Subscription, 1)
	subscriptions[0].Source.Topic = "test"
	subscriptions[0].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 18: cycle in one source and target config(self-cycle, normal and function)
	subscriptions[0].Source.Topic = "test"
	subscriptions[0].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 19: cycle in one source and target config(self-cycle, normal and wildcard)
	subscriptions[0].Source.Topic = "+"
	subscriptions[0].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 20: cycle in one source and target config(self-cycle, wildcard and normal)
	subscriptions[0].Source.Topic = "#"
	subscriptions[0].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 21: cycle in one source and target config(self-cycle, wildcard and function)
	subscriptions[0].Source.Topic = "+"
	subscriptions[0].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 22: cycle in one source and target config(self-cycle, wildcard and function)
	subscriptions[0].Source.Topic = "#"
	subscriptions[0].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 23: cycle in one source and target config(self-cycle, wildcard and remote)
	subscriptions[0].Source.Topic = "#"
	subscriptions[0].Target.Topic = "remote/iothub/test"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 24: cycle in one source and target config(self-cycle, remote and remote)
	subscriptions[0].Source.Topic = "remote/iothub/test"
	subscriptions[0].Target.Topic = "remote/iothub/test"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 25: cycle in multiple source and target config(self-cycle, normal topic)
	subscriptions = make([]Subscription, 4)
	subscriptions[0].Source.Topic = "point"
	subscriptions[0].Target.Topic = "test"
	subscriptions[1].Source.Topic = "test"
	subscriptions[1].Target.Topic = "test"
	subscriptions[2].Source.Topic = "test"
	subscriptions[2].Target.Topic = "sayhi"
	subscriptions[3].Source.Topic = "bmi"
	subscriptions[3].Target.Topic = "bmo"
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 26: cycle in multiple source and target config(self-cycle, normal && wildcard topic)
	subscriptions[1].Source.Topic = "test/#"
	subscriptions[1].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 27: cycle in multiple source and target config(normal topic)
	subscriptions[1].Source.Topic = "sayhi"
	subscriptions[1].Target.Topic = "point"
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 28: cycle in multiple source and target config(normal && wildcard)
	subscriptions = make([]Subscription, 4)
	subscriptions[0].Source.Topic = "test"
	subscriptions[0].Target.Topic = "point"
	subscriptions[1].Source.Topic = "test/from/#"
	subscriptions[1].Target.Topic = "remote/iothub/test"
	subscriptions[2].Source.Topic = "remote/iothub/test"
	subscriptions[2].Target.Topic = "point"
	subscriptions[3].Source.Topic = "point"
	subscriptions[3].Target.Topic = "test/from"
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 29: cycle in multiple source and target config(normal && function)
	subscriptions = subscriptions[:len(subscriptions)-1]
	var subcycle Subscription
	subcycle.Source.Topic = "point"
	subcycle.Target.Topic = "test"
	subscriptions = append(subscriptions, subcycle)
	err = subscriptionsValidate(subscriptions, "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 30: cycle in multiple source and target config(normal && normal)
	subscriptions = make([]Subscription, 4)
	subscriptions[0].Source.Topic = "test"
	subscriptions[0].Target.Topic = "point"
	subscriptions[1].Source.Topic = "test"
	subscriptions[1].Target.Topic = "sayhi"
	subscriptions[2].Source.Topic = "sayhi"
	subscriptions[2].Target.Topic = "func"
	subscriptions[3].Source.Topic = "func"
	subscriptions[3].Target.Topic = "test"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 31: cycle in multiple source and target config(remote && remote)
	subscriptions[0].Source.Topic = "remote/iothub/test"
	subscriptions[0].Target.Topic = "point"
	subscriptions[1].Source.Topic = "point"
	subscriptions[1].Target.Topic = "sayhi"
	subscriptions[2].Source.Topic = "sayhi"
	subscriptions[2].Target.Topic = "remote/iothub/test"
	subscriptions[3].Source.Topic = "test"
	subscriptions[3].Target.Topic = "func"
	err = subscriptionsValidate(subscriptions, "")
	assert.Equal(t, fmt.Sprintf("found cycle in source and target config"), err.Error())

	// round 32: no cycle found in multiple source and target config
	subscriptions[0].Source.Topic = "test"
	subscriptions[0].Target.Topic = "point"
	subscriptions[1].Source.Topic = "point"
	subscriptions[1].Target.Topic = "remote/awshub/aws_test"
	subscriptions[2].Source.Topic = "remote/iothubtest"
	subscriptions[2].Target.Topic = "filted"
	subscriptions[3].Source.Topic = "sayhi"
	subscriptions[3].Target.Topic = "point"
	err = subscriptionsValidate(subscriptions, "")
	assert.NoError(t, err)
}
