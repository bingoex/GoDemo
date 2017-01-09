package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	AppIDPC = 1
	AppIDIM = 3
)

var gListenAddr = flag.String("addr", ":40002", "listen addr")
var gLogFile = flag.String("log", "/data/log/status.log", "log file to write to")

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	flag.Parse()

	file, err := os.OpenFile(*gLogFile, os.O_APPEND|os.O_WRONLY, 0666)
	panicOnErr(err)

	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
}

func RunCmd(cmd string, timeout time.Duration) (out string, err error) {
	var command *exec.Cmd

	args := strings.Split(cmd, " ")
	if len(args) > 1 {
		command = exec.Command(args[0], args[1:]...)
	} else {
		command = exec.Command(args[0])
	}

	type cmdResult struct {
		result string
		err    error
	}

	// start a new goroutine
	result := make(chan cmdResult)
	go func() { out, err := command.Output(); result <- cmdResult{string(out), err} }()

	// wait result with timeout
	select {
	case cmdOutput := <-result:
		return cmdOutput.result, cmdOutput.err

	case <-time.After(timeout):
		command.Process.Kill()
		return "", errors.New("timeout")
	}
}

func responseNotOnline(conn *net.UDPConn, clientAddr *net.UDPAddr) {
	conn.WriteToUDP([]byte(fmt.Sprintf("Status:%d", 20)), clientAddr)
}

func RecoverFromErr() {
	err := recover()
	if err == nil {
		return
	}

	buf := make([]byte, 1<<16)
	writen := runtime.Stack(buf, false)

	log.Printf("real_err(%+v), backtrace:\n%s\n", err, buf[0:writen])
}

func ServerClient(buf []byte, conn *net.UDPConn, clientAddr *net.UDPAddr) {
	defer RecoverFromErr()

	key := string(buf)
	log.Printf("quering online status for key:%s\n", key)

	onlineInfo, err := RunCmd("/home/tools/show_info.sh "+key, time.Second*2)
	if err != nil {
		log.Printf("Check online info failed for key(%s) errr:%s\n", key, err)
		responseNotOnline(conn, clientAddr)
	}

	appInfoIndex := strings.Index(onlineInfo, "App:")
	if appInfoIndex == -1 {
		log.Printf("no app info found in (%s)\n", onlineInfo)
		responseNotOnline(conn, clientAddr)
	}

	// 分隔出每个app进行解析
	for _, appInfo := range strings.Split(onlineInfo[appInfoIndex:len(onlineInfo)], "App:") {
		var (
			notCare            string
			appid              int64
			status             int64
			ptl                int64
			has5003            bool
			instanceFlagIsZero bool
			has5004            bool
		)

		if len(strings.Trim(appInfo, "\n \t")) == 0 {
			continue
		}

		lines := strings.Split(appInfo, "\n")
		lineCnt := len(lines)

		if lineCnt < 15 { // invalidate app info
			log.Printf("app info(%s) is too short!\n", strings.Join(lines, ""))
			continue
		}

		// Parse appid
		if _, err := fmt.Sscanf(lines[0], "%d len:%s", &appid, &notCare); err != nil {
			log.Printf("parse appid(%s) failed: %v\n", lines[0], err)
			continue
		}
		// 过滤掉非PC和IM
		if appid != AppIDPC && appid != AppIDIM {
			// log.Printf("appid(%d) not validate", appid)
			continue
		}

		for i := 0; i < lineCnt; i++ {
			if strings.Contains(lines[i], "dwInstanceFlag=0x0") {
				instanceFlagIsZero = true
				for oi := i; oi < lineCnt; oi++ {
					// find until "Field:5003"
					if strings.Contains(lines[oi], "Field:5003") {
						if oi+2 >= lineCnt {
							log.Println("filed 5003 has not value")
							goto parse_err
						}

						if _, err := fmt.Sscanf(lines[oi+2],
							"0000000:  %x %s", &status, &notCare); err != nil {
							log.Println("parse filed 5003 failed:", err)
							goto parse_err
						}

						has5003 = true
					}

					if strings.Contains(lines[oi], "Field:5004") {
						if oi+2 >= lineCnt {
							log.Println("filed 5004 has not value")
							goto parse_err
						}

						if _, err := fmt.Sscanf(lines[oi+2],
							"0000000:  %s %s %x %s", &notCare, &notCare, &ptl, &notCare); err != nil {
							log.Println("parse filed 5004 failed:", err)
							goto parse_err
						}

						has5004 = true
					}
				}
			}
		}

		if instanceFlagIsZero && has5003 && has5004 {
			response := fmt.Sprintf("Status:%d\nPtl:%d", status, ptl)
			log.Printf("response:'%s'\n", response)
			conn.WriteToUDP([]byte(response), clientAddr)
			return
		} else {
			goto parse_err
		}
	}

parse_err:
	log.Println("parse error")
	responseNotOnline(conn, clientAddr)
}

func main() {
	la, err := net.ResolveUDPAddr("udp4", *gListenAddr)
	panicOnErr(err)

	conn, err := net.ListenUDP("udp", la)
	panicOnErr(err)

	for {
		buf := make([]byte, 2048)
		if n, clientAddr, err := conn.ReadFromUDP(buf); err != nil {
			log.Printf("Read error:%s form remote(%s)\n", err, clientAddr)
		} else {
			go ServerClient(buf[0:n], conn, clientAddr)
		}
	}
}
