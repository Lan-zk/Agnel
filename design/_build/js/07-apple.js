/* 07-apple.js — 方案 05 的手感层。
   资料面板是一块可以从底部拖走的材料：跟手、可中断、松手后按速度交接。 */

(function () {
  if (document.body.dataset.motion !== 'apple') return;

  var detail = document.querySelector('[data-detail]');
  var scrim = document.querySelector('[data-scrim]');
  if (!detail) return;

  var REDUCED = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  var DIM = 0.55;              // 橡皮筋系数
  var DECAY = 0.998;           // 动量投影的衰减
  var PROJECTION = DECAY / (1 - DECAY);

  var y = null;                // 当前展示位置（px，0 = 完全展开）
  var drag = null;
  var raf = null;

  function height() {
    return Math.max(detail.getBoundingClientRect().height, 1);
  }

  function setY(v) {
    y = v;
    detail.style.transform = 'translate3d(0,' + v.toFixed(2) + 'px,0)';
  }

  /* 收起时清掉内联位移，交给 CSS 的 translateY(100%)，高度变了也不会露出边 */
  function park() {
    y = null;
    detail.style.transform = '';
  }

  /* 临界阻尼弹簧：damping 1.0 不回弹，damping 0.85 用于动量收尾 */
  function spring(from, to, velocity, response, damping, done) {
    if (raf) cancelAnimationFrame(raf);
    if (REDUCED) { setY(to); if (done) done(); return; }
    var omega = (2 * Math.PI) / response;
    var x = from, v = velocity, last = performance.now();
    function step(now) {
      var dt = Math.min(0.064, (now - last) / 1000);
      last = now;
      var n = Math.max(1, Math.ceil(dt * 240));
      var h = dt / n;
      for (var i = 0; i < n; i++) {
        var a = -omega * omega * (x - to) - 2 * damping * omega * v;
        v += a * h;
        x += v * h;
      }
      setY(x);
      if (Math.abs(x - to) < 0.4 && Math.abs(v) < 8) { setY(to); raf = null; if (done) done(); return; }
      raf = requestAnimationFrame(step);
    }
    raf = requestAnimationFrame(step);
  }

  /* 越界回弹：overshoot*dim*0.55/(dim+0.55*|overshoot|) */
  function rubber(overshoot, dim) {
    return overshoot * dim * DIM / (dim + DIM * Math.abs(overshoot));
  }

  function ensureHandle() {
    if (detail.querySelector('.sheet-handle')) return;
    var h = document.createElement('div');
    h.className = 'sheet-handle';
    h.setAttribute('role', 'button');
    h.setAttribute('tabindex', '0');
    h.setAttribute('aria-label', '拖动关闭资料面板');
    detail.insertBefore(h, detail.firstChild);
  }

  function open() {
    ensureHandle();
    var h = height();
    var from = y === null ? h : y;
    setY(from);
    spring(from, 0, 0, 0.36, 1);
  }

  function close(velocity) {
    var h = height();
    spring(y === null ? h : y, h, velocity || 0, 0.4, 0.85, park);
    if (scrim) scrim.classList.remove('is-on');
  }

  /* 包一层开关函数：面板的显隐交给弹簧，而不是 CSS 过渡 */
  var baseOpen = window.openDetail;
  window.openDetail = function (on) {
    if (on) {
      open();
      if (baseOpen) baseOpen(true);
    } else {
      if (baseClose) baseClose();
      close(0);
    }
  };
  var baseClose = window.closeDetail;
  window.closeDetail = function () {
    if (baseClose) baseClose();
    close(0);
  };

  /* 每次重绘资料内容后，把把手放回最上面；面板关着时继续收在屏幕外 */
  new MutationObserver(function () {
    ensureHandle();
    if (!document.body.classList.contains('detail-open')) park();
  }).observe(detail, { childList: true });
  ensureHandle();

  /* 拖动：1:1 跟手，速度采样后按投影决定去留 */
  detail.addEventListener('pointerdown', function (e) {
    var handle = e.target.closest && e.target.closest('.sheet-handle');
    if (!handle) return;
    if (e.pointerType === 'mouse' && e.button !== 0) return;
    if (raf) { cancelAnimationFrame(raf); raf = null; }
    handle.setPointerCapture(e.pointerId);
    drag = {
      id: e.pointerId,
      startY: e.clientY,
      base: y === null ? 0 : y,
      lastY: e.clientY,
      lastT: performance.now(),
      velocity: 0
    };
    detail.classList.add('is-dragging');
    e.preventDefault();
  });

  detail.addEventListener('pointermove', function (e) {
    if (!drag || e.pointerId !== drag.id) return;
    var dy = e.clientY - drag.startY;
    var dim = height();
    var next = drag.base + dy;
    if (next < 0) next = rubber(next, dim);
    if (next > dim) next = rubber(next - dim, dim) + dim;
    setY(next);
    var now = performance.now();
    var dt = now - drag.lastT;
    if (dt > 8) {
      drag.velocity = (e.clientY - drag.lastY) / dt * 1000;
      drag.lastY = e.clientY;
      drag.lastT = now;
    }
  });

  function endDrag(e) {
    if (!drag || (e && e.pointerId !== drag.id)) return;
    var v = drag.velocity;
    var dim = height();
    var projected = (y || 0) + (v / 1000) * PROJECTION;
    drag = null;
    detail.classList.remove('is-dragging');
    if (projected > dim * 0.28 || v > 700) {
      if (baseClose) baseClose();
      close(v);
    } else {
      spring(y, 0, v, 0.36, 1);
    }
  }

  detail.addEventListener('pointerup', endDrag);
  detail.addEventListener('pointercancel', endDrag);

  detail.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') { if (baseClose) baseClose(); close(0); }
  });

  /* 初始位置交给 CSS 收在屏幕外；内容还没排版时量不到高度 */
  window.addEventListener('resize', function () {
    if (!drag && (!document.body.classList.contains('detail-open'))) park();
  });
})();
