/* 用 Chrome DevTools Protocol 截图：可先执行一段脚本再拍。
   用法：node cdp.js <url> <out.png> [脚本文件] [宽] [高] [等待毫秒] */
const { spawn } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');

const CHROME = '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const [, , url, out, scriptFile, W = '1440', H = '900', waitMs = '700'] = process.argv;
const PORT = 9333 + Math.floor(Math.random() * 400);

const profile = fs.mkdtempSync(path.join(require('node:os').tmpdir(), 'cdp-'));
const chrome = spawn(CHROME, [
  '--headless=new',
  '--remote-debugging-port=' + PORT,
  '--user-data-dir=' + profile,
  '--disable-gpu',
  '--hide-scrollbars',
  '--force-device-scale-factor=1',
  '--no-first-run',
  '--no-default-browser-check',
  '--window-size=' + W + ',' + H,
  'about:blank'
], { stdio: 'ignore' });

const sleep = ms => new Promise(r => setTimeout(r, ms));

async function targets() {
  for (let i = 0; i < 60; i++) {
    try {
      const r = await fetch('http://127.0.0.1:' + PORT + '/json/list');
      const list = await r.json();
      const page = list.find(t => t.type === 'page');
      if (page) return page;
    } catch (e) { /* 还没起来 */ }
    await sleep(150);
  }
  throw new Error('chrome 没起来');
}

(async () => {
  const page = await targets();
  const ws = new WebSocket(page.webSocketDebuggerUrl);
  let id = 0;
  const waiting = new Map();
  const events = [];
  await new Promise(res => ws.addEventListener('open', res));
  ws.addEventListener('message', ev => {
    const msg = JSON.parse(ev.data);
    if (msg.id && waiting.has(msg.id)) {
      const { resolve, reject } = waiting.get(msg.id);
      waiting.delete(msg.id);
      msg.error ? reject(new Error(JSON.stringify(msg.error))) : resolve(msg.result);
    } else if (msg.method) {
      events.push(msg.method);
    }
  });
  const send = (method, params = {}) => new Promise((resolve, reject) => {
    const mid = ++id;
    waiting.set(mid, { resolve, reject });
    ws.send(JSON.stringify({ id: mid, method, params }));
  });

  await send('Page.enable');
  await send('Runtime.enable');
  await send('Emulation.setDeviceMetricsOverride', {
    width: Number(W), height: Number(H), deviceScaleFactor: 1, mobile: false
  });
  await send('Page.navigate', { url });
  for (let i = 0; i < 80; i++) {
    if (events.includes('Page.loadEventFired')) break;
    await sleep(100);
  }
  await sleep(Number(waitMs));

  if (scriptFile && scriptFile !== '-') {
    const expr = fs.readFileSync(scriptFile, 'utf8');
    const r = await send('Runtime.evaluate', { expression: expr, awaitPromise: true, returnByValue: true });
    if (r.exceptionDetails) {
      console.error('脚本出错：', JSON.stringify(r.exceptionDetails.exception && r.exceptionDetails.exception.description || r.exceptionDetails.text));
    } else if (r.result && r.result.value !== undefined) {
      console.log('脚本结果：', JSON.stringify(r.result.value));
    }
    await sleep(Number(waitMs));
  }

  const dragArg = process.env.DRAG;   // "x1,y1,x2,y2"
  if (dragArg) {
    const [x1, y1, x2, y2] = dragArg.split(',').map(Number);
    await send('Input.dispatchMouseEvent', { type: 'mousePressed', x: x1, y: y1, button: 'left', buttons: 1, clickCount: 1 });
    for (let i = 1; i <= 12; i++) {
      const t = i / 12;
      await send('Input.dispatchMouseEvent', { type: 'mouseMoved', x: x1 + (x2 - x1) * t, y: y1 + (y2 - y1) * t, button: 'left', buttons: 1 });
      await sleep(16);
    }
    await send('Input.dispatchMouseEvent', { type: 'mouseReleased', x: x2, y: y2, button: 'left', buttons: 0, clickCount: 1 });
    await sleep(600);
    console.log('拖动完成');
  }

  const shot = await send('Page.captureScreenshot', { format: 'png' });
  fs.mkdirSync(path.dirname(out), { recursive: true });
  fs.writeFileSync(out, Buffer.from(shot.data, 'base64'));
  console.log('已保存', out);
  ws.close();
  chrome.kill();
  process.exit(0);
})().catch(err => {
  console.error(err);
  chrome.kill();
  process.exit(1);
});
