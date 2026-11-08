#!/bin/bash
# 给每套原型截一张图：sh design/_build/shot.sh [宽] [高]
CHROME="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
W=${1:-1440}
H=${2:-900}
OUT="${OUT_DIR:-/tmp/shot}"
mkdir -p "$OUT"
for d in design/0*/; do
  name=$(basename "$d")
  "$CHROME" --headless=new --disable-gpu --hide-scrollbars \
    --window-size="$W,$H" --virtual-time-budget=3000 \
    --screenshot="$OUT/$name-$W.png" "file://$PWD/$d/index.html" >/dev/null 2>&1
  echo "$OUT/$name-$W.png"
done
