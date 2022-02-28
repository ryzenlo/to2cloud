# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.137.0/containers/go/.devcontainer/base.Dockerfile
ARG VARIANT="1.15"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

RUN  sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list && \
     sed -i  's/security.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list && \
     echo 'deb http://ppa.launchpad.net/ansible/ansible/ubuntu bionic main' > /etc/apt/sources.list.d/ansible.list

# [Optional] Install a version of Node.js using nvm for front end dev
ARG INSTALL_NODE="false"
ARG NODE_VERSION="lts/*"
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

ENV DEBIAN_FRONTEND=noninteractive

RUN  apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 93C4A3FD7BB9C367 && \
     apt-get update && apt-get install -y ansible openssh-client netcat-openbsd sqlite3 && \
     rm -rf /var/lib/apt/lists/* && apt-get clean

# [Optional] Uncomment the next line to use go get to install anything else you need
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1