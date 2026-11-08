// 打开资料面板，回报弹簧后的位置与把手坐标
(function () {
  const btn = document.querySelector('[data-open-detail]');
  btn.click();
  return new Promise(res => setTimeout(() => {
    const d = document.querySelector('[data-detail]');
    const h = d.querySelector('.sheet-handle');
    const hr = h.getBoundingClientRect();
    res({
      open: document.body.classList.contains('detail-open'),
      transform: d.style.transform,
      sheetTop: Math.round(d.getBoundingClientRect().top),
      vh: window.innerHeight,
      handleX: Math.round(hr.left + hr.width / 2),
      handleY: Math.round(hr.top + hr.height / 2)
    });
  }, 800));
})();
