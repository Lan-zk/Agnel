/* ============================================================
   练习模块：单词四式 + 句子/经文练习 + 即时反馈 + 错词收集
   ============================================================ */

const practice = {
  mode: 'word-copy',        // word-copy 跟打 | word-dict 默写 | word-listen 听写 | word-quiz 自测
  scope: 'bank',            // bank 收藏词 | verse 经句
  queue: [],
  slots: false,
  index: 0,
  typed: '',
  wrongWords: [],
  session: { right: 0, wrong: 0, start: 0 },
  showTranslation: true,
  peek: false,
  item: null,
  inputs: [],

  /* ---------- 题库构建 ---------- */
  buildQueue() {
    if (this.scope === 'verse') {
      const out = [];
      for (const key of Object.keys(DATA.chapters)) {
        const ch = DATA.chapters[key];
        const book = DATA.books.find(b => b.id === ch.book);
        ch.verses.forEach(v => {
          out.push({
            kind: 'sentence',
            en: v.en,
            zh: v.zh,
            source: book.zh + ' ' + ch.chapter + ':' + v.v,
            words: this.keyWords(v.en)
          });
        });
      }
      this.queue = out.slice(0, 40);
    } else {
      const words = Object.keys(store.data.bank).filter(k => {
        const w = store.data.bank[k];
        return w.stage !== 'known' || (w.lastReview && (Date.now() - w.lastReview) > 3 * 86400000);
      });
      const all = words.length ? words : Object.keys(store.data.bank);
      this.queue = all.map(k => {
        const w = store.data.bank[k];
        return {
          kind: 'word',
          word: k,
          zh: (w.senses && w.senses[0]) || '',
          ipa: w.ipa || '',
          source: (w.contexts && w.contexts[0] && w.contexts[0].source) || '',
          en: w.word
        };
      });
    }
    this.index = 0;
    this.session = { right: 0, wrong: 0, start: Date.now() };
    this.wrongWords = [];
  },

  keyWords(sentence) {
    return sentence.split(/\s+/).map(w => w.replace(/[^A-Za-z'’-]/g, '')).filter(Boolean);
  },

  /* ---------- 侧栏 ---------- */
  side() {
    const parts = ['<div class="side-scroll">'];
    parts.push('<div class="side-group"><h2 class="side-title">单词</h2><ul class="filter-list">');
    [['word-copy', '跟打', '照着词形敲一遍'], ['word-dict', '默写', '看着中文写英文'], ['word-listen', '听写', '只听声音写'], ['word-quiz', '自测', '四选一']]
      .forEach(m => {
        parts.push('<li><button type="button" class="filter-item' + (this.mode === m[0] ? ' is-on' : '') + '" data-pmode="' + m[0] + '">' +
          '<span class="fm-name">' + m[1] + '<span class="fm-hint">' + m[2] + '</span></span></button></li>');
      });
    parts.push('</ul></div>');

    parts.push('<div class="side-group"><h2 class="side-title">句子与经文</h2><ul class="filter-list">');
    [['sentence-copy', '跟打整句'], ['sentence-cloze', '填空'], ['verse-typed', '逐句默写']]
      .forEach(m => {
        parts.push('<li><button type="button" class="filter-item' + (this.mode === m[0] ? ' is-on' : '') + '" data-pmode="' + m[0] + '">' +
          '<span class="fm-name">' + m[1] + '</span></button></li>');
      });
    parts.push('</ul></div>');

    parts.push('<div class="side-group"><h2 class="side-title">材料</h2><ul class="filter-list">');
    parts.push('<li><button type="button" class="filter-item' + (this.scope === 'bank' ? ' is-on' : '') + '" data-pscope="bank"><span class="fm-name">用收藏词练习<span class="fm-hint">' + Object.keys(store.data.bank).length + ' 个</span></span></button></li>');
    parts.push('<li><button type="button" class="filter-item' + (this.scope === 'verse' ? ' is-on' : '') + '" data-pscope="verse"><span class="fm-name">用经句练习<span class="fm-hint">' + '40 句</span></span></button></li>');
    parts.push('</ul></div>');
    parts.push('</div>');
    return parts.join('');
  },

  /* ---------- 主区 ---------- */
  render() {
    const pane = $('[data-pane="practice"]');
    pane.innerHTML = '';

    if (!this.queue.length) this.buildQueue();
    if (!this.queue.length) {
      pane.append(el('p', { class: 'empty-note', text: '还没有可练的内容，先在阅读里收几个生词。' }));
      return;
    }
    this.item = this.queue[this.index];

    const head = el('div', { class: 'pane-head practice-head' });
    head.append(el('div', { class: 'practice-progress' }, [
      el('span', { class: 'pp-count', text: (this.index + 1) + ' / ' + this.queue.length }),
      el('span', { class: 'pp-bar' }, el('span', { class: 'pp-fill', style: 'width:' + Math.round(((this.index) / this.queue.length) * 100) + '%' }))
    ]));
    head.append(el('div', { class: 'practice-head-right' }, [
      el('span', { class: 'pp-stat', text: '对 ' + this.session.right }),
      el('span', { class: 'pp-stat is-wrong', text: '错 ' + this.session.wrong })
    ]));
    pane.append(head);

    const card = el('div', { class: 'practice-card' });
    switch (this.mode) {
      case 'word-quiz': this.renderQuiz(card); break;
      case 'sentence-cloze': this.renderCloze(card); break;
      case 'sentence-copy':
      case 'verse-typed': this.renderSentence(card); break;
      default: this.renderWord(card);
    }
    pane.append(card);
    pane.append(this.renderFoot());
  },

  renderFoot() {
    const foot = el('div', { class: 'practice-foot' });
    foot.append(el('button', {
      type: 'button', class: 'btn btn-quiet', text: '重来',
      onclick: () => { this.typed = ''; this.render(); }
    }));
    foot.append(el('button', {
      type: 'button', class: 'btn btn-quiet', text: '跳过',
      onclick: () => this.next(true)
    }));
    foot.append(el('button', {
      type: 'button', class: 'btn', text: this.index === this.queue.length - 1 ? '完成' : '下一个',
      onclick: () => this.next(false)
    }));
    return foot;
  },

  renderWord(card) {
    const it = this.item;
    const mode = this.mode;
    const masked = mode === 'word-dict' || mode === 'word-listen';
    const reveal = masked && this.peek;

    const prompt = el('div', { class: 'pc-prompt' });
    if (mode === 'word-listen') {
      prompt.append(el('button', {
        type: 'button', class: 'listen-btn', html: icon('speaker'),
        'aria-label': '播放', onclick: () => speak(it.word)
      }));
      if (reveal) prompt.append(el('h2', { class: 'pc-word', text: it.word }));
    } else {
      prompt.append(el('h2', { class: 'pc-word' + (masked && !reveal ? ' is-hidden' : '') },
        masked && !reveal ? it.word.replace(/[A-Za-z'’-]/g, '·') : it.word));
      if (it.ipa && !masked) prompt.append(el('p', { class: 'pc-ipa', text: it.ipa }));
    }
    if (mode === 'word-dict' || mode === 'word-copy') {
      if (it.zh) prompt.append(el('p', { class: 'pc-zh', text: it.zh }));
    }
    if (it.source) prompt.append(el('p', { class: 'pc-src', text: it.source }));
    card.append(prompt);

    const zone = el('div', { class: 'typing-zone' });
    this.typed = '';
    zone.append(this.renderChars(this.targetText(), '', null, true));
    card.append(zone);

    card.append(this.hiddenInput());
    card.append(el('p', { class: 'pc-tip', text: this.hintText(mode) }));
  },

  hintText(mode) {
    if (mode === 'word-dict') return '看中文把英文写出来，按 Tab 看一眼拼写';
    if (mode === 'word-listen') return '听发音拼出这个词，按 Tab 看一眼';
    return '照着上面的词敲一遍';
  },

  /* 默写 / 听写时输入区不暴露答案，按 Tab 才亮出 */
  displayTarget() {
    const truth = this.targetText();
    const masked = this.mode === 'word-dict' || this.mode === 'word-listen';
    if (!masked || this.peek) return truth;
    return truth.replace(/[A-Za-z'’-]/g, '_');
  },

  renderSentence(card) {
    const it = this.item;
    const showZh = this.showTranslation && this.mode !== 'verse-typed';
    const prompt = el('div', { class: 'pc-prompt' });
    prompt.append(el('p', { class: 'pc-sentence-en', text: it.en }));
    if (showZh) prompt.append(el('p', { class: 'pc-sentence-zh', text: it.zh }));
    if (it.source) prompt.append(el('p', { class: 'pc-src', text: it.source }));
    card.append(prompt);

    const zone = el('div', { class: 'typing-zone' });
    this.typed = '';
    zone.append(this.renderChars(it.en, ''));
    card.append(zone);

    card.append(el('div', { class: 'pc-tools' }, [
      el('button', {
        type: 'button', class: 'icon-btn', html: icon('speaker'), 'aria-label': '朗读',
        onclick: () => speak(it.en)
      }),
      el('button', {
        type: 'button', class: 'toggle-btn' + (showZh ? ' is-on' : ''), text: '中文',
        onclick: () => { this.showTranslation = !this.showTranslation; this.render(); }
      }),
      el('button', {
        type: 'button', class: 'toggle-btn', text: '看答案',
        onclick: () => { this.peek = !this.peek; this.render(); }
      })
    ]));
    card.append(this.hiddenInput());
    card.append(el('p', { class: 'pc-tip', text: this.mode === 'verse-typed' ? '逐句默写：把整句写下来' : '照着上面的句子敲一遍' }));
  },

  renderCloze(card) {
    const it = this.item;
    const words = it.en.split(/\s+/);
    const blanks = [];
    const picked = new Set();
    const n = Math.max(2, Math.min(4, Math.floor(words.length / 5)));
    let seed = words.length;
    while (picked.size < n && seed > 0) {
      seed = (seed * 7 + 13) % Math.max(1, words.length);
      picked.add(seed % words.length);
    }
    const sorted = Array.from(picked).sort((a, b) => a - b);
    sorted.forEach(i => blanks.push(i));

    const prompt = el('div', { class: 'pc-prompt' });
    prompt.append(el('p', { class: 'pc-sentence-en', text: it.en }));
    if (this.showTranslation) prompt.append(el('p', { class: 'pc-sentence-zh', text: it.zh }));
    if (it.source) prompt.append(el('p', { class: 'pc-src', text: it.source }));
    card.append(prompt);

    const zone = el('div', { class: 'cloze-zone' });
    const parts = [];
    words.forEach((w, i) => {
      if (blanks.indexOf(i) >= 0) {
        const slot = el('input', {
          type: 'text', class: 'cloze-slot', 'aria-label': '第 ' + (blanks.indexOf(i) + 1) + ' 个空',
          dataset: { answer: w.replace(/[^A-Za-z'’-]/g, ''), index: String(i) }
        });
        slot.addEventListener('input', () => slot.classList.toggle('is-filled', !!slot.value));
        this.inputs.push(slot);
        parts.push(slot);
      } else {
        parts.push(el('span', { class: 'cloze-word', text: w }));
      }
    });
    zone.append(parts);
    card.append(zone);

    card.append(el('div', { class: 'pc-tools' }, [
      el('button', {
        type: 'button', class: 'icon-btn', html: icon('speaker'), 'aria-label': '朗读',
        onclick: () => speak(it.en)
      }),
      el('button', {
        type: 'button', class: 'toggle-btn' + (this.showTranslation ? ' is-on' : ''), text: '中文',
        onclick: () => { this.showTranslation = !this.showTranslation; this.render(); }
      }),
      el('button', {
        type: 'button', class: 'btn', text: '检查',
        onclick: () => this.checkCloze()
      })
    ]));
    card.append(el('p', { class: 'pc-tip', text: '把空补上，检查一次' }));
  },

  checkCloze() {
    let right = 0;
    let wrong = 0;
    this.inputs.forEach(slot => {
      const answer = slot.dataset.answer;
      const given = slot.value.trim().replace(/[^A-Za-z'’-]/g, '');
      if (!given) { slot.classList.remove('is-right', 'is-wrong'); return; }
      if (given.toLowerCase() === answer.toLowerCase()) {
        slot.classList.add('is-right');
        slot.classList.remove('is-wrong');
        right++;
      } else {
        slot.classList.add('is-wrong');
        slot.classList.remove('is-right');
        wrong++;
      }
    });
    this.session.right += right;
    this.session.wrong += wrong;
    const it = this.item;
    this.inputs.forEach(slot => {
      const given = slot.value.trim().replace(/[^A-Za-z'’-]/g, '').toLowerCase();
      const answer = slot.dataset.answer.toLowerCase();
      if (given === answer && answer) this.markWord(answer, true);
      else if (given) this.markWord(answer, false);
    });
    toast(right && !wrong ? '都对上了' : (wrong ? '有 ' + wrong + ' 个没对' : '空着的没算分'), wrong ? 'warning' : 'success');
    this.renderPracticeDetail();
  },

  renderQuiz(card) {
    const it = this.item;
    const correct = (it.zh || '').trim() ||
      (DATA.dict[it.word] && DATA.dict[it.word].senses[0]) || it.word;
    const pool = Object.keys(store.data.bank).filter(k => k !== it.word);
    const all = pool.length > 3 ? pool : Object.keys(DATA.dict).filter(k => k !== it.word);
    const distractors = [];
    let i = 0;
    while (distractors.length < 3 && all.length && i < 60) {
      const cand = all[(it.word.length + i * 3) % all.length];
      const meaning = (bankEntry(cand) && bankEntry(cand).senses && bankEntry(cand).senses[0]) ||
        (DATA.dict[cand] && DATA.dict[cand].senses[0]);
      if (cand !== it.word && meaning && meaning !== correct && distractors.indexOf(meaning) < 0) {
        distractors.push(meaning);
      }
      i++;
    }
    const spare = ['成为', '保持', '留下'];
    while (distractors.length < 3) distractors.push(spare[distractors.length]);
    const choices = distractors.concat([correct]);
    const off = it.word.length % 3;
    const shuffled = choices.slice(off).concat(choices.slice(0, off));

    const prompt = el('div', { class: 'pc-prompt' });
    prompt.append(el('h2', { class: 'pc-word', text: it.word }));
    prompt.append(el('p', { class: 'pc-zh', text: '这个词是什么意思？' }));
    card.append(prompt);

    const opts = el('div', { class: 'quiz-opts' });
    shuffled.forEach(text => {
      opts.append(el('button', {
        type: 'button', class: 'quiz-opt', text: text,
        onclick: e => this.answerQuiz(e.currentTarget, text, correct, it.word)
      }));
    });
    card.append(opts);
    card.append(el('p', { class: 'pc-tip', text: '选一个' }));
  },

  answerQuiz(btn, text, correct, word) {
    const opts = btn.parentElement;
    const right = text === correct;
    opts.querySelectorAll('.quiz-opt').forEach(o => {
      o.disabled = true;
      if (o.textContent === correct) o.classList.add('is-right');
    });
    if (!right) btn.classList.add('is-wrong');
    this.session[right ? 'right' : 'wrong']++;
    this.markWord(word, right);
    toast(right ? '对了' : '不对，正确答案已标出', right ? 'success' : 'error');
    setTimeout(() => this.next(!right), 850);
  },

  /* slots = true 时不显示参考文字，只留一格一格的空位（单词练习用） */
  renderChars(target, typed, display, slots) {
    const zone = el('div', { class: 'chars' + (slots ? ' is-slots' : '') });
    const shown = display || target;
    const givenLen = typed.length;
    for (let i = 0; i < target.length; i++) {
      const ch = target[i];
      const given = typed[i];
      let cls = 'ch';
      if (ch === ' ') cls += ' ch-space';
      if (given !== undefined) {
        cls += given === ch ? ' is-right' : ' is-wrong';
        zone.append(el('span', { class: cls, text: given }));
      } else if (slots) {
        cls += ' is-slot';
        zone.append(el('span', { class: cls, text: '\u00a0' }));
      } else {
        zone.append(el('span', { class: cls, text: shown[i] === ' ' ? ' ' : shown[i] }));
      }
      if (i + 1 === givenLen) zone.append(el('span', { class: 'caret' }));
    }
    if (!givenLen && !slots) zone.append(el('span', { class: 'caret' }));
    return zone;
  },

  hiddenInput() {
    const input = el('input', {
      type: 'text',
      class: 'hidden-input',
      'aria-label': '在这里输入',
      autocomplete: 'off',
      autocapitalize: 'off',
      autocorrect: 'off',
      spellcheck: 'false'
    });
    input.addEventListener('input', () => {
      const v = input.value;
      input.value = '';
      if (v) this.handleInput(v[v.length - 1]);
    });
    input.addEventListener('keydown', e => {
      if (e.key === 'Tab') {
        e.preventDefault();
        this.peek = !this.peek;
        this.render();
        const again = $('[data-pane="practice"] .hidden-input');
        if (again) again.focus();
      } else if (e.key === 'Backspace') {
        e.preventDefault();
        this.typed = this.typed.slice(0, -1);
        this.paint();
      } else if (e.key === 'Enter') {
        e.preventDefault();
        this.next(false);
      }
    });
    setTimeout(() => input.focus(), 0);
    return input;
  },

  targetText() {
    const it = this.item;
    if (!it) return '';
    if (it.kind === 'word') return it.word;
    return it.en;
  },

  handleInput(ch) {
    const it = this.item;
    const target = this.targetText();
    if (this.typed.length >= target.length) return;
    const expect = target[this.typed.length];
    this.typed += ch;
    if (ch !== expect) {
      this.session.wrong++;
      this.shake();
      if (it && it.kind === 'word') this.markWord(it.word, false);
    } else if (this.typed.length === target.length) {
      this.session.right++;
      if (it && it.kind === 'word') this.markWord(it.word, true);
      if (it) this.collectWrong(it);
    }
    this.paint();
    if (this.typed.length >= target.length) {
      setTimeout(() => this.next(false), 620);
    }
  },

  paint() {
    const zone = $('[data-pane="practice"] .typing-zone .chars');
    if (!zone) return;
    const target = this.targetText();
    this.slots = this.mode === 'word-copy' || this.mode === 'word-dict' || this.mode === 'word-listen';
    const fresh = this.renderChars(target, this.typed, null, this.slots);
    zone.replaceWith(fresh);
    const head = $('[data-pane="practice"] .practice-head');
    if (head) {
      head.querySelectorAll('.pp-stat')[0].textContent = '对 ' + this.session.right;
      head.querySelectorAll('.pp-stat')[1].textContent = '错 ' + this.session.wrong;
    }
  },

  shake() {
    const zone = $('[data-pane="practice"] .typing-zone');
    if (!zone) return;
    zone.classList.remove('is-shake');
    void zone.offsetWidth;
    zone.classList.add('is-shake');
  },

  collectWrong(item) {
    if (item.kind !== 'word') return;
    const w = bankEntry(item.word);
    if (w && w.wrong > 0 && this.wrongWords.indexOf(item.word) < 0) this.wrongWords.push(item.word);
    this.renderPracticeDetail();
  },

  markWord(word, right) {
    const w = bankEntry(word);
    if (!w) return;
    if (right) { w.right = (w.right || 0) + 1; if (w.stage === 'new') w.stage = 'learning'; }
    else { w.wrong = (w.wrong || 0) + 1; }
    w.lastReview = Date.now();
    store.save();
    bus.emit('bank:changed', word);
  },

  next(skipped) {
    if (skipped && this.item && this.item.kind === 'word') this.markWord(this.item.word, false);
    this.index++;
    this.typed = '';
    this.peek = false;
    this.inputs = [];
    if (this.index >= this.queue.length) {
      this.index = this.queue.length - 1;
      this.renderSummary();
      return;
    }
    this.render();
  },

  renderSummary() {
    const pane = $('[data-pane="practice"]');
    pane.innerHTML = '';
    const total = this.session.right + this.session.wrong;
    const rate = total ? Math.round((this.session.right / total) * 100) : 0;
    const mins = Math.max(1, Math.round((Date.now() - this.session.start) / 60000));
    const card = el('div', { class: 'practice-card summary-card' });
    card.append(el('h2', { class: 'summary-title', text: '这一轮结束了' }));
    const stats = el('dl', { class: 'summary-list' });
    [['答对', this.session.right], ['答错', this.session.wrong],
     ['正确率', rate + '%'], ['用时', mins + ' 分钟']].forEach(s => {
      stats.append(el('dt', { text: s[0] }));
      stats.append(el('dd', { text: String(s[1]) }));
    });
    card.append(stats);
    if (this.wrongWords.length) {
      card.append(el('h3', { class: 'wd-h3', text: '这一轮错的词' }));
      const list = el('ul', { class: 'wrong-chips' });
      this.wrongWords.forEach(w => list.append(el('li', { class: 'wrong-chip', text: w })));
      card.append(list);
    }
    card.append(el('button', {
      type: 'button', class: 'btn', text: '再来一轮',
      onclick: () => { this.buildQueue(); this.render(); }
    }));
    pane.append(card);
  },

  /* ---------- 右栏：记录 ---------- */
  detail() {
    const total = this.session.right + this.session.wrong;
    const rate = total ? Math.round((this.session.right / total) * 100) : 0;
    const parts = ['<div class="detail-scroll">'];
    parts.push('<section class="detail-block">');
    parts.push('<h2 class="detail-title">本轮</h2>');
    parts.push('<p class="detail-meta">正确率 ' + rate + '%</p>');
    parts.push('<p class="detail-meta">对 ' + this.session.right + ' · 错 ' + this.session.wrong + ' · 第 ' + (this.index + 1) + ' / ' + this.queue.length + ' 题</p>');
    parts.push('</section>');

    const wrongList = Object.keys(store.data.bank).filter(k => (store.data.bank[k].wrong || 0) > 0);
    parts.push('<section class="detail-block">');
    parts.push('<h2 class="detail-title">错词本</h2>');
    if (!wrongList.length) {
      parts.push('<p class="detail-empty">还没有错词。写错的词会自动收进这里。</p>');
    } else {
      parts.push('<ul class="mini-words">');
      wrongList.forEach(k => {
        const w = store.data.bank[k];
        parts.push('<li><button type="button" class="mini-word" data-open-word="' + esc(k) + '">' +
          '<span class="mini-word-w">' + esc(k) + '</span>' +
          '<span class="mini-word-zh">错 ' + w.wrong + ' 次</span>' +
          '</button></li>');
      });
      parts.push('</ul>');
    }
    parts.push('</section></div>');
    return parts.join('');
  },

  renderPracticeDetail() {
    paintDetail(this.detail());
  }
};

MODULES.practice = practice;
