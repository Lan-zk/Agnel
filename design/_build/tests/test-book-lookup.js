(() => {
  const btns = [...document.querySelectorAll('.seg-btn')];
  btns.find(x => x.textContent.trim() === '书页对开').click();
  const words = [...document.querySelectorAll('.page .w')];
  const w = words.find(x => /created/i.test(x.textContent)) || words[0];
  if (w) w.click();
  const lk = document.querySelector('.lookup');
  const ctx = lk ? lk.querySelector('.lookup-context') : null;
  const src = lk ? lk.querySelector('.lookup-src') : null;
  const out = {
    clicked: w ? w.textContent : null,
    lookupOpen: !!lk,
    src: src ? src.textContent.trim() : null,
    ctx: ctx ? ctx.textContent.replace(/\s+/g, ' ').trim().slice(0, 200) : null,
    rows: lk ? lk.querySelectorAll('.lookup-row').length : 0
  };
  const before = document.querySelector('.book-nav-now').textContent;
  const next = [...document.querySelectorAll('.book-nav .pager-btn')].find(b => b.textContent.includes('后一页'));
  next.click();
  out.navBefore = before;
  out.navAfter = document.querySelector('.book-nav-now').textContent;
  out.pageNum = document.querySelector('.page-verso .page-num').textContent;
  out.firstRow = document.querySelector('.page-verso .vtext').textContent.slice(0, 30);
  // 键盘翻页
  document.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }));
  out.navAfterKey = document.querySelector('.book-nav-now').textContent;
  return out;
})()
