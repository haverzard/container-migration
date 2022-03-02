package backend

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NTHU-LSALAB/DRAGON/pkg/common/migration"
	"github.com/NTHU-LSALAB/DRAGON/pkg/controller.v1/DRAGON/cluster"
)

type MigrationHandler interface {
	AddMigration(obj interface{})
}

func (hb *HaverzardBackend) HandleMigration(w http.ResponseWriter, req *http.Request) {
	d := json.NewDecoder(req.Body)
	mr := &migration.MigrationObject{}
	err := d.Decode(mr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hb.mh.AddMigration(mr)
	fmt.Fprintf(w, "hello\n")
}

func (hb *HaverzardBackend) GetClusterInfo(w http.ResponseWriter, req *http.Request) {
	nodeRes, err := cluster.SyncClusterResource()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(nodeRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(data))
}