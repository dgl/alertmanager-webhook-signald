package signald

type Typed interface {
	Type() string
	SetType(string)
	New() interface{}
}

var typeMap = map[string]Typed{
	"version": &Version{},
	"subscribed": &Subscribed{},
}

type Request struct {
  Type string `json:"type"`
}

func (r *Request) SetType(t string) {
	r.Type = t
}

type Send struct {
	Request
	Username string `json:"username"`
	RecipientNumber string `json:"recipientNumber,omitempty"`
	RecipientGroupID string `json:"recipientGroupId,omitempty"`
	MessageBody string `json:"messageBody"`
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
	ID int `json:"id"`
	Author string `json:"author"`
	Text string `json:"text"`
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

type Response struct {
	Type string `json:"type"`
}

func (r *Response) SetType(t string) {
	r.Type = t
}

type Version struct {
	Response
	Data map[string]interface{}
}

func (s Version) Type() string {
	return "version"
}

func (s Version) New() interface{} {
	return &Version{}
}

type Subscribed struct {
	Response
	Data map[string]interface{}
}

func (s Subscribed) Type() string {
	return "subscribed"
}

func (s Subscribed) New() interface{} {
	return &Subscribed{}
}
