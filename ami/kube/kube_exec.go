package kube

import (
	"context"
	"io"
	"net/http"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/utils"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/baetyl/baetyl/v2/ami"
)

func (k *kubeImpl) RemoteCommand(option *ami.DebugOptions, pipe ami.Pipe) error {
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

func (k *kubeImpl) RemoteWebsocket(ctx context.Context, option *ami.DebugOptions, pipe ami.Pipe) error {
	return ami.RemoteWebsocket(ctx, option, pipe)
}

func (k *kubeImpl) RemoteLogs(option *ami.LogsOptions, pipe ami.Pipe) error {
	req := k.cli.core.RESTClient().Get().Resource("pods").
		Name(option.Name).Namespace(option.Namespace).SubResource("log")

	opt := &coreV1.PodLogOptions{
		Follow:       option.Follow,
		Previous:     option.Previous,
		SinceSeconds: option.SinceSeconds,
		Timestamps:   option.Timestamps,
		TailLines:    option.TailLines,
	}
	if option.Container != "" {
		opt.Container = option.Container
	}
	if option.LimitBytes != nil && *option.LimitBytes > int64(0) {
		opt.LimitBytes = option.LimitBytes
	}

	req.VersionedParams(
		opt,
		scheme.ParameterCodec,
	)

	reader, err := req.Stream(context.TODO())
	if err != nil {
		return errors.Trace(err)
	}
	defer reader.Close()
	_, err = io.Copy(pipe.OutWriter, reader)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (k *kubeImpl) UpdateNodeLabels(name string, labels map[string]string) error {
	defer utils.Trace(k.log.Debug, "UpdateNodeLabels")()
	n, err := k.cli.core.Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return errors.Trace(err)
	}
	n.Labels = labels
	n, err = k.cli.core.Nodes().Update(context.TODO(), n, metav1.UpdateOptions{})
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}
