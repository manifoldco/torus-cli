package socket

import "strconv"
import "net"
import "encoding/json"

type Connection struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
	id      int
}

type Client interface {
	Read() (*Message, error)
	Write(m *Message) error
	Close() error
	String() string
}

func NewConnection(c net.Conn, id int) Client {
	return &Connection{
		conn:    c,
		id:      id,
		encoder: json.NewEncoder(c),
		decoder: json.NewDecoder(c),
	}
}

func (c *Connection) Read() (*Message, error) {
	m := &Message{}
	if err := c.decoder.Decode(m); err != nil {
		return nil, err
	}

	return m, nil
}

func (c *Connection) Write(m *Message) error {
	if err := c.encoder.Encode(m); err != nil {
		return err
	}

	return nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) String() string {
	return strconv.Itoa(c.id)
}
