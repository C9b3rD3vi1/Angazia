var AngaziaModal = (function () {
  var overlay = null;
  var modalEl = null;
  var resolver = null;

  function open(content, opts) {
    close();
    opts = opts || {};
    overlay = document.createElement('div');
    overlay.className = 'angazia-modal-overlay';
    overlay.style.cssText = 'position:fixed;inset:0;z-index:9999;background:rgba(0,0,0,0.5);display:flex;align-items:center;justify-content:center;padding:20px;opacity:0;transition:opacity 0.2s ease;';

    modalEl = document.createElement('div');
    modalEl.className = 'angazia-modal';
    modalEl.style.cssText = 'background:var(--s1,#fff);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,12px);width:100%;max-width:' + (opts.maxWidth || '520px') + ';max-height:90vh;overflow-y:auto;box-shadow:0 20px 60px rgba(0,0,0,0.3);transform:scale(0.95) translateY(10px);transition:transform 0.2s ease;';

    if (opts.title) {
      var hdr = document.createElement('div');
      hdr.className = 'angazia-modal-header';
      hdr.style.cssText = 'display:flex;align-items:center;justify-content:space-between;padding:20px 24px 0;';
      hdr.innerHTML = '<h3 style="font-family:var(--fh,\'Inter\',sans-serif);font-size:18px;font-weight:600;color:var(--text,#111);margin:0;">' + escapeHtml(opts.title) + '</h3>'
        + (opts.closable !== false ? '<button class="angazia-modal-close" style="background:none;border:none;color:var(--muted,#6b7280);font-size:24px;cursor:pointer;padding:0;line-height:1;">&times;</button>' : '');
      modalEl.appendChild(hdr);
      var closeBtn = hdr.querySelector('.angazia-modal-close');
      if (closeBtn) closeBtn.addEventListener('click', close);
    }

    var body = document.createElement('div');
    body.className = 'angazia-modal-body';
    body.style.cssText = 'padding:' + (opts.title ? '16px 24px 24px' : '24px') + ';';
    if (typeof content === 'string') body.innerHTML = content;
    else body.appendChild(content);
    modalEl.appendChild(body);

    if (opts.footer) {
      var ft = document.createElement('div');
      ft.className = 'angazia-modal-footer';
      ft.style.cssText = 'display:flex;justify-content:flex-end;gap:12px;padding:0 24px 20px;';
      if (typeof opts.footer === 'string') ft.innerHTML = opts.footer;
      else opts.footer.forEach(function (btn) { ft.appendChild(createFooterBtn(btn)); });
      modalEl.appendChild(ft);
    }

    overlay.appendChild(modalEl);
    document.body.appendChild(overlay);
    document.body.style.overflow = 'hidden';

    requestAnimationFrame(function () {
      overlay.style.opacity = '1';
      modalEl.style.transform = 'scale(1) translateY(0)';
    });

    overlay.addEventListener('click', function (e) {
      if (opts.backdropClose !== false && e.target === overlay) {
        if (opts.closable !== false) close();
      }
    });

    document.addEventListener('keydown', onKeyDown);
  }

  function close(result) {
    document.removeEventListener('keydown', onKeyDown);
    if (overlay) {
      overlay.style.opacity = '0';
      modalEl.style.transform = 'scale(0.95) translateY(10px)';
      setTimeout(function () {
        if (overlay && overlay.parentNode) overlay.parentNode.removeChild(overlay);
        overlay = null;
        modalEl = null;
        document.body.style.overflow = '';
      }, 200);
    }
    if (resolver) { resolver(result); resolver = null; }
  }

  function onKeyDown(e) {
    if (e.key === 'Escape') close();
  }

  function createFooterBtn(cfg) {
    var b = document.createElement('button');
    b.textContent = cfg.text || '';
    b.className = cfg.className || '';
    b.style.cssText = 'padding:10px 20px;border-radius:var(--radius,10px);font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;font-weight:500;cursor:pointer;transition:all 0.15s;border:' + (cfg.variant === 'primary' ? 'none' : '1px solid var(--border,#e5e7eb)') + ';background:' + (cfg.variant === 'primary' ? 'var(--accent,#00e5a0)' : 'var(--s2,#f3f4f6)') + ';color:' + (cfg.variant === 'primary' ? '#050a0a' : 'var(--text,#111)') + ';';
    if (cfg.variant === 'danger') { b.style.background = 'var(--danger,#ef4444)'; b.style.color = '#fff'; b.style.border = 'none'; }
    if (cfg.variant === 'ghost') { b.style.background = 'transparent'; b.style.color = 'var(--muted,#6b7280)'; b.style.border = 'none'; }
    if (cfg.disabled) { b.style.opacity = '0.5'; b.style.cursor = 'not-allowed'; }
    b.addEventListener('click', function (e) {
      if (cfg.disabled) return;
      if (cfg.action) cfg.action(e, close);
      else close(cfg.value);
    });
    return b;
  }

  function alert(msg, title) {
    return new Promise(function (resolve) {
      open('<p style="color:var(--text,#111);font-size:14px;line-height:1.6;margin:0;">' + escapeHtml(msg) + '</p>', {
        title: title || 'Notice',
        footer: [{ text: 'OK', variant: 'primary', action: function (e, c) { c(); resolve(); } }],
      });
    });
  }

  function confirm(msg, title) {
    return new Promise(function (resolve) {
      open('<p style="color:var(--text,#111);font-size:14px;line-height:1.6;margin:0;">' + escapeHtml(msg) + '</p>', {
        title: title || 'Confirm',
        footer: [
          { text: 'Cancel', variant: 'ghost', action: function (e, c) { c(); resolve(false); } },
          { text: 'Confirm', variant: 'danger', action: function (e, c) { c(); resolve(true); } },
        ],
      });
    });
  }

  function prompt(msg, defaultValue, title) {
    return new Promise(function (resolve) {
      var container = document.createElement('div');
      container.innerHTML = '<p style="color:var(--text,#111);font-size:14px;line-height:1.6;margin:0 0 12px;">' + escapeHtml(msg) + '</p>';
      var input = document.createElement('input');
      input.type = 'text';
      input.value = defaultValue || '';
      input.style.cssText = 'width:100%;padding:10px 14px;background:var(--s2,#f3f4f6);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,10px);color:var(--text,#111);font-size:14px;outline:none;box-sizing:border-box;';
      input.addEventListener('keydown', function (e) {
        if (e.key === 'Enter') { e.preventDefault(); confirmBtn.click(); }
      });
      container.appendChild(input);
      var confirmBtn;
      open(container, {
        title: title || 'Input',
        footer: [
          { text: 'Cancel', variant: 'ghost', action: function (e, c) { c(); resolve(null); } },
          { text: 'OK', variant: 'primary', action: function (e, c) { c(); resolve(input.value); } },
        ],
      });
    });
  }

  function setLoading(loading) {
    if (!modalEl) return;
    var btns = modalEl.querySelectorAll('.angazia-modal-footer button');
    btns.forEach(function (b) {
      b.disabled = loading;
      b.style.opacity = loading ? '0.6' : '1';
      b.style.cursor = loading ? 'not-allowed' : 'pointer';
    });
    if (loading) {
      var spinner = document.createElement('span');
      spinner.className = 'angazia-modal-spinner';
      spinner.style.cssText = 'display:block;width:24px;height:24px;border:3px solid var(--border,#e5e7eb);border-top-color:var(--accent,#00e5a0);border-radius:50%;animation:angazia-spin 0.6s linear infinite;margin:12px auto;';
      modalEl.querySelector('.angazia-modal-body').appendChild(spinner);
    } else {
      var s = modalEl.querySelector('.angazia-modal-spinner');
      if (s) s.remove();
    }
  }

  var styleInjected = false;
  function injectStyles() {
    if (styleInjected) return;
    styleInjected = true;
    var s = document.createElement('style');
    s.textContent = '@keyframes angazia-spin{to{transform:rotate(360deg)}}';
    document.head.appendChild(s);
  }
  injectStyles();

  function escapeHtml(t) {
    if (!t) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(t));
    return d.innerHTML;
  }

  return {
    open: open,
    close: close,
    alert: alert,
    confirm: confirm,
    prompt: prompt,
    setLoading: setLoading,
  };
})();
