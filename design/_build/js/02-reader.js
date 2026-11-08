/* ============================================================
   阅读模块：逐节对照、点词查词、划词查词、阅读进度
   ============================================================ */

const reader = {
  unit: 'gen-1',
  openBook: 'gen',
  mode: '上下对照',
  modes: ['上下对照', '书页对开', '英文', '中文'],
  bookIndex: 0,
  bookPageCount: 1,

  chaptersOf(bookId) {
    return Object.keys(DATA.chapters)
      .filter(key => DATA.chapters[key].book === bookId)
      .sort((a, b) => DATA.chapters[a].chapter - DATA.chapters[b].chapter);
  },

  currentUnit() {
    if (this.unit.indexOf('art:') === 0) {
      return DATA.articles.find(a => a.id === this.unit.slice(4));
    }
    return DATA.chapters[this.unit] || DATA.chapters['gen-1'];
  },

  currentLabel() {
    const u = this.currentUnit();
    if (!u) return '';
    if (u.paras) return u.title;
    const book = DATA.books.find(b => b.id === u.book);
    return (book ? book.zh : '') + ' ' + u.chapter;
  },

  // ---- 侧栏：书卷 / 章 / 文章 ----
  side() {
    const parts = [];
    parts.push('<div class="side-scroll">');
    parts.push('<div class="side-group">');
    parts.push('<h2 class="side-title">经文</h2>');
    parts.push('<div class="book-list">');
    for (const book of DATA.books) {
      const chapters = this.chaptersOf(book.id);
      if (!chapters.length) continue;
      const open = this.openBook === book.id;
      parts.push('<div class="book' + (open ? ' is-open' : '') + '">');
      parts.push('<button type="button" class="book-item' + (open ? ' is-on' : '') + '" data-book="' + book.id + '" aria-expanded="' + (open ? 'true' : 'false') + '">' +
        '<span>' + esc(book.zh) + '</span>' +
        '<span class="book-count">' + chapters.length + '</span>' +
        '</button>');
      if (open) {
        parts.push('<div class="chapter-list">');
        for (const key of chapters) {
          const ch = DATA.chapters[key];
          const on = this.unit === key;
          parts.push('<button type="button" class="chapter-chip' + (on ? ' is-on' : '') + '" data-unit="' + key + '"' +
            (on ? ' aria-current="true"' : '') + '>' + ch.chapter + '</button>');
        }
        parts.push('</div>');
      }
      parts.push('</div>');
    }
    parts.push('</div>');
    parts.push('</div>');

    parts.push('<div class="side-group">');
    parts.push('<h2 class="side-title">文章</h2>');
    parts.push('<div class="book-list">');
    for (const art of DATA.articles) {
      const on = this.unit === 'art:' + art.id;
      parts.push('<button type="button" class="book-item' + (on ? ' is-on' : '') + '" data-unit="art:' + art.id + '">' +
        '<span>' + esc(art.title) + '</span></button>');
    }
    parts.push('</div>');
    parts.push('</div>');
    parts.push('</div>');
    return parts.join('');
  },

  // ---- 主区：正文 ----
  render() {
    const pane = $('[data-pane="read"]');
    const u = this.currentUnit();
    if (!u) return;
    pane.innerHTML = '';
    pane.dataset.readerMode = this.mode;
    document.body.dataset.readerMode = this.mode;

    const head = el('div', { class: 'reader-head' });
    head.append(el('div', { class: 'reader-ref' }, [
      el('h1', { class: 'reader-title', text: this.currentLabel() }),
      el('p', { class: 'reader-sub', text: u.paras ? (u.paras.length + ' 段') : (u.verses.length + ' 节') })
    ]));

    const controls = el('div', { class: 'reader-controls' });
    const seg = el('div', { class: 'seg', role: 'group', 'aria-label': '对照方式' });
    this.modes.forEach(m => {
      const b = el('button', {
        type: 'button',
        class: 'seg-btn' + (this.mode === m ? ' is-on' : ''),
        text: m,
        'aria-pressed': this.mode === m ? 'true' : 'false',
        onclick: () => {
          this.mode = m;
          this.bookIndex = 0;
          store.data.settings.mode = m;
          store.save();
          this.render();
        }
      });
      seg.append(b);
    });
    controls.append(seg);
    controls.append(el('button', {
      type: 'button',
      class: 'icon-btn',
      html: icon('speaker'),
      'aria-label': '朗读本段',
      onclick: () => speak(this.readableText(u))
    }));
    head.append(controls);
    pane.append(head);

    const body = el('div', { class: 'reader-body' });
    body.dataset.mode = this.mode;

    const showEn = this.mode !== '中文';
    const showZh = this.mode !== '英文';

    if (this.mode === '书页对开') {
      // 书页对开：内容由 bookPaginate 摊页
    } else if (u.paras) {
      u.paras.forEach((p, i) => {
        const sec = el('section', { class: 'para' });
        if (showEn) sec.append(el('p', { class: 'para-en', lang: 'en', html: this.markup(p.en) }));
        if (showZh) sec.append(el('p', { class: 'para-zh', lang: 'zh-Hans', text: p.zh }));
        sec.addEventListener('click', e => this.onWordClick(e));
        body.append(sec);
      });
    } else {
      u.verses.forEach(v => {
        const sec = el('section', { class: 'verse', dataset: { verse: v.v } });
        sec.append(el('span', { class: 'verse-n', text: String(v.v) }));
        const text = el('div', { class: 'verse-body' });
        if (showEn) {
          text.append(el('p', {
            class: 'verse-en',
            lang: 'en',
            html: this.markup(v.en)
          }));
        }
        if (showZh) {
          text.append(el('p', {
            class: 'verse-zh',
            lang: 'zh-Hans',
            html: esc(v.zh)
          }));
        }
        sec.append(text);
        sec.addEventListener('click', e => this.onWordClick(e));
        body.append(sec);
      });
    }

    pane.append(body);
    if (this.mode === '书页对开') this.bookPaginate(u, body);
    pane.append(this.pager(u));

    this.restoreScroll();
    this.saveUnit();
  },

  pager(u) {
    const wrap = el('nav', { class: 'reader-pager' });
    const units = [];
    for (const book of DATA.books) units.push.apply(units, this.chaptersOf(book.id));
    DATA.articles.forEach(a => units.push('art:' + a.id));
    const idx = units.indexOf(this.unit);
    const prev = idx > 0 ? units[idx - 1] : null;
    const next = idx < units.length - 1 ? units[idx + 1] : null;
    const labelOf = key => {
      if (key.indexOf('art:') === 0) {
        const a = DATA.articles.find(x => x.id === key.slice(4));
        return a ? a.title : '';
      }
      const c = DATA.chapters[key];
      if (!c) return '';
      const b = DATA.books.find(x => x.id === c.book);
      return b.zh + ' ' + c.chapter;
    };
    if (prev) {
      wrap.append(el('button', {
        type: 'button', class: 'pager-btn', onclick: () => this.open(prev),
        html: icon('left') + '<span>' + esc(labelOf(prev)) + '</span>'
      }));
    } else {
      wrap.append(el('span', { class: 'pager-spacer' }));
    }
    wrap.append(el('span', { class: 'pager-now', text: (idx + 1) + ' / ' + units.length }));
    if (next) {
      wrap.append(el('button', {
        type: 'button', class: 'pager-btn pager-next', onclick: () => this.open(next),
        html: '<span>' + esc(labelOf(next)) + '</span>' + icon('right')
      }));
    }
    return wrap;
  },

  // ---------------- 书页对开：仿真实对开本 ----------------

  bookMeta(u) {
    const book = DATA.books.find(b => b.id === u.book) || { zh: '', en: '' };
    return {
      zh: book.zh,
      en: book.en,
      abbr: book.abbr || (book.en || '').slice(0, 3),
      chapter: u.chapter || 1
    };
  },

  // 排成行：每节一行，并记住段号与段落标题
  bookRows(u) {
    const rows = [];
    if (u.paras) {
      u.paras.forEach((p, i) => rows.push({ en: p.en, zh: p.zh, v: 0, showSec: 0, title: '', i: i }));
      return rows;
    }
    const secs = u.sections || [];
    let prev = 0;
    u.verses.forEach((v, i) => {
      let sec = null;
      for (const s of secs) {
        if (v.v >= s.from && v.v <= s.to) { sec = s; break; }
      }
      const num = sec ? sec.n : 0;
      const fresh = num && num !== prev;
      rows.push({
        en: v.en,
        zh: v.zh,
        v: v.v,
        showSec: fresh ? num : 0,
        title: fresh && sec && sec.titleEn ? sec.titleEn : '',
        i: i
      });
      prev = num;
    });
    return rows;
  },

  bookHead(u, side, meta) {
    if (u.paras) return u.title;
    // 左页英文在前，右页中文在前；窄屏一栏里也是英文在上
    return side === 'recto'
      ? meta.zh + ' · ' + meta.en + ' ' + meta.chapter
      : meta.en + ' · ' + meta.zh + ' ' + meta.chapter;
  },

  // 窄屏放不下两页，就一栏一栏地读：英文在上、中文在下，仍按节对齐
  bookPage(u, side, rows, meta, pageNo, notes) {
    const both = side === 'both';
    const lang = side === 'recto' ? 'zh' : 'en';
    const page = el('div', { class: 'page page-' + side });

    const top = el('div', { class: 'page-top' });
    top.append(el('span', { class: 'page-run', text: this.bookHead(u, side, meta) }));
    page.append(top);
    if (pageNo) page.append(el('span', { class: 'page-num', text: String(pageNo) }));

    if (!u.paras && !both) {
      const badge = el('div', { class: 'page-badge' });
      badge.append(el('span', { class: 'pb-zh', text: meta.zh.slice(0, 1) }));
      badge.append(el('span', { class: 'pb-en', text: meta.abbr }));
      page.append(badge);
    }

    const text = el('div', { class: 'page-text' });
    rows.forEach(row => {
      const r = el('div', { class: 'vrow', dataset: row.v ? { verse: String(row.v) } : {} });
      r.append(el('span', { class: 'vsec', text: row.showSec ? String(row.showSec) : '' }));
      const cell = el('div', { class: 'vcell' });
      if (lang === 'en' && row.title) cell.append(el('p', { class: 'vtitle', text: row.title }));
      const p = el('p', { class: 'vtext v-' + lang });
      if (lang === 'en') {
        p.setAttribute('lang', 'en');
        p.innerHTML = this.markup(row.en);
      } else {
        p.setAttribute('lang', 'zh-Hans');
        p.textContent = row.zh;
      }
      if (row.v && !both) p.append(el('sup', { class: 'vn', text: String(row.v) }));
      cell.append(p);
      if (both) {
        const q = el('p', { class: 'vtext v-zh' });
        q.setAttribute('lang', 'zh-Hans');
        q.textContent = row.zh;
        if (row.v) q.append(el('sup', { class: 'vn', text: String(row.v) }));
        cell.append(q);
      }
      r.append(cell);
      text.append(r);
    });
    text.addEventListener('click', e => this.onWordClick(e));
    page.append(text);

    if (notes && notes.length) {
      const box = el('div', { class: 'page-notes' });
      notes.forEach(n => box.append(el('p', { class: 'fn', text: n })));
      page.append(box);
    }
    return page;
  },

  bookProbe(u, rows, meta, narrow) {
    const probe = el('div', { class: 'spread is-probe' + (narrow ? ' is-narrow' : '') });
    if (narrow) probe.append(this.bookPage(u, 'both', rows, meta, 0, []));
    else {
      probe.append(this.bookPage(u, 'verso', rows, meta, 0, []));
      probe.append(this.bookPage(u, 'recto', rows, meta, 0, []));
    }
    return probe;
  },

  // 按页高把整章切成若干对开，再把两栏的行高钉齐
  bookPaginate(u, host) {
    const rows = this.bookRows(u);
    const meta = this.bookMeta(u);

    // 先摊一份探针页：两栏都放整章，用来量宽度和行高
    let narrow = window.innerWidth <= 900;
    let probe = this.bookProbe(u, rows, meta, narrow);
    const old = host.querySelector('.spread');
    if (old) host.replaceChild(probe, old);
    else host.append(probe);
    if (!narrow && probe.getBoundingClientRect().width < 820) {
      // 版心不够宽，两页会挤在一起，改成一栏一栏读
      narrow = true;
      const narrowProbe = this.bookProbe(u, rows, meta, true);
      host.replaceChild(narrowProbe, probe);
      probe = narrowProbe;
    }

    // 页高按视口里剩下的地方算，整页刚好摊在一屏里
    const scroller = probe.closest('.main');
    let top = probe.getBoundingClientRect().top;
    if (scroller) {
      const sr = scroller.getBoundingClientRect();
      top = top - sr.top + scroller.scrollTop;
    }
    const room = scroller ? scroller.clientHeight : window.innerHeight;
    const h = Math.max(340, Math.min(860, room - top - 64));
    host.style.setProperty('--page-h', h + 'px');

    const leftRows = narrow ? probe.querySelectorAll('.vrow') : probe.querySelectorAll('.page-verso .vrow');
    const rightRows = narrow ? probe.querySelectorAll('.vrow') : probe.querySelectorAll('.page-recto .vrow');
    const heights = [];
    const gaps = [];
    for (let i = 0; i < rows.length; i++) {
      const a = leftRows[i] ? leftRows[i].getBoundingClientRect().height : 0;
      const b = rightRows[i] ? rightRows[i].getBoundingClientRect().height : 0;
      heights.push(Math.max(a, b));
      const cs = leftRows[i] ? getComputedStyle(leftRows[i]) : null;
      gaps.push(cs ? (parseFloat(cs.marginBottom) || 0) : 0);
    }
    // 版心净高：clientHeight 含内边距，摊页只按内边距以内的部分算
    const tbox = probe.querySelector('.page-text');
    const tcs = tbox ? getComputedStyle(tbox) : null;
    const avail = tbox
      ? Math.max(0, tbox.clientHeight - (parseFloat(tcs.paddingTop) || 0) - (parseFloat(tcs.paddingBottom) || 0))
      : 0;

    // 分页：每页的脚注要占地方，装不下就少放几行，脚注跟着它提到的那一节走
    this.probeEl = probe;
    const pages = [];
    let i = 0;
    while (i < rows.length) {
      let take = rows.length - i;
      let pageRows = rows.slice(i);
      let notes = { en: [], zh: [] };
      let reserve = 0;
      for (;;) {
        pageRows = rows.slice(i, i + take);
        notes = this.bookNotes(u, pageRows, i + take >= rows.length);
        reserve = Math.max(this.notesHeight(notes.zh), this.notesHeight(notes.en));
        let total = 0;
        pageRows.forEach(r => { total += heights[r.i] + gaps[r.i]; });
        if (total + reserve <= avail || take <= 1) break;
        take -= 1;
      }
      pages.push({ rows: pageRows, notes: notes });
      i += take;
    }
    if (!pages.length) pages.push({ rows: rows.slice(0, 1), notes: { en: [], zh: [] } });

    const idx = Math.max(0, Math.min(this.bookIndex, pages.length - 1));
    this.bookIndex = idx;
    this.bookPageCount = pages.length;

    const pageRows = pages[idx].rows;
    const notes = pages[idx].notes;
    const base = this.bookBaseNo(u, idx);

    const book = el('div', { class: 'spread' + (narrow ? ' is-narrow' : '') });
    if (narrow) {
      book.append(this.bookPage(u, 'both', pageRows, meta, base, notes.en.concat(notes.zh)));
    } else {
      book.append(this.bookPage(u, 'verso', pageRows, meta, base, notes.en));
      book.append(this.bookPage(u, 'recto', pageRows, meta, base + 1, notes.zh));
    }
    const vs = book.querySelectorAll(narrow ? '.vrow' : '.page-verso .vrow');
    const es = narrow ? [] : book.querySelectorAll('.page-recto .vrow');
    pageRows.forEach((row, k) => {
      const h = heights[row.i];
      if (vs[k]) vs[k].style.height = h + 'px';
      if (es[k]) es[k].style.height = h + 'px';
    });
    host.replaceChild(book, probe);

    const oldNav = host.querySelector('.book-nav');
    if (oldNav) oldNav.remove();
    host.append(this.bookNav(pages.length, idx, base));
  },

  bookBaseNo(u, spreadIdx) {
    if (u.paras) return 1 + 2 * spreadIdx;
    const list = this.chaptersOf(u.book);
    const ci = Math.max(0, list.indexOf(this.unit));
    return 2 + 2 * (ci + spreadIdx);
  },

  // 脚注块要占多高：在探针页里量一次
  notesHeight(list) {
    if (!list || !list.length) return 0;
    const probe = this.probeEl;
    if (!probe) return 0;
    const box = probe.querySelector('.page-text');
    const width = box ? Math.round(box.getBoundingClientRect().width) : 320;
    const d = document.createElement('div');
    d.className = 'page-notes';
    d.style.cssText = 'position:absolute;left:-9999px;top:0;width:' + width + 'px;visibility:hidden;';
    list.forEach(n => d.appendChild(el('p', { class: 'fn', text: n })));
    probe.appendChild(d);
    const h = d.getBoundingClientRect().height;
    d.remove();
    return h;
  },

  // 脚注挂在它提到的那一节所在的那一页上
  bookNotes(u, pageRows, isLast) {
    const notes = u.notes || {};
    const pick = list => {
      const out = [];
      (list || []).forEach(n => {
        const m = String(n).match(/(\d+):(\d+)/);
        if (!m) {
          if (isLast) out.push(n);
          return;
        }
        if (pageRows.some(r => r.v === Number(m[2]))) out.push(n);
      });
      return out;
    };
    return { en: pick(notes.en), zh: pick(notes.zh) };
  },

  bookNav(count, idx, base) {
    const nav = el('nav', { class: 'book-nav' });
    const prev = el('button', {
      type: 'button',
      class: 'pager-btn',
      html: icon('left') + '<span>前一页</span>',
      onclick: () => this.bookGo(-1)
    });
    if (idx <= 0) prev.setAttribute('disabled', 'disabled');
    nav.append(prev);
    nav.append(el('span', {
      class: 'book-nav-now',
      text: (idx + 1) + ' / ' + count + ' · 第 ' + base + '–' + (base + 1) + ' 页'
    }));
    const next = el('button', {
      type: 'button',
      class: 'pager-btn',
      html: '<span>后一页</span>' + icon('right'),
      onclick: () => this.bookGo(1)
    });
    if (idx >= count - 1) next.setAttribute('disabled', 'disabled');
    nav.append(next);
    return nav;
  },

  bookGo(delta) {
    const next = Math.max(0, Math.min(this.bookIndex + delta, this.bookPageCount - 1));
    if (next === this.bookIndex) return;
    this.bookIndex = next;
    const body = $('[data-pane="read"] .reader-body');
    if (body) this.bookPaginate(this.currentUnit(), body);
  },

  onKey(e) {
    if (this.mode !== '书页对开') return;
    if (store.data.settings.active !== 'read') return;
    const tag = document.activeElement ? document.activeElement.tagName : '';
    if (tag === 'INPUT' || tag === 'TEXTAREA') return;
    if (e.key === 'ArrowLeft') { this.bookGo(-1); e.preventDefault(); }
    else if (e.key === 'ArrowRight') { this.bookGo(1); e.preventDefault(); }
  },

  // 窗口变了，重新摊页
  onResize() {
    if (this.mode !== '书页对开') return;
    if (store.data.settings.active !== 'read') return;
    const body = document.querySelector('[data-pane="read"] .reader-body');
    if (!body || !body.querySelector('.spread')) return;
    this.bookPaginate(this.currentUnit(), body);
  },

  // 取纯文本，去掉内嵌的节号
  plainText(node) {
    if (!node) return '';
    const clone = node.cloneNode(true);
    clone.querySelectorAll('.vsup, .vn').forEach(n => n.remove());
    return clone.textContent.trim();
  },

  // 把英文拆成词与标点，词可点
  markup(text) {
    const re = /([A-Za-z][A-Za-z'’-]*)/g;
    let out = '';
    let last = 0;
    let m;
    while ((m = re.exec(text)) !== null) {
      if (m.index > last) out += esc(text.slice(last, m.index));
      const raw = m[0];
      const key = raw.toLowerCase();
      const saved = inBank(key);
      const stage = saved ? bankEntry(key).stage : '';
      out += '<span class="w' + (saved ? ' w-saved is-' + stage : '') + '" data-w="' + esc(key) + '" role="button" tabindex="0">' + esc(raw) + '</span>';
      last = m.index + raw.length;
    }
    if (last < text.length) out += esc(text.slice(last));
    return out;
  },

  readableText(u) {
    if (u.paras) return u.paras.map(p => p.en).join(' ');
    return u.verses.map(v => v.en).join(' ');
  },

  onWordClick(e) {
    const span = e.target.closest('.w');
    if (!span) return;
    if (window.getSelection && String(window.getSelection()).trim().length > 1) return;
    const key = span.dataset.w;
    const ctx = this.contextFor(span, key);
    lookup.open(key, ctx, span.getBoundingClientRect());
  },

  contextFor(span, key) {
    const block = span.closest('.verse, .para, .vrow');
    let en = null;
    let zh = null;
    if (block && block.classList.contains('vrow')) {
      // 窄屏时一行里两种语言都在；宽屏时两种语言分居两页，要跨页取
      const ownEn = block.querySelector('.v-en');
      const ownZh = block.querySelector('.v-zh');
      if (ownEn && ownZh) {
        en = ownEn;
        zh = ownZh;
      } else {
        const page = block.closest('.page');
        const verso = page && page.classList.contains('page-verso');
        const other = document.querySelector('.spread .page-' + (verso ? 'recto' : 'verso'));
        const mine = page ? Array.prototype.indexOf.call(page.querySelectorAll('.vrow'), block) : -1;
        const otherRow = other && mine >= 0 ? other.querySelectorAll('.vrow')[mine] : null;
        en = verso ? ownEn : (otherRow ? otherRow.querySelector('.v-en') : null);
        zh = verso ? (otherRow ? otherRow.querySelector('.v-zh') : null) : ownZh;
      }
    } else {
      en = block ? block.querySelector('.verse-en, .para-en') : null;
      zh = block ? block.querySelector('.verse-zh, .para-zh') : null;
    }
    let source = this.currentLabel();
    if (block && block.dataset.verse) {
      const u = this.currentUnit();
      const book = DATA.books.find(b => b.id === u.book);
      source = (book ? book.zh : '') + ' ' + u.chapter + ':' + block.dataset.verse;
    }
    return {
      source: source,
      sentence: this.plainText(en),
      zh: this.plainText(zh)
    };
  },

  // 划词：选中文本后浮层查词
  onSelection() {
    const sel = window.getSelection();
    if (!sel || sel.isCollapsed) return;
    const text = sel.toString().trim();
    if (!text || text.length > 60) return;
    const range = sel.getRangeAt(0);
    const block = range.commonAncestorContainer.nodeType === 1
      ? range.commonAncestorContainer
      : range.commonAncestorContainer.parentElement;
    if (!block || !block.closest('.verse, .para, .vrow')) { lookup.close(); return; }
    const rect = range.getBoundingClientRect();
    const phrase = text.replace(/^[^\w]+|[^\w]+$/g, '');
    if (!phrase) return;
    const ctx = this.contextFor(block.querySelector('.w') || block, phrase);
    lookup.open(phrase, ctx, rect);
  },

  open(unit) {
    this.unit = unit;
    this.bookIndex = 0;
    const u = DATA.chapters[unit];
    if (u) this.openBook = u.book;
    this.render();
    this.refreshSide();
    bus.emit('reader:changed', unit);
  },

  // 侧栏展开 / 选中状态变了，单独刷一次侧栏（不动正文滚动位置）
  refreshSide() {
    const side = $('[data-side]');
    if (side) side.innerHTML = this.side();
  },

  saveUnit() {
    store.data.progress.unit = this.unit;
    store.save();
  },

  restoreScroll() {
    const p = store.data.progress;
    if (p.unit !== this.unit || !p.scroll) return;
    const main = $('[data-main]');
    if (main) main.scrollTop = p.scroll;
  },

  rememberScroll() {
    const main = $('[data-main]');
    if (main && this.unit === store.data.progress.unit) {
      store.data.progress.scroll = main.scrollTop;
      store.save();
    }
  },

  // ---- 右栏：本周词与快速跳转 ----
  detail() {
    const u = this.currentUnit();
    const parts = [];
    parts.push('<div class="detail-scroll">');
    parts.push('<section class="detail-block">');
    parts.push('<h2 class="detail-title">这一章</h2>');
    if (u.paras) {
      u.paras.forEach((p, i) => {
        parts.push('<button type="button" class="detail-line" data-goto-para="' + i + '">' +
          '<span class="detail-line-n">' + (i + 1) + '</span>' +
          '<span class="detail-line-t">' + esc(p.en.slice(0, 42)) + (p.en.length > 42 ? '…' : '') + '</span>' +
          '</button>');
      });
    } else {
      u.verses.forEach(v => {
        parts.push('<button type="button" class="detail-line" data-goto-verse="' + v.v + '">' +
          '<span class="detail-line-n">' + v.v + '</span>' +
          '<span class="detail-line-t">' + esc(v.en.slice(0, 42)) + (v.en.length > 42 ? '…' : '') + '</span>' +
          '</button>');
      });
    }
    parts.push('</section>');

    const savedHere = Object.keys(store.data.bank).filter(k => {
      const ctxs = store.data.bank[k].contexts || [];
      return ctxs.some(c => (c.source || '').indexOf(this.currentLabel()) === 0);
    });
    parts.push('<section class="detail-block">');
    parts.push('<h2 class="detail-title">这里收过的词</h2>');
    if (!savedHere.length) {
      parts.push('<p class="detail-empty">读到生词点一下就能收进词库，句子会一起留着。</p>');
    } else {
      parts.push('<ul class="mini-words">');
      savedHere.forEach(k => {
        const b = store.data.bank[k];
        parts.push('<li><button type="button" class="mini-word" data-open-word="' + esc(k) + '">' +
          '<span class="mini-word-w">' + esc(k) + '</span>' +
          '<span class="mini-word-zh">' + esc((b.senses && b.senses[0]) || '') + '</span>' +
          '</button></li>');
      });
      parts.push('</ul>');
    }
    parts.push('</section>');
    parts.push('</div>');
    return parts.join('');
  }
};

MODULES.read = reader;
