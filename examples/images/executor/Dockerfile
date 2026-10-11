FROM debian:latest

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

COPY debian.sources /etc/apt/sources.list.d/debian.sources

# PentAGI 任务容器的通用增强执行环境。
RUN printf 'Acquire::ForceIPv4 "true";\n' > /etc/apt/apt.conf.d/99force-ipv4 \
    && apt-get update \
    && apt-get install --yes --no-install-recommends \
        bash \
        bash-completion \
        build-essential \
        bzip2 \
        ca-certificates \
        curl \
        dnsutils \
        file \
        git \
        iproute2 \
        iputils-ping \
        jq \
        less \
        lsof \
        nano \
        netcat-openbsd \
        openssh-client \
        procps \
        psmisc \
        python3 \
        python3-pip \
        python3-venv \
        socat \
        tar \
        traceroute \
        unzip \
        vim-tiny \
        wget \
        xz-utils \
        zip \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /work

CMD ["bash"]
