package socket

import "github.com/satori/go.uuid"

type Body struct {
	Message    string `json:"message,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
	Token      string `json:"token,omitempty"`
}

type Headers struct {
	ReplyId string `json:"reply_id,omitempty"`
	Success bool   `json:"success"`
}

type Message struct {
	Type    string   `json:"type"`
	Id      string   `json:"id"`
	Command string   `json:"command,omitempty"`
	Headers *Headers `json:"headers"`
	Body    *Body    `json:"body,omitempty"`
}

// TODO: Errors should have codes or atleast types
func CreateError(msg string, req *Message) *Message {
	reply := &Message{Type: "error"}
	reply.Id = uuid.NewV4().String()

	if req != nil {
		reply.Headers = &Headers{ReplyId: req.Id, Success: false}
	}

	reply.Body = &Body{Message: msg}
	return reply
}

// TODO: Create a reply to a request
func CreateReply(req *Message) *Message {
	reply := &Message{Type: "reply"}
	reply.Id = uuid.NewV4().String()
	reply.Headers = &Headers{ReplyId: req.Id, Success: true}
	reply.Body = &Body{}

	return reply
}
