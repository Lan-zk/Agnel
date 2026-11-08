#!/bin/sh
# 启动 Agnel。
#   ./run.sh            打开打包好的 build/Agnel.app
#   ./run.sh dev        用 viva dev 跑（热重载；这台机器上开窗尺寸也更稳）
cd "$(dirname "$0")" || exit 1

if [ "$1" = "dev" ]; then
	# viva dev 的窗口属于 viva 宿主进程，裸跑时 Dock 只能用通用 exec 图标。
	# 把宿主放进一个临时 .app 里，Dock 就会用这个 bundle 的图标。
	dev_app="build/Agnel Dev.app"
	rm -rf "$dev_app"
	mkdir -p "$dev_app/Contents/MacOS" "$dev_app/Contents/Resources"
	cp "${GOPATH:-$HOME/go}/bin/viva" "$dev_app/Contents/MacOS/viva"
	if [ -f "build/Agnel.app/Contents/Resources/AppIcon.icns" ]; then
		cp "build/Agnel.app/Contents/Resources/AppIcon.icns" "$dev_app/Contents/Resources/AppIcon.icns"
	else
		iconset="$(mktemp -d)/AppIcon.iconset"
		mkdir -p "$iconset"
		for size in 16 32 128 256 512; do
			sips -z "$size" "$size" assets/icon.png --out "$iconset/icon_${size}x${size}.png" >/dev/null
			double=$((size * 2))
			sips -z "$double" "$double" assets/icon.png --out "$iconset/icon_${size}x${size}@2x.png" >/dev/null
		done
		iconutil -c icns "$iconset" -o "$dev_app/Contents/Resources/AppIcon.icns"
		rm -rf "$(dirname "$iconset")"
	fi
	cat > "$dev_app/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key><string>en</string>
	<key>CFBundleExecutable</key><string>viva</string>
	<key>CFBundleIconFile</key><string>AppIcon</string>
	<key>CFBundleIdentifier</key><string>dev.agnel.reader.devhost</string>
	<key>CFBundleName</key><string>Agnel</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>0.1.0</string>
	<key>CFBundleVersion</key><string>1</string>
	<key>LSMinimumSystemVersion</key><string>12.0</string>
	<key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST
	codesign --force --deep -s - "$dev_app" >/dev/null 2>&1 || true
	exec "$dev_app/Contents/MacOS/viva" dev .
fi

if [ -d "build/Agnel.app" ]; then
	open "build/Agnel.app"
else
	./build/agnel
fi
