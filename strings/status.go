package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
	"utils"
)

const (
	APPID1 = 1
	APPID2 = 7
)

var gListenAddr = flag.String("l", ":40002", "listen addr")

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
}

func responseNotOnline(conn *net.UDPConn, clientAddr *net.UDPAddr) {
	conn.WriteToUDP([]byte(fmt.Sprintf("Status:%d", 20)), clientAddr)
}

func ServerClient(buf []byte, conn *net.UDPConn, clientAddr *net.UDPAddr) {
	key := string(buf)
	log.Printf("quering status for key:%s\n", key)

	statuInfos, err := RunCmd("/xx/xx/xx.sh "+key, time.Second*2)
	if err != nil {
		log.Printf("Check status info failed for key(%s) errr:%s\n", key, err)
		responseNotOnline(conn, clientAddr)
	}

	appInfoIndex := strings.Index(statuInfos, "App:")
	if appInfoIndex == -1 {
		log.Printf("no app info found in (%s)\n", statuInfos)
		responseNotOnline(conn, clientAddr)
	}

	// 分隔出每个app进行解析
	for _, appInfo := range strings.Split(statuInfos[appInfoIndex:len(statuInfos)], "App:") {
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

		if appid != APPID1 && appid != APPID2 {
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

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
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
