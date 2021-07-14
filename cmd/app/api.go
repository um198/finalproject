package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/um198/finalproject/pkg/model"
	"github.com/um198/finalproject/pkg/security"
)


// 
func (s *Server) handleFileDownloadZip(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		stat := model.FileInfo{}

		err := json.NewDecoder(request.Body).Decode(&stat)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)
		body, err := ioutil.ReadFile("files/" + stat.Name)

		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		zipFile, err := zipWriter.Create(stat.Name)
		if err != nil {
			fmt.Println(err)
		}

		_, err = zipFile.Write(body)
		if err != nil {
			fmt.Println(err)
		}

		err = zipWriter.Close()
		if err != nil {
			fmt.Println(err)
		}

		ioutil.WriteFile(stat.Name+".zip", buf.Bytes(), 0777)
		writer.Header().Set("Content-Type", "Content-Type: application/octet-stream")
		writer.Header().Set("Content-Disposition", "attachment;filename=\""+stat.Name+".zip"+"\"")
		_, err = writer.Write(buf.Bytes())
		if err != nil {
			log.Println(err)
		}

	}
}


// handleGetConfirmationCode - получение кода конфирмации, если по каким-то причинам
// сервис отправки кода потвержления не смог отпрвить код. Только в целях тестирования.
func (s *Server) handleGetConfirmationCode(writer http.ResponseWriter, request *http.Request) {
	

		stat := model.Email{}

		err := json.NewDecoder(request.Body).Decode(&stat)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		code:=s.securitySvc.UserCodeEmail(request.Context(),stat.Email)

		status := map[string]string{"code": strconv.Itoa(int(code)) }
		writeData(writer, status, err, 0)
	

}


// handleFileInfo  - информация о файле
func (s *Server) handleFileInfo(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		stat := model.FileInfo{}

		err := json.NewDecoder(request.Body).Decode(&stat)
		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		fileStat, err := os.Stat("files/" + stat.Name)
		if err != nil {
			log.Fatal(err)
		}

		stat = model.FileInfo{
			Name:    fileStat.Name(),
			Size:    fileStat.Size(),
			Mode:    fileStat.Mode().String(),
			ModTime: fileStat.ModTime(),
			IsDir:   fileStat.IsDir(),
		}

		if err != nil {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		writeData(writer, stat, err, 0)
	}

}



 
// Отправка ответа клиенту.
func writeData(writer http.ResponseWriter, item interface{}, err error, code int) {

	if errors.Is(err, model.ErrNotFound) {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if code != 0 {
		writer.WriteHeader(code)
	}
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}

	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
