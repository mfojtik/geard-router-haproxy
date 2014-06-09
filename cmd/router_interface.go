package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"

	//"bufio"
)

var (
	messageBus string
)

func sendMessage(msg interface{}) {
	messageBus += fmt.Sprintf("%s\n", msg)
}

func sendError(err error) {
	messageBus += fmt.Sprintf(":ERR %s\n", err)
}

func execute(cmd *exec.Cmd) (output string, err error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		sendError(err)
	} else {
		output = string(out)
	}
	return
}

func writeHaproxyConfig() (err error) {
	cmd := exec.Command("/var/lib/haproxy/bin/write_haproxy_config")
	if _, err = execute(cmd); err != nil {
		sendError(err)
	}
	return
}

func startServer() {
	cmd := exec.Command("haproxy", "-f", "/var/lib/haproxy/conf/haproxy.config", "-p", "/var/lib/haproxy/run/haproxy.pid")
	execute(cmd)
}

func stopServer() {
	pid, err := ioutil.ReadFile("/var/lib/haproxy/run/haproxy.pid")
	if err != nil {
		cmd := exec.Command("haproxy", "-f", "/var/lib/haproxy/conf/haproxy.config", "-p", "/var/lib/haproxy/run/haproxy.pid", "-st", string(pid))
		execute(cmd)
	}
}

func reloadServer() (err error) {
	oldPid, err := ioutil.ReadFile("/var/lib/haproxy/run/haproxy.pid")
	var cmd *exec.Cmd

	if err != nil {
		cmd = exec.Command("haproxy", "-f", "/var/lib/haproxy/conf/haproxy.config", "-p", "/var/lib/haproxy/run/haproxy.pid")
	} else {
		cmd = exec.Command("haproxy", "-f", "/var/lib/haproxy/conf/haproxy.config", "-p", "/var/lib/haproxy/run/haproxy.pid", "-sf", string(oldPid))
	}

	_, err = execute(cmd)
	return
}

func commandServer(c net.Conn) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := strings.Trim(string(buf[0:nr]), "\n ")
		fmt.Printf("[SERVER] <- '%s'", data)

		switch data {
		case "Start":
			sendMessage(":OK")
			writeHaproxyConfig()
			startServer()
		case "Stop":
			sendMessage(":OK")
			writeHaproxyConfig()
			stopServer()
		case "Reload":
			sendMessage(":OK")
			if err := writeHaproxyConfig(); err != nil {
				continue
			}
			reloadServer()
		default:
			sendError(errors.New("Accepted commands: 'Start', 'Stop' and 'Reload'."))
		}

		fmt.Printf("[SERVER] -> '%s'", messageBus)
		_, err = c.Write([]byte(messageBus))
		messageBus = ""
		if err != nil {
			fmt.Printf("Write error: %s", err)
		}
	}
}

func main() {
	filename := "/var/lib/haproxy/run/router_interface.sock"
	rerr := os.Remove(filename)
	if rerr != nil {
		println("Failed to remove unix socket file", rerr)
		os.Exit(1)
	}
	l, err := net.Listen("unix", filename)
	if err != nil {
		println("listen error", err)
		os.Exit(1)
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			println("accept error", err)
			os.Exit(1)
		}

		go commandServer(fd)
	}
}
