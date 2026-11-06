package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Client struct {
	socketPath string
}

func NewClient(socketPath ...string) *Client {
	sock := GetSocketPath()
	if len(socketPath) > 0 && socketPath[0] != "" {
		sock = socketPath[0]
	}
	return &Client{socketPath: sock}
}

func (c *Client) Send(req Request) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("could not connect to quickcast daemon at %s: %w", c.socketPath, err)
	}
	defer conn.Close()

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	if _, err := conn.Write(data); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var resp Response
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, scanner.Err()
}
