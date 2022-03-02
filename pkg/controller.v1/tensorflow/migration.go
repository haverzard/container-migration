package tensorflow

import (
	"github.com/NTHU-LSALAB/DRAGON/pkg/common/migration"
	log "github.com/sirupsen/logrus"
)

// When a pod is created, enqueue the job that manages it and update its expectations.
func (tc *TFController) AddMigration(obj interface{}) {
	migrationObj, ok := obj.(*migration.MigrationObject)
	if !ok {
		log.Errorf("enqueueMigration: Cannot interpret argument migrationObject as *MigrationObject? Am I wrong?: %#v", obj)
		return
	}

	tc.WorkQueue.Add(migrationObj)

	return
}
