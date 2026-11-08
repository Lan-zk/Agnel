// 英语基础：切到「词根与词族」页，检查三种关系都画出来了
(() => {
  const errs = [];
  window.addEventListener('error', e => errs.push(String(e.message)));
  goModule('basics');
  const tab = [...document.querySelectorAll('[data-btab]')]
    .find(x => /词根/.test(x.textContent || ''));
  if (tab) tab.click();
  const svg = document.querySelector('.root-graph svg');
  return {
    tabFound: !!tab,
    tabLabel: tab ? tab.textContent.trim() : null,
    svg: !!svg,
    origins: document.querySelectorAll('.rg-origin').length,
    roots: document.querySelectorAll('.rg-root').length,
    formLinks: document.querySelectorAll('.rg-link.is-form').length,
    derivLinks: document.querySelectorAll('.rg-link.is-deriv').length,
    formNodes: document.querySelectorAll('.rg-form').length,
    derivNodes: document.querySelectorAll('.rg-deriv').length,
    colLabels: [...document.querySelectorAll('.rg-col-label')].map(n => n.textContent.trim()),
    errors: errs,
  };
})();
