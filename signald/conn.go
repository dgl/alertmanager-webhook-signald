// Package signald implements a simple client for the Signald protocol.
package signald

import (
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"strings"
)

// Unfortunately signald-go is not written in idiomatic Go and panics if unable
// to connect, meaning restarts of signald will kill the client process too,
// without complex panic handling. This client manages reconnecting if needed.

const (
  DefaultPath = "/var/run/signald/signald.sock"
)

// Client represents a connection to a signald server.
type Client struct {
  path string
  conn net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
}

// New returns a client and attempts to connect on DefaultPath.
// If an error is returned the client is still valid and calling Encode will
// attempt to reconnect.
func New() (*Client, error) {
  return NewPath(DefaultPath)
}

// NewPath is as new, but allows specifying the path to connect to.
func NewPath(path string) (*Client, error) {
  client := &Client{
    path: path,
  }
  return client, client.connect()
}

func (c *Client) connect() error {
	c.conn = nil
  conn, err := net.Dial("unix", c.path)
  if err != nil {
		return err
	}
	c.conn = conn
	c.encoder = json.NewEncoder(c.conn)
	c.decoder = json.NewDecoder(c.conn)
	return nil
}

func (c *Client) Encode(req interface{}) error {
	 request, ok := req.(*Request)
	 if !ok {
		return errors.New("Argument to Encode not convertable to a *signald.Request")
	}
	t := strings.ToLower(reflect.TypeOf(req).Name())
	request.Type = t
	if c.conn == nil {
		err := c.connect()
		if err != nil {
			return err
		}
	}
	return c.encoder.Encode(req)
}

func (c *Client) Decode(res *Response) error {
	if c.decoder == nil {
		// We only connect when trying to send
		return errors.New("Not connected")
	}
	// XXX: need to interrupt if reconnected
	return c.decoder.Decode(res)
}
