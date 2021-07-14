package app

import (
	
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/um198/finalproject/pkg/dropbox"
	"github.com/um198/finalproject/pkg/model"
	"github.com/um198/finalproject/pkg/security"
	"golang.org/x/crypto/bcrypt"
)



// handleActivate - активация пользователя.
func (s *Server) handleActivate(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			fcode, _ := strconv.Atoi(request.FormValue("code"))
			code := s.securitySvc.UserCode(request.Context(), authID)
			println("fcode", fcode)
			println(code)
			if fcode == code {
				s.securitySvc.ActivateUser(request.Context(), authID)
				http.Redirect(writer, request, "/list", 302)
				return
			}
			fmt.Fprintln(writer, "Invalid")
		}
	}
}


// handleActivation - страница активации пользоватлья
func (s *Server) handleActivation(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			data := s.securitySvc.UserEmail(request.Context(), authID)

			list := struct {
				List string
			}{
				List: data,
			}
			err := runTemplate(writer, "templates/activation.html", list)
			if err != nil {
				os.Exit(2)
			}
		}

	}
}


// активен ли пользователь?
func Active(writer http.ResponseWriter, request *http.Request,authID int) bool {
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			return false
		}

	}
	return true		
}


// 
func (s *Server) handleApi(writer http.ResponseWriter, request *http.Request) {

	authID, folder := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
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
		err = runTemplate(writer, "templates/api.html", list)

	}
}


// handleDeleteUser -  удаление пользователья. Доступ только Админу.
func (s *Server) handleDeleteUser(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)
	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		if s.securitySvc.Admin(request.Context(), authID) {
			id := request.FormValue("id")
			s.filesSvc.DeleteUser(request.Context(), id)
			s.filesSvc.History(request.Context(), strconv.Itoa(int(authID)), "Удалено", id, "true", "0", "0")

			http.Redirect(writer, request, "/list", 302)
		}
	}
}


// handleLogout - удаление сеанча пользователья
func (s *Server) handleLogout(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("session")
	if err != nil {
		http.Redirect(writer, request, "/", 302)
		return
	}

	cookie.MaxAge = -1
	http.SetCookie(writer, cookie)
	http.Redirect(writer, request, "/", 302)
}


// handleRestore - восстановление удаленноно файла на стороне Дропбокса.
func (s *Server) handleRestore(writer http.ResponseWriter, request *http.Request) {
	authID, folder := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		id := request.FormValue("id")
		name := request.FormValue("name")
		rev := request.FormValue("rev")

		var err error
		if s.securitySvc.Admin(request.Context(), authID) {
			_, err = dropbox.DoRestore("/"+name, rev)
		} else {
			_, err = dropbox.DoRestore("/"+folder+"/"+name, rev)
		}

		if err != nil {
			log.Print(err)
			return
		}

		s.filesSvc.RestoreHistory(request.Context(), id, "Восстановлен")
		http.Redirect(writer, request, "/history", 302)

	}
}


// handleHistory - история.
func (s *Server) handleHistory(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		historyData, _ := s.filesSvc.GetHistory(request.Context(), authID, "s")
		list := struct {
			List []*model.History
		}{
			List: historyData,
		}
		err := runTemplate(writer, "templates/history.html", list)
		if err != nil {
			os.Exit(2)
		}

	}

}


// handleChekEmail - проверка Email адреса при регистрации.
func (s *Server) handleChekEmail(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	email := string(body)

	if email == "" {
		return
	}

	hasMail, err := s.filesSvc.ChekEmail(request.Context(), email)
	if err != nil || hasMail == "" {
		fmt.Fprint(writer, "false")
		return
	}

	fmt.Fprint(writer, "true")

}


