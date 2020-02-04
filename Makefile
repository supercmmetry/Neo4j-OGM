.phony: build
build:
	PKG_CONFIG_PATH=/usr/local/share/pkgconfig
	go build --tags seabolt_static -o bin/main .

.phony: run
run:
	PKG_CONFIG_PATH=/usr/local/share/pkgconfig
	@~/.air -d -c .air.conf