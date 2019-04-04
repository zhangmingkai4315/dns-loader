# dns-loader
go语言实现的dns负载测试工具

### 开发环境

如需开发，可以使用docker-envirments文件夹下的docker镜像启动一个bind dns server
启动后服务器的地址可以通过`docker inspect $(Your docker id)`查看

启动docker环境命令：

```
docker build -t named-root .
docker run -d named-root
```

### 1 启动程序

#### 1.1 参数说明

程序运行通过命令模式执行，执行程序二进制后，有多个模式可供选择，分别是adhoc, master和agent模式.

```shell
Usage:
  dnsloader [flags]
  dnsloader [command]

Available Commands:
  adhoc       Run dnsloader in adhoc mode
  agent       Run dnsloader in agent mode
  help        Help about any command
  master      Run dnsloader in master mode
  version     Print version of dnsloader

Flags:
  -h, --help   help for dnsloader
```


#### 1.1  adhoc模式
 
该模式下无需指定配置文件, 所有的配置通过命令行来输入，仅仅执行一次发包，单机模式，如需要中断直接运行ctrl+C即可退出。支持的命令如下：


```shell
Usage:
  dnsloader adhoc [flags]

Flags:
  -d, --domain string      domain name
  -D, --duration int       send out dns traffic duration (default 60)
  -h, --help               help for adhoc
  -p, --port int           dns server port (default 53)
  -Q, --qps int            qps for dns traffic (default 100)
  -q, --querytype string   random dns query type (default "A")
  -r, --random int         prefix random subdomain length (default 5)
  -R, --randomtype         random dns query type
  -s, --server string      dns server ip
```

**实例** 

下面通过adhoc命令向baidu.com发送随机五个字符长度的域名比如```abcde.baidu.com```， QPS=100, 域名服务器地址为8.8.8.8。

```
dnsload adhoc -d baidu.com -Q 100 -s 8.8.8.8

2017/11/15 20:58:04 new dns loader client success, start send packet...
2017/11/15 20:58:04 Start generate dns packet
2017/11/15 20:58:04 init dns loader client configuration success
...
[===>                ] 
...
2017/11/15 20:59:04 [Result]total packets sum:60001
2017/11/15 20:59:04 [Result]runing time 1m0s
2017/11/15 20:59:04 [Result]status nxdomain:60001 [100.00%]


```
#### 1.2  master模式

该模式下需要指定模式类型为`master`和配置文件，运行后通过web浏览器来管理发包请求，默认登入用户名和密码为admin/admin,登入配置发包类型和模式后点击开始即可，该模式下可以添加agent，只需要agent端运行agent模式即可（见1.3）
```

```
启动后访问：http://localhost:9889

#### 1.3  agent模式

该模式需要指定模式类型为`agent`和配置文件，可放入后台运行，该模式不会开启任何的web页面，但是会提供http接口供master服务器端调用
```
go run dns-loader.go -t agent -c config.ini 
2017/11/15 21:15:12 load configuration from file:config.ini
2017/11/15 21:15:12 start agent server listen on 8998 for master connect
2017/11/15 21:15:12 agent server route init success
...
```

