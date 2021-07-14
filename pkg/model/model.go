package model

import (
	"errors"
	"time"
)

var ErrNoSuchUser = errors.New("no such user")
var ErrInvalidPassword = errors.New("no such user")
var ErrInternal = errors.New("internal error")
var ErrExpireToken = errors.New("token expired")
var ErrNotFound = errors.New("item not found")

type Users struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Active    bool      `json:"active"`
	Folder    string    `json:"folder"`
	Created   time.Time `json:"created"`
}

type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"mode_time"`
	IsDir   bool      `json:"is_dir"`
}

type History struct {
	ID        int64  `json:"id"`
	User_id   string `json:"user_id"`
	Operation string `json:"operation"`
	File      string `json:"file"`
	Restore   string `json:"restore"`
	Created   string `json:"created"`
	Rev       string `json:"Rev"`
}

type Metadata struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	PathDisplay    string    `json:"path_display"`
	ClientModified time.Time `json:"client_modified"`
	Size           uint64    `json:"size"`
	Tag            string    `json:".tag"`
	Rev            string    `json:"rev"`
}

type Fpath struct {
	Path string `json:"path"`
}

type List struct {
	Name string `json:"name"`
	Rev  string `json:"rev"`
	Link string `json:"link"`
}

type Fname struct {
	Entries []*List
}

type Rename struct {
	From_path string `json:"from_path"`
	To_path   string `json:"to_path"`
}

type Restore struct {
	Path string `json:"path"`
	Rev  string `json:"rev"`
}

type SearchQuery struct {
	Query   string        `json:"query"`
	Options *SearchOtions `json:"options"`
}

type SearchOtions struct {
	Path         string `json:"path"`
	FilenameOnly bool   `json:"filename_only"`
}

type SearchOutput struct {
	Matches []*SearchMatch `json:"matches"`
}

type SearchMatch struct {
	Metadata *SearchFileName `json:"metadata"`
}

type SearchFileName struct {
	Tag      string `json:".tag"`
	Metadata *Metadata
}

type Email struct {
	Email string `json:"email"`
}
type Confirm struct {
	Sender      *Email   `json:"sender"`
	To          []*Email `json:"to"`
	TextContent string   `json:"textContent"`
	Subject     string   `json:"subject"`
}

type Item struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
