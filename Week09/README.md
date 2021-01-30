### 实现
实现一个echo服务器：
- 服务器逻辑：
    - 主线程监听accept事件，建立连接后将conn扔到group.GO运行
    - handler 启动两个goroutine, 一个负责读，一个负责写，conn是线程安全的
    - 读goroutine通过chan将数据传递给写gorontine,写goroutine写回conn
    
- 客户端逻辑：
    - 主线程建立连接
    - 启动两个goroutine, 一个负责读，一个负责写
    - 写goroutine将控制台输入传输到server（按行划分）
    - 读goroutine读取sever返回打印到控制台
    
### 启动
- 启动服务端
```shell script
   go run ./main.go -mode=server
```

- 启动客户端
```shell script
   go run ./main.go -mode=client
```

- 关闭
```
    Ctrl + C
```