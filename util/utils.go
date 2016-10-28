package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"time"
)

func GetInnerIp() string {
	/*
		cmdStr := "/bin/bash -c if'config eth1 | awk '/addr:/{sub(/addr:/,\"\",$2);print $2;exit}'"
		cmdOutput, RunCmdErr := RunCmd(cmdStr, time.Duration(2*60)*time.Second)
		if RunCmdErr != nil {
			cmdOutput = ""
		}
	*/
	cmd := exec.Command("/bin/bash", "-c", `ifconfig eth1 | awk '/addr:/{sub(/addr:/,"",$2);print $2;exit}'`)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ""
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Start: ", err.Error())
		return ""
	}

	cmdOutput, err := ioutil.ReadAll(stdout)
	if err != nil {
		return ""
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("ReadAll stdout: ", err.Error())
		return ""
	}

	if len(cmdOutput) <= 1 {
		fmt.Println("ReadAll empty")
		return ""
	}
	return string(cmdOutput[0 : len(cmdOutput)-1])
}

func IPAddrByInterfaceName(name string) (ip string, err error) {
	interfaces, _ := net.Interfaces()
	for _, inter := range interfaces {
		if addrs, err := inter.Addrs(); err == nil {
			for _, addr := range addrs {
				if inter.Name == name {
					if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
						return ip.String(), nil
					}
				}
			}
		}
	}

	return "", errors.New(fmt.Sprintf("%s not found", name))
}

func RunCmd(cmd string, timeout time.Duration) (out string, err error) {
	var command *exec.Cmd

	args := strings.Split(cmd, " ")
	fmt.Printf("%q\n", args)

	if len(args) > 1 {
		command = exec.Command(args[0], args[1:]...)
	} else {
		command = exec.Command(args[0])
	}

	result := make(chan string)

	var stderr bytes.Buffer
	command.Stderr = &stderr

	go func() {
		out, err := command.Output()
		if err != nil {
			fmt.Println("error :", err, "----", stderr.String())
		}
		result <- string(out)
	}()

	select {
	case cmdOutput := <-result:
		fmt.Println("select cmdOutput")
		return cmdOutput, nil

	case <-time.After(timeout):
		fmt.Println("select timeout")
		command.Process.Kill()
		return "", errors.New(fmt.Sprintf("exec cmd(%s) timeout", cmd))
	}

	fmt.Println("end select")
	return "nothing", nil
}
