# GoDemo
A series of go's demo  

## network
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

## strings
 - strings.Index
 - strings.Split
 - strings.Trim
 - fmt.Sscanf
 - strings.Contains
 
## point
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
status进程（服务器、cmd、channel、正则表达式）
 - flag.String
 - log
 - command
 - strings.Split
 - strings.Trim
 - fmt.Sscanf
 - strings.Contains
 - ResolveUDPAddr
 - net.ListenUDP
 - net.ResolveUDPAddr
 - ReadFromUDP
 - WriteToUDP
