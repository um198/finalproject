package app

import (
	"fmt"
	"github.com/um198/finalproject/pkg/dropbox"
	"github.com/um198/finalproject/pkg/model"
	"github.com/um198/finalproject/pkg/security"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)


// handleDropboxRestore - восстановление файлов на стороне Дропбокса
func (s *Server) handleDropboxRestore(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		name := request.FormValue("name")
		rev := request.FormValue("rev")

		_, err := dropbox.DoRestore("/"+folder+"/"+name, rev)
		if err != nil {
			log.Print(err)
			return
		}

		s.filesSvc.RestoreHistory(request.Context(), id, "Восстановлен")
		http.Redirect(writer, request, "/dropbox/history", 302)

	}
}


// handleDropboxList - список файлов в папке пользователя.
func (s *Server) handleDropboxList(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		if s.securitySvc.Admin(request.Context(), authID) {
			data, err := s.filesSvc.GetUsers(request.Context())
			if err != nil {
				log.Print(err)
				return
			}
			list := struct {
				List []*model.Users
			}{
				List: data,
			}
			err = runTemplate(writer, "templates/dropbox/index2.html", list)
			if err != nil {
				os.Exit(2)
			}
			return
		}

		data, err := dropbox.DoList("/" + folder)
		if err != nil {
			log.Print(err)
			return
		}
		list := struct {
			List []*model.List
		}{
			List: data,
		}
		err = runTemplate(writer, "templates/dropbox/index.html", list)
		if err != nil {
			os.Exit(2)
		}
	}

}


//handleDropboxUpload - выгрузка файлов на сервер Дропбокса.
func (s *Server) handleDropboxUpload(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		file, header, err := request.FormFile("file")
		if err != nil {
			fmt.Fprintf(writer, "validate: %s\n", err)
			fmt.Println(err)
			return
		}
		defer file.Close()

		byteContainer, err := ioutil.ReadAll(file)

		_, err = dropbox.DoUpload("/"+folder+"/"+header.Filename, byteContainer)
		if err != nil {
			log.Print(err)
			return
		}

		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Загружен", header.Filename, "false", "0", "d")
		http.Redirect(writer, request, "/dropbox/list", 302)
	}
}


// handleDropboxDownload - загрузка файлов.
func (s *Server) handleDropboxDownload(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")

		data, err := dropbox.DoDownload("/" + folder + "/" + id)
		if err != nil {
			log.Print(err)
			return
		}
		writer.Header().Set("Content-Type", "Content-Type: application/octet-stream")
		writer.Header().Set("Content-Disposition", "attachment;filename=\""+id+"\"")
		_, err = writer.Write(data)
		if err != nil {
			log.Println(err)
		}
		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Скачан", id, "false", "0", "d")
	}
}


//handleDropboxDelete - удаление файлов с папки пользователья
func (s *Server) handleDropboxDelete(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		rev := request.FormValue("rev")

		var err error

		_, err = dropbox.DoDelete("/" + folder + "/" + id)
		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Удален", id, "true", rev, "d")

		if err != nil {
			log.Print(err)
			return
		}

		http.Redirect(writer, request, "/dropbox/list", 302)
	}
}


//handleDropboxMove - переименование (перемещение) файлов
func (s *Server) handleDropboxMove(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		id2 := request.FormValue("id2")

		_, err := dropbox.DoMove("/"+folder+"/"+id2, "/"+folder+"/"+id)
		if err != nil {
			log.Print(err)
			return
		}
		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Переименован", "Из "+id2+" в "+id, "true", "0", "d")
		http.Redirect(writer, request, "/dropbox/list", 302)
	}
}


//handleDropboxRename - переименование (перемещение) файлов
func (s *Server) handleDropboxRename(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		id2 := request.FormValue("id2")

		_, err := dropbox.DoMove("/"+folder+"/"+id2, "/"+folder+"/"+id)
		if err != nil {
			log.Print(err)
			return
		}

		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Переименован", "Из "+id2+" в "+id, "true", "0", "d")
		http.Redirect(writer, request, "/dropbox/list", 302)
	}
}


// handleDropboxHistory -  история операций файлов на стороне сервера Дропбокса
func (s *Server) handleDropboxHistory(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		historyData, _ := s.filesSvc.GetHistory(request.Context(), authID, "d")
		list := struct {
			List []*model.History
		}{
			List: historyData,
		}
		err := runTemplate(writer, "templates/dropbox/history.html", list)
		if err != nil {
			os.Exit(2)
		}

	}

}
