package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PodManager struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewPodManager creates a new PodManager using in-cluster config
func NewPodManager(namespace string) (*PodManager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &PodManager{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// CreateCrawlPod creates a new crawl4ai pod with a unique name
func (pm *PodManager) CreateCrawlPod(ctx context.Context, podName string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: pm.namespace,
			Labels: map[string]string{
				"app":       "crawl4ai",
				"managed-by": "gocrawl",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            "crawl4ai",
					Image:           "unclecode/crawl4ai:0.7.5",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 11235,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "dshm",
							MountPath: "/dev/shm",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "dshm",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumMemory,
							SizeLimit: resourceQuantity("3Gi"),
						},
					},
				},
			},
		},
	}

	createdPod, err := pm.clientset.CoreV1().Pods(pm.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	return createdPod, nil
}

// WaitForPodReady waits for the pod to be ready with a timeout
func (pm *PodManager) WaitForPodReady(ctx context.Context, podName string, timeout time.Duration) error {
	fmt.Printf("[K8S] Polling for pod %s readiness...\n", podName)
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		pod, err := pm.clientset.CoreV1().Pods(pm.namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("[K8S] Error getting pod %s: %v\n", podName, err)
			return false, err
		}

		fmt.Printf("[K8S] Pod %s - Phase: %s\n", podName, pod.Status.Phase)

		// Check if pod is running and ready
		if pod.Status.Phase == corev1.PodRunning {
			for _, condition := range pod.Status.Conditions {
				fmt.Printf("[K8S] Pod %s - Condition: %s = %s (Reason: %s, Message: %s)\n",
					podName, condition.Type, condition.Status, condition.Reason, condition.Message)
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					fmt.Printf("[K8S] Pod %s is ready!\n", podName)
					return true, nil
				}
			}
		}

		// Check if pod failed
		if pod.Status.Phase == corev1.PodFailed {
			fmt.Printf("[K8S] Pod %s failed! Reason: %s\n", podName, pod.Status.Reason)
			return false, fmt.Errorf("pod failed to start: %s", pod.Status.Reason)
		}

		// Log container statuses
		for _, containerStatus := range pod.Status.ContainerStatuses {
			fmt.Printf("[K8S] Container %s - Ready: %v, Started: %v\n",
				containerStatus.Name, containerStatus.Ready, *containerStatus.Started)
			if containerStatus.State.Waiting != nil {
				fmt.Printf("[K8S] Container %s waiting: %s - %s\n",
					containerStatus.Name, containerStatus.State.Waiting.Reason, containerStatus.State.Waiting.Message)
			}
			if containerStatus.State.Terminated != nil {
				fmt.Printf("[K8S] Container %s terminated: %s - %s\n",
					containerStatus.Name, containerStatus.State.Terminated.Reason, containerStatus.State.Terminated.Message)
			}
		}

		return false, nil
	})
}

// DeletePod deletes the specified pod
func (pm *PodManager) DeletePod(ctx context.Context, podName string) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := pm.clientset.CoreV1().Pods(pm.namespace).Delete(ctx, podName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	return nil
}

// GetPodIP returns the IP address of the pod
func (pm *PodManager) GetPodIP(ctx context.Context, podName string) (string, error) {
	pod, err := pm.clientset.CoreV1().Pods(pm.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod: %w", err)
	}

	if pod.Status.PodIP == "" {
		return "", fmt.Errorf("pod IP not yet assigned")
	}

	return pod.Status.PodIP, nil
}

// Helper function to create resource quantity
func resourceQuantity(value string) *resource.Quantity {
	q := resource.MustParse(value)
	return &q
}
