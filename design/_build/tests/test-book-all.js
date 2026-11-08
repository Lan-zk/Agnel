// 每一卷每一章都摊一次对开，看有没有报错、有没有被裁掉
(async () => {
  const errs = [];
  window.addEventListener('error', e => errs.push(String(e.message)));
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const out = [];
  reader.mode = '书页对开';
  for (const b of DATA.books) {
    for (const k of reader.chaptersOf(b.id)) {
      reader.open(k);
      await sleep(140);
      const pages = [...document.querySelectorAll('.page')];
      const clipped = pages.some(p => {
        const t = p.querySelector('.page-text');
        const last = t ? t.querySelector('.vrow:last-child') : null;
        if (!t || !last) return false;
        return last.getBoundingClientRect().bottom > t.getBoundingClientRect().bottom + 1;
      });
      out.push({
        k,
        pages: pages.length,
        rows: document.querySelectorAll('.page-verso .vrow').length,
        enRows: document.querySelectorAll('.page-recto .vrow').length,
        secs: document.querySelectorAll('.page-verso .vsec').length,
        notes: document.querySelectorAll('.fn').length,
        clipped,
        nav: (document.querySelector('.book-nav-now') || {}).textContent || ''
      });
    }
  }
  return { out, errs };
})()
