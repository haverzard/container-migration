package backend

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type HaverzardBackend struct {
	mh MigrationHandler
}

func NewHaverzardBackend(mh MigrationHandler) *HaverzardBackend {
	return &HaverzardBackend{mh}
}

func (hb *HaverzardBackend) Start(stopCh <-chan struct{}) {
	http.HandleFunc("/migrate", hb.HandleMigration)
	http.HandleFunc("/cluster-info", hb.GetClusterInfo)
	server := &http.Server{Addr: ":8769"}
	go func() {
		server.ListenAndServe()
	}()

	log.Info("Starting backend server")
	<-stopCh
	log.Info("Shutting down backend server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Error on shutting down backend server")
	}
}
