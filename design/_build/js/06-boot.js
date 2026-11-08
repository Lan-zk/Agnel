/* ============================================================
   启动：装配外壳、模块切换、跨模块事件
   ============================================================ */

function buildBar() {
  const bar = $('[data-bar]');
  if (!bar) return;

  const sideToggle = el('button', {
    type: 'button', class: 'icon-btn side-toggle', html: icon('menu'),
    'aria-label': '打开导航', onclick: () => document.body.classList.toggle('side-open')
  });
  bar.append(sideToggle);

  const brand = el('a', { class: 'brand', href: '#', 'aria-label': '回到阅读' });
  brand.append(el('span', { class: 'brand-mark', 'aria-hidden': 'true' }));
  brand.append(el('span', { class: 'brand-name', text: 'Agnel' }));
  brand.addEventListener('click', e => {
    e.preventDefault();
    reader.open(store.data.progress.unit || 'gen-1');
    goModule('read');
  });
  bar.append(brand);

  const nav = el('nav', { class: 'module-nav', 'aria-label': '模块' });
  MODULE_ORDER.forEach(m => {
    nav.append(el('button', {
      type: 'button',
      class: 'module-btn',
      dataset: { module: m.id },
      html: icon(m.icon) + '<span>' + esc(m.label) + '</span>'
    }));
  });
  bar.append(nav);

  const tools = el('div', { class: 'bar-tools' });
  tools.append(el('button', {
    type: 'button', class: 'icon-btn', html: icon('list'),
    dataset: { openDetail: '1' },
    'aria-label': '打开资料面板', title: '资料面板',
    onclick: () => openDetail(true)
  }));
  tools.append(el('button', {
    type: 'button', class: 'icon-btn', html: icon('speaker'),
    'aria-label': '朗读开关', title: '朗读开关',
    onclick: e => {
      store.data.settings.sound = !store.data.settings.sound;
      store.save();
      e.currentTarget.classList.toggle('is-off', !store.data.settings.sound);
      toast(store.data.settings.sound ? '朗读已开' : '朗读已关');
    }
  }));
  tools.append(el('button', {
    type: 'button', class: 'icon-btn', html: icon('refresh'),
    'aria-label': '清空演示数据', title: '清空演示数据',
    onclick: () => {
      store.reset();
      toast('演示数据已清空');
      renderModule(store.data.settings.active);
    }
  }));
  bar.append(tools);
}

