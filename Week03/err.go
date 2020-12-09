package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main()  {

	// 假设开启10个http server
	group, ctx := errgroup.WithContext(context.Background())
	servers := make([]*http.Server, 0, 10)

	for i:=0;i<10;i++ {
		server := &http.Server{Addr:fmt.Sprintf(":808%d",i), Handler:IndexHandler{}}
		servers = append(servers, server)
		group.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 64<<10)
					buf = buf[:runtime.Stack(buf, false)]
					log.Printf("panic in server(%s) proc, err: %v, stack: %s\n", server.Addr, r, buf)
				}
			}()

			err := server.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				log.Printf("server(%s) closed\n", server.Addr)
				return nil
			}
			if err != nil {
				log.Printf("server %v stop for err（%+v）\n", server, err)
				return err
			}
			return nil
		})
	}

  // 有一个goroutine负责监听系统信号
	group.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10)
				buf = buf[:runtime.Stack(buf, false)]
				log.Printf("panic in signal proc, err: %v, stack: %s\n", r, buf)
			}
			// 优雅关停
			shutDownCtx, _ := context.WithTimeout(context.Background(), time.Second)
			for _, server := range servers {
				server.Shutdown(shutDownCtx)
			}
		}()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

		for {
			select {
			case s := <- c:
				log.Printf("app get a signal %s\n", s.String())
				return nil
			case <-ctx.Done():
				// errgroup 中有线程报错退出，则全部关闭
				log.Printf("app context done \n",)
				return nil
			}
		}
	})

	// group.Go(func() error {
	// 	// 模拟http server 退出, 利用errgroup return err -> ctx.Done
	// 	time.Sleep(10 * time.Second)
	// 	return errors.New("fake err")
	// })

	if err := group.Wait(); err != nil {
		log.Printf("app stop for err(%+v)\n", err)
		return
	}
	log.Println("app stop")
}

type IndexHandler struct {}

func (handler IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(100 * time.Millisecond)
	if _, err := w.Write([]byte("hello world")); err != nil {
		log.Printf("index handle err(%v)\n", err)
	}
}
