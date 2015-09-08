all: wrapper

wrapper: wrapper.c
	gcc -g -o wrapper wrapper.c

