FROM --platform=linux/amd64 ubuntu 

RUN apt-get -y update && apt-get -y install gcc make golang curl

ADD ./ /app
WORKDIR /app

