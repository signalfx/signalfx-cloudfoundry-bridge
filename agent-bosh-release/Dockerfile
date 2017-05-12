FROM ubuntu:trusty

RUN apt-get update -q &&\
    apt-get install -y -qq software-properties-common python-software-properties apt-transport-https &&\
    add-apt-repository -y ppa:signalfx/collectd-release &&\
    add-apt-repository -y ppa:signalfx/collectd-plugin-release &&\
    apt-get update -q &&\
    apt-get install -y -qq parallel wget

# Make dpkg think there aren't any packages installed so we get all dependencies
RUN > /var/lib/dpkg/status &&\
    mkdir /tmp/stage

WORKDIR /tmp/stage

ENV packages="collectd collectd-core signalfx-collectd-plugin"
RUN apt-get --print-uris --reinstall --yes install -qq $packages | awk '{print $1}' | sed -e "s/^'//" -e "s/'$//" > package_urls

# Download them all in parallel
RUN cat package_urls | parallel --gnu "wget --timeout=10 --tries=3 {}"

# Extract all deb packages to a fake root dir
RUN mkdir signalfx-collectd-bundle &&\
    echo "Unpacking all .deb packages into fake root..." &&\
    find . -name "*.deb" -exec dpkg-deb -x \{\} signalfx-collectd-bundle \; &&\
    rm -rf *.deb

CMD tar -czf - signalfx-collectd-bundle
