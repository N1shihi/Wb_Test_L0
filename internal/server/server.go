package server

import (
	"Wb_Test_L0/internal/config"
	"Wb_Test_L0/internal/handler"
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	router *gin.Engine
	cfg    *config.Config
	db     *sql.DB
	http   *http.Server
}

func New(cfg *config.Config, db *sql.DB) *Server {
	r := gin.Default()
	return &Server{
		router: r,
		cfg:    cfg,
		db:     db,
	}
}

func (s *Server) MapRoutes(h *handler.Handler) {
	s.router.GET("/order/:id", h.GetOrder)
	s.router.GET("/", h.GetHtml)
	s.router.Static("/static", "./web")
}

func (s *Server) Run(h *handler.Handler) error {
	s.MapRoutes(h)
	addr := s.cfg.Server.Port
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
