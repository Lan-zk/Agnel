/* ============================================================
   核心：DOM 助手 / 图标 / 本地状态 / 路由 / 查词浮层 / 提示条
   ============================================================ */

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

function el(tag, attrs, children) {
  const node = document.createElement(tag);
  const a = attrs || {};
  for (const key of Object.keys(a)) {
    const val = a[key];
    if (val === null || val === undefined || val === false) continue;
    if (key === 'class') node.className = val;
    else if (key === 'text') node.textContent = val;
    else if (key === 'html') node.innerHTML = val;
    else if (key === 'dataset') Object.assign(node.dataset, val);
    else if (key === 'value') node.value = val;
    else if (key === 'checked') node.checked = !!val;
    else if (key.slice(0, 2) === 'on' && typeof val === 'function') node.addEventListener(key.slice(2).toLowerCase(), val);
    else node.setAttribute(key, val === true ? '' : String(val));
  }
  const list = Array.isArray(children) ? children : (children === undefined ? [] : [children]);
  for (const child of list) {
    if (child === null || child === undefined || child === false) continue;
    node.append(child.nodeType ? child : document.createTextNode(String(child)));
  }
  return node;
}

const CN_DIGITS = ['零', '一', '二', '三', '四', '五', '六', '七', '八', '九'];
function cnNum(n) {
  if (n < 10) return CN_DIGITS[n];
  if (n < 20) return '十' + (n % 10 ? CN_DIGITS[n % 10] : '');
  const tens = Math.floor(n / 10);
  return CN_DIGITS[tens] + '十' + (n % 10 ? CN_DIGITS[n % 10] : '');
}

