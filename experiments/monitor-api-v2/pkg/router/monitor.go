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
	PodCtl *controller.PodController
}

func NewMonitorRouter(podCtl *controller.PodController) *MonitorRouter {
	return &MonitorRouter{PodCtl: podCtl}
}

func RegisterMonitorRouter(mr *MonitorRouter) {
	http.HandleFunc("/monitor", mr.MonitorEvaluation)
}

func (mr *MonitorRouter) DecideMigration(pod *model.Pod) {
	originalCategory := pod.Category
	pod.Speculate()
	log.Printf("Speculated Category: %v, Original: %v", pod.Category, originalCategory)
	if pod.Category == model.Converged {
		log.Printf("Migrating pod %v", pod.Name)
		body, err := json.Marshal(map[string]string{
			"pod":  pod.Name,
			"node": utils.NODE_NAME,
		})
		if err != nil {
			log.Fatalf("Error on Migration: %v", err)
		}
		http.Post(utils.SERVER_ENDPOINT+"/hello", "application/json", bytes.NewBuffer(body))
	}
}

func (mr *MonitorRouter) MonitorEvaluation(w http.ResponseWriter, req *http.Request) {
	log.Println("Received a request")
	d := json.NewDecoder(req.Body)
	eval := &model.EvaluationObject{}
	err := d.Decode(eval)
	if err != nil {
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
