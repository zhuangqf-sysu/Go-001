package echo

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// todo context 传递 matedata + 日志链路 + 超时控制
type Server struct {
	Addr string

	listener net.Listener
	group    *errgroup.Group
	doneChan chan struct{}

	connList sync.Map
}

func NewServer(addr string) *Server {
	group, _ := errgroup.WithContext(context.Background())

	return &Server{
		Addr:     addr,
		group:    group,
		doneChan: make(chan struct{}),
		connList: sync.Map{},
	}
}

func (server *Server) Shutdown() error {
	log.Printf("server(%s) stopping....\n", server.Addr)
	server.listener.Close()
	server.doneChan <- struct{}{}

	server.connList.Range(func(key, _ interface{}) bool {
		if conn, ok := key.(net.Conn); ok && conn != nil {
			conn.Close()
		}
		return true
	})

	return server.group.Wait()
}

func (server *Server) ListenAndServe() error {
	listen, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	defer listen.Close()
	log.Printf("server listenning on(%s)\n", server.Addr)
	server.listener = listen
	for {
		conn, err := listen.Accept()
		if err != nil {
			_, ok := <-server.doneChan
			if ok {
				log.Printf("echo(%s) closed\n", server.Addr)
				return nil
			}
			log.Printf("echo(%s) accept err(%v) retry...\n", server.Addr, err)
			// todo config
			time.Sleep(10 * time.Millisecond)
		}

		server.connList.Store(conn, struct{}{})
		log.Printf("connected by %s\n", conn.RemoteAddr())
		server.run(func() error {
			defer conn.Close()
			server.handler(conn)
			return nil
		})
	}
}

// todo ctx
// 包装errgroup.Go()
func (server *Server) run(f func() error) {
	server.group.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				log.Printf("panic in echo proc, err: %v, stack: %s\n", r, buf)
			}
		}()
		return f()
	})
}

func (server *Server) handler(conn net.Conn) {
	defer conn.Close()

	group, _ := errgroup.WithContext(context.Background())
	msgChan := make(chan string, 1024)

	// 处理读
	group.Go(func() error {
		defer close(msgChan)

		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadSlice('\n')
			if err != nil && err != io.EOF {
				log.Printf("read from %s error(%v)\n", conn.RemoteAddr(), err)
				return err
			}
			if err == io.EOF {
				return nil
			}
			msgChan <- string(line)
		}
	})

	// 处理写
	group.Go(func() error {
		for msg := range msgChan {
			if _, err := conn.Write([]byte(msg)); err != nil {
				log.Printf("write to %s(%s) error(%v)\n", conn.RemoteAddr(), msg, err)
			}
		}
		return nil
	})

	_ = group.Wait()
	server.connList.Delete(conn)
}