function esc(str) {
  return String(str).replace(/[&<>"]/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]));
}

/* --- 图标：统一 1.6 线宽的线性图标，不使用符号字符充当图标 --- */
const ICONS = {
  book: 'M4 5.2A2.2 2.2 0 0 1 6.2 3H19v15.4H6.2A2.2 2.2 0 0 0 4 20.6z|M4 18.6A2.2 2.2 0 0 1 6.2 16.4H19',
  search: 'M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14z|M20 20l-4.1-4.1',
  star: 'M12 4.2l2.3 4.9 5.2.7-3.8 3.7 1 5.3-4.7-2.6-4.7 2.6 1-5.3L4.5 9.8l5.2-.7z',
  layers: 'M12 3.2 3.6 7.6 12 12l8.4-4.4z|M3.6 12.2 12 16.6l8.4-4.4|M3.6 16.6 12 21l8.4-4.4',
  pen: 'M4.5 19.5l.9-4 9.6-9.6 3.1 3.1-9.6 9.6z|M13.6 7.3l3.1 3.1',
  type: 'M4.5 19 8.6 6.4 12.7 19|M6.4 15.2h4.4|M14.5 19l3.2-9.8 3.2 9.8|M15.7 16.6h4',
  speaker: 'M4.5 9.4h3.2L12 5.6v12.8L7.7 14.6H4.5z|M15.4 9.4a3.6 3.6 0 0 1 0 5.2|M18 6.8a7.2 7.2 0 0 1 0 10.4',
  close: 'M6.4 6.4 17.6 17.6|M17.6 6.4 6.4 17.6',
  check: 'M5 12.6 9.8 17.4 19 6.8',
  left: 'M14.5 5 7.5 12l7 7',
  right: 'M9.5 5l7 7-7 7',
  arrow: 'M4 12h15|M13 6l6 6-6 6',
  plus: 'M12 5.2v13.6|M5.2 12h13.6',
  filter: 'M4 6.5h16|M7 12v6.4l-2.6-1.6V12z',
  eye: 'M2.8 12S6.3 6.4 12 6.4 21.2 12 21.2 12 17.7 17.6 12 17.6 2.8 12 2.8 12z|M12 9.6a2.4 2.4 0 1 0 0 4.8 2.4 2.4 0 0 0 0-4.8',
  trash: 'M4.8 7h14.4|M9.4 7V4.8h5.2V7|M6.8 7l.9 12.2h8.6L17.2 7',
  refresh: 'M20 12a8 8 0 1 1-2.4-5.7|M20.4 4v4.4H16',
  list: 'M8.4 6.4h11.2|M8.4 12h11.2|M8.4 17.6h11.2|M4.6 6.4h.01|M4.6 12h.01|M4.6 17.6h.01',
  clock: 'M12 3.4a8.6 8.6 0 1 0 0 17.2 8.6 8.6 0 0 0 0-17.2z|M12 7.2V12l3.4 2',
  book2: 'M5 5.2A2.2 2.2 0 0 1 7.2 3H19v18H7.2A2.2 2.2 0 0 1 5 18.8z|M9 8h6.5|M9 11.4h4.5',
  sparkle: 'M12 3.6l1.7 5.1 5.1 1.7-5.1 1.7L12 17.2l-1.7-5.1L5.2 10.4l5.1-1.7z',
  translate: 'M4 6.4h9|M8.5 4.4v2|M11.6 6.4c-.6 4.4-3.2 7.6-6.6 9.2|M6.6 10.6c1.4 2.6 3.6 4.4 6.2 5.2|M13.4 20l3.8-9.4L21 20|M14.8 17h4.8',
  branch: 'M7 4.6v9.2|M7 19.4v.01|M7 4.6a1.6 1.6 0 1 0 0-.01|M7 13.8a1.6 1.6 0 1 0 0-.01|M17 4.6a1.6 1.6 0 1 0 0-.01|M17 4.6c0 4-3.4 4.4-6.4 5.4-1.6.5-3.6 1-3.6 3.4',
  ear: 'M8 8.4a4 4 0 1 1 8 0c0 2.4-1.6 3.4-2.8 4.6-1 1-1.2 2-1.2 3a2.6 2.6 0 0 1-5.2 0|M10.2 8.8a1.8 1.8 0 1 1 3.6 0',
  grid: 'M4.6 4.6h6v6h-6zM13.4 4.6h6v6h-6zM4.6 13.4h6v6h-6zM13.4 13.4h6v6h-6z'
};

function icon(name, cls) {
  const d = ICONS[name] || '';
  const paths = d.split('|').map(p => `<path d="${p}"/>`).join('');
  return `<svg class="ic${cls ? ' ' + cls : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${paths}</svg>`;
}

/* --- 本地状态（原型用 localStorage 模拟离线持久化） --- */
const STORE_KEY = 'bes.demo.v1';
const store = {
  data: {
    bank: {},
    progress: { unit: 'gen-1', anchor: 0, scroll: 0 },
    settings: { mode: '上下对照', sound: true, active: 'read' }
  },
  load() {
    let raw = null;
    try {
      raw = localStorage.getItem(STORE_KEY);
    } catch (err) { /* 无痕模式或存储不可用时继续以内存态运行 */ }
    if (!raw) { this.seedDemo(); return; }
    try {
      if (raw) {
        const parsed = JSON.parse(raw);
        this.data.bank = parsed.bank || {};
        this.data.progress = Object.assign(this.data.progress, parsed.progress || {});
        this.data.settings = Object.assign(this.data.settings, parsed.settings || {});
      }
    } catch (err) { /* 无痕模式或存储不可用时继续以内存态运行 */ }
  },
  save() {
    try { localStorage.setItem(STORE_KEY, JSON.stringify(this.data)); } catch (err) { /* 忽略写入失败 */ }
  },
  reset() {
    this.data.bank = {};
    this.data.progress = { unit: 'gen-1', anchor: 0, scroll: 0 };
    this.data.settings.mode = '上下对照';
    this.save();
  },
  seedDemo() {
    DEMO_BANK.forEach(d => {
      const entry = DATA.dict[d.word];
      if (!entry || this.data.bank[d.word]) return;
      this.data.bank[d.word] = {
        word: d.word,
        ipa: entry.ipa || '',
        pos: entry.pos || '',
        senses: (entry.senses || []).slice(),
        forms: (entry.forms || []).slice(),
        derivations: (entry.derivations || []).slice(),
        etymology: entry.etymology || '',
        collocations: (entry.collocations || []).slice(),
        root: entry.root || '',
        stage: d.stage,
        contexts: [d.ctx],
        added: Date.now() - Math.floor(Math.random() * 6e7),
        right: d.stage === 'known' ? 3 : 0,
        wrong: 0
      };
    });
    this.data.demo = true;
    this.save();
  }
};

const STAGES = {
  new: { label: '生词', next: 'learning' },
  learning: { label: '学习中', next: 'known' },
  known: { label: '已掌握', next: 'new' },
  ignored: { label: '忽略', next: 'new' }
};

function stageOf(word) {
  const entry = store.data.bank[String(word).toLowerCase()];
  return entry ? entry.stage : '';
}

function inBank(word) {
  return !!store.data.bank[String(word).toLowerCase()];
}

function bankEntry(word) {
  return store.data.bank[String(word).toLowerCase()] || null;
}

/* --- 事件总线：模块之间通过它联动 --- */
const bus = (() => {
  const handlers = {};
  return {
    on(name, fn) { (handlers[name] = handlers[name] || []).push(fn); },
    emit(name, payload) { (handlers[name] || []).forEach(fn => fn(payload)); }
  };
})();

/* --- 提示条 --- */
function toast(message, kind) {
  let box = $('.toast');
  if (!box) {
    box = el('div', { class: 'toast', role: 'status', 'aria-live': 'polite' });
    document.body.append(box);
  }
  const item = el('div', { class: 'toast-item' + (kind ? ' is-' + kind : ''), text: message });
  box.append(item);
  requestAnimationFrame(() => item.classList.add('is-in'));
  setTimeout(() => {
    item.classList.remove('is-in');
    setTimeout(() => item.remove(), 260);
  }, 2800);
}

/* --- 朗读（系统语音，离线可用） --- */
function speak(text, rate) {
  if (!store.data.settings.sound) { toast('朗读已关闭', 'warning'); return; }
  if (!('speechSynthesis' in window)) { toast('当前浏览器不支持朗读', 'warning'); return; }
  window.speechSynthesis.cancel();
  const utter = new SpeechSynthesisUtterance(text);
  utter.lang = 'en-US';
  utter.rate = rate || 0.92;
  window.speechSynthesis.speak(utter);
}

/* --- 词典检索 --- */
function resolveWord(key) {
  const lower = String(key).toLowerCase();
  if (DATA.dict[lower]) return { entry: DATA.dict[lower], via: null, owner: lower };
  for (const name of Object.keys(DATA.dict)) {
    const entry = DATA.dict[name];
    if (entry.forms && entry.forms.indexOf(lower) >= 0) return { entry, via: 'form', owner: name };
    if (entry.derivations && entry.derivations.indexOf(lower) >= 0) return { entry, via: 'deriv', owner: name };
  }
  return { entry: null, via: null, owner: null };
}

function rootOf(id) {
  return DATA.roots.find(r => r.root === id) || null;
}

/* --- 查词浮层 --- */
const lookup = {
  node: null,
  anchorRect: null,
  key: '',

  ensure() {
    if (this.node) return this.node;
    this.node = el('div', { class: 'lookup', role: 'dialog', 'aria-label': '查词', hidden: true });
    document.body.append(this.node);
    document.addEventListener('mousedown', e => {
      if (!this.node || this.node.hidden) return;
      if (this.node.contains(e.target)) return;
      if (e.target.closest && e.target.closest('.w')) return;
      this.close();
    });
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape' && this.node && !this.node.hidden) { this.close(); }
    });
    window.addEventListener('resize', () => this.reposition());
    document.addEventListener('scroll', () => this.reposition(), true);
    return this.node;
  },

  open(key, ctx, anchorRect) {
    const node = this.ensure();
    this.key = key;
    this.anchorRect = anchorRect;
    this.ctx = ctx || {};
    node.innerHTML = this.build(key, ctx || {});
    node.hidden = false;
    node.classList.remove('is-in');
    requestAnimationFrame(() => {
      this.reposition();
      node.classList.add('is-in');
    });
    node.querySelector('.lookup-close').focus();
  },

  close() {
    if (!this.node) return;
    this.node.hidden = true;
    this.node.classList.remove('is-in');
    this.anchorRect = null;
  },

  reposition() {
    if (!this.node || this.node.hidden || !this.anchorRect) return;
    const rect = this.anchorRect;
    const pad = 14;
    const box = this.node.getBoundingClientRect();
    let top = rect.bottom + 12;
    if (top + box.height > window.innerHeight - pad) {
      const above = rect.top - 12 - box.height;
      top = above > pad ? above : Math.max(pad, window.innerHeight - box.height - pad);
    }
    let left = rect.left + rect.width / 2 - box.width / 2;
    left = Math.min(Math.max(pad, left), window.innerWidth - box.width - pad);
    this.node.style.left = Math.round(left) + 'px';
    this.node.style.top = Math.round(top) + 'px';
  },

  build(key, ctx) {
    const lower = String(key).toLowerCase();
    const { entry, via, owner } = resolveWord(lower);
    const saved = bankEntry(lower);
    const stage = saved ? saved.stage : 'new';

    const parts = [];

    parts.push(
      `<div class="lookup-head">` +
        `<span class="lookup-src">${esc(ctx.source || '本地词条')}</span>` +
        `<span class="lookup-stage" data-role="stage">` +
          `<button type="button" data-stage="new" aria-label="标记为生词">生词</button>` +
          `<button type="button" data-stage="learning" aria-label="标记为学习中">学习中</button>` +
          `<button type="button" data-stage="known" aria-label="标记为已掌握">已掌握</button>` +
        `</span>` +
        `<button type="button" class="lookup-close" aria-label="关闭">${icon('close')}</button>` +
      `</div>`
    );

    parts.push(
      `<div class="lookup-word">` +
        `<h2>${esc(key)}</h2>` +
        (entry ? `<span class="lookup-ipa">${esc(entry.ipa)}</span>` : '') +
        `<button type="button" class="icon-btn" data-act="speak" aria-label="朗读">${icon('speaker')}</button>` +
      `</div>`
    );

    if (entry && entry.pos) parts.push(`<p class="lookup-pos">${esc(entry.pos)}</p>`);

    if (via === 'form' && owner !== lower) {
      parts.push(`<p class="lookup-note">这是 <b>${esc(owner)}</b> 的词形变化形式。</p>`);
    } else if (via === 'deriv' && owner !== lower) {
      parts.push(`<p class="lookup-note">这是 <b>${esc(owner)}</b> 的派生词。</p>`);
    }

    if (entry) {
      const senses = entry.senses.map(s => `<li>${esc(s)}</li>`).join('');
      parts.push(`<ul class="lookup-senses">${senses}</ul>`);

      const rows = [];
      if (entry.forms && entry.forms.length) {
        rows.push(row('词形变化', entry.forms.map(w => `<a href="#" data-w-link="${esc(w)}">${esc(w)}</a>`).join(' ')));
      }
      if (entry.derivations && entry.derivations.length) {
        rows.push(row('派生关系', entry.derivations.map(w => `<a href="#" data-w-link="${esc(w)}">${esc(w)}</a>`).join(' ')));
      }
      if (entry.etymology) rows.push(row('词源', esc(entry.etymology)));
      if (entry.collocations && entry.collocations.length) {
        rows.push(row('常见搭配', entry.collocations.map(c => `<span class="phrase">${esc(c)}</span>`).join(' ')));
      }
      if (rows.length) parts.push(`<dl class="lookup-rows">${rows.join('')}</dl>`);
    } else {
      parts.push(`<p class="lookup-note">词库还没有这个词。你可以自己加一条释义。</p>`);
    }

    const addForm = entry ? '' :
      `<form class="lookup-add" data-role="add-form">` +
        `<input type="text" name="gloss" placeholder="写一条你自己的释义" aria-label="释义" />` +
        `<button type="submit" class="btn btn-quiet">加入词库</button>` +
      `</form>`;

    parts.push(
      `<div class="lookup-actions">` +
        `<button type="button" class="btn" data-act="save">${saved ? '已收藏' : '收藏这个词'}</button>` +
        `<button type="button" class="btn btn-quiet" data-act="practice">练习这个词</button>` +
      `</div>` + addForm
    );

    if (ctx.sentence) {
      parts.push(
        `<div class="lookup-context">` +
          `<p class="ctx-en">${esc(ctx.sentence)}</p>` +
          (ctx.zh ? `<p class="ctx-zh">${esc(ctx.zh)}</p>` : '') +
        `</div>`
      );
    }

    parts.push('<div class="lookup-foot"><span>Esc 关闭</span></div>');

    return parts.join('');
  },

  syncStage() {
    if (!this.node) return;
    const buttons = this.node.querySelectorAll('[data-stage]');
    const current = bankEntry(this.key);
    const stage = current ? current.stage : 'new';
    buttons.forEach(b => {
      const isOn = b.dataset.stage === stage;
      b.classList.toggle('is-on', isOn);
      b.setAttribute('aria-pressed', isOn ? 'true' : 'false');
    });
    const saveBtn = this.node.querySelector('[data-act="save"]');
    if (saveBtn) saveBtn.textContent = current ? '已收藏' : '收藏这个词';
  }
};

