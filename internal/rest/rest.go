package rest

import (
	"net/http"

	"github.com/vpomo/industrial-mcp/internal/domain/repository"
	"github.com/vpomo/industrial-mcp/internal/domain/service"
	"github.com/vpomo/industrial-mcp/internal/infrastructure/driver"
)

type Server struct {
	mux               *http.ServeMux
	dataSourceHandler *DataSourceHandler
	exposedTagHandler *ExposedTagHandler
}

func NewServer(
	dsRepo repository.DataSourceRepository,
	etRepo repository.ExposedTagRepository,
) *Server {
	driverMgr := driver.NewDriverManager()
	driverMgr.Register(driver.NewOPCUADriver())
	driverMgr.Register(driver.NewMQTTDriver())
	driverMgr.Register(driver.NewModbusDriver())
	driverMgr.Register(driver.NewBACnetDriver())

	dsService := service.NewDataSourceService(dsRepo, driverMgr)
	etService := service.NewExposedTagService(etRepo, dsService)

	return &Server{
		mux:               http.NewServeMux(),
		dataSourceHandler: NewDataSourceHandler(dsService),
		exposedTagHandler: NewExposedTagHandler(etService),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) RegisterRoutes() {
	s.mux.HandleFunc("POST /api/v1/data-sources", s.dataSourceHandler.Create)
	s.mux.HandleFunc("GET /api/v1/data-sources", s.dataSourceHandler.List)
	s.mux.HandleFunc("GET /api/v1/data-sources/get", s.dataSourceHandler.Get)
	s.mux.HandleFunc("DELETE /api/v1/data-sources", s.dataSourceHandler.Delete)
	s.mux.HandleFunc("POST /api/v1/data-sources/connect", s.dataSourceHandler.Connect)
	s.mux.HandleFunc("POST /api/v1/data-sources/disconnect", s.dataSourceHandler.Disconnect)
	s.mux.HandleFunc("POST /api/v1/data-sources/scan", s.dataSourceHandler.Scan)

	s.mux.HandleFunc("POST /api/v1/tags", s.exposedTagHandler.Create)
	s.mux.HandleFunc("GET /api/v1/tags", s.exposedTagHandler.List)
	s.mux.HandleFunc("GET /api/v1/tags/get", s.exposedTagHandler.Get)
	s.mux.HandleFunc("DELETE /api/v1/tags", s.exposedTagHandler.Delete)
	s.mux.HandleFunc("GET /api/v1/tags/read", s.exposedTagHandler.ReadValue)
	s.mux.HandleFunc("POST /api/v1/tags/write", s.exposedTagHandler.WriteValue)
}
