var AngaziaPagination = (function () {
  function render(container, state, onChange) {
    if (typeof container === 'string') container = document.querySelector(container);
    if (!container) return;

    var total = state.total || 0;
    var limit = state.limit || 20;
    var page = state.page || 1;
    var totalPages = Math.max(1, Math.ceil(total / limit));
    if (page < 1) page = 1;
    if (page > totalPages) page = totalPages;

    container.innerHTML = '';
    if (totalPages <= 1 && !state.always) return;

    var wrap = document.createElement('div');
    wrap.className = 'angazia-pagination';
    wrap.style.cssText = 'display:flex;align-items:center;justify-content:space-between;gap:16px;padding:16px 0;flex-wrap:wrap;';

    var info = document.createElement('div');
    info.className = 'pagination-info';
    info.style.cssText = 'font-family:var(--fm,\'Inter\',sans-serif);font-size:12px;color:var(--muted,#6b7280);';
    var from = total === 0 ? 0 : (page - 1) * limit + 1;
    var to = Math.min(page * limit, total);
    info.textContent = total > 0 ? 'Showing ' + from + '\u2013' + to + ' of ' + total : 'No results';

    var btns = document.createElement('div');
    btns.className = 'pagination-buttons';
    btns.style.cssText = 'display:flex;align-items:center;gap:4px;';

    function addBtn(label, pageNum, disabled, isActive) {
      var b = document.createElement('button');
      b.textContent = label;
      b.disabled = disabled;
      b.className = 'page-btn' + (isActive ? ' active' : '');
      b.style.cssText = 'min-width:36px;height:36px;padding:0 10px;border:1px solid ' + (isActive ? 'var(--accent,#00e5a0)' : 'var(--border,#e5e7eb)') + ';border-radius:var(--radius,8px);background:' + (isActive ? 'var(--accent,#00e5a0)' : 'var(--s1,#fff)') + ';color:' + (isActive ? '#050a0a' : 'var(--text,#111)') + ';font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;font-weight:' + (isActive ? '600' : '400') + ';cursor:' + (disabled ? 'not-allowed' : 'pointer') + ';opacity:' + (disabled ? '0.4' : '1') + ';transition:all 0.15s;';
      if (!disabled && !isActive) {
        b.addEventListener('mouseenter', function () { b.style.borderColor = 'var(--accent,#00e5a0)'; b.style.background = 'var(--s2,#f3f4f6)'; });
        b.addEventListener('mouseleave', function () { b.style.borderColor = 'var(--border,#e5e7eb)'; b.style.background = 'var(--s1,#fff)'; });
      }
      if (!disabled) {
        b.addEventListener('click', function () {
          if (onChange) onChange(pageNum);
        });
      }
      btns.appendChild(b);
    }

    addBtn('\u00AB', 1, page <= 1);
    addBtn('\u2039', page - 1, page <= 1);

    var pages = getPageRange(page, totalPages, 5);
    pages.forEach(function (p) {
      if (p === null) {
        var dot = document.createElement('span');
        dot.textContent = '\u2026';
        dot.style.cssText = 'padding:0 4px;color:var(--muted,#6b7280);font-size:13px;';
        btns.appendChild(dot);
      } else {
        addBtn(String(p), p, false, p === page);
      }
    });

    addBtn('\u203A', page + 1, page >= totalPages);
    addBtn('\u00BB', totalPages, page >= totalPages);

    wrap.appendChild(info);
    wrap.appendChild(btns);
    container.appendChild(wrap);
  }

  function getPageRange(current, total, maxVisible) {
    if (total <= maxVisible) {
      var r = [];
      for (var i = 1; i <= total; i++) r.push(i);
      return r;
    }
    var side = Math.floor((maxVisible - 3) / 2);
    var start = Math.max(2, current - side);
    var end = Math.min(total - 1, current + side);
    if (current - side <= 2) end = maxVisible - 2;
    if (current + side >= total - 1) start = total - maxVisible + 3;
    var pages = [1];
    if (start > 2) pages.push(null);
    for (var j = start; j <= end; j++) pages.push(j);
    if (end < total - 1) pages.push(null);
    if (total > 1) pages.push(total);
    return pages;
  }

  function calcOffset(page, limit) {
    return ((page || 1) - 1) * (limit || 20);
  }

  return {
    render: render,
    calcOffset: calcOffset,
  };
})();
