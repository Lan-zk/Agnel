(() => {
  const btns = [...document.querySelectorAll('.seg-btn')];
  const b = btns.find(x => x.textContent.trim() === '书页对开');
  if (!b) return { error: 'no book button', modes: btns.map(x => x.textContent.trim()) };
  b.click();
  const pages = [...document.querySelectorAll('.page')];
  const nav = document.querySelector('.book-nav-now');
  const r1 = document.querySelector('.page-verso .vrow');
  const r2 = document.querySelector('.page-recto .vrow');
  const badge = document.querySelector('.page-badge');
  const heads = [...document.querySelectorAll('.page-run')].map(n => n.textContent);
  const nums = [...document.querySelectorAll('.page-num')].map(n => n.textContent);
  return {
    pages: pages.length,
    rows: document.querySelectorAll('.page-verso .vrow').length,
    enRows: document.querySelectorAll('.page-recto .vrow').length,
    rowHeights: [r1 ? r1.getBoundingClientRect().height : 0, r2 ? r2.getBoundingClientRect().height : 0],
    heads: heads,
    nums: nums,
    badge: badge ? badge.textContent : '',
    sec: document.querySelector('.vsec') ? document.querySelector('.vsec').textContent : '',
    notes: document.querySelectorAll('.fn').length,
    nav: nav ? nav.textContent : '',
    pageH: pages[0] ? Math.round(pages[0].getBoundingClientRect().height) : 0,
    pageW: pages[0] ? Math.round(pages[0].getBoundingClientRect().width) : 0,
    colW: pages[0] ? Math.round(pages[0].querySelector('.page-text').getBoundingClientRect().width) : 0,
    zhChars: (() => {
      const p = [...document.querySelectorAll('.page')].find(x => x.querySelector('.v-zh'));
      const z = p ? p.querySelector('.v-zh') : null;
      return z ? Math.round(z.getBoundingClientRect().width / 14.6) : 0;
    })(),
    leftLang: document.querySelector('.page-verso .v-en') ? 'en' : (document.querySelector('.page-verso .v-zh') ? 'zh' : ''),
    rightLang: document.querySelector('.page-recto .v-en') ? 'en' : (document.querySelector('.page-recto .v-zh') ? 'zh' : ''),
    titleOn: document.querySelector('.vtitle') ? (document.querySelector('.page-verso .vtitle') ? 'verso' : 'recto') : ''
  };
})()
