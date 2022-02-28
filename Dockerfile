FROM golang:1.15

ENV DEBIAN_FRONTEND=noninteractive

RUN  sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list && \
     sed -i  's/security.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list && \
     echo 'deb http://ppa.launchpad.net/ansible/ansible/ubuntu bionic main' > /etc/apt/sources.list.d/ansible.list
 
RUN  apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 93C4A3FD7BB9C367 && \
     apt-get update && apt-get install -y ansible openssh-client netcat-openbsd sqlite3 && \
     rm -rf /var/lib/apt/lists/* && apt-get clean

# setup golang proxy 
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /go/src/workspace

COPY . /go/src/workspace

RUN touch sqlite/to2cloud.db && sqlite3 sqlite/to2cloud.db < sqlite/database.dump

RUN go build -o /go/bin/to2cloud cmd/web/main.go

EXPOSE 9000

CMD ["/go/bin/to2cloud"]