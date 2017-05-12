collectd-rootfs-blob:
	rm -rf tmp/collectd-rootfs.tar.gz
	bash make-collectd-rootfs-blob
	bosh add blob tmp/collectd-rootfs.tar.gz signalfx-collectd

runc-blob:
	rm -rf tmp/runc-linux-amd64
	bash make-runc-blob
	bosh add blob tmp/runc-linux-amd64 runc

bosh-dev-release: #collectd-bundle
	echo yes | bosh delete deployment signalfx-agent
	bosh releases | grep '\<signalfx-agent\>' && echo yes | bosh delete release --force signalfx-agent || true
	echo yes | bosh reset release
	echo | bosh create release --force
	bosh upload release dev_releases/signalfx-agent/signalfx-agent-0+dev.1.yml

bosh-dev-deploy: bosh-dev-release
	bosh target https://192.168.50.4:25555
	bosh deployment bosh-lite/manifest.yml
	sed -i -e "s/SIGNALFX_API_TOKEN/$(shell cat .sfx-token)/" bosh-lite/manifest.yml
	echo yes | bosh deploy

.PHONY: collectd-bundle
