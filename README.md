# dns-loader

DNS Benchmark test tools, build with a manager webui for remote control. And you can also add agents to do the benchmark job. 


### 1 Usage

#### 1.1 Download 

You can download the build binary from release page, no need to install. Just copy the binary to your server or pc and start the app in command line.


```shell
$ dns-loader 
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

#### 1.2  adhoc

adhoc mode is the basic running mode of dns-loader, just like dnsperf but no perf files support, all dns query domain and type will be generated from arguments.

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

**example** 

send to dns server 127.0.0.1(default port is 53) ,query domain is test with prefix random subdomin length 5(just like xjsjf.test, adfnd.test), max query persecond is 100000. query type default is random you can set it to A or AAAA as you wish. default duration is 60s

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
#### 1.3  master

master mode will allow user set the bench arguments in web ui, default webui link is http://HOST:9889, the user/password is set in config.ini file. 

```
Usage:
  dns-loader master [flags]

Flags:
  -h, --help            help for master
      --config string   config file (default is $HOME/config.ini)
      --dbfile string   database file for dns loader app(create automatic) (default "app.db")
```


#### 1.4  agent

start agent host in any host which can talk with master host, it will listen the command and do query job. you need add the connection agent ip and port in master webui, after that master and agents can do query job as the same.

```
Run dns-loader in agent mode, receive job from master and gen dns packets

Usage:
  dns-loader agent [flags]

Flags:
  -h, --help            help for agent
      --host string   ipaddress for start agent (default "0.0.0.0")
      --port string   port to listen (default "8998")

```
