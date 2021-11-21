package main

import (
	"fmt"
	"net/http"

	"github.com/haverzard/ta/pkg/controller"
	"github.com/haverzard/ta/pkg/router"
)

func main() {
	podCtl := controller.NewPodController()
	monitor := router.NewMonitorRouter(podCtl)
	router.RegisterMonitorRouter(monitor)
	fmt.Println("Start server on port 8081")
	http.ListenAndServe(":8081", nil)
}
