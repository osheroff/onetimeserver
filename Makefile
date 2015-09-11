all:  wrapper/onetimeserver onetimeserver-go install-local

wrapper/onetimeserver: wrapper/wrapper.c
	gcc -g -o wrapper/onetimeserver-wrapper wrapper/wrapper.c

onetimeserver-go:
	go install github.com/osheroff/onetimeserver/...

install-local:
	mkdir -p bin
	cp wrapper/onetimeserver-wrapper ./bin
	cp ${GOPATH}/bin/onetimeserver-go ./bin


