package backend

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NTHU-LSALAB/DRAGON/pkg/common/migration"
)

type MigrationHandler interface {
	AddMigration(obj interface{})
}

func (hb *HaverzardBackend) hello(w http.ResponseWriter, req *http.Request) {
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
