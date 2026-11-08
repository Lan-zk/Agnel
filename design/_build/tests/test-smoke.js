// 全模块冒烟：四个模块 + 书页对开 + 查词浮层
(function () {
  const errs = [];
  window.addEventListener('error', e => errs.push(String(e.message)));
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const q = s => document.querySelector(s);

  return (async function () {
    const out = {};

    // 书页对开
    reader.mode = '书页对开';
    reader.render();
    await sleep(250);
    out.mode = document.body.dataset.readerMode;
    out.pages = document.querySelectorAll('.page').length;
    out.rows = document.querySelectorAll('.page-verso .vrow').length;
    out.rightRows = document.querySelectorAll('.page-recto .vrow').length;
    out.leftLang = q('.page-verso .v-en') ? 'en' : 'zh';
    out.rightLang = q('.page-recto .v-en') ? 'en' : 'zh';
    out.head = q('.page-run') ? q('.page-run').textContent.trim() : '';
    out.nav = q('.book-nav-now') ? q('.book-nav-now').textContent.trim() : '';
    out.pageH = q('.page') ? Math.round(q('.page').getBoundingClientRect().height) : 0;

    // 点词查词（对开页里的英文，两页里哪一页都行）
    const words = [...document.querySelectorAll('.page .v-en .w')];
    const w = words.find(x => x.textContent.trim().toLowerCase() === 'created') || words[0];
    if (w) w.click();
    await sleep(300);
    out.lookupOpen = !!q('.lookup');
    out.lookupHasCtx = !!(q('.lookup-context') && q('.lookup-context').textContent.trim());
    out.lookupRows = document.querySelectorAll('.lookup-row').length;
    out.lookupSrc = q('.lookup-src') ? q('.lookup-src').textContent.trim() : '';
    lookup.close();

    // 翻页
    const next = [...document.querySelectorAll('.book-nav .pager-btn')].find(b => /后一页/.test(b.textContent));
    if (next) next.click();
    await sleep(250);
    out.navAfter = q('.book-nav-now') ? q('.book-nav-now').textContent.trim() : '';

    // 词库
    goModule('bank');
    await sleep(200);
    out.bankRows = document.querySelectorAll('.word-row').length;
    const first = q('.word-row');
    if (first) { first.click(); await sleep(200); }
    out.detailBlocks = document.querySelectorAll('.detail-block').length;

    // 练习
    goModule('practice');
    await sleep(250);
    out.practiceCard = !!q('.practice-card');

    // 英语基础
    goModule('basics');
    await sleep(250);
    out.phonicCells = document.querySelectorAll('.phonic-cell').length;
    out.rgLinks = document.querySelectorAll('.rg-link').length;

    // 回到阅读，确认 body 属性已复位
    goModule('read');
    await sleep(200);
    out.modeAfter = document.body.dataset.readerMode;

    out.errors = errs;
    return out;
  })();
})();
