package client

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func New(conn net.Conn) *Client {
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (c *Client) ReadResponse() (int, string, error) {
	return ftp.ReadResponse(c.reader)
}

func (c *Client) Cmd(cmd string) (int, string, error) {
	if err := ftp.SendCmd(c.conn, cmd); err != nil {
		return 0, "", err
	}
	return c.ReadResponse()
}

func (c *Client) OpenDataConn() (net.Listener, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	args := ftp.BuildPORTArgs(lis.Addr())
	code, msg, err := c.Cmd("PORT " + args)
	if err != nil {
		lis.Close()
		return nil, fmt.Errorf("PORT failed: %v", err)
	}
	if code != ftp.StatusOK {
		lis.Close()
		return nil, fmt.Errorf("PORT failed: %s", msg)
	}
	return lis, nil
}

func (c *Client) dataTransfer(cmd string) ([]byte, error) {
	lis, err := c.OpenDataConn()
	if err != nil {
		return nil, err
	}
	defer lis.Close()

	code, msg, err := c.Cmd(cmd)
	if err != nil {
		return nil, err
	}
	if code != ftp.StatusStartingDataTransfer && code != ftp.StatusAlreadyOpen {
		return nil, fmt.Errorf("unexpected response: %s", msg)
	}

	dc, err := lis.Accept()
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dc)
	dc.Close()

	c.ReadResponse()
	return data, err
}

func (c *Client) List() ([]byte, error) {
	return c.dataTransfer("NLST")
}

func (c *Client) Get(filename string) ([]byte, error) {
	return c.dataTransfer("RETR " + filename)
}

func (c *Client) Put(filename string, data []byte) error {
	lis, err := c.OpenDataConn()
	if err != nil {
		return err
	}
	defer lis.Close()

	code, msg, err := c.Cmd("STOR " + filename)
	if err != nil {
		return err
	}
	if code != ftp.StatusStartingDataTransfer && code != ftp.StatusAlreadyOpen {
		return fmt.Errorf("unexpected response: %s", msg)
	}

	dc, err := lis.Accept()
	if err != nil {
		return err
	}
	dc.Write(data)
	dc.Close()

	c.ReadResponse()
	return nil
}

func (c *Client) Close() {
	c.Cmd("QUIT")
	c.conn.Close()
}
