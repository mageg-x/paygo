#!/bin/bash
set -e

APP_NAME="GoPay"
ARCH="${1:-arm64}"
BIN_NAME="gopay-macos-${ARCH}"
OUTPUT="${APP_NAME}-macos-${ARCH}.zip"

echo "Build macOS GUI binary: ${ARCH}"
cd server
CGO_ENABLED=1 GOOS=darwin GOARCH=${ARCH} go build -tags gui -ldflags="-s -w" -o "${BIN_NAME}" ./src
cd ..

rm -rf "${APP_NAME}.app"
mkdir -p "${APP_NAME}.app/Contents/MacOS"
mkdir -p "${APP_NAME}.app/Contents/Resources"

cp "server/${BIN_NAME}" "${APP_NAME}.app/Contents/MacOS/${APP_NAME}"
chmod +x "${APP_NAME}.app/Contents/MacOS/${APP_NAME}"

if [ -f "server/src/systray/icon.png" ]; then
  sips -s format icns server/src/systray/icon.png --out "${APP_NAME}.app/Contents/Resources/${APP_NAME}.icns" 2>/dev/null || true
fi

cat > "${APP_NAME}.app/Contents/Info.plist" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>GoPay</string>
  <key>CFBundleIdentifier</key>
  <string>com.gopay.app</string>
  <key>CFBundleName</key>
  <string>GoPay</string>
  <key>CFBundleDisplayName</key>
  <string>GoPay</string>
  <key>CFBundleVersion</key>
  <string>1</string>
  <key>CFBundleShortVersionString</key>
  <string>1.0.0</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>LSMinimumSystemVersion</key>
  <string>11.0</string>
  <key>LSUIElement</key>
  <true/>
</dict>
</plist>
EOF

rm -f "${OUTPUT}"
zip -r "${OUTPUT}" "${APP_NAME}.app"
echo "Output: ${OUTPUT}"

