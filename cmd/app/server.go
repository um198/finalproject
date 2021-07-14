package app

import (
	"github.com/gorilla/mux"
	"github.com/um198/finalproject/cmd/app/middleware"
	"github.com/um198/finalproject/pkg/files"
	"github.com/um198/finalproject/pkg/security"
	"net/http"
)

type Server struct {
	mux         *mux.Router
	securitySvc *security.Service
	filesSvc    *files.Service
}

func NewServer(mux *mux.Router, securitySvc *security.Service, filesSvc *files.Service) *Server {
	return &Server{
		mux:         mux,
		securitySvc: securitySvc,
		filesSvc:    filesSvc,
	}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

const (
	GET    = "GET"
	POST   = "POST"
	DELETE = "DELETE"
)

func (s *Server) Init() {
	s.mux.Use(middleware.Logger)
	authenticateMd := middleware.Authenticate(s.securitySvc.IDByToken)
	s.mux.Use(authenticateMd)

	s.mux.HandleFunc("/", s.handleHome).Methods(GET)
	s.mux.HandleFunc("/login", s.handleLogin).Methods(GET)
	s.mux.HandleFunc("/2/login", s.handleLoginProcess).Methods(POST)
	s.mux.HandleFunc("/logout", s.handleLogout).Methods(GET)
	s.mux.HandleFunc("/list", s.handleFilesList).Methods(GET)
	s.mux.HandleFunc("/register", s.handleRegistration).Methods(POST)
	s.mux.HandleFunc("/signup", s.handleSingup).Methods(GET)
	s.mux.HandleFunc("/chekemail", s.handleChekEmail).Methods(POST)
	s.mux.HandleFunc("/history", s.handleHistory).Methods(GET)
	s.mux.HandleFunc("/killuser", s.handleDeleteUser).Methods(POST)
	
	s.mux.HandleFunc("/activation", s.handleActivation).Methods(GET)
	s.mux.HandleFunc("/activate", s.handleActivate).Methods(POST)
	
	// Файлы на нашем сервере
	filesSubrouter := s.mux.PathPrefix("/api/files").Subrouter()
	filesSubrouter.HandleFunc("/getcode", s.handleGetConfirmationCode).Methods(POST)
	filesSubrouter.HandleFunc("/get_info", s.handleFileInfo).Methods(POST)
	filesSubrouter.HandleFunc("/download_zip", s.handleFileDownloadZip).Methods(POST)
	filesSubrouter.HandleFunc("/delete", s.handleFileDelete).Methods(POST)
	filesSubrouter.HandleFunc("/download", s.handleFileDownload).Methods(GET)
	filesSubrouter.HandleFunc("/upload", s.handleFileUpload).Methods(POST)
	filesSubrouter.HandleFunc("/move", s.handleFileRename).Methods(POST)
	filesSubrouter.HandleFunc("/restore", s.handleFileRestore).Methods(POST)

	// Файлы на сервере Дропбокса
	dropboxSubrouter := s.mux.PathPrefix("/dropbox").Subrouter()
	dropboxSubrouter.HandleFunc("/list", s.handleDropboxList).Methods(GET)
	dropboxSubrouter.HandleFunc("/delete", s.handleDropboxDelete).Methods(POST)
	dropboxSubrouter.HandleFunc("/download", s.handleDropboxDownload).Methods(GET)
	dropboxSubrouter.HandleFunc("/upload", s.handleDropboxUpload).Methods(POST)
	dropboxSubrouter.HandleFunc("/move", s.handleDropboxRename).Methods(POST)
	dropboxSubrouter.HandleFunc("/history", s.handleDropboxHistory).Methods(GET)
	dropboxSubrouter.HandleFunc("/restore", s.handleDropboxRestore).Methods(POST)
}