function row(label, html) {
  return `<div class="lookup-row"><dt>${esc(label)}</dt><dd>${html}</dd></div>`;
}

/* --- 路由 --- */
const MODULES = {};
const MODULE_ORDER = [
  { id: 'read', label: '阅读', icon: 'book' },
  { id: 'bank', label: '词库', icon: 'layers' },
  { id: 'practice', label: '练习', icon: 'pen' },
  { id: 'basics', label: '英语基础', icon: 'type' }
];

/* 详情面板：抽屉形态时带一个关闭按钮，常驻列形态由样式隐藏 */
function paintDetail(html) {
  const detail = $('[data-detail]');
  if (!detail) return;
  const body = html || '';
  detail.innerHTML = body
    ? '<button type="button" class="detail-close" data-close-detail aria-label="关闭面板">' + icon('close') + '</button>' + body
    : '';
  detail.classList.toggle('is-empty', !body.trim());
}

function renderModule(id) {
  const mod = MODULES[id];
  if (!mod) return;
  $$('[data-pane]').forEach(pane => { pane.hidden = pane.dataset.pane !== id; });
  document.body.dataset.readerMode = id === 'read' ? reader.mode : '';
  const side = $('[data-side]');
  if (side) side.innerHTML = mod.side ? mod.side() : '';
  paintDetail(mod.detail ? mod.detail() : '');
  $$('[data-module]').forEach(btn => {
    const on = btn.dataset.module === id;
    btn.classList.toggle('is-on', on);
    btn.setAttribute('aria-current', on ? 'page' : 'false');
  });
  const main = $('[data-main]');
  if (main) main.scrollTop = 0;
  if (mod.render) mod.render();
  bus.emit('module:rendered', id);
}

