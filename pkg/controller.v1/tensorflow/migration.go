/* haverzard */
package tensorflow

import (
	"github.com/NTHU-LSALAB/DRAGON/pkg/common/migration"
	log "github.com/sirupsen/logrus"
)

// When a migration request is received, enqueue it as migration job.
func (tc *TFController) AddMigration(obj interface{}) {
	migrationEv, ok := obj.(*migration.MigrationEvent)
	if !ok {
		log.Errorf("enqueueMigration: Cannot interpret argument obj as *MigrationEvent? Am I wrong?: %#v", obj)
		return
	}

	tc.WorkQueue.Add(migrationEv)

	return
}

/* haverzard */
