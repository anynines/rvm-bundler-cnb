FROM ubuntu:bionic

ARG package_args='--allow-downgrades --allow-remove-essential --allow-change-held-packages --no-install-recommends'

RUN echo "debconf debconf/frontend select noninteractive" | debconf-set-selections && \
  export DEBIAN_FRONTEND=noninteractive && \
  apt-get -y $package_args update && \
  apt-get -y $package_args install locales && \
  locale-gen en_US.UTF-8 && \
  update-locale LANG=en_US.UTF-8 LANGUAGE=en_US.UTF-8 LC_ALL=en_US.UTF-8 && \
  apt-get -y $package_args install \
  libgmp-dev \
  libyaml-0-2 \
  tzdata \
  libreadline7 \
  build-essential \
  gawk \
  autoconf \
  automake \
  bison \
  sqlite3 \
  ca-certificates \
  curl \
  git \
  jq \
  libssl1.1 \
  libssl1.0.0 \
  openssl \
  xz-utils \
  graphicsmagick-libmagick-dev-compat \
  libffi-dev \
  libgdbm-dev \
  libmagickwand-dev \
  libmysqlclient-dev \
  libncurses5-dev \
  libpq-dev \
  libreadline-dev \
  libsqlite3-dev \
  libssl-dev \
  libyajl-dev \
  libyaml-dev \
  ruby-dev \
  zlib1g-dev

RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*

RUN curl -sL -o /usr/local/bin/yj https://github.com/sclevine/yj/releases/latest/download/yj-linux \
  && chmod +x /usr/local/bin/yj
