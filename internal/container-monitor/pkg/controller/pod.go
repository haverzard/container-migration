package controller

import (
	"sync"
	"time"

	"github.com/haverzard/ta/pkg/model"
)

type PodController struct {
	mu   sync.Mutex
	Pods map[string]*model.Pod
}

func NewPodController() *PodController {
	pods := make(map[string]*model.Pod)
	return &PodController{Pods: pods}
}

func (podCtl *PodController) Find(name string) *model.Pod {
	podCtl.mu.Lock()
	pod, ok := podCtl.Pods[name]
	podCtl.mu.Unlock()
	if ok {
		pod.AccessedAt = time.Now()
		return pod
	}
	return nil
}

func (podCtl *PodController) CreatePod(name string) *model.Pod {
	pod := model.NewPod(name)
	podCtl.mu.Lock()
	podCtl.Pods[name] = pod
	podCtl.mu.Unlock()
	return pod
}

func (podCtl *PodController) GarbageCollection() {
	currentTime := time.Now()

	// Copy pod mappings
	pods := make(map[string]*model.Pod)
	podCtl.mu.Lock()
	defer podCtl.mu.Unlock()
	for k, pod := range podCtl.Pods {
		if currentTime.Unix()-pod.AccessedAt.Unix() < 60 {
			pods[k] = pod
		}
	}
	podCtl.Pods = pods
}
