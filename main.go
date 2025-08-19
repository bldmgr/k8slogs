package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// PodLogConfig holds configuration for log collection
type PodLogConfig struct {
	Namespace     string
	PodNames      []string
	OutputDir     string
	TailLines     *int64
	Follow        bool
	Previous      bool
	SinceSeconds  *int64
	Timestamps    bool
	ContainerName string // Optional: specific container name, empty for all containers
}

// GetPodLogsToFile retrieves logs from specified pods and saves them to text files
func GetPodLogsToFile(clientset *kubernetes.Clientset, config PodLogConfig) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	ctx := context.Background()

	for _, podName := range config.PodNames {
		fmt.Printf("Collecting logs for pod: %s\n", podName)

		// Get pod details to check containers
		pod, err := clientset.CoreV1().Pods(config.Namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Warning: failed to get pod %s: %v\n", podName, err)
			continue
		}

		// Determine which containers to get logs from
		var containers []string
		if config.ContainerName != "" {
			containers = []string{config.ContainerName}
		} else {
			// Get all containers in the pod
			for _, container := range pod.Spec.Containers {
				containers = append(containers, container.Name)
			}
			// Also include init containers if they exist
			for _, initContainer := range pod.Spec.InitContainers {
				containers = append(containers, initContainer.Name)
			}
		}

		// Collect logs from each container
		for _, containerName := range containers {
			if err := collectContainerLogs(clientset, config, podName, containerName); err != nil {
				fmt.Printf("Warning: failed to collect logs for pod %s, container %s: %v\n",
					podName, containerName, err)
			}
		}
	}

	return nil
}

// collectContainerLogs collects logs from a specific container in a pod
func collectContainerLogs(clientset *kubernetes.Clientset, config PodLogConfig, podName, containerName string) error {
	// Prepare log options
	logOptions := &v1.PodLogOptions{
		Container:    containerName,
		Follow:       config.Follow,
		Previous:     config.Previous,
		Timestamps:   config.Timestamps,
		TailLines:    config.TailLines,
		SinceSeconds: config.SinceSeconds,
	}

	// Get logs
	req := clientset.CoreV1().Pods(config.Namespace).GetLogs(podName, logOptions)
	logs, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get logs stream: %w", err)
	}
	defer logs.Close()

	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.log", podName, containerName, timestamp)

	// Handle case where pod has only one container (cleaner filename)
	if len(strings.Split(containerName, "-")) == 1 && containerName != "istio-proxy" {
		filename = fmt.Sprintf("%s_%s.log", podName, timestamp)
	}

	filepath := filepath.Join(config.OutputDir, filename)

	// Create and write to file
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer file.Close()

	// Add header to log file
	header := fmt.Sprintf("=== Pod: %s | Container: %s | Collected: %s ===\n\n",
		podName, containerName, time.Now().Format("2006-01-02 15:04:05"))
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Copy logs to file
	_, err = io.Copy(file, logs)
	if err != nil {
		return fmt.Errorf("failed to write logs to file: %w", err)
	}

	fmt.Printf("  ✓ Saved logs to: %s\n", filepath)
	return nil
}

// CreateKubernetesClient creates a Kubernetes client from kubeconfig or in-cluster config
func CreateKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (for pods running inside cluster)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// ListPodsInNamespace returns a list of all pods in the specified namespace
func ListPodsInNamespace(clientset *kubernetes.Clientset, namespace string) ([]v1.Pod, error) {
	// Get pods from the specified namespace
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %v", namespace, err)
	}

	// Return the list of pods
	return pods.Items, nil
}

// GetKubernetesClient creates and returns a Kubernetes clientset
// It tries in-cluster config first, then falls back to kubeconfig file
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (when running inside a pod)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %v", err)
		}
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	return clientset, nil
}

// GetPodNames returns a slice of pod names from a list of pods
func GetPodNames(pods []v1.Pod) []string {
	podNames := make([]string, len(pods))
	for i, pod := range pods {
		podNames[i] = pod.Name
	}
	return podNames
}

// GetPodNamesInNamespace is a convenience function that returns just the pod names
// as a slice of strings for a given namespace
func GetPodNamesInNamespace(clientset *kubernetes.Clientset, namespace string) ([]string, error) {
	pods, err := ListPodsInNamespace(clientset, namespace)
	if err != nil {
		return nil, err
	}
	return GetPodNames(pods), nil
}

func readArgsWithFlags() string {
	fmt.Println("\n=== Pull pods logs from Kubernetes ===")

	name := flag.String("n", "", "namespace to use")

	// Parse the flags
	flag.Parse()

	// Get remaining non-flag arguments
	remainingArgs := flag.Args()
	if len(remainingArgs) > 0 {
		fmt.Printf("Non-flag arguments: %v\n", remainingArgs)
	}

	return *name
}

func main() {

	// Create Kubernetes client
	clientset, err := CreateKubernetesClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Specify the namespace
	namespace := readArgsWithFlags() // Change this to your desired namespace

	pods, err := ListPodsInNamespace(clientset, namespace)
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	podNames := GetPodNames(pods)

	// Configure log collection
	tailLines := int64(1000) // Get last 1000 lines
	config := PodLogConfig{
		Namespace:     namespace, // Change to your namespace
		PodNames:      podNames,
		OutputDir:     "./pod-logs",
		TailLines:     &tailLines,
		Follow:        false, // Set to true for live streaming
		Previous:      false, // Set to true to get logs from previous container instance
		Timestamps:    true,  // Include timestamps in logs
		ContainerName: "",    // Empty string gets all containers, or specify specific container
	}

	// Collect logs
	if err := GetPodLogsToFile(clientset, config); err != nil {
		fmt.Printf("Error collecting logs: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Log collection completed!")
}

// Function for collecting logs with custom time range
func GetPodLogsWithTimeRange(clientset *kubernetes.Clientset, namespace string, podNames []string,
	outputDir string, sinceTime time.Time) error {

	sinceSeconds := int64(time.Since(sinceTime).Seconds())

	config := PodLogConfig{
		Namespace:    namespace,
		PodNames:     podNames,
		OutputDir:    outputDir,
		SinceSeconds: &sinceSeconds,
		Timestamps:   true,
		Follow:       false,
	}

	return GetPodLogsToFile(clientset, config)
}
