package functions

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	clusterData    map[string][][]string
	clusterMetrics map[string]map[string]string
	namespaceList  []corev1.Namespace
	pods           []corev1.Pod
	events         []corev1.Event
)

func Overview(c *gin.Context) {

	Namespace = c.Query("namespace")
	c.HTML(http.StatusOK, "overview.html", gin.H{})
}

func WsOverview(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
	}

	data := make(map[string]interface{})
	data["namespaces"] = namespaceList
	data["pods"] = pods
	data["events"] = events
	// data["irregularMetrics"] = namespaceList
	// for true {
	// 	err = conn.WriteJSON(data)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	time.Sleep(5 * time.Second)
	// }
	go func() {
		pollCluster(conn)
	}()

}

func pollCluster(conn *websocket.Conn) {

	clientset := Clientset
	// metricsClientset := functions.MetricsClientset

	// clusterData = make(map[string][][]string)

	// clusterMetrics = make(map[string]map[string]string)

	for {
		fmt.Println("Polling cluster...")

		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		namespaceList = namespaces.Items

		clusterPods, err := clientset.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		clusterEvents, err := clientset.CoreV1().Events(Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		// clusterData["pods"] = convertPodsToData(pods.Items)
		// clusterData["events"] = convertEventsToData(events.Items)

		// clusterMetrics = convertPodsUsageToData(metricsClientset.MetricsV1beta1(), pods.Items)

		pods = clusterPods.Items
		events = clusterEvents.Items

		data := make(map[string]interface{})
		data["namespaces"] = namespaceList
		data["pods"] = pods
		data["events"] = events
		err = conn.WriteJSON(data)
		if err != nil {
			// netErr, ok := err.(*net.OpError)
			// if ok && netErr.Err == syscall.EPIPE {
			// 	fmt.Println("Connection broken, stopping writes.")

			// }
			fmt.Println("Breaking connection: " + err.Error())
			break
		}
		time.Sleep(RefreshTime * time.Second)
	}

}
