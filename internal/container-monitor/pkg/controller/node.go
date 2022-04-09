package controller

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sync"

	"github.com/haverzard/ta/pkg/model"
	"github.com/haverzard/ta/pkg/utils"
)

func getNodeScoreByResource(node *model.NodeResource) float64 {
	return float64(node.CpuMaxRequest)/float64(node.CpuTotal)*0.7 + float64(node.MemMaxRequest)/float64(node.MemTotal)*0.3
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
	defer nc.mu.Unlock()
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
}

func (nc *NodeController) IsOverload() bool {
	nc.mu.Lock()
	defer nc.mu.Unlock()
	nodeResources := nc.CopyResources()
	current := (*nodeResources)[utils.NODE_NAME]
	if current == nil {
		return false
	}
	nodeScore := getNodeScoreByResource(current)
	scores := make([]float64, 0)
	sum := float64(0)
	n := 0
	for _, nodeRes := range *nodeResources {
		score := getNodeScoreByResource(nodeRes)
		scores = append(scores, score)
		sum += score
		n++
	}
	mean := sum / float64(n)
	sum = 0
	for _, score := range scores {
		dx := (score - mean)
		sum = dx * dx
	}
	std := math.Sqrt(sum / float64(n))
	return (nodeScore - mean) > (utils.OVERLOAD_THREESHOLD * std)

	// for nodeName, nodeRes := range *nodeResources {
	// 	if nodeName != utils.NODE_NAME {
	// 		score := getNodeScoreByResource(nodeRes)
	// 		dscore := nodeScore - score
	// 		if nodeScore > 1 && score < 1 {
	// 			return true
	// 		}
	// 		if dscore > utils.OVERLOAD_THREESHOLD {
	// 			return true
	// 		}
	// 	}
	// }
}
