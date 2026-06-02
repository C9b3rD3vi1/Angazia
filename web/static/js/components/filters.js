var AngaziaFilters = (function () {
  var instanceId = 0;

  function create(container, config, onChange) {
    if (typeof container === 'string') container = document.querySelector(container);
    if (!container) return null;

    var id = ++instanceId;
    var state = {};
    var debounceTimer = null;
    var wrapper = document.createElement('div');
    wrapper.className = 'angazia-filters angazia-filters-' + id;
    wrapper.style.cssText = 'background:var(--s1,#fff);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,12px);padding:20px;margin-bottom:20px;';

    var header = document.createElement('div');
    header.style.cssText = 'display:flex;align-items:center;justify-content:space-between;margin-bottom:16px;';
    var title = document.createElement('span');
    title.style.cssText = 'font-family:var(--fh,\'Inter\',sans-serif);font-size:14px;font-weight:600;color:var(--text,#111);';
    title.textContent = 'Filters';
    var clearBtn = document.createElement('button');
    clearBtn.textContent = 'Clear all';
    clearBtn.style.cssText = 'background:none;border:none;color:var(--accent,#00e5a0);font-size:12px;font-weight:500;cursor:pointer;padding:4px 8px;border-radius:6px;transition:background 0.15s;display:none;';
    clearBtn.addEventListener('mouseenter', function () { clearBtn.style.background = 'var(--s2,#f3f4f6)'; });
    clearBtn.addEventListener('mouseleave', function () { clearBtn.style.background = 'transparent'; });
    clearBtn.addEventListener('click', function () { reset(); triggerChange(); });
    header.appendChild(title);
    header.appendChild(clearBtn);
    wrapper.appendChild(header);

    var grid = document.createElement('div');
    grid.className = 'filters-grid';
    grid.style.cssText = 'display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:12px;';
    wrapper.appendChild(grid);

    var fields = {};
    config.forEach(function (cfg) {
      var group = document.createElement('div');
      group.className = 'filter-group filter-group-' + cfg.key;
      group.style.cssText = 'display:flex;flex-direction:column;gap:4px;';

      var label = document.createElement('label');
      label.textContent = cfg.label || cfg.key;
      label.style.cssText = 'font-family:var(--fm,\'Inter\',sans-serif);font-size:11px;font-weight:500;color:var(--muted,#6b7280);text-transform:uppercase;letter-spacing:0.05em;';
      group.appendChild(label);

      var input;
      if (cfg.type === 'select') {
        input = document.createElement('select');
        input.style.cssText = 'padding:8px 12px;background:var(--s2,#f3f4f6);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,8px);color:var(--text,#111);font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;outline:none;cursor:pointer;';
        var optAll = document.createElement('option');
        optAll.value = '';
        optAll.textContent = 'All';
        input.appendChild(optAll);
        (cfg.options || []).forEach(function (o) {
          var opt = document.createElement('option');
          opt.value = o.value;
          opt.textContent = o.label || o.value;
          input.appendChild(opt);
        });
        input.addEventListener('change', function () {
          state[cfg.key] = input.value || undefined;
          showClearIfNeeded();
          debounce(triggerChange);
        });
      } else if (cfg.type === 'range') {
        var rangeWrap = document.createElement('div');
        rangeWrap.style.cssText = 'display:flex;align-items:center;gap:8px;';
        input = document.createElement('input');
        input.type = 'number';
        input.placeholder = (cfg.placeholderMin) || 'Min';
        input.style.cssText = 'flex:1;padding:8px 10px;background:var(--s2,#f3f4f6);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,8px);color:var(--text,#111);font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;outline:none;width:100%;box-sizing:border-box;';
        var inputMax = document.createElement('input');
        inputMax.type = 'number';
        inputMax.placeholder = (cfg.placeholderMax) || 'Max';
        inputMax.style.cssText = 'flex:1;padding:8px 10px;background:var(--s2,#f3f4f6);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,8px);color:var(--text,#111);font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;outline:none;width:100%;box-sizing:border-box;';
        rangeWrap.appendChild(input);
        rangeWrap.appendChild(inputMax);
        group.appendChild(rangeWrap);
        var onChangeFn = function () {
          var min = input.value ? Number(input.value) : undefined;
          var max = inputMax.value ? Number(inputMax.value) : undefined;
          if (min !== undefined || max !== undefined) state[cfg.key] = { min: min, max: max };
          else delete state[cfg.key];
          showClearIfNeeded();
          debounce(triggerChange);
        };
        input.addEventListener('input', onChangeFn);
        inputMax.addEventListener('input', onChangeFn);
      } else if (cfg.type === 'checkbox') {
        var checkboxWrap = document.createElement('div');
        checkboxWrap.style.cssText = 'display:flex;flex-wrap:wrap;gap:8px;';
        (cfg.options || []).forEach(function (o) {
          var chkId = 'f-' + id + '-' + cfg.key + '-' + o.value;
          var cbWrap = document.createElement('label');
          cbWrap.style.cssText = 'display:flex;align-items:center;gap:6px;cursor:pointer;font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;color:var(--text,#111);';
          var cb = document.createElement('input');
          cb.type = 'checkbox';
          cb.id = chkId;
          cb.value = o.value;
          cb.style.cssText = 'accent-color:var(--accent,#00e5a0);cursor:pointer;';
          cb.addEventListener('change', function () {
            var selected = [];
            checkboxWrap.querySelectorAll('input[type=checkbox]').forEach(function (c) {
              if (c.checked) selected.push(c.value);
            });
            if (selected.length) state[cfg.key] = selected;
            else delete state[cfg.key];
            showClearIfNeeded();
            debounce(triggerChange);
          });
          cbWrap.appendChild(cb);
          var lbl = document.createElement('span');
          lbl.textContent = o.label || o.value;
          cbWrap.appendChild(lbl);
          checkboxWrap.appendChild(cbWrap);
          if (!fields[cfg.key]) fields[cfg.key] = [];
          fields[cfg.key].push(cb);
        });
        group.appendChild(checkboxWrap);
      } else {
        input = document.createElement('input');
        input.type = 'text';
        input.placeholder = cfg.placeholder || 'Search ' + (cfg.label || cfg.key).toLowerCase() + '\u2026';
        input.style.cssText = 'padding:8px 12px;background:var(--s2,#f3f4f6);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,8px);color:var(--text,#111);font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;outline:none;width:100%;box-sizing:border-box;';
        input.addEventListener('input', function () {
          state[cfg.key] = input.value || undefined;
          showClearIfNeeded();
          debounce(triggerChange);
        });
        group.appendChild(input);
      }

      if (cfg.type === 'select' || cfg.type === 'text') group.appendChild(input);
      grid.appendChild(group);
    });

    container.appendChild(wrapper);

    function showClearIfNeeded() {
      var hasFilters = Object.keys(state).length > 0;
      clearBtn.style.display = hasFilters ? '' : 'none';
    }

    function debounce(fn) {
      if (debounceTimer) clearTimeout(debounceTimer);
      debounceTimer = setTimeout(fn, 300);
    }

    function triggerChange() {
      if (onChange) onChange(getValues());
    }

    function getValues() {
      var out = {};
      Object.keys(state).forEach(function (k) {
        if (state[k] !== undefined && state[k] !== null) out[k] = state[k];
      });
      return out;
    }

    function setValues(vals) {
      state = {};
      config.forEach(function (cfg) {
        var v = vals[cfg.key];
        if (v === undefined || v === null) return;
        state[cfg.key] = v;
        var el = grid.querySelector('.filter-group-' + cfg.key);
        if (!el) return;
        if (cfg.type === 'select') {
          var sel = el.querySelector('select');
          if (sel) sel.value = v;
        } else if (cfg.type === 'text') {
          var inp = el.querySelector('input');
          if (inp) inp.value = v;
        } else if (cfg.type === 'range') {
          var inputs = el.querySelectorAll('input[type=number]');
          if (inputs.length >= 2) {
            if (v.min !== undefined) inputs[0].value = v.min;
            if (v.max !== undefined) inputs[1].value = v.max;
          }
        } else if (cfg.type === 'checkbox') {
          var cbs = fields[cfg.key] || [];
          cbs.forEach(function (cb) {
            cb.checked = Array.isArray(v) && v.indexOf(cb.value) !== -1;
          });
        }
      });
      showClearIfNeeded();
    }

    function reset() {
      state = {};
      config.forEach(function (cfg) {
        var el = grid.querySelector('.filter-group-' + cfg.key);
        if (!el) return;
        if (cfg.type === 'select') {
          var sel = el.querySelector('select');
          if (sel) sel.value = '';
        } else if (cfg.type === 'text') {
          var inp = el.querySelector('input');
          if (inp) inp.value = '';
        } else if (cfg.type === 'range') {
          var inputs = el.querySelectorAll('input[type=number]');
          inputs.forEach(function (i) { i.value = ''; });
        } else if (cfg.type === 'checkbox') {
          var cbs = fields[cfg.key] || [];
          cbs.forEach(function (cb) { cb.checked = false; });
        }
      });
      showClearIfNeeded();
    }

    return { getValues: getValues, setValues: setValues, reset: reset, el: wrapper };
  }

  return { create: create };
})();
