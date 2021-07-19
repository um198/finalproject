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


//handleFileUpload - выгрузка файла на серве
func (s *Server) handleFileUpload(writer http.ResponseWriter, request *http.Request) {
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

		err = s.filesSvc.SaveFile("files/"+folder+"/"+header.Filename, byteContainer)
		if err != nil {
			log.Print("SaveFile ", err)
			return
		}
		
		// Горутина паралельно отправляет на сервер Дроппи.
		go func() {
			_, err = dropbox.DoUpload("/"+folder+"/"+header.Filename, byteContainer)
			if err != nil {
				log.Print(err)
				return
			}

		}()

		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Загружен", header.Filename, "false", "0", "s")

		http.Redirect(writer, request, "/list", 302)
	}
}


// handleFileRestore - метод для восстановления удаленного файла.
func (s *Server) handleFileRestore(writer http.ResponseWriter, request *http.Request) {
	
	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		name := request.FormValue("name")

		// Ищем файла у себя в Корзине.
		_, err := os.Stat("files/Deleted/" + folder + "/" + name)
		if !os.IsNotExist(err) {
			println("File exist")
			err = os.Rename("files/Deleted/"+folder+"/"+name, "files/"+folder+"/"+name)
			if err != nil {
				log.Print(err)
				return
			}
			s.filesSvc.RestoreHistory(request.Context(), id, "Восстановлен")
			http.Redirect(writer, request, "/history", 302)
			return
		}

		
		if !connected() {
			fmt.Fprint(writer, "Нет Интернет подключения")
			return
		}

		// Ищем файла в Дропбоксе.
		// Если файл удален на сервре Дропбокса то можно его восстановвить в разделе Истории.
		// 
		fileToRestore, err := dropbox.DoSearch("/"+folder, name)
		if err != nil {
			log.Print(err)
			return
		}

		if fileToRestore == nil {
			fmt.Fprintln(writer, "Файл не найден в Корзине. На севере Дропбокса тоже не найден. ")
			fmt.Fprintln(writer, "Попробуйте перейти на серис Дропбокса раздел Истории и там восстановить файл. ")
			fmt.Fprintln(writer, "Только тогда можно будет восстановить файл на нашем сервере с серера Дропбокса.")
			return
		}

		if fileToRestore.Name == name {

			// Скачиваем 
			data, err := dropbox.DoDownload("/" + folder + "/" + name)
			if err != nil {
				log.Print(err)
				return
			}

			// и сохраняем
			err = s.filesSvc.SaveFile("files/"+folder+"/"+name, data)
			if err != nil {
				log.Print("SaveFile ", err)
				return
			}

			s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Загружен", id, "false", "0", "d")
			s.filesSvc.RestoreHistory(request.Context(), id, "Восстановлен")
			http.Redirect(writer, request, "/history", 302)
			return
		}
		fmt.Fprint(writer, "Not found!")
	}
}


// handleFileRename - метод для переименования файлов.
func (s *Server) handleFileRename(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		id2 := request.FormValue("id2")

		err := os.Rename("files/"+folder+"/"+id2, "files/"+folder+"/"+id)
		if err != nil {
			log.Print(err)
			return
		}

		// Горутина сразу переименовывает на стороне Дропбокса.
		go func() {
			_, err := dropbox.DoMove("/"+folder+"/"+id2, "/"+folder+"/"+id)
			if err != nil {
				log.Print(err)
				return
			}
		}()

		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Переименован", "Из "+id2+" в "+id, "true", "0", "s")
		http.Redirect(writer, request, "/list", 302)
	}
}


// handleFileDelete - удаляет файлы.
func (s *Server) handleFileDelete(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		var err error

		_, err = os.Stat("files/deleted/" + folder)
		if os.IsNotExist(err) {
			err = os.Mkdir("files/deleted/"+folder, 0755)
			println("Mkdir ")
			if err != nil {
				log.Print(err)
				return
			}
		}
		err = os.Rename("files/"+folder+"/"+id, "files/deleted/"+folder+"/"+id)
		if err != nil {
			log.Print(err)
			return
		}

		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Удален", id, "true", "0", "s")

		http.Redirect(writer, request, "/list", 302)
	}
}

// handleFilesList - список файлов пользователья.
func (s *Server) handleFilesList(writer http.ResponseWriter, request *http.Request) {

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
			err = runTemplate(writer, "templates/admin.html", list)
			if err != nil {
				os.Exit(2)
			}
			return
		}

		var data []*model.List
		files, err := ioutil.ReadDir("files/" + folder)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			fmt.Println(f.Name())
			data = append(data, &model.List{Name: f.Name()})
		}

		list := struct {
			List []*model.List
		}{
			List: data,
		}
		err= runTemplate(writer, "templates/index.html", list)
		if err != nil {
			os.Exit(2)
		}
	}

}

// handleFileDownload - скачивание файлов.
func (s *Server) handleFileDownload(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}

		id := request.FormValue("id")

		body, err := ioutil.ReadFile("files/" + folder + "/" + id)

		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "Content-Type: application/octet-stream")
		writer.Header().Set("Content-Disposition", "attachment;filename=\""+id+"\"")
		_, err = writer.Write(body)
		if err != nil {
			log.Println(err)
		}
		s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Скачан", id, "false", "0", "s")
	}
}


// Проверяем интернет
func connected() (ok bool) {
	_, err := http.Get("https://alif.academy/")
	if err != nil {
		return false
	}
	return true
}
