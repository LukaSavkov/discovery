package service

import (
	"context"
	"fmt"
	"github.com/c12s/discovery/heartbeat"
	"github.com/c12s/discovery/storage"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strings"
	"time"
)

type Service struct {
	w       heartbeat.Heartbeat
	r       *mux.Router
	db      storage.DB
	address string
}

func (ns *Service) sub(ctx context.Context) {
	ns.w.Watch(ctx, func(msg string) {
		ns.db.Store(ctx, msg)
	})
}

func createBaseRouter(version string) *mux.Router {
	r := mux.NewRouter().StrictSlash(false)
	prefix := strings.Join([]string{"/api", version}, "/")
	return r.PathPrefix(prefix).Subrouter()
}

func (s *Service) setupEndpoints() {
	d := s.r.PathPrefix("/discovery").Subrouter()
	d.HandleFunc("/discover", s.discovery()).Methods("GET")

}

func (s *Service) discovery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["service"]; !ok {
			sendErrorMessage(w, "missing service name", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		service := r.URL.Query()["service"][0]
		address, err := s.db.Get(ctx, service)
		if err != nil {
			sendErrorMessage(w, "invalid request try again latter", http.StatusBadRequest)
			return
		}
		defer cancel()

		data, err := resp(service, address)
		if err != nil {
			sendErrorMessage(w, "invalid request try again latter", http.StatusBadRequest)
			return
		}

		sendJSONResponse(w, data)
	}
}

func Run(v, address string, db storage.DB, w heartbeat.Heartbeat) {
	server := &Service{
		db:      db,
		r:       createBaseRouter(v),
		address: address,
		w:       w,
	}
	server.setupEndpoints()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.sub(ctx)

	fmt.Println("Server Started")
	http.ListenAndServe(server.address, handlers.LoggingHandler(os.Stdout, server.r))
}
