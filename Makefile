.PHONY: build-linux build-linux-gui build-windows-gui build-macos-gui generate-icon clean

build-linux:
	cd server && go build -o gopay-linux-amd64 ./src

build-linux-gui: generate-icon
	cd server && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags gui -o gopay-linux-gui-amd64 ./src

generate-icon:
	cp assets/gopay.png server/src/systray/icon.png
	cd server/src/systray && go generate -tags gui .

build-windows-gui: generate-icon
	cd server/src && rsrc -ico systray/icon.ico -o windows/rsrc.syso
	cd server && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags gui -ldflags "-H=windowsgui -s -w" -o gopay-windows-amd64.exe ./src

build-macos-gui: generate-icon
	cd server && CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -tags gui -ldflags "-s -w" -o gopay-macos-amd64 ./src
	cd server && CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags gui -ldflags "-s -w" -o gopay-macos-arm64 ./src

clean:
	rm -f server/src/windows/rsrc.syso
	rm -f server/gopay-linux-amd64 server/gopay-linux-gui-amd64
	rm -f server/gopay-windows-amd64.exe
	rm -f server/gopay-macos-amd64 server/gopay-macos-arm64
