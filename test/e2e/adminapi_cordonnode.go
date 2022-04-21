package e2e

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"context"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[Admin API] Cordon Node action", func() {
	BeforeEach(skipIfNotInDevelopmentEnv)

	It("should cordon then uncordon a selected Node", func() {
		ctx := context.Background()

		By("picking a worker node to cordon")
		nodes, err := clients.Kubernetes.CoreV1().Nodes().List(ctx, metav1.ListOptions{
			LabelSelector: "node-role.kubernetes.io/worker",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(nodes.Items).NotTo(HaveLen(0))
		node := nodes.Items[0]
		log.Infof("selected node: %s", node.Name)

		By("verifying cordon action completes without error")
		resp, err := adminRequest(ctx, http.MethodPost, "/admin"+resourceIDFromEnv()+"/cordonnode", url.Values{"vmName": []string{node.Name}, "unschedulable": []string{"true"}}, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		By("verifying pod deployed to cordoned node causes an error")
		err = deploySamplePod(node.Name)
		Expect(err).NotTo(HaveOccurred())

		By("verifying uncordon action completes without error")
		resp, err = adminRequest(ctx, http.MethodPost, "/admin"+resourceIDFromEnv()+"/cordonnode", url.Values{"vmName": []string{node.Name}, "unschedulable": []string{"false"}}, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		By("verifying pod deployed to uncordoned node complete without error")
		err = deploySamplePod(node.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})

func deploySamplePod(nodeName string) error {
	ctx := context.Background()
	namespace := "default"
	name := nodeName
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "cli",
					Image: "image-registry.openshift-image-registry.svc:5000/openshift/cli",
					Command: []string{
						"/bin/sh",
						"-c",
						"uptime -s",
					},
				},
			},
			RestartPolicy: "Never",
			NodeName:      nodeName,
		},
	}

	// Create
	_, err := clients.Kubernetes.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	// Defer Delete
	defer func() {
		err := clients.Kubernetes.CoreV1().Pods(namespace).Delete(ctx, nodeName, metav1.DeleteOptions{})
		if err != nil {
			log.Error("Could not delete test Pod")
		}
	}()

	return nil
}
