all:  wrapper onetimeserver-go install-local

wrapper: wrapper/wrapper.c wrapper/wrapper
	gcc -g -o wrapper/onetimeserver-wrapper wrapper/wrapper.c

onetimeserver-go:
	go install github.com/osheroff/onetimeserver/...

DIR=${HOME}/.onetimeserver/$(shell uname -s)-$(shell uname -m)

install-local:
	mkdir -p $(DIR)
	cp wrapper/wrapper $(DIR)
	cp ${GOPATH}/bin/onetimeserver-go $(DIR)


