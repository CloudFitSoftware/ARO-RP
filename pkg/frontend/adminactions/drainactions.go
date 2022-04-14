package adminactions

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"

	"github.com/Azure/ARO-RP/pkg/api"
	"github.com/Azure/ARO-RP/pkg/env"
	"github.com/Azure/ARO-RP/pkg/util/restconfig"
)

// DrainActions are those that involve k8s objects, and thus depend upon k8s clients being createable
type DrainActions interface {
	CordonNode(ctx context.Context, nodeName string, unschedulable bool) error
	DrainNode(nodeName string) error
}

type drainActions struct {
	log *logrus.Entry
	oc  *api.OpenShiftCluster

	kubernetescli kubernetes.Interface
}

// NewDrainActions returns a drainActions
func NewDrainActions(log *logrus.Entry, env env.Interface, oc *api.OpenShiftCluster) (DrainActions, error) {
	restConfig, err := restconfig.RestConfig(env, oc)
	if err != nil {
		return nil, err
	}

	kubernetescli, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &drainActions{
		log: log,
		oc:  oc,

		kubernetescli: kubernetescli,
	}, nil
}

func (d *drainActions) CordonNode(ctx context.Context, nodeName string, unschedulable bool) error {

	node, err := d.kubernetescli.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	drainer := &drain.Helper{
		Ctx:                 ctx,
		Client:              d.kubernetescli,
		Force:               true,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Timeout:             60 * time.Second,
		DeleteEmptyDirData:  true,
		DisableEviction:     true,
		OnPodDeletedOrEvicted: func(pod *corev1.Pod, usingEviction bool) {
			log.Printf("deleted pod %s/%s", pod.Namespace, pod.Name)
		},
		Out:    log.Writer(),
		ErrOut: log.Writer(),
	}

	return drain.RunCordonOrUncordon(drainer, node, unschedulable)
}

func (d *drainActions) DrainNode(nodeName string) error {

	drainer := &drain.Helper{
		Client:              d.kubernetescli,
		Force:               true,
		GracePeriodSeconds:  -1,
		IgnoreAllDaemonSets: true,
		Timeout:             60 * time.Second,
		DeleteEmptyDirData:  true,
		DisableEviction:     true,
		OnPodDeletedOrEvicted: func(pod *corev1.Pod, usingEviction bool) {
			log.Printf("deleted pod %s/%s", pod.Namespace, pod.Name)
		},
		Out:    log.Writer(),
		ErrOut: log.Writer(),
	}

	return drain.RunNodeDrain(drainer, nodeName)
}
