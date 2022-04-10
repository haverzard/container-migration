package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/haverzard/ta/internal/ticker"
	"github.com/haverzard/ta/pkg/controller"
	"github.com/haverzard/ta/pkg/router"
	"github.com/haverzard/ta/pkg/utils"
)

func RunBackgroundTask(nodeCtl *controller.NodeController, podCtl *controller.PodController) {
	log.Printf("%v - just ticked", time.Now())
	jt := ticker.NewJobTicker(0, 1, 0)
	for {
		<-jt.T.C
		log.Println("Getting cluster info...")
		nodeCtl.RetrieveGlobalInfo()
		podCtl.GarbageCollection()
		jt.UpdateJobTicker()
	}
}

func main() {
	utils.LoadEnv()
	podCtl := controller.NewPodController()
	nodeCtl := controller.NewNodeController()
	monitor := router.NewMonitorRouter(nodeCtl, podCtl)
	router.RegisterMonitorRouter(monitor)
	go RunBackgroundTask(nodeCtl, podCtl)
	fmt.Println("Start server on port 8081")
	http.ListenAndServe(":8081", nil)
}
