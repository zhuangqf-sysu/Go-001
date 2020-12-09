1、创建一个errgroup(group + ctx)
2、通过errgroup启动10个http server
3、通过errgroup启动goroutine(g1)负责监听系统信号和ctx.Done事件
4、g1负责http server的优雅关停
5、errgroup中一旦有goroutine返回err, 内部会调用ctx.cannel() -> ctx.Done() -> g1 -> http server优雅关停

详见err.go(可运行)
