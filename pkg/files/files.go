package files

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/um198/finalproject/pkg/dropbox"
	"github.com/um198/finalproject/pkg/model"
	"github.com/um198/finalproject/pkg/security"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}


// SaveFile - сохранение файла на диск.
func (s *Service) SaveFile(filename string, file []byte) (err error) {
	dst, _ := os.Create(filename)
	defer dst.Close()
	_, err = dst.Write(file)

	if err != nil {
		return err
	}
	return nil
}


// GetUsers - получение пользователей админов. Отображаются на админпанельке.
func (s *Service) GetUsers(ctx context.Context) ([]*model.Users, error) {
	items := make([]*model.Users, 0)
	query := "SELECT id, firstname, lastname, active, email, folder FROM users WHERE admin=false"

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println(err)
			return items, nil
		}
		return nil, model.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &model.Users{}
		err = rows.Scan(&item.ID, &item.FirstName, &item.LastName, &item.Active, &item.Email, &item.Folder)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	if err == pgx.ErrNoRows {
		return nil, model.ErrNoSuchUser
	}

	if err != nil {
		log.Print(err)
		return nil, model.ErrInternal
	}
	return items, nil
}


// случайне цифра, коды акторизации
func digits() int64 {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}
	return n.Int64()
}


// Save - сохранение пользователья.
func (s *Service) Save(ctx context.Context, item *model.Users) (string, error) {
	var err error
	folder := uuid.New().String()
	code:=digits()
	active:="false"
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (firstname, lastname, email, password,folder,code, active) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;
		`, item.FirstName, item.LastName, item.Email, item.Password, folder, code,active).Scan(&item.ID)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", model.ErrInternal
	}

	if err != nil {
		log.Print(err)
		return "", model.ErrNotFound
	}

	dropbox.DoFolder("/" + folder)

	err = os.Mkdir("files/"+folder, 0755)
	if err != nil {
		log.Print(err)
		return "", model.ErrInternal
	}

	r,err:=security.SendConfirmCode(item.Email,strconv.Itoa(int(code)))
	if err != nil {
		log.Print(err)
		return "", model.ErrInternal
	}

	println(r)

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", model.ErrInternal
	}

	token := hex.EncodeToString(buffer)

	_, err = s.pool.Exec(ctx, `INSERT INTO users_tokens (token, user_id) VALUES ($1, $2)`, token, item.ID)
	if err != nil {
		log.Print(err)
		return "", model.ErrInternal
	}

	return token, nil
}


// Удаление.
func (s *Service) DeleteUser(ctx context.Context, id string) (Admin bool) {

	_, err := s.pool.Exec(ctx, `DELETE FROM users_tokens WHERE user_id = $1`, id)
	if err != nil {
		log.Fatal(err)
		return false
	}

	_, err = s.pool.Exec(ctx, `DELETE FROM history WHERE user_id = $1;`, id)
	if err != nil {
		log.Fatal(err)
		return false
	}
	_, err = s.pool.Exec(ctx, `DELETE FROM users WHERE id = $1;`, id)
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}


// Получение исрории.
func (s *Service) GetHistory(ctx context.Context, id int64, server string) ([]*model.History, error) {

	items := make([]*model.History, 0)

	query := "SELECT id, user_id, operation, created, file, restore, rev FROM history WHERE server='"+
	server+"' AND user_id ='" + 
	strconv.Itoa(int(id)) + "' ORDER BY created DESC"

	rows, err := s.pool.Query(ctx, query)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Print(err)
			return items, nil
		}

		return nil, model.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &model.History{}
		err = rows.Scan(&item.ID, &item.User_id, &item.Operation, &item.Created, 
			&item.File, &item.Restore, &item.Rev)
		if err != nil {
			log.Print(err)
			return nil, err
		}

		items = append(items, item)
	}
	if err == pgx.ErrNoRows {
		return nil, model.ErrNoSuchUser
	}
	if err != nil {
		log.Print(err)
		return nil, model.ErrInternal
	}
	return items, nil
}


// Обновление статуса удаленного файла на востановленный при востановление.
func (s *Service) RestoreHistory(ctx context.Context, id, operation string) error {

	_, err := s.pool.Exec(ctx, `
	UPDATE history SET 
	operation=$1 WHERE id=$2`, operation, id)

	if err == pgx.ErrNoRows {
		return model.ErrNoSuchUser
	}
	if err != nil {
		log.Print(err)
		return model.ErrInternal
	}
	return nil
}


// History  - Добавление истории операций пользователья в базу данных.
func (s *Service) History(ctx context.Context, user_id, operation, file, restore, rev, server string) error {
	loc, _ := time.LoadLocation("Asia/Tashkent")
	cTime := time.Now().In(loc).Format("02.01.2006 15:04:05")
	println(cTime)
	_, err := s.pool.Exec(ctx, `
	INSERT INTO history (user_id, operation, file, restore, created, rev, server) 
	VALUES ($1, $2, $3,$4, $5,$6,$7)`, user_id, operation, file, restore, cTime, rev, server)

	if err == pgx.ErrNoRows {
		return model.ErrNoSuchUser
	}
	if err != nil {
		log.Print(err)
		return model.ErrInternal
	}
	return nil
}


// ChekEmail - проверка Email понадобиться на форме ругистрации пользователья
func (s *Service) ChekEmail(ctx context.Context, email string) (string, error) {
	var hasEmail string
	println(email)
	err := s.pool.QueryRow(ctx, `SELECT email FROM users WHERE email =$1`, email).Scan(&hasEmail)
	println(hasEmail)
	if err == pgx.ErrNoRows {
		return "", model.ErrNoSuchUser
	}
	if err != nil {
		log.Print(err)
		return "", model.ErrInternal
	}
	return hasEmail, nil
}
