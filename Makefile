.PHONY: build-linux build-linux-gui build-windows-gui build-macos-gui generate-icon clean

build-linux:
	cd server && go build -o paygo-linux-amd64 ./src

build-linux-gui: generate-icon
	cd server && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags gui -o paygo-linux-gui-amd64 ./src

generate-icon:
	cp assets/paygo.png server/src/systray/icon.png
	cd server/src/systray && go generate -tags gui .

build-windows-gui: generate-icon
	cd server/src && rsrc -ico systray/icon.ico -o windows/rsrc.syso
	cd server && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags gui -ldflags "-H=windowsgui -s -w" -o paygo-windows-amd64.exe ./src

build-macos-gui: generate-icon
	cd server && CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -tags gui -ldflags "-s -w" -o paygo-macos-amd64 ./src
	cd server && CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags gui -ldflags "-s -w" -o paygo-macos-arm64 ./src

clean:
	rm -f server/src/windows/rsrc.syso
	rm -f server/paygo-linux-amd64 server/paygo-linux-gui-amd64
	rm -f server/paygo-windows-amd64.exe
	rm -f server/paygo-macos-amd64 server/paygo-macos-arm64
