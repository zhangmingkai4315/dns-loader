# dns-loader
go语言实现的dns负载测试工具


### 1 启动程序

#### 1.1 参数说明

程序运行通过命令模式执行，执行程序二进制后，有多个模式可供选择，分别是adhoc, master和agent模式.

```shell
Usage:
  dns-loader [flags]
  dns-loader [command]

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
  dns-loader adhoc [flags]

Flags:
  -d, --domain string      domain name
  -D, --duration int       duration for send dns traffic (default 60s)
  -h, --help               help for adhoc
  -p, --port int           dns server port (default 53)
  -Q, --qps int            qps for dns traffic (default 100)
  -q, --querytype string   random dns query type (default "" random query type)
  -r, --random int         prefix random subdomain length (default 5)
  -s, --server string      dns server ip
```

**实例** 

下面通过adhoc命令向baidu.com发送随机五个字符长度的域名比如```abcde.baidu.com```， QPS=100, 域名服务器地址为8.8.8.8。

```
./dns-loader adhoc -d test -s 127.0.0.1 -Q 100000
INFO[0000] new dns loader client success
INFO[0000] dns packet info :[domain=test,length=5,type=1]
INFO[0000] config the dns loader success
INFO[0000] dnsloader server info : server:127.0.0.1|port:53
INFO[0000] initialize load gernerator[qps=100000, durations=1m0s,timeout=1s]
INFO[0000] checking the parameters
INFO[0000] check the parameters success. (timeout=1s, qps=100000, duration=1m0s)
INFO[0000] starting dns loader generator
INFO[0000] setting throttle 10µs
INFO[0000] create new thread to receive dns data from server
INFO[0000] start push packets to dns server and will stop at 1m0s later...
INFO[0060] prepare to stop load test [context deadline exceeded]
INFO[0060] doing calculation work
INFO[0060] total packets sum:2617542                     result=true
INFO[0060] runing time 1m0.000113166s                    result=true
INFO[0060] status Success:21627 [0.83]                   result=true
INFO[0060] status unknown:2595915 [99.17]                result=true
INFO[0060] stop success!

```
#### 1.2  master模式

该模式下需要指定模式类型为`master`和对应的配置文件，运行后通过web浏览器来管理发包请求，默认登入用户名和密码为admin/admin,登入配置发包类型和模式后点击开始即可，该模式下可以添加agent，只需要agent端运行agent模式即可。
```
Usage:
  dns-loader master [flags]

Flags:
      --config string   config file (default is $HOME/config.ini)
  -h, --help            help for master
```

启动后访问：http://localhost:9889

#### 1.3  agent模式

该模式需要指定模式类型为`agent`和配置文件，可放入后台运行，该模式不会开启任何的web页面，但是会提供http接口供master服务器端调用
```
Run dns-loader in agent mode, receive job from master and gen dns packets

Usage:
  dns-loader agent [flags]

Flags:
      --config string   config file (default is $HOME/config.ini) (default "-c")
  -h, --help            help for agent
```

### 开发环境

如需开发，可以启动一个权威DNS服务器用于测试使用, 权威配置文件在docker目录中conf文件夹下
```
docker-compose up
```
