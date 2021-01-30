package echo

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"

	"golang.org/x/sync/errgroup"
)

type Client struct {
	Addr string
	Done chan struct{}

	input io.ReadCloser
	conn  net.Conn
	group *errgroup.Group
}

func NewClient(addr string, input io.ReadCloser) *Client {
	group, _ := errgroup.WithContext(context.Background())
	return &Client{
		Addr:  addr,
		Done:  make(chan struct{}),
		input: input,
		group: group,
	}
}

func (client *Client) Connect() error {
	defer func() {
		client.Done <- struct{}{}
	}()
	conn, err := net.Dial("tcp", client.Addr)
	if err != nil {
		return err
	}
	client.conn = conn
	log.Println("input anything")

	// 处理输入
	client.group.Go(func() error {
		defer client.Shutdown()
		reader := bufio.NewReader(client.input)
		for {
			line, err := reader.ReadSlice('\n')
			if err != nil {
				log.Printf("read from input error(%v)\n", err)
				return err
			}
			_, err = conn.Write(line)
			if err != nil {
				log.Printf("write to server(%s) error(%v)\n", conn.RemoteAddr(), err)
				return err
			}
		}
	})

	// 处理输出
	client.group.Go(func() error {
		defer client.Shutdown()
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadSlice('\n')
			if err != nil {
				log.Printf("read from %s error(%v)\n", conn.RemoteAddr(), err)
				return err
			}
			log.Printf("read from %s:%s", conn.RemoteAddr(), line)
		}
	})
	if err := client.group.Wait(); err != nil {
		return err
	}
	return nil
}

func (client *Client) Shutdown() {
	client.input.Close()
	client.conn.Close()
}
