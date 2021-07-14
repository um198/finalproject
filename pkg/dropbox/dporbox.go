package dropbox

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/um198/finalproject/pkg/model"
)

// Нужно было установить токенты в переменных окружения.
var dropToken="T63FZ88f1sIAAAAAAAAAAXFzghST_n9_jJ7HoY5X4ZW1igovminzkLvU-oWR564g"

// Ищем на стороне Дропбокса файлы.
func DoSearch(path, filename string) (*model.Metadata, error) {
	in := model.SearchQuery{Query: filename, Options: &model.SearchOtions{Path: path, FilenameOnly: true}}
	body, err := json.Marshal(in)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/search_v2", bytes.NewReader(body))
	if err != nil {
		log.Print(err)
		fmt.Println(err.Error())
		return nil, err
	}
	// Нужно было установить токенты в переменных окружения.
	//req.Header.Set("Authorization", "Bearer AHqnzf-VJ5YAAAAAAAAAAUSwsOhLNj7o7Xfha4VxgJ2vOFhle2mG4ggetVLq5muG")
	req.Header.Set("Authorization", "Bearer "+dropToken)
	req.Header.Set("Content-Type", "application/json")

	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	var file model.SearchOutput
	err = json.Unmarshal(body, &file)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var find *model.Metadata

	if len(file.Matches) != 0 {
		find = file.Matches[0].Metadata.Metadata
	}

	return find, err
}

// Скачиваем файлы.
func DoDownload(filename string) ([]byte, error) {
	body, err := Do("https://content.dropboxapi.com/2/files/download", filename, nil, true)
	return body, err
}


// Выгрузка файлов на сервер Дропбокса
func DoUpload(filename string, file []byte) (*model.List, error) {
	out := &model.List{}
	body, err := Do("https://content.dropboxapi.com/2/files/upload", filename, file, true)
	err = json.Unmarshal(body, out)

	if err != nil {
		log.Print(err)
		return nil, err
	}
	return out, err

}


// Получение списка файлов пользователья с папки на Дропбоксе.
func DoList(filename string) ([]*model.List, error) {
	out := &model.Fname{}
	body, err := Do("https://api.dropboxapi.com/2/files/list_folder", filename, nil, true)

	err = json.Unmarshal(body, out)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return out.Entries, err
}


// Восстановление файлов.
func DoRestore(path, rev string) (string, error) {
	in := model.Restore{Path: path, Rev: rev}
	body, err := json.Marshal(in)
	if err != nil {
		log.Print(err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.dropboxapi.com//2/files/restore", bytes.NewReader(body))
	if err != nil {
		log.Print(err)
		fmt.Println(err.Error())
		return "", err
	}

	//req.Header.Set("Authorization", "Bearer AHqnzf-VJ5YAAAAAAAAAAUSwsOhLNj7o7Xfha4VxgJ2vOFhle2mG4ggetVLq5muG")
	req.Header.Set("Authorization", "Bearer "+dropToken)
	
	req.Header.Set("Content-Type", "application/json")

	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return string(body), err
}


// Удаление файлов.
func DoDelete(filename string) (*model.List, error) {

	out := &model.List{}
	body, err := Do("https://api.dropboxapi.com/2/files/delete", filename, nil, true)
	err = json.Unmarshal(body, out)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	return out, err
}


// создание папки пользователья при регистрации
func DoFolder(filename string) (string, error) {
	out := &model.List{}
	body, err := Do("https://api.dropboxapi.com/2/files/create_folder_v2", filename, nil, true)
	err = json.Unmarshal(body, out)
	if err != nil {
		log.Print(err)
		return "", err
	}

	return out.Name, err
}


// Переименование или перемещение файлов
func DoMove(from_path, to_path string) (string, error) {
	in := model.Rename{From_path: from_path, To_path: to_path}
	body, err := json.Marshal(in)
	if err != nil {
		log.Print(err)
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.dropboxapi.com/2/files/move_v2", bytes.NewReader(body))
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	// req.Header.Set("Authorization", "Bearer AHqnzf-VJ5YAAAAAAAAAAUSwsOhLNj7o7Xfha4VxgJ2vOFhle2mG4ggetVLq5muG")
	req.Header.Set("Authorization", "Bearer "+dropToken)
	req.Header.Set("Content-Type", "application/json")

	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return "nil", err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return string(body), err

}


// Эскиз картинок если такие имеются на сервере.
func DoGetThumbnail(filename string) (metadata io.Reader, err error) {
	body, err := Do("https://content.dropboxapi.com/2/files/get_thumbnail", filename, nil,false)
	if err != nil {
		log.Print(err)
		return 
	}
	metadata = base64.NewDecoder(base64.StdEncoding, bytes.NewReader(body) )
	
	return
}


// Получение ссылки на файл.
func DoGetLink(filename string) (string, error) {
	out := &model.List{}
	body, err := Do("https://api.dropboxapi.com/2/files/get_temporary_link", filename, nil,true)
	err = json.Unmarshal(body, out)
	if err != nil {
		log.Print(err)
		return "", err
	}
	return out.Link, err
}

// Поделиться файлом.
func DoShare(filename string, file []byte) (*model.List, error) {
	out := &model.List{}
	body, err := Do("https://content.dropboxapi.com/sharing/create_shared_link_with_settings", filename, file,true)
	err = json.Unmarshal(body, out)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return out, err

}


// Метаданные о файле.
func DoGetMetadata(filename string) (metadata *model.Metadata, err error) {
	body, err := Do("https://api.dropboxapi.com/2/files/get_metadata", filename, nil,true)
	println(string(body))
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		log.Print(err)
		return 
	}
	return
}


// Функция для отправки запросов на API Dropbox.
func Do(path string, filename string, file []byte, hasbody bool) ([]byte, error) {
	in := model.Fpath{Path: filename}
	method := "POST"
	if strings.HasSuffix(path, "download") {
		method = "GET"
	}

	body, err := json.Marshal(in)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var req *http.Request

	
	if hasbody {
		req, err = http.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if strings.HasPrefix(path, "https://content.dropboxapi.com") {
		req.Header.Set("Dropbox-API-Arg", string(body))
	} else {
		req.Header.Set("Content-Type", "application/json")
	}

	// req.Header.Set("Authorization", "Bearer T63FZ88f1sIAAAAAAAAAAXFzghST_n9_jJ7HoY5X4ZW1igovminzkLvU-oWR564g")
	req.Header.Set("Authorization", "Bearer "+dropToken)

	if file != nil {
		closer := ioutil.NopCloser(bytes.NewReader(file))
		req.Body = closer
		req.ContentLength = int64(len(file))
		req.Header.Set("Content-Type", "application/octet-stream")
		fmt.Println("has file")
	}

	var client http.Client
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return body, err
}
