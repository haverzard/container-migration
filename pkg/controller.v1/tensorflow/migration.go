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
	// enqueue := false
	// if jc.Option.KubeShareSupport {
	// 	enqueue = false
	// 	if val, okk := job.GetAnnotations()["DRAGON_KUBESHARE"]; okk && val == "true" {
	// 		enqueue = true
	// 	}
	// } else {
	// 	enqueue = true
	// 	if val, okk := job.GetAnnotations()["DRAGON_KUBESHARE"]; okk && val == "true" {
	// 		enqueue = false
	// 	}
	// }
	// if !enqueue {
	// 	return
	// }

	// jobKey, err := controller.KeyFunc(job)
	// if err != nil {
	// 	logger.Infof("Failed to get the jobkey: %v", err)
	// 	return
	// }

	// if _, ok := pod.Labels[jc.Controller.GetReplicaTypeLabelKey()]; !ok {
	// 	logger.Infof("This pod maybe not created by %v", jc.Controller.ControllerName())
	// 	return
	// }

	// rtype := pod.Labels[jc.Controller.GetReplicaTypeLabelKey()]
	// expectationPodsKey := GenExpectationPodsKey(jobKey, rtype)

	// jc.Expectations.CreationObserved(expectationPodsKey)
	// TODO: we may need add backoff here
	tc.WorkQueue.Add(migrationObj)

	return
}
