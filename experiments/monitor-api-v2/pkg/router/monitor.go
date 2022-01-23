package router

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/haverzard/ta/pkg/controller"
	"github.com/haverzard/ta/pkg/model"
	"github.com/haverzard/ta/pkg/utils"
)

type MonitorRouter struct {
	NodeCtl *controller.NodeController
	PodCtl  *controller.PodController
}

func NewMonitorRouter(nodeCtl *controller.NodeController, podCtl *controller.PodController) *MonitorRouter {
	return &MonitorRouter{NodeCtl: nodeCtl, PodCtl: podCtl}
}

func RegisterMonitorRouter(mr *MonitorRouter) {
	http.HandleFunc("/monitor", mr.MonitorEvaluation)
}

func (mr *MonitorRouter) DecideMigration(pod *model.Pod) {
	// originalCategory := pod.Category
	pod.Speculate()
	// log.Printf("Speculated Category: %v, Original: %v", pod.Category, originalCategory)
	if pod.Category == model.Converged || mr.NodeCtl.IsOverload() {
		if pod.NextRetry != 0 {
			pod.NextRetry--
			return
		}
		log.Printf("Migrating pod %v with category %v", pod.Name, pod.Category)
		body, err := json.Marshal(map[string]string{
			"pod":  pod.Name,
			"node": utils.NODE_NAME,
		})
		if err != nil {
			log.Fatalf("Error on Migration: %v", err)
		}
		http.Post(utils.SERVER_ENDPOINT+"/migrate", "application/json", bytes.NewBuffer(body))
		pod.NextRetry = 20
	}
}

func (mr *MonitorRouter) MonitorEvaluation(w http.ResponseWriter, req *http.Request) {
	d := json.NewDecoder(req.Body)
	eval := &model.EvaluationObject{}
	if err := d.Decode(eval); err != nil {
		log.Fatalf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var pod *model.Pod = mr.PodCtl.Find(eval.Pod)
	if pod == nil {
		pod = mr.PodCtl.CreatePod(eval.Pod)
	}
	pod.AddEvaluation(eval)
	w.Write([]byte("Ok"))
	go mr.DecideMigration(pod)
}
