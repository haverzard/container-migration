/* haverzard */
package backend

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type MigrationBackend struct {
	mh MigrationHandler
}

/*
	Initialize migration microservice
*/
func NewMigrationBackend(mh MigrationHandler) *MigrationBackend {
	return &MigrationBackend{mh}
}

/*
	Start migration microservice
*/
func (mb *MigrationBackend) Start(stopCh <-chan struct{}) {
	// Register endpoints
	http.HandleFunc("/migrate", mb.HandleMigration)
	http.HandleFunc("/cluster-info", mb.GetClusterInfo)

	// Start server
	server := &http.Server{Addr: ":8769"}
	go func() {
		server.ListenAndServe()
	}()

	log.Info("Starting backend server")
	<-stopCh
	log.Info("Shutting down backend server")

	// Shutdown server upon termination
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Error on shutting down backend server")
	}
}

/* haverzard */
