package controller

import (
	"time"

	"github.com/haverzard/ta/pkg/model"
)

type PodController struct {
	Pods map[string]*model.Pod
}

func NewPodController() *PodController {
	pods := make(map[string]*model.Pod)
	return &PodController{Pods: pods}
}

func (podCtl *PodController) Find(name string) *model.Pod {
	pod, ok := podCtl.Pods[name]
	if ok {
		pod.AccessedAt = time.Now()
		return pod
	}
	return nil
}

func (podCtl *PodController) CreatePod(name string) *model.Pod {
	pod := model.NewPod(name)
	podCtl.Pods[name] = pod
	return pod
}

func (podCtl *PodController) GarbageCollection() {
	time := time.Now()

	// Copy pod mappings
	pods := make(map[string]*model.Pod)
	for k, pod := range podCtl.Pods {
		if time.Unix()-pod.AccessedAt.Unix() < 3600 {
			pods[k] = pod
		}
	}
}
