(() => {
  // 切到「书页对开」，再选诗篇，检查章数据与脚注是否接上
  const seg = [...document.querySelectorAll('.seg-btn')].find(x => x.textContent.trim() === '书页对开');
  if (!seg) return { error: 'no book button' };
  seg.click();
  const ps = [...document.querySelectorAll('.book-item')].find(x => /诗篇/.test(x.textContent));
  if (ps) ps.click();
  const chip = document.querySelector('.chapter-chip');
  if (chip) chip.click();
  const pages = [...document.querySelectorAll('.page')];
  const nav = document.querySelector('.book-nav-now');
  const rows = [...document.querySelectorAll('.page-verso .vrow')];
  const secCells = [...document.querySelectorAll('.page-verso .vsec')].map(n => n.textContent);
  const notes = [...document.querySelectorAll('.fn')].map(n => n.textContent);
  const last = rows[rows.length - 1];
  return {
    clicked: !!ps, chip: !!chip,
    pages: pages.length,
    rows: rows.length,
    heads: [...document.querySelectorAll('.page-run')].map(n => n.textContent),
    nums: [...document.querySelectorAll('.page-num')].map(n => n.textContent),
    badge: (document.querySelector('.page-badge') || {}).textContent || '',
    secs: secCells,
    notes: notes,
    lastVerse: last ? last.querySelector('.v-zh').textContent.slice(-12) : '',
    nav: nav ? nav.textContent : '',
    pageH: pages[0] ? Math.round(pages[0].getBoundingClientRect().height) : 0,
    colW: pages[0] ? Math.round(pages[0].querySelector('.page-text').getBoundingClientRect().width) : 0,
    clipped: (() => {
      if (!last) return false;
      const t = pages[0].querySelector('.page-text');
      const lb = last.getBoundingClientRect().bottom;
      const tb = t.getBoundingClientRect().bottom;
      return lb > tb + 1;
    })()
  };
})()