function goModule(id) {
  if (!MODULES[id]) return;
  // 换模块时把资料面板收起来，别让它盖住新模块（第五套是底部浮层）
  if (document.body.classList.contains('detail-open') && typeof closeDetail === 'function') closeDetail();
  store.data.settings.active = id;
  store.save();
  renderModule(id);
  if (location.hash !== '#/' + id) history.replaceState(null, '', '#/' + id);
}

function openDetail(open) {
  const detail = $('[data-detail]');
  if (!detail) return;
  document.body.classList.toggle('detail-open', !!open);
}

function setStage(word, stage) {
  const key = String(word).toLowerCase();
  const entry = bankEntry(key);
  if (!entry) return;
  entry.stage = stage;
  store.save();
  lookup.syncStage();
  bus.emit('bank:changed', key);
}

function toggleSave(word, ctx) {
  const key = String(word).toLowerCase();
  const existing = bankEntry(key);
  if (existing) {
    delete store.data.bank[key];
    store.save();
    toast(`已从词库移除 ${key}`);
  } else {
    const { entry } = resolveWord(key);
    const src = ctx || {};
    store.data.bank[key] = {
      word: key,
      ipa: entry ? entry.ipa : '',
      pos: entry ? entry.pos : (src.gloss || ''),
      senses: entry ? entry.senses.slice() : (src.gloss ? [src.gloss] : []),
      forms: entry ? entry.forms.slice() : [],
      derivations: entry ? entry.derivations.slice() : [],
      etymology: entry ? entry.etymology : '',
      collocations: entry ? entry.collocations.slice() : [],
      root: entry ? entry.root : '',
      stage: 'new',
      contexts: src.sentence ? [{ source: src.source || '', en: src.sentence, zh: src.zh || '' }] : [],
      added: Date.now(),
      right: 0,
      wrong: 0
    };
    store.save();
    toast(`已收藏 ${key}`);
  }
  lookup.syncStage();
  bus.emit('bank:changed', key);
}

