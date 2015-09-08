all:  wrapper/onetimeserver onetimeserver-go

wrapper/onetimeserver: wrapper/wrapper.c
	gcc -o wrapper/onetimeserver wrapper/wrapper.c

onetimeserver-go:
	go install github.com/osheroff/onetimeserver/...

