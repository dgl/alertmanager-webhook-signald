package signald

type Typed interface {
	Type() string
	SetType(string)
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
	RecipientNumber string `json:"recipientNumber"`
	RecipientGroupID string `json:"recipientGroupId"`
	MessageBody string `json:"messageBody"`
	//Attachments []Attachment `json:"attachments"`
	Quote Quote `json:"quote"`
}

func (s Send) Type() string {
	return "send"
}

type Quote struct {
	ID int `json:"id"`
	Author string `json:"author"`
	Text string `json:"text"`
}

type Response struct {
	Type string `json:"type"`
}
