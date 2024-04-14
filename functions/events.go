package functions

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Event(c *gin.Context) {

	Namespace = c.Query("namespace")

	c.HTML(http.StatusOK, "events.html", gin.H{})
}

func EventsTimeline(c *gin.Context) {

	Namespace = c.Query("namespace")

	c.HTML(http.StatusOK, "eventsTimeline.html", gin.H{})
}

func GetEvent(c *gin.Context) {
	clientset := Clientset
	namespace := c.Query("namespace")
	name := c.Query("name")
	eventData, err := clientset.CoreV1().Events(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"data": eventData})
}

func WsEvents(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err.Error())
	}
	go func() {
		pollEvents(conn)
	}()
}

func pollEvents(conn *websocket.Conn) {
	clientset := Clientset
	for {
		fmt.Print("Polling events...")

		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		namespaceList = namespaces.Items

		events, err := clientset.CoreV1().Events(Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		data := make(map[string]interface{})
		data["events"] = events.Items
		data["namespaces"] = namespaceList

		err = conn.WriteJSON(data)
		if err != nil {
			panic(err.Error())
		}
		time.Sleep(RefreshTime * time.Second)
	}
}
