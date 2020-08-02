package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	http *http.Server
}

func New(port int, healthChecks ...func() error) *Service {
	return &Service{
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: buildHandler(healthChecks),
		},
	}
}

func (s *Service) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Info("health check Service: begin run")

	go func() {
		defer wg.Done()
		log.Debug("health check service addr:", s.http.Addr)
		err := s.http.ListenAndServe()
		log.Info("health check service: end run > ", err)
	}()

	go func() {
		<-ctx.Done()
		sdCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		err := s.http.Shutdown(sdCtx)
		if err != nil {
			log.Info("health check service shutdown (", err, ")")
		}
	}()
}

func buildHandler(healthChecks []func() error) http.Handler {
	handler := http.NewServeMux()
	var checks = func(w http.ResponseWriter, _ *http.Request) { serveCheck(w, healthChecks) }
	handler.HandleFunc("/", checks)
	handler.HandleFunc("/health", checks)
	handler.HandleFunc("/ready", checks)
	return handler
}

func serveCheck(w http.ResponseWriter, checks []func() error) {
	writtenHeader := false
	for _, check := range checks {
		if err := check(); err != nil {
			if !writtenHeader {
				w.WriteHeader(http.StatusInternalServerError)
				writtenHeader = true
			}
			w.Write([]byte(err.Error())) // nolint
			w.Write([]byte("\n\n"))      // nolint
		}
	}

	if !writtenHeader {
		w.WriteHeader(http.StatusNoContent)
	}
}
