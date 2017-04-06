# GoDemo
提供一些列常用简单Demo及工具类库，旨在快速开发时提供参考

## network
tcp\udp网络首发包demo
 - net.Dial
 - net.ResolveTCPAddr
 - net.ListenTCP
 - Accept
 - net.Conn (type)
 - conn.RemoteAddr().String()
 - Write
 - Read
 - Close
 - <br>
 - net.DialUDP
 - net.UDPAddr
 - net.IPv4
 - ReadFromUDP
 - WriteToUDP
 - net.ListenUDP

## util
 工具类库（共享内存系统调用、获取本机内网ip，根据网卡名字获取ip，执行本机命令行）
 - exec.Command
 - cmd.StdoutPipe (cmd is type  \*cmd 
 - cmd.Start
 - ioutil.ReadAll
 - cmd.Wait
 - <br>
 - net.Interfaces
 - inter.Addrs (inter is type \*Interface)
 - net.ParseCIDR
 - <br>
 - strings.Split
 - command.Output
 - select
 - time.After(timeout) (timeout is type \*time.Duration)
 
## point
 指针类、原子操作demo、系统调用
 - unsafe.Pointer
 - uintptr (type)
 - syscall.RawSyscall
 - sync.Once
 - atomic.CompareAndSwapUint64
 - atomic.StoreUint64
 - atomic.LoadUint64
 - atomic.AddUint64

## panic
 - defer
 - panic
 - recover

## channel
 管道demo
 - make
 - go

## json
 - json.MarshalIndent
 - json.Unmarshal

## http
 http服务器demo
 - runtime.Stack
 - http.ListenAndServe
 - http.Error
 - r.URL.Path (r is type \*http.Request)
 - http.StatusInternalServerError
 
## rpc
 - net.Listen
 - rpc.NewServer
 - newServer.Register (newServer is type \*Server)
 - newServer.ServeConn
 - <br>
 - rpc.Dial
 - client.Call (client is type \*Client)

## daemon srv
status进程（服务器、cmd、channel、正则表达式、字符串操作）
 - flag.String
 - ResolveUDPAddr
 - net.ListenUDP
 - net.ResolveUDPAddr
 - ReadFromUDP
 - go
 - defer/panic
 - time.Second
 - log
 - WriteToUDP
 - exec.Cmd/exec.Command/command.Process.Kill
 - time.After
 - strings.Index
 - strings.Split
 - strings.Trim
 - fmt.Sscanf
 - strings.Contains
 - goto
