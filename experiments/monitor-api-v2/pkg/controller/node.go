package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/haverzard/ta/pkg/model"
	"github.com/haverzard/ta/pkg/utils"
)

func getNodeScoreByResource(node *model.NodeResource) float64 {
	return float64(node.CpuUB)/float64(node.CpuTotal)*0.7 + float64(node.MemUB)/float64(node.MemTotal)*0.3
}

type NodeController struct {
	mu            sync.Mutex
	NodeResources map[string]*model.NodeResource
}

func (nc *NodeController) CopyResources() *map[string]*model.NodeResource {
	copy := make(map[string]*model.NodeResource, len(nc.NodeResources))
	for k, v := range nc.NodeResources {
		copy[k] = v.DeepCopy()
	}
	return &copy
}

func NewNodeController() *NodeController {
	nodeRes := make(map[string]*model.NodeResource)
	return &NodeController{
		NodeResources: nodeRes,
	}
}

func (nc *NodeController) RetrieveGlobalInfo() {
	nc.mu.Lock()
	res, err := http.Get(utils.SERVER_ENDPOINT + "/cluster-info")
	if err != nil {
		log.Fatalf("Error: %v", err)
		return
	}
	d := json.NewDecoder(res.Body)
	var clusterInfo map[string]*model.NodeResource
	if err := d.Decode(&clusterInfo); err != nil {
		log.Fatalf("Error: %v", err)
		return
	}
	nc.NodeResources = clusterInfo
	log.Printf("Data: %v\n", clusterInfo)
	nc.mu.Unlock()
}

func (nc *NodeController) IsOverload() bool {
	nc.mu.Lock()
	nodeResources := nc.CopyResources()
	current := (*nodeResources)[utils.NODE_NAME]
	if current == nil {
		return false
	}
	nodeScore := getNodeScoreByResource(current)
	for nodeName, nodeRes := range *nodeResources {
		if nodeName != utils.NODE_NAME {
			score := getNodeScoreByResource(nodeRes)
			dscore := nodeScore - score
			if nodeScore > 1 && score < 1 {
				return true
			}
			if dscore > utils.OVERLOAD_THREESHOLD {
				return true
			}
		}
	}
	nc.mu.Unlock()
	return false
}
