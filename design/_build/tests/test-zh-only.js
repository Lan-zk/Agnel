// 切到「中文」模式，检查英文是否真的被隐藏
(function () {
  reader.mode = '中文';
  reader.render();
  const en = document.querySelector('.verse-en');
  const zh = document.querySelector('.verse-zh');
  return {
    mode: document.body.dataset.readerMode,
    enExists: !!en,
    zhExists: !!zh,
    zhText: zh ? zh.textContent.slice(0, 30) : '',
    bodyCols: getComputedStyle(document.querySelector('.reader-body')).gridTemplateColumns
  };
})();
