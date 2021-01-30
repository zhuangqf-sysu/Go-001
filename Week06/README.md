代码地址：https://github.com/zhuangqf-sysu/go-tools/tree/master/window

- Event 事件
- Sink 计算结果，伴随窗口滑动动态变更
- CountWindow为固定大小的滑动窗口
    - 每次针对进出的event流进行计算
    - 线程不安全
- TimeWindow 时间窗口
    - 划分为多个小时间片
    - 每个时间片按照该时间片的事件流累计结果 -> Sink
    - 每个时间片的Sink作为一个Event流入整体的window
    - 线程安全
    
具体使用见test