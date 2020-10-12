package kube

import (
	"net/http"

	"github.com/baetyl/baetyl-go/v2/errors"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/baetyl/baetyl/v2/ami"
)

func (k *kubeImpl) RemoteCommand(option ami.DebugOptions, pipe ami.Pipe) error {
	req := k.cli.core.RESTClient().Post().Resource("pods").
		Name(option.Name).Namespace(option.Namespace).SubResource("exec")

	opt := &coreV1.PodExecOptions{
		Command: option.Command,
		Stdin:   pipe.InReader != nil,
		Stdout:  pipe.OutWriter != nil,
		Stderr:  pipe.OutWriter != nil,
		TTY:     true,
	}
	if option.Container != "" {
		opt.Container = option.Container
	}

	req.VersionedParams(
		opt,
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(k.cli.kubeConfig, http.MethodPost, req.URL())
	if err != nil {
		return errors.Trace(err)
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  pipe.InReader,
		Stdout: pipe.OutWriter,
		Stderr: pipe.OutWriter,
		Tty:    true,
	})
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
