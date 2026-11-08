/* ============================================================
   英语基础：音标 / 语法句法 / 词根与词族关系 / 基础词汇短语
   ============================================================ */

const basics = {
  tab: 'phonics',

  tabs: [
    { id: 'phonics', label: '音标' },
    { id: 'grammar', label: '语法句法' },
    { id: 'roots', label: '词根与词族' },
    { id: 'vocab', label: '基础词汇' }
  ],

  side() {
    const parts = ['<div class="side-scroll"><div class="side-group"><ul class="filter-list">'];
    this.tabs.forEach(t => {
      parts.push('<li><button type="button" class="filter-item' + (this.tab === t.id ? ' is-on' : '') + '" data-btab="' + t.id + '">' +
        '<span class="fm-name">' + t.label + '</span></button></li>');
    });
    parts.push('</ul></div></div>');
    return parts.join('');
  },

  render() {
    const pane = $('[data-pane="basics"]');
    pane.innerHTML = '';
    const head = el('div', { class: 'pane-head basics-head' });
    head.append(el('h1', { class: 'basics-title', text: this.tabs.find(t => t.id === this.tab).label }));
    pane.append(head);

    if (this.tab === 'phonics') this.renderPhonics(pane);
    else if (this.tab === 'grammar') this.renderGrammar(pane);
    else if (this.tab === 'roots') this.renderRoots(pane);
    else this.renderVocab(pane);
  },

  renderPhonics(pane) {
    const group = (title, list, note) => {
      const sec = el('section', { class: 'phonic-group' });
      sec.append(el('h2', { class: 'group-title', text: title }));
      if (note) sec.append(el('p', { class: 'group-note', text: note }));
      const grid = el('div', { class: 'phonic-grid' });
      list.forEach(item => {
        const cell = el('div', { class: 'phonic-cell' });
        cell.append(el('button', {
          type: 'button', class: 'phonic-ipa', text: item.ipa,
          'aria-label': '朗读 ' + item.ipa, onclick: () => speak(item.sample)
        }));
        cell.append(el('span', { class: 'phonic-sample', text: item.sample }));
        cell.append(el('span', { class: 'phonic-zh', text: item.zh }));
        grid.append(cell);
      });
      sec.append(grid);
      pane.append(sec);
    };
    group('元音', DATA.phonics.vowels, '点一下音标会读出例词。');
    group('辅音', DATA.phonics.consonants, '同一个位置的两组要分清：清音不振动声带，浊音振动。');
  },

  renderGrammar(pane) {
    const list = el('div', { class: 'grammar-list' });
    DATA.grammar.forEach(g => {
      const item = el('article', { class: 'grammar-item' });
      item.append(el('h3', { class: 'grammar-point', text: g.point }));
      item.append(el('p', { class: 'grammar-note', text: g.note }));
      const ex = el('div', { class: 'grammar-ex' });
      ex.append(el('p', { class: 'grammar-en', text: g.en }));
      ex.append(el('p', { class: 'grammar-zh', text: g.zh }));
      ex.append(el('p', { class: 'grammar-from', text: g.from }));
      item.append(ex);
      list.append(item);
    });
    pane.append(list);
  },

  renderRoots(pane) {
    pane.append(el('p', { class: 'group-note group-note-top', text: '实线是同一个词的词形变化，虚线是由它派生出的新词；词源另标，不和派生混在一起。' }));

    const legend = el('div', { class: 'root-legend' });
    legend.append(el('span', { class: 'legend-item' }, [
      el('span', { class: 'legend-line is-form' }),
      el('span', { text: '词形变化' })
    ]));
    legend.append(el('span', { class: 'legend-item' }, [
      el('span', { class: 'legend-line is-deriv' }),
      el('span', { text: '派生关系' })
    ]));
    legend.append(el('span', { class: 'legend-item' }, [
      el('span', { class: 'legend-node is-origin' }),
      el('span', { text: '词源' })
    ]));
    pane.append(legend);

    DATA.roots.forEach(root => {
      pane.append(this.rootGraph(root));
    });
  },

  rootGraph(root) {
    const wrap = el('section', { class: 'root-graph' });
    wrap.append(el('div', { class: 'root-head' }, [
      el('h3', { class: 'root-name', text: root.root + '-' }),
      el('span', { class: 'root-gloss', text: root.gloss })
    ]));

    const forms = root.members.forms;
    const derivs = root.members.derivations;
    const rowH = 34;
    const padY = 20;
    const colX = [12, 156, 312];
    const width = colX[2] + 168;
    const F = forms.length;
    const D = derivs.length;
    const height = padY * 2 + (F + D) * rowH;

    const ns = 'http://www.w3.org/2000/svg';
    const svg = document.createElementNS(ns, 'svg');
    svg.setAttribute('class', 'root-svg');
    svg.setAttribute('viewBox', '0 0 ' + width + ' ' + height);
    svg.setAttribute('width', width);
    svg.setAttribute('height', height);
    svg.setAttribute('role', 'img');
    svg.setAttribute('aria-label', root.root + ' 词族关系');

    const formY = i => padY + i * rowH + rowH / 2;
    const derivY = i => padY + (F + i) * rowH + rowH / 2;
    const yRoot = padY + F * rowH / 2;

    const mk = (tag, attrs, text) => {
      const n = document.createElementNS(ns, tag);
      Object.keys(attrs).forEach(k => n.setAttribute(k, attrs[k]));
      if (text !== undefined) n.textContent = text;
      return n;
    };

    // 词源
    const originY = 4;
    svg.append(mk('rect', { class: 'rg-origin', x: colX[0], y: originY, width: 128, height: 22, rx: 4 }));
    svg.append(mk('text', { class: 'rg-origin-text', x: colX[0] + 64, y: originY + 15, 'text-anchor': 'middle' }, root.origin));
    svg.append(mk('path', {
      class: 'rg-origin-link',
      d: 'M' + (colX[0] + 64) + ' ' + (originY + 22) + ' V' + (yRoot - 15)
    }));

    // 根
    svg.append(mk('rect', { class: 'rg-root', x: colX[0], y: yRoot - 15, width: 104, height: 30, rx: 5 }));
    svg.append(mk('text', { class: 'rg-root-text', x: colX[0] + 52, y: yRoot + 5, 'text-anchor': 'middle' }, root.root));

    // 列标题
    svg.append(mk('text', { class: 'rg-col-label', x: colX[1], y: padY - 6 }, '词形变化'));
    svg.append(mk('text', { class: 'rg-col-label', x: colX[2], y: padY + F * rowH - 6 }, '派生关系'));

    // 词形变化：实线，从根节点右侧出发
    forms.forEach((w, i) => {
      const y = formY(i);
      svg.append(mk('path', {
        class: 'rg-link is-form',
        d: 'M' + (colX[0] + 104) + ' ' + yRoot +
           ' C' + (colX[0] + 130) + ' ' + yRoot +
           ' ' + (colX[1] - 28) + ' ' + y + ' ' + colX[1] + ' ' + y
      }));
      svg.append(mk('rect', { class: 'rg-form', x: colX[1], y: y - 12, width: 108, height: 24, rx: 4 }));
      svg.append(mk('text', { class: 'rg-form-text', x: colX[1] + 9, y: y + 5 }, w));
    });

    // 派生关系：虚线，从根节点下方绕开词形一列
    derivs.forEach((d, i) => {
      const y = derivY(i);
      svg.append(mk('path', {
        class: 'rg-link is-deriv',
        d: 'M' + (colX[0] + 52) + ' ' + (yRoot + 15) +
           ' C' + (colX[0] + 52) + ' ' + (yRoot + 85) +
           ' ' + (colX[2] - 70) + ' ' + (y - 10) +
           ' ' + colX[2] + ' ' + y
      }));
      svg.append(mk('rect', { class: 'rg-deriv', x: colX[2], y: y - 12, width: 160, height: 24, rx: 4 }));
      svg.append(mk('text', { class: 'rg-deriv-text', x: colX[2] + 9, y: y + 5 }, d.word + ' · ' + d.gloss));
    });

    const scroller = el('div', { class: 'root-svg-wrap' });
    scroller.append(svg);
    wrap.append(scroller);
    return wrap;
  },

  detail() {
    const parts = ['<div class="detail-scroll">'];
    if (this.tab === 'phonics') {
      parts.push('<section class="detail-block"><h2 class="detail-title">怎么看音标</h2>');
      parts.push('<p class="detail-meta">点任意一个音标，会读出它的例词。</p>');
      parts.push('<p class="detail-meta">同一组里的两个符号是同一个发音位置。</p>');
      parts.push('<p class="detail-meta">清音不振动声带，浊音振动，别把两个记混。</p>');
      parts.push('</section>');
      parts.push('<section class="detail-block"><h2 class="detail-title">分成两部分</h2>');
      parts.push('<p class="detail-meta">元音 ' + DATA.phonics.vowels.length + ' 个：嘴形自然，舌位靠前。</p>');
      parts.push('<p class="detail-meta">辅音 ' + DATA.phonics.consonants.length + ' 个：按位置成对出现。</p>');
      parts.push('</section>');
    } else if (this.tab === 'grammar') {
      parts.push('<section class="detail-block"><h2 class="detail-title">怎么用这一栏</h2>');
      parts.push('<p class="detail-meta">每一条先看例句，再看它出现在句子的哪个位置。</p>');
      parts.push('<p class="detail-meta">英文的语序比中文松，修饰成分常常前置。</p>');
      parts.push('<p class="detail-meta">一条只讲一个意思，不要把整段当成一个结构。</p>');
      parts.push('</section>');
      parts.push('<section class="detail-block"><h2 class="detail-title">共 ' + DATA.grammar.length + ' 条</h2>');
      parts.push('<p class="detail-meta">例句取自本产品的正文，可以回到阅读里找同一句。</p>');
      parts.push('</section>');
    } else if (this.tab === 'roots') {
      parts.push('<section class="detail-block"><h2 class="detail-title">三种关系</h2>');
      parts.push('<p class="detail-meta">词形变化：同一个词的不同形式，例如 create / created。</p>');
      parts.push('<p class="detail-meta">派生关系：由它造出的新词，例如 create → creation。</p>');
      parts.push('<p class="detail-meta">词源：这个词来自哪里，和上面两种不是一回事。</p>');
      parts.push('</section><section class="detail-block"><h2 class="detail-title">词根</h2><ul class="mini-words">');
      DATA.roots.forEach(r => parts.push('<li class="mini-word"><span class="mini-word-w">' + r.root + '-</span><span class="mini-word-zh">' + esc(r.gloss) + '</span></li>'));
      parts.push('</ul></section>');
    } else {
      parts.push('<section class="detail-block"><h2 class="detail-title">怎么用这一栏</h2>');
      parts.push('<p class="detail-meta">这些是最常用的动词、名词和短语，可以脱离阅读单独看。</p>');
      parts.push('<p class="detail-meta">点一个词：词库里有的直接打开，没有的会弹出释义，可以顺手收进来。</p>');
      parts.push('</section>');
    }
    parts.push('</div>');
    return parts.join('');
  },

  renderVocab(pane) {
    const groups = [
      { key: 'verbs', title: '常用动词' },
      { key: 'nouns', title: '常用名词' },
      { key: 'phrases', title: '常用短语' }
    ];
    groups.forEach(g => {
      const sec = el('section', { class: 'vocab-group' });
      sec.append(el('h2', { class: 'group-title', text: g.title }));
      const list = el('ul', { class: 'vocab-list' });
      DATA.basics[g.key].forEach(item => {
        list.append(el('li', { class: 'vocab-item' }, [
          el('button', {
            type: 'button', class: 'vocab-w', text: item.w,
            onclick: () => {
              const b = bankEntry(item.w);
              if (b) { goModule('bank'); bank.select(item.w); }
              else {
                const r = resolveWord(item.w);
                const ctx = { source: '基础词汇', gloss: item.zh };
                const rect = (event && event.currentTarget ? event.currentTarget : document.querySelector('.basics-title')).getBoundingClientRect();
                lookup.open(item.w, ctx, rect);
              }
            }
          }),
          el('span', { class: 'vocab-zh', text: item.zh })
        ]));
      });
      sec.append(list);
      pane.append(sec);
    });
  }
};

MODULES.basics = basics;
