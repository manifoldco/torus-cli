package main

import "io"
import "log"
import "fmt"

type Router struct {
	client  Client
	session Session
	config  *Config
}

func NewRouter(client Client, cfg *Config, session Session) *Router {
	return &Router{client: client, config: cfg, session: session}
}

func (r *Router) process() {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("Client[%s] caught a panic: %v", r.client, p)

			// TODO: Use our own internal error objects so we can filter out
			// errors properly
			err := r.client.Write(CreateError("Internal Error", nil))
			if err != nil {
				log.Printf("Client[%s] caught err on write: %s", r.client, err)
			}
		}

		if err := r.client.Close(); err != nil {
			log.Printf("Client[%s] could not close: %s", r.client, err)
		}
	}()

	log.Printf("Client[%s] has connected", r.client)
	for {
		m, err := r.client.Read()
		if err != nil {
			if err == io.EOF {
				log.Printf("Client[%s] closed the socket", r.client)
				break
			}

			panic(err)
		}

		log.Printf("Client[%s] processing request[%s] from: %s",
			r.client, m.Id, m.Command)

		switch m.Command {
		case "status":
			err = r.status(m)
		case "get":
			err = r.get(m)
		case "set":
			err = r.set(m)
		case "logout":
			err = r.logout(m)
		case "version":
			err = r.version(m)
		default:
			msg := fmt.Sprintf("Unknown Command: %s", m.Command)
			err = r.client.Write(CreateError(msg, m))
		}

		if err != nil {
			log.Printf("Client[%s] error processing request[%s]: %s",
				r.client, m.Id, err)
			panic(err)
		}
	}
}

func (r *Router) status(m *Message) error {
	hasToken := r.session.HasToken()
	hasPassphrase := r.session.HasPassphrase()

	reply := CreateReply(m)
	reply.Body.HasToken = &hasToken
	reply.Body.HasPassphrase = &hasPassphrase

	log.Printf(
		"Client[%s] has retrieved session status: %s", r.client, r.session)
	return r.client.Write(reply)
}

func (r *Router) set(m *Message) error {
	if len(m.Body.Passphrase) == 0 && len(m.Body.Token) == 0 {
		return r.client.Write(CreateError("Missing value", m))
	}

	if len(m.Body.Passphrase) > 0 {
		log.Printf("Client[%s] has set the passphrase", r.client)
		r.session.SetPassphrase(m.Body.Passphrase)
	}
	if len(m.Body.Token) > 0 {
		log.Printf("Client[%s] has set the token", r.client)
		r.session.SetToken(m.Body.Token)
	}

	reply := CreateReply(m)

	log.Printf("Client[%s] has set the value", r.client)
	return r.client.Write(reply)
}

func (r *Router) get(m *Message) error {
	reply := CreateReply(m)
	reply.Body.Passphrase = r.session.GetPassphrase()
	reply.Body.Token = r.session.GetToken()

	log.Printf("Client[%s] has retrieved the value", r.client)
	return r.client.Write(reply)
}

func (r *Router) logout(m *Message) error {
	reply := CreateReply(m)
	r.session.Logout()

	log.Printf("Client[%s] has logged us out", r.client)
	return r.client.Write(reply)
}

func (r *Router) version(m *Message) error {
	reply := CreateReply(m)
	reply.Body.Version = version

	log.Printf(
		"Client[%s] has asked for the version: %s", r.client, r.config.Version)
	return r.client.Write(reply)
}
