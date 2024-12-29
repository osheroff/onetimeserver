FROM --platform=linux/amd64 ubuntu 

RUN apt-get -y update && apt-get -y install gcc make golang curl less vim gdb

ENV GOPATH=/root/go/
WORKDIR /app

