package ami

import (
	"fmt"
	"io"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kl "k8s.io/apimachinery/pkg/labels"
)

func (k *kubeImpl) FetchLog(ns, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error) {
	deploy, err := k.cli.app.Deployments(ns).Get(service, metav1.GetOptions{})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if deploy == nil {
		return nil, fmt.Errorf("service doesn't exist")
	}
	ls := kl.Set{}
	selector := deploy.Spec.Selector.MatchLabels
	err = copier.Copy(&ls, &selector)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pods, err := k.cli.core.Pods(ns).List(metav1.ListOptions{
		LabelSelector: ls.String(),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if pods == nil || len(pods.Items) == 0 {
		return nil, errors.New("no pod or more than one pod exists")
	}
	s, err := k.cli.core.Pods(ns).GetLogs(pods.Items[0].Name, k.toLogOptions(tailLines, sinceSeconds)).Stream()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return s, nil
}

func (k *kubeImpl) toLogOptions(tailLines, sinceSeconds int64) *corev1.PodLogOptions {
	logOptions := &corev1.PodLogOptions{
		Follow:     k.conf.LogConfig.Follow,
		Previous:   k.conf.LogConfig.Previous,
		Timestamps: k.conf.LogConfig.TimeStamps,
	}
	if tailLines > 0 {
		logOptions.TailLines = &tailLines
	}
	if sinceSeconds > 0 {
		logOptions.SinceSeconds = &sinceSeconds
	}
	return logOptions
}