//handleLoginProcess - процесс логина.
func (s *Server) handleLoginProcess(writer http.ResponseWriter, request *http.Request) {

	email := request.FormValue("email")
	password := request.FormValue("paswword")

	if email == "" || password == "" {
		fmt.Fprintln(writer, "Invalid")
		return
	}

	token, err := s.securitySvc.Chek(request.Context(), email, password)
	if err != nil || token == "" {
		fmt.Fprintln(writer, "Invalid")
		return
	}

	cookie := &http.Cookie{
		Name:   "session",
		Value:  token,
		Path:   "/",
		MaxAge: 60 * 60 * 24,
	}
	http.SetCookie(writer, cookie)

	http.Redirect(writer, request, "/list", 302)

}


// handleSingup - регистрация.
func (s *Server) handleSingup(writer http.ResponseWriter, request *http.Request) {
	err := runTemplate(writer, "templates/register.html", nil)
	if err != nil {
		os.Exit(2)
	}
}


//handleLogin - логин.
func (s *Server) handleLogin(writer http.ResponseWriter, request *http.Request) {
	err := runTemplate(writer, "templates/login.html", nil)
	if err != nil {
		os.Exit(2)
	}
}


//
func (s *Server) handleHome(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			// http.Redirect(writer, request, "/activation", 302)
		}
		http.Redirect(writer, request, "/list", 302)
	}

}


// handleRegistration - генерируем токен и сразу записываем в куки.
func (s *Server) handleRegistration(writer http.ResponseWriter, request *http.Request) {

	item := &model.Users{
		FirstName: request.FormValue("firstname"),
		LastName:  request.FormValue("lastname"),
		Email:     request.FormValue("email"),
		Password:  request.FormValue("paswword"),
	}

	if item.FirstName == "" || item.LastName == "" || item.Email == "" || item.Password == "" {
		fmt.Fprintln(writer, "Invalid")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item.Password = string(hashed)

	token, err := s.filesSvc.Save(request.Context(), item)
	if err != nil || token == "" {
		log.Println(err)
		http.Redirect(writer, request, "/login", 302)
		return
	}

	cookie := &http.Cookie{
		Name:   "session",
		Value:  token,
		Path:   "/",
		MaxAge: 60 * 60 * 24,
	}
	http.SetCookie(writer, cookie)

	http.Redirect(writer, request, "/list", 302)

}


// runTemplate - запуск HTML шаблонов.
func runTemplate(writer http.ResponseWriter, templ string, list interface{}) error {
	tmpl, err := template.ParseFiles(templ)
	if err != nil {
		log.Println(err)
		log.Print("Не могу найти шаблоны. Убедитесь, что находитесь в корневом каталоге проекта.")
		return err
	}
	err = tmpl.Execute(writer, list)
	if err != nil {
		log.Println(err)
		log.Print("Не могу найти шаблоны. Убедитесь, что находитесь в корневом каталоге проекта.")
		return err
	}

	return nil

}


// handleDropbox
func (s *Server) handleDropbox(writer http.ResponseWriter, request *http.Request) {
	authID, _ := security.Auth(writer, request)

	if authID != 0 {
		active := security.IsActive(writer, request)
		if !active {
			http.Redirect(writer, request, "/activation", 302)
		}
		f := request.URL.Query().Get("func")
		name := request.URL.Query().Get("name")
		if f == "getmetadata" {
			GetMetadata(writer, name)
		}

		if f == "get_thumbnail" {
			d, err := dropbox.DoGetThumbnail(name)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Fprintln(writer, d)
		}

	}
}


// GetMetadata
func GetMetadata(writer http.ResponseWriter, name string) {
	metadata, err := dropbox.DoGetMetadata(name)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprintln(writer, metadata.ID+"<br>")
	fmt.Fprintln(writer, "File name: ", metadata.Name, "<br>")
	fmt.Fprintln(writer, "Display name: ", metadata.PathDisplay, "<br>")
	fmt.Fprintln(writer, "ClientModified: ", metadata.ClientModified, "<br>")
	fmt.Fprintln(writer, "Size: ", metadata.Size, "<br>")
}
