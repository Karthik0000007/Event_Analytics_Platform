package main

import (
	"net/http"

	"github.com/Karthik0000007/Event_Analytics_Platform/internal/api"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/config"
	"github.com/Karthik0000007/Event_Analytics_Platform/internal/logging"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.ServiceName)

	logger.Info("starting service", map[string]any{
		"port": cfg.Port,
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: api.NewRouter(),
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("server stopped", map[string]any{
			"error": err.Error(),
		})
	}
}
