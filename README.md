# dns-loader
go语言实现的dns负载测试工具

### 1 启动程序

```
go run dns-loader.go
```

### 2 性能优化

##### 函数性能

使用原生库来展示调用栈

```
go tool pprof -seconds=5 localhost:8080/debug/pprof/profile
> web
```

使用uber的库来展示调用栈
```
go get github.com/uber/go-torch
git clone https://github.com/brendangregg/FlameGraph.git
go-torch -t 5
firefox torch.svg

```

