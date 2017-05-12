FROM ubuntu:yakkety
COPY sources.list /etc/apt
RUN apt-get update && apt-get install -y \
    python2.7 \
    python-pip \
    openjdk-8-jdk-headless \
    curl \
    unzip \
# BOSH CLI dependencies.
    build-essential \
    ruby \
    ruby-dev \
    libxml2-dev \
    libsqlite3-dev \
    libxslt1-dev \
    libpq-dev \
    libmysqlclient-dev \
    zlib1g-dev \
    m4

RUN pip install tile-generator>=6.0.0
RUN gem install bosh_cli --no-ri --no-rdoc

RUN mkdir -p /opt
WORKDIR /opt

RUN mkdir -p /opt/build
COPY build.gradle settings.gradle gradlew VERSION /opt/build/
COPY gradle /opt/build/gradle

# Populate dependencies.
WORKDIR /opt/build
RUN ./gradlew --console plain --no-daemon --refresh-dependencies shadowJar
