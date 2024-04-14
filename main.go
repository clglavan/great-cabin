package main

import (
	"github.com/clglavan/great-cabin/functions"
	"github.com/gin-gonic/gin"
)

func main() {
	functions.InitializeCluster()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")

	r.GET("/", functions.Overview)
	r.GET("/wsOverview", functions.WsOverview)

	r.GET("/pods", functions.Pods)
	r.GET("/wsPods", functions.WsPods)
	r.GET("/pod", functions.GetPod)
	r.POST("/pod", functions.UpdatePod)
	r.DELETE("/pod", functions.DeletePod)
	r.GET("/podLogs", functions.PodLogs)

	r.GET("/deployments", functions.Deployment)
	r.GET("/deployment", functions.GetDeployment)
	r.POST("/deployment", functions.UpdateDeployment)
	r.DELETE("/deployment", functions.DeleteDeployment)
	r.GET("/wsDeployments", functions.WsDeployments)

	r.GET("/event", functions.GetEvent)
	r.GET("/events", functions.Event)
	r.GET("/events-timeline", functions.EventsTimeline)
	r.GET("/wsEvents", functions.WsEvents)

	r.Run()
}