function bindLookup(root) {
  root.addEventListener('click', e => {
    const stageBtn = e.target.closest('[data-stage]');
    if (stageBtn) {
      const key = lookup.key;
      if (!inBank(key)) toggleSave(key, lookup.ctx);
      setStage(key, stageBtn.dataset.stage);
      return;
    }
    const link = e.target.closest('[data-w-link]');
    if (link) {
      e.preventDefault();
      const w = link.dataset.wLink;
      const target = document.querySelector(`[data-w="${CSS.escape(w.toLowerCase())}"]`);
      const rect = target ? target.getBoundingClientRect() : (lookup.anchorRect || { left: window.innerWidth / 2, right: window.innerWidth / 2, top: 60, bottom: 60 });
      lookup.open(w, lookup.ctx, rect);
      return;
    }
    const act = e.target.closest('[data-act]');
    if (!act) return;
    const kind = act.dataset.act;
    if (kind === 'speak') speak(lookup.key);
    else if (kind === 'save') toggleSave(lookup.key, lookup.ctx);
    else if (kind === 'practice') {
      bus.emit('practice:queue', lookup.key);
      lookup.close();
    }
  });
  root.addEventListener('submit', e => {
    const form = e.target.closest('[data-role="add-form"]');
    if (!form) return;
    e.preventDefault();
    const gloss = form.querySelector('[name="gloss"]').value.trim();
    if (!gloss) { toast('先写一条释义', 'warning'); return; }
    toggleSave(lookup.key, Object.assign({}, lookup.ctx, { gloss }));
    toast(`${lookup.key} 已加入词库`);
  });
}
