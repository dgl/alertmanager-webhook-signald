package signald

import (
	"fmt"
)

type Typed interface {
	Type() string
	SetType(string)
	SetID(int)
	New() interface{}
}

var typeMap = map[string]Typed{
	"send":       &Send{},
	"version":    &Version{},
	"subscribe":  &Subscribe{},
	"subscribed": &Subscribed{},
	"get_user":   &GetUser{},
	"user":       &User{},
}

type Request struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (r *Request) SetType(t string) {
	r.Type = t
}

func (r *Request) SetID(id int) {
	r.ID = fmt.Sprintf("%d", id)
}

type Send struct {
	Request
	Username         string `json:"username"`
	RecipientNumber  string `json:"recipientNumber,omitempty"`
	RecipientGroupID string `json:"recipientGroupId,omitempty"`
	MessageBody      string `json:"messageBody"`
	//Attachments []Attachment `json:"attachments"`
	Quote Quote `json:"quote,omitempty"`
}

func (s Send) Type() string {
	return "send"
}

func (s Send) New() interface{} {
	return &Send{}
}

type Quote struct {
	ID     int    `json:"id"`
	Author string `json:"author"`
	Text   string `json:"text"`
}

type Subscribe struct {
	Request
	Username string `json:"username"`
}

func (s Subscribe) Type() string {
	return "subscribe"
}

func (s Subscribe) New() interface{} {
	return &Subscribe{}
}

type GetUser struct {
	Request
	Username        string `json:"username"`
	RecipientNumber string `json:"recipientNumber,omitempty"`
}

func (s GetUser) Type() string {
	return "get_user"
}

func (s GetUser) New() interface{} {
	return &GetUser{}
}

type Response struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (r *Response) SetType(t string) {
	r.Type = t
}

func (r *Response) SetID(id int) {
	r.ID = fmt.Sprintf("%d", id)
}

type Version struct {
	Response
	Data map[string]string
}

func (s Version) Type() string {
	return "version"
}

func (s Version) New() interface{} {
	return &Version{}
}

type Subscribed struct {
	Response
	Data map[string]string
}

func (s Subscribed) Type() string {
	return "subscribed"
}

func (s Subscribed) New() interface{} {
	return &Subscribed{}
}

type User struct {
	Response
	Data map[string]interface{}
}

func (s User) New() interface{} {
	return &User{}
}

func (s User) Type() string {
	return "user"
}
