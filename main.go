package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type PageData struct {
	// K8sPods []v1.Pod
	ClusterData    map[string][][]string
	ClusterMetrics map[string]map[string]string
}

var (
	clusterData    map[string][][]string
	clusterMetrics map[string]map[string]string
	// mux         sync.Mutex
)

func main() {
	go func() {
		pollCluster()
	}()
	http.HandleFunc("/", greatCabin)
	log.Fatal(http.ListenAndServe(":81", nil))

}

func greatCabin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	// w.Write(clusterData)

	data := PageData{
		ClusterData:    clusterData,
		ClusterMetrics: clusterMetrics,
	}
	// fmt.Println(clusterData)
	tmpl := template.Must(template.ParseFiles("index.html"))
	// tmpl.ParseGlob("views/assets/*")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
	// return
}

func pollCluster() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	metricsClientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	clusterData = make(map[string][][]string)

	clusterMetrics = make(map[string]map[string]string)

	for {
		fmt.Println("Polling cluster...")
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		events, err := clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		clusterData["pods"] = convertPodsToData(pods.Items)

		clusterData["events"] = convertEventsToData(events.Items)

		clusterMetrics = convertPodsUsageToData(metricsClientset.MetricsV1beta1(), pods.Items)

		// fmt.Println(clusterMetrics) // test

		// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// fmt.Print("\033[H\033[2J")
		// mux.Lock()

		// mux.Unlock()

		time.Sleep(10 * time.Second)
	}

}

func convertEventsToData(events []v1.Event) [][]string {
	var data [][]string
	for _, event := range events {
		if event.Type == "Warning" {
			eventData := []string{event.Name, event.Namespace, event.Source.Component, event.Message}
			data = append(data, eventData)
		}
	}
	return data
}

func convertPodsToData(pods []v1.Pod) [][]string {
	var data [][]string
	// fmt.Println(pods)
	for _, pod := range pods {
		// Convert the PodPhase enum to a string
		phase := string(pod.Status.Phase)
		// fmt.Println(pod)
		if phase != "Running" && phase != "Succeeded" {
			// Extract relevant information from the pod into a slice of strings
			podData := []string{pod.Name, pod.Namespace, phase, pod.Status.Message}
			// fmt.Println(podData)
			// Append the slice of strings to the final data slice
			data = append(data, podData)
		}
	}
	return data
}

func convertPodsUsageToData(metricsClient v1beta1.PodMetricsesGetter, pods []v1.Pod) map[string]map[string]string {
	var data map[string]map[string]string

	data = make(map[string]map[string]string)

	for _, pod := range pods {
		if pod.Status.Phase == "Running" {
			podMetrics, err := metricsClient.PodMetricses(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Failed to get metrics for pod %s/%s: %v\n", pod.Namespace, pod.Name, err)
				continue
			}
			for i, container := range podMetrics.Containers {
				var metricsStatus int = 0
				containerName := container.Name
				cpuUsage := container.Usage[v1.ResourceCPU]
				memoryUsage := container.Usage[v1.ResourceMemory]

				cpuRequests := pod.Spec.Containers[i].Resources.Requests[v1.ResourceCPU]
				cpuPercentage := float64(cpuUsage.MilliValue()) / float64(memoryUsage.Value()) * 100
				cpuLimits := pod.Spec.Containers[i].Resources.Limits[v1.ResourceCPU]
				memoryRequests := pod.Spec.Containers[i].Resources.Requests[v1.ResourceMemory]
				memoryLimits := pod.Spec.Containers[i].Resources.Limits[v1.ResourceMemory]

				cpuUsageStr := strconv.FormatInt(int64(cpuUsage.MilliValue()), 10) + "m"
				cpuPercentageStr := strconv.FormatInt(int64(cpuPercentage), 10) + "%"
				memoryUsageStr := strconv.FormatInt(int64(memoryUsage.Value()/(1024*1024)), 10) + "Mi"

				if int64(cpuUsage.MilliValue()) > int64(cpuRequests.MilliValue()) && int64(cpuRequests.MilliValue()) > 0 {
					// fmt.Println(int64(cpuUsage.MilliValue()))
					// fmt.Println(int64(cpuRequests.MilliValue()))
					metricsStatus = 1
				}

				// row := []string{
				// 	strconv.Itoa(metricsStatus),
				// 	pod.Namespace,
				// 	pod.Name,
				// 	containerName,
				// 	cpuUsageStr,
				// 	cpuRequests.String(),
				// 	cpuLimits.String(),
				// 	memoryUsageStr,
				// 	memoryRequests.String(),
				// 	memoryLimits.String(),
				// }

				data[pod.Name+"-"+containerName] = make(map[string]string)
				data[pod.Name+"-"+containerName]["metricStatus"] = strconv.Itoa(metricsStatus)
				data[pod.Name+"-"+containerName]["namespace"] = pod.Namespace
				data[pod.Name+"-"+containerName]["podName"] = pod.Name
				data[pod.Name+"-"+containerName]["containerName"] = containerName
				data[pod.Name+"-"+containerName]["cpuUsage"] = cpuUsageStr
				data[pod.Name+"-"+containerName]["cpuRequests"] = cpuRequests.String()
				data[pod.Name+"-"+containerName]["cpuPercentage"] = cpuPercentageStr
				data[pod.Name+"-"+containerName]["cpuLimits"] = cpuLimits.String()
				data[pod.Name+"-"+containerName]["memoryUsage"] = memoryUsageStr
				data[pod.Name+"-"+containerName]["memoryRequests"] = memoryRequests.String()
				data[pod.Name+"-"+containerName]["memoryPercentage"] = memoryRequests.String()
				data[pod.Name+"-"+containerName]["memoryLimits"] = memoryLimits.String()

				// fmt.Println(row)
				// data = append(data, row)
			}
		}
	}
	// fmt.Println(data)
	return data
}
