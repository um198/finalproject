package security

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/um198/finalproject/cmd/app/middleware"
	"github.com/um198/finalproject/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}


// ActivateUser -  активауия пользователья.
func (s *Service) ActivateUser(ctx context.Context, id int64) bool {
	_, err := s.pool.Exec(ctx, `UPDATE users SET active=true WHERE id = $1`, id)
	if err != nil {
		log.Print(err)
		return false
	}
	
	return true
}


// UserCodeEmail - код активации по email адрусу пользователья.
func (s *Service) UserCodeEmail(ctx context.Context, email string) (code int64) {
	err := s.pool.QueryRow(ctx, `SELECT code FROM users WHERE email = $1`, email).Scan(&code)
	if err != nil {
		log.Print(err)
		return 0
	}
	
	return code
}


// UserCode - код активации пользователья.
func (s *Service) UserCode(ctx context.Context, id int64) (code int) {
	err := s.pool.QueryRow(ctx, `SELECT code FROM users WHERE id = $1`, id).Scan(&code)
	if err != nil {
		log.Print(err)
		return 0
	}
	
	return code
}


// UserEmail - email пользователья.
func (s *Service) UserEmail(ctx context.Context, id int64) (email string) {
	err := s.pool.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, id).Scan(&email)
	if err != nil {
		log.Print(err)
		return ""
	}
	log.Println("Admin ", email)
	return email
}


// Admin ли пользователь?
func (s *Service) Admin(ctx context.Context, id int64) (Admin bool) {
	err := s.pool.QueryRow(ctx, `SELECT admin FROM users WHERE id = $1`, id).Scan(&Admin)
	if err != nil {
		log.Print(err)
		return false
	}
	
	return Admin
}


// Проверка логи на и пароль и сохранение токена.
func (s *Service) Chek(ctx context.Context, email string, password string) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id, password FROM users WHERE email =$1`, email).Scan(&id, &hash)

	if err == pgx.ErrNoRows {
		return "", model.ErrNoSuchUser
	}
	if err != nil {
		return "", model.ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Print(err)
		return "", model.ErrInvalidPassword
	}

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", model.ErrInternal
	}

	token = hex.EncodeToString(buffer)

	_, err = s.pool.Exec(ctx, `INSERT INTO users_tokens (token, user_id) VALUES ($1, $2)`, token, id)
	if err != nil {
		log.Print(err)
		return "", model.ErrInternal
	}

	return token, nil
}


// IDByToken - фукция для мидалваре.
func (s *Service) IDByToken(ctx context.Context, token string) (int64, string, error, bool) {
	var id int64
	var folder string
	var active bool

	err := s.pool.QueryRow(ctx, `
	SELECT user_id, folder, active FROM users_tokens, users WHERE 
	users.id=users_tokens.user_id and
	token = $1
	`, token).Scan(&id, &folder, &active)

	if err == pgx.ErrNoRows {
		return 0, "", nil, false
	}

	if err != nil {
		log.Print(err)
		return 0, "", model.ErrInternal, false
	}

	return id, folder, nil, active
}


// Проверки авторизации пользователья.
func Auth(writer http.ResponseWriter, request *http.Request) (int64, string) {
	id, err := middleware.Authentication(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return 0, ""
	}

	if id == 0 {
		http.Redirect(writer, request, "/login", 302)
		return 0, ""
	}

	folder, err := middleware.Folder(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return 0, ""
	}

	if id == 0 {
		http.Redirect(writer, request, "/login", 302)
		return 0, ""
	}
	
	return id, folder

}


// Активен ли пользоваетль?
func IsActive(writer http.ResponseWriter, request *http.Request) bool {

	active, err := middleware.UserActive(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return false
	}

	if !active {
		return false
	}

	return true

}


// SendConfirmCode - отправка кода активации на Email пользователья.
// Используется API сервиса Sendinblue.
// Если не подведет Sendinblue то коды должны отправиться на почту.
func SendConfirmCode(reciver, code string) (string, error) {
	in := model.Confirm{
		Sender:      &model.Email{Email: "ConfirmEmail@AlifServer.com"},
		To:          []*model.Email{{Email: reciver}},
		TextContent: "You confirmation ckde is: " + code,
		Subject:     "Confirm You Emal",
	}

	body, err := json.Marshal(in)
	if err != nil {
		log.Print(err)
		return "", err
	}

	println(string(body))

	req, err := http.NewRequest("POST", "https://api.sendinblue.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	// Как-то неправльно
	req.Header.Set("api-key", "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")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")

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
