/* ============================================================
   词库模块：收藏词、词形/派生/词源、语境、练习记录
   ============================================================ */

const bank = {
  filter: 'all', // all | new | learning | known | wrong
  query: '',
  selected: null,

  words() {
    return Object.keys(store.data.bank).map(k => store.data.bank[k]);
  },

  filtered() {
    const q = this.query.trim().toLowerCase();
    return this.words().filter(w => {
      if (this.filter === 'wrong' && !(w.wrong > 0)) return false;
      if (this.filter === 'new' && w.stage !== 'new') return false;
      if (this.filter === 'learning' && w.stage !== 'learning') return false;
      if (this.filter === 'known' && w.stage !== 'known') return false;
      if (!q) return true;
      const hay = [w.word, w.ipa, w.pos].concat(w.senses || []).concat(w.contexts || []).map(c => c.en || '').join(' ');
      return hay.toLowerCase().indexOf(q) >= 0;
    }).sort((a, b) => (b.added || 0) - (a.added || 0));
  },

  counts() {
    const all = this.words();
    return {
      all: all.length,
      new: all.filter(w => w.stage === 'new').length,
      learning: all.filter(w => w.stage === 'learning').length,
      known: all.filter(w => w.stage === 'known').length,
      wrong: all.filter(w => (w.wrong || 0) > 0).length
    };
  },

  side() {
    const c = this.counts();
    const items = [
      { id: 'all', label: '全部', n: c.all },
      { id: 'new', label: '生词', n: c.new },
      { id: 'learning', label: '学习中', n: c.learning },
      { id: 'known', label: '已掌握', n: c.known },
      { id: 'wrong', label: '错词', n: c.wrong }
    ];
    const parts = ['<div class="side-scroll"><div class="side-group">'];
    parts.push('<h2 class="side-title">筛选</h2>');
    parts.push('<ul class="filter-list">');
    items.forEach(it => {
      parts.push('<li><button type="button" class="filter-item' + (this.filter === it.id ? ' is-on' : '') + '" data-filter="' + it.id + '">' +
        '<span>' + esc(it.label) + '</span><span class="filter-n">' + it.n + '</span></button></li>');
    });
    parts.push('</ul>');
    parts.push('</div>');

    parts.push('<div class="side-group">');
    parts.push('<h2 class="side-title">来源</h2>');
    const bySource = {};
    this.words().forEach(w => (w.contexts || []).forEach(c => {
      const src = c.source || '手动添加';
      bySource[src] = (bySource[src] || 0) + 1;
    }));
    const keys = Object.keys(bySource).sort((a, b) => bySource[b] - bySource[a]).slice(0, 8);
    if (!keys.length) {
      parts.push('<p class="detail-empty">词库还是空的。</p>');
    } else {
      parts.push('<ul class="src-list">');
      keys.forEach(k => {
        parts.push('<li><span class="src-name">' + esc(k) + '</span><span class="filter-n">' + bySource[k] + '</span></li>');
      });
      parts.push('</ul>');
    }
    parts.push('</div></div>');
    return parts.join('');
  },

  render() {
    const pane = $('[data-pane="bank"]');
    pane.innerHTML = '';

    const head = el('div', { class: 'pane-head' });
    const field = el('div', { class: 'search-field' });
    field.append(el('span', { class: 'search-icon', html: icon('search') }));
    const input = el('input', {
      type: 'search',
      class: 'search-input',
      placeholder: '搜单词、释义或句子',
      'aria-label': '搜索词库',
      value: this.query
    });
    input.addEventListener('input', e => {
      this.query = e.target.value;
      this.renderList();
    });
    field.append(input);
    head.append(field);
    if (store.data.demo) {
      head.append(el('span', { class: 'count-chip', text: '示例词条' }));
    }
    head.append(el('button', {
      type: 'button',
      class: 'btn btn-quiet',
      html: icon('plus') + '<span>手动添加</span>',
      onclick: () => this.openAdd()
    }));
    pane.append(head);

    const list = el('div', { class: 'word-list' });
    list.dataset.role = 'list';
    pane.append(list);
    this.renderList();
  },

  renderList() {
    const list = $('[data-pane="bank"] [data-role="list"]');
    if (!list) return;
    const rows = this.filtered();
    list.innerHTML = '';
    if (!rows.length) {
      list.append(el('p', { class: 'empty-note', text: this.words().length ? '这里没有符合条件的词。' : '词库还是空的，在阅读里点一个生词就能收进来。' }));
      return;
    }
    rows.forEach(w => {
      const row = el('button', {
        type: 'button',
        class: 'word-row' + (this.selected === w.word ? ' is-on' : ''),
        dataset: { word: w.word }
      });
      row.append(el('span', { class: 'word-main' }, [
        el('span', { class: 'word-text', text: w.word }),
        el('span', { class: 'word-ipa', text: w.ipa || '' })
      ]));
      row.append(el('span', { class: 'word-zh', text: (w.senses && w.senses[0]) || (w.pos || '') }));
      const tags = [];
      tags.push(el('span', { class: 'tag tag-' + w.stage, text: stageLabel(w.stage) }));
      if (w.wrong > 0) tags.push(el('span', { class: 'tag tag-wrong', text: '错 ' + w.wrong }));
      row.append(el('span', { class: 'word-tags' }, tags));
      list.append(row);
    });
  },

  stageLabel(s) {
    return (STAGES[s] && STAGES[s].label) || s;
  },

  select(word) {
    this.selected = word;
    this.renderList();
    const w = bankEntry(word);
    if (!w) return;
    paintDetail(this.wordDetail(w));
    openDetail(true);
  },

  wordDetail(w) {
    const parts = ['<div class="detail-scroll">'];
    parts.push('<section class="detail-block word-detail">');
    parts.push('<div class="wd-top">');
    parts.push('<h2 class="wd-word">' + esc(w.word) + '</h2>');
    if (w.ipa) parts.push('<span class="wd-ipa">' + esc(w.ipa) + '</span>');
    parts.push('</div>');
    if (w.pos) parts.push('<p class="wd-pos">' + esc(w.pos) + '</p>');

    parts.push('<div class="wd-stage">');
    ['new', 'learning', 'known'].forEach(s => {
      parts.push('<button type="button" class="seg-btn' + (w.stage === s ? ' is-on' : '') + '" data-stage-btn="' + s + '" data-word="' + esc(w.word) + '" aria-pressed="' + (w.stage === s) + '">' + stageLabel(s) + '</button>');
    });
    parts.push('</div>');

    if (w.senses && w.senses.length) {
      parts.push('<ul class="wd-senses">');
      w.senses.forEach(s => parts.push('<li>' + esc(s) + '</li>'));
      parts.push('</ul>');
    }

    const groups = [
      { label: '词形变化', items: w.forms || [], cls: 'forms' },
      { label: '派生关系', items: w.derivations || [], cls: 'deriv' },
      { label: '词源', items: w.etymology ? [w.etymology] : [], cls: 'ety' },
      { label: '常见搭配', items: w.collocations || [], cls: 'coll' }
    ].filter(g => g.items.length);

    if (groups.length) {
      parts.push('<dl class="wd-rows">');
      groups.forEach(g => {
        parts.push('<div class="wd-row wd-row-' + g.cls + '"><dt>' + g.label + '</dt><dd>' +
          g.items.map(i => '<span class="wd-chip">' + esc(i) + '</span>').join('') + '</dd></div>');
      });
      parts.push('</dl>');
    }

    if (w.contexts && w.contexts.length) {
      parts.push('<h3 class="wd-h3">原句</h3>');
      w.contexts.forEach(c => {
        parts.push('<blockquote class="wd-ctx">');
        if (c.source) parts.push('<cite class="wd-ctx-src">' + esc(c.source) + '</cite>');
        parts.push('<p class="wd-ctx-en">' + esc(c.en) + '</p>');
        if (c.zh) parts.push('<p class="wd-ctx-zh">' + esc(c.zh) + '</p>');
        parts.push('</blockquote>');
      });
    }

    const stats = [];
    if (w.right) stats.push('答对 ' + w.right);
    if (w.wrong) stats.push('答错 ' + w.wrong);
    if (stats.length) parts.push('<p class="wd-stats">' + stats.join(' · ') + '</p>');

    parts.push('<div class="wd-actions">');
    parts.push('<button type="button" class="btn" data-practice-word="' + esc(w.word) + '">用这个词练习</button>');
    parts.push('<button type="button" class="btn btn-quiet" data-speak-word="' + esc(w.word) + '">朗读</button>');
    parts.push('<button type="button" class="btn btn-quiet" data-remove-word="' + esc(w.word) + '">移出词库</button>');
    parts.push('</div>');
    parts.push('</section></div>');
    return parts.join('');
  },

  openAdd() {
    paintDetail(
      '<div class="detail-scroll"><section class="detail-block">' +
      '<h2 class="detail-title">手动添加</h2>' +
      '<form class="add-form" data-role="add">' +
      '<label class="field"><span class="field-label">单词</span><input name="word" required /></label>' +
      '<label class="field"><span class="field-label">音标</span><input name="ipa" placeholder="/ˈeksæmpl/" /></label>' +
      '<label class="field"><span class="field-label">释义</span><input name="gloss" required placeholder="你的理解" /></label>' +
      '<label class="field"><span class="field-label">出处（可选）</span><input name="source" placeholder="书名或文章" /></label>' +
      '<div class="add-actions"><button type="submit" class="btn">加入词库</button>' +
      '<button type="button" class="btn btn-quiet" data-cancel-add>取消</button></div>' +
      '</form></section></div>'
    );
    openDetail(true);
    const field = $('[data-detail] [name="word"]');
    if (field) field.focus();
  }
};

function stageLabel(s) {
  return (STAGES[s] && STAGES[s].label) || s;
}

MODULES.bank = bank;
