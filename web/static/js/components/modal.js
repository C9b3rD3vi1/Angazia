var AngaziaModal = (function () {
  var overlay = null;
  var modalEl = null;
  var resolver = null;

  function open(content, opts) {
    close();
    opts = opts || {};
    overlay = document.createElement('div');
    overlay.className = 'angazia-modal-overlay';

    modalEl = document.createElement('div');
    modalEl.className = 'angazia-modal';

    if (opts.title) {
      var hdr = document.createElement('div');
      hdr.className = 'angazia-modal-header';
      hdr.innerHTML = '<h3>' + escapeHtml(opts.title) + '</h3>'
        + (opts.closable !== false ? '<button class="angazia-modal-close">&times;</button>' : '');
      modalEl.appendChild(hdr);
      var closeBtn = hdr.querySelector('.angazia-modal-close');
      if (closeBtn) closeBtn.addEventListener('click', close);
    }

    var body = document.createElement('div');
    body.className = 'angazia-modal-body';
    if (typeof content === 'string') body.innerHTML = content;
    else body.appendChild(content);
    modalEl.appendChild(body);

    if (opts.footer) {
      var ft = document.createElement('div');
      ft.className = 'angazia-modal-footer';
      if (typeof opts.footer === 'string') ft.innerHTML = opts.footer;
      else opts.footer.forEach(function (btn) { ft.appendChild(createFooterBtn(btn)); });
      modalEl.appendChild(ft);
    }

    overlay.appendChild(modalEl);
    document.body.appendChild(overlay);
    document.body.style.overflow = 'hidden';

    requestAnimationFrame(function () {
      overlay.classList.add('active');
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
      overlay.classList.remove('active');
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
    var cls = 'angazia-modal-btn';
    if (cfg.variant === 'primary') cls += ' primary';
    else if (cfg.variant === 'danger') cls += ' danger';
    else if (cfg.variant === 'ghost') cls += ' ghost';
    if (cfg.className) cls += ' ' + cfg.className;
    b.className = cls;
    if (cfg.disabled) { b.disabled = true; b.classList.add('disabled'); }
    b.addEventListener('click', function (e) {
      if (cfg.disabled) return;
      if (cfg.action) cfg.action(e, close);
      else close(cfg.value);
    });
    return b;
  }

  function alert(msg, title) {
    return new Promise(function (resolve) {
      open('<p class="angazia-modal-msg">' + escapeHtml(msg) + '</p>', {
        title: title || 'Notice',
        footer: [{ text: 'OK', variant: 'primary', action: function (e, c) { c(); resolve(); } }],
      });
    });
  }

  function confirm(msg, title) {
    return new Promise(function (resolve) {
      open('<p class="angazia-modal-msg">' + escapeHtml(msg) + '</p>', {
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
      container.innerHTML = '<p class="angazia-modal-msg" style="margin-bottom:12px;">' + escapeHtml(msg) + '</p>';
      var input = document.createElement('input');
      input.type = 'text';
      input.value = defaultValue || '';
      input.className = 'angazia-modal-input';
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
      b.classList.toggle('disabled', loading);
    });
    if (loading) {
      var spinner = document.createElement('span');
      spinner.className = 'angazia-modal-spinner';
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
