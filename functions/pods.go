package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml "k8s.io/apimachinery/pkg/util/yaml"
)

type PodUpdate struct {
	Data string `json:"data"`
}

func PodLogs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err.Error())
	}

	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")
	podLogOpts := corev1.PodLogOptions{
		Follow: true,
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(name, &podLogOpts)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return
	}
	defer podLogs.Close()

	buf := make([]byte, 1024)
	for {
		n, err := podLogs.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, buf[:n])
		if err != nil {
			return
		}
	}
}

func Pods(c *gin.Context) {

	Namespace = c.Query("namespace")

	c.HTML(http.StatusOK, "pods.html", gin.H{})
}

func GetPod(c *gin.Context) {
	// clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")
	podData, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"data": podData})

}

func UpdatePod(c *gin.Context) {
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")

	var podUpdate PodUpdate
	if err := c.ShouldBindJSON(&podUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// podData := []byte(podUpdate.Data)
	// if err := yaml.Unmarshal(podData, &pod); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	podData := []byte(podUpdate.Data)
	jsonData, err := yaml.ToJSON(podData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := json.Unmarshal(jsonData, &pod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedPod, err := clientset.CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedPod})
}

func DeletePod(c *gin.Context) {
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")

	err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func WsPods(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		pollPods(conn)
	}()
}

func pollPods(conn *websocket.Conn) {
	clientset := Clientset

	for {
		fmt.Println("Polling pods...")

		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		namespaceList = namespaces.Items

		pods, err := clientset.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		data := make(map[string]interface{})
		data["pods"] = pods.Items
		data["namespaces"] = namespaceList

		err = conn.WriteJSON(data)
		if err != nil {
			panic(err.Error())
		}
		time.Sleep(RefreshTime * time.Second)
	}
}

func DeletePods(c *gin.Context) {
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")

	err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
