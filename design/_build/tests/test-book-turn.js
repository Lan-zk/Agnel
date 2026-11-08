// 把每一章逐个对开翻到底：核对行数 = 节数、脚注跟着它引的那一节走、没有报错
(async () => {
  const errs = [];
  window.addEventListener('error', e => errs.push(String(e.message)));
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  reader.mode = '书页对开';
  const out = [];
  for (const b of DATA.books) {
    for (const k of reader.chaptersOf(b.id)) {
      reader.open(k);
      await sleep(150);
      const ch = DATA.chapters[k];
      let rows = 0, notes = 0, spreads = 0, titles = 0, clipped = false;
      for (let guard = 0; guard < 40; guard++) {
        const pages = [...document.querySelectorAll('.page')];
        rows += document.querySelectorAll('.page-verso .vrow').length;
        notes += document.querySelectorAll('.fn').length;
        titles += document.querySelectorAll('.vtitle').length;
        spreads++;
        for (const p of pages) {
          const t = p.querySelector('.page-text');
          const last = t ? t.querySelector('.vrow:last-child') : null;
          if (t && last && last.getBoundingClientRect().bottom > t.getBoundingClientRect().bottom + 1) clipped = true;
        }
        const before = reader.bookIndex;
        reader.bookGo(1);
        await sleep(90);
        if (reader.bookIndex === before) break;
      }
      out.push({
        k,
        spreads,
        rows,
        verses: ch.verses.length,
        rowsOk: rows === ch.verses.length,
        notes,
        notesExpected: (ch.notes ? ch.notes.en.length + ch.notes.zh.length : 0),
        titles,
        sections: ch.sections.length,
        clipped,
      });
    }
  }
  return { out, errs };
})()