function bindShell() {
  // 模块切换
  document.addEventListener('click', e => {
    const btn = e.target.closest('[data-module]');
    if (!btn) return;
    goModule(btn.dataset.module);
    if (window.innerWidth < 760) closeDetail();
  });

  // 侧栏交互
  document.addEventListener('click', e => {
    const filter = e.target.closest('[data-filter]');
    if (filter) {
      bank.filter = filter.dataset.filter;
      renderModule('bank');
      return;
    }
    const unit = e.target.closest('[data-unit]');
    if (unit) {
      reader.open(unit.dataset.unit);
      return;
    }
    const bookBtn = e.target.closest('[data-book]');
    if (bookBtn) {
      const id = bookBtn.dataset.book;
      reader.openBook = reader.openBook === id ? null : id;
      reader.render();
      reader.refreshSide();
      return;
    }
    const pmode = e.target.closest('[data-pmode]');
    if (pmode) {
      practice.mode = pmode.dataset.pmode;
      practice.scope = practice.mode === 'word-copy' || practice.mode === 'word-dict' ||
        practice.mode === 'word-listen' || practice.mode === 'word-quiz'
        ? (practice.scope === 'verse' ? 'verse' : 'bank') : 'verse';
      practice.buildQueue();
      renderModule('practice');
      return;
    }
    const pscope = e.target.closest('[data-pscope]');
    if (pscope) {
      practice.scope = pscope.dataset.pscope;
      practice.buildQueue();
      renderModule('practice');
      return;
    }
    const btab = e.target.closest('[data-btab]');
    if (btab) {
      basics.tab = btab.dataset.btab;
      renderModule('basics');
      return;
    }
    const stageBtn = e.target.closest('[data-stage-btn]');
    if (stageBtn) {
      setStage(stageBtn.dataset.word, stageBtn.dataset.stageBtn);
      bank.select(stageBtn.dataset.word);
      return;
    }
    const wordRow = e.target.closest('.word-row');
    if (wordRow) {
      bank.select(wordRow.dataset.word);
      return;
    }
    const openWord = e.target.closest('[data-open-word]');
    if (openWord) {
      const w = bankEntry(openWord.dataset.openWord);
      if (w) {
        bank.select(openWord.dataset.openWord);
        return;
      }
    }
    const practiceWord = e.target.closest('[data-practice-word]');
    if (practiceWord) {
      practice.scope = 'bank';
      practice.mode = 'word-copy';
      practice.buildQueue();
      practice.index = Math.max(0, practice.queue.findIndex(i => i.word === practiceWord.dataset.practiceWord));
      goModule('practice');
      return;
    }
    const speakWord = e.target.closest('[data-speak-word]');
    if (speakWord) { speak(speakWord.dataset.speakWord); return; }
    const removeWord = e.target.closest('[data-remove-word]');
    if (removeWord) {
      const key = removeWord.dataset.removeWord;
      delete store.data.bank[key];
      store.save();
      toast(key + ' 已移出词库');
      bank.selected = null;
      renderModule('bank');
      return;
    }
    const cancelAdd = e.target.closest('[data-cancel-add]');
    if (cancelAdd) { renderModule('bank'); return; }
    const gotoVerse = e.target.closest('[data-goto-verse]');
    if (gotoVerse) {
      const target = $('[data-pane="read"] .verse[data-verse="' + gotoVerse.dataset.gotoVerse + '"]');
      if (target) target.scrollIntoView({ block: 'center', behavior: 'smooth' });
      return;
    }
    const gotoPara = e.target.closest('[data-goto-para]');
    if (gotoPara) {
      const target = $('[data-pane="read"] .para:nth-of-type(' + (Number(gotoPara.dataset.gotoPara) + 1) + ')');
      if (target) target.scrollIntoView({ block: 'center', behavior: 'smooth' });
    }
  });

  // 手动添加词条
  document.addEventListener('submit', e => {
    const form = e.target.closest('[data-role="add"]');
    if (!form) return;
    e.preventDefault();
    const data = new FormData(form);
    const word = String(data.get('word') || '').trim().toLowerCase();
    const gloss = String(data.get('gloss') || '').trim();
    if (!word || !gloss) return;
    if (store.data.bank[word]) { toast('这个词已经在词库里了', 'warning'); return; }
    store.data.bank[word] = {
      word: word,
      ipa: String(data.get('ipa') || '').trim(),
      pos: '',
      senses: [gloss],
      forms: [],
      derivations: [],
      etymology: '',
      collocations: [],
      root: '',
      stage: 'new',
      contexts: data.get('source') ? [{ source: String(data.get('source')).trim(), en: '', zh: '' }] : [],
      added: Date.now(),
      right: 0,
      wrong: 0
    };
    store.save();
    toast(word + ' 已加入词库');
    bank.selected = word;
    renderModule('bank');
    bank.select(word);
  });

  // 划词查词
  document.addEventListener('mouseup', () => {
    if (reader) reader.onSelection();
  });

  // 键盘：练习中按 Tab 偷看
  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') {
      if (document.body.classList.contains('side-open')) {
        document.body.classList.remove('side-open');
        return;
      }
      const detail = $('[data-detail]');
      if (detail && document.body.classList.contains('detail-open') && window.innerWidth < 1100) closeDetail();
      return;
    }
    if (reader) reader.onKey(e);
  });

  // 窗口尺寸变了，书页对开要重新摊页
  let resizeTimer = null;
  window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => { if (reader) reader.onResize(); }, 160);
  });

  // 右栏开合（窄屏时是抽屉）
  document.addEventListener('click', e => {
    if (e.target.closest('[data-close-detail]')) { closeDetail(); return; }
    if (e.target.closest('[data-open-detail]')) { openDetail(true); return; }
    if (e.target.closest('[data-scrim]')) { closeDetail(); return; }
  });

  bus.on('practice:queue', word => {
    practice.scope = 'bank';
    practice.mode = 'word-copy';
    practice.buildQueue();
    const i = practice.queue.findIndex(x => x.word === word);
    practice.index = i >= 0 ? i : 0;
    goModule('practice');
  });

  bus.on('bank:changed', () => {
    // 词库变化时刷新词库视图与阅读标记
    const pane = $('[data-pane="bank"]');
    if (pane && !pane.hidden) bank.renderList();
    const rd = $('[data-pane="read"]');
    if (rd && !rd.hidden) {
      const main = $('[data-main]');
      const y = main ? main.scrollTop : 0;
      reader.render();
      if (main) main.scrollTop = y;
    }
    const mod = MODULES[store.data.settings.active];
    if (mod && mod.detail) paintDetail(mod.detail());
  });

  bus.on('reader:changed', () => {
    if (store.data.settings.active === 'read') paintDetail(reader.detail());
  });
}

function closeDetail() {
  document.body.classList.remove('detail-open');
}

function boot() {
  store.load();
  reader.unit = store.data.progress.unit || 'gen-1';
  reader.mode = reader.modes.indexOf(store.data.settings.mode) >= 0
    ? store.data.settings.mode
    : '上下对照';
  buildBar();
  bindShell();
  bindLookup(document.body);
  const hash = (location.hash || '').replace('#/', '');
  const start = MODULES[hash] ? hash
    : (MODULES[store.data.settings.active] ? store.data.settings.active : 'read');
  renderModule(start);
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', boot);
} else {
  boot();
}
