package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	// "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yaml "k8s.io/apimachinery/pkg/util/yaml"
)

type DeploymentUpdate struct {
	Data string `json:"data"`
}

func Deployment(c *gin.Context) {

	Namespace = c.Query("namespace")

	c.HTML(http.StatusOK, "deployments.html", gin.H{})
}

func GetDeployment(c *gin.Context) {
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")
	deploymentData, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"data": deploymentData})
}

func UpdateDeployment(c *gin.Context) {
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")

	var deploymentUpdate DeploymentUpdate
	if err := c.ShouldBindJSON(&deploymentUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	deploymentData := []byte(deploymentUpdate.Data)
	jsonData, err := yaml.ToJSON(deploymentData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := json.Unmarshal(jsonData, &deployment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedDeployment, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updatedDeployment})
}

func DeleteDeployment(c *gin.Context) {
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")

	err := clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		panic(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func WsDeployments(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		pollDeployments(conn)
	}()
}

func pollDeployments(conn *websocket.Conn) {
	clientset := Clientset

	for {
		fmt.Print("Polling deployments...")

		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		namespaceList = namespaces.Items
		deployments, err := clientset.AppsV1().Deployments(Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		data := make(map[string]interface{})
		data["deployments"] = deployments.Items
		data["namespaces"] = namespaceList

		err = conn.WriteJSON(data)
		if err != nil {
			panic(err.Error())
		}
		time.Sleep(RefreshTime * time.Second)
	}
}
