var AngaziaSearch = (function () {
  function init(inputSelector, opts) {
    var input;
    if (typeof inputSelector === 'string') {
      input = document.querySelector(inputSelector);
    } else {
      input = inputSelector;
    }
    if (!input) return null;

    opts = opts || {};
    var endpoint = opts.endpoint || '/api/search';
    var minChars = opts.minChars || 2;
    var debounceMs = opts.debounceMs || 350;
    var onSelect = opts.onSelect || function (item) {};
    var renderItem = opts.renderItem || null;
    var maxResults = opts.maxResults || 10;
    var noResultsText = opts.noResultsText || 'No results found';

    var debounceTimer = null;
    var currentRequest = null;
    var dropdown = null;
    var activeIndex = -1;
    var results = [];

    function createDropdown() {
      if (dropdown) return;
      dropdown = document.createElement('div');
      dropdown.className = 'angazia-search-dropdown';
      dropdown.style.cssText = 'position:absolute;top:100%;left:0;right:0;z-index:999;background:var(--s1,#fff);border:1px solid var(--border,#e5e7eb);border-radius:var(--radius,10px);margin-top:4px;box-shadow:0 12px 40px rgba(0,0,0,0.12);max-height:360px;overflow-y:auto;display:none;';
      dropdown.addEventListener('mouseenter', function () { input.dataset.searchFocus = 'dropdown'; });
      dropdown.addEventListener('mouseleave', function () { input.dataset.searchFocus = 'input'; });
      input.parentNode.style.position = 'relative';
      input.parentNode.appendChild(dropdown);
    }

    function showDropdown() {
      if (!dropdown) createDropdown();
      dropdown.style.display = 'block';
    }

    function hideDropdown() {
      if (dropdown) { dropdown.style.display = 'none'; dropdown.innerHTML = ''; }
      activeIndex = -1;
      input.dataset.searchFocus = 'input';
    }

    function renderResults(items) {
      if (!dropdown) createDropdown();
      dropdown.innerHTML = '';
      results = items;

      if (!items || items.length === 0) {
        var empty = document.createElement('div');
        empty.style.cssText = 'padding:16px 20px;font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;color:var(--muted,#6b7280);text-align:center;';
        empty.textContent = noResultsText;
        dropdown.appendChild(empty);
        showDropdown();
        return;
      }

      items.forEach(function (item, idx) {
        var el = document.createElement('div');
        el.className = 'angazia-search-item';
        el.style.cssText = 'padding:10px 16px;cursor:pointer;display:flex;align-items:center;gap:12px;transition:background 0.1s;' + (idx === items.length - 1 ? '' : 'border-bottom:1px solid var(--border,#e5e7eb);');
        el.addEventListener('mouseenter', function () {
          setActive(idx);
        });
        el.addEventListener('mousedown', function (e) {
          e.preventDefault();
          selectItem(item);
        });

        if (renderItem) {
          el.innerHTML = renderItem(item);
        } else {
          var icon = document.createElement('span');
          icon.style.cssText = 'width:32px;height:32px;border-radius:8px;background:var(--s2,#f3f4f6);display:flex;align-items:center;justify-content:center;font-size:14px;flex-shrink:0;';
          icon.textContent = '🔍';
          el.appendChild(icon);
          var textWrap = document.createElement('div');
          textWrap.style.cssText = 'flex:1;min-width:0;';
          var title = document.createElement('div');
          title.style.cssText = 'font-family:var(--fm,\'Inter\',sans-serif);font-size:13px;font-weight:500;color:var(--text,#111);';
          title.textContent = item.title || item.label || item.name || '';
          textWrap.appendChild(title);
          if (item.subtitle || item.description) {
            var sub = document.createElement('div');
            sub.style.cssText = 'font-size:11px;color:var(--muted,#6b7280);margin-top:2px;';
            sub.textContent = item.subtitle || item.description;
            textWrap.appendChild(sub);
          }
          el.appendChild(textWrap);
          if (item.meta) {
            var meta = document.createElement('span');
            meta.style.cssText = 'font-size:11px;color:var(--accent,#00e5a0);font-weight:500;';
            meta.textContent = item.meta;
            el.appendChild(meta);
          }
        }

        dropdown.appendChild(el);
      });

      activeIndex = -1;
      showDropdown();
    }

    function setActive(idx) {
      if (!dropdown) return;
      var items = dropdown.querySelectorAll('.angazia-search-item');
      items.forEach(function (el, i) {
        el.style.background = i === idx ? 'var(--s2,#f3f4f6)' : '';
        el.style.borderLeft = i === idx ? '3px solid var(--accent,#00e5a0)' : '3px solid transparent';
      });
      activeIndex = idx;
    }

    function selectItem(item) {
      hideDropdown();
      if (onSelect) onSelect(item);
    }

    function doSearch(term) {
      if (!term || term.length < minChars) {
        hideDropdown();
        return;
      }

      if (currentRequest && typeof currentRequest.abort === 'function') {
        currentRequest.abort();
      }

      activeIndex = -1;

      if (opts.searchFn) {
        var result = opts.searchFn(term);
        if (result && typeof result.then === 'function') {
          result.then(function (data) {
            renderResults((data.results || data.items || data).slice(0, maxResults));
          }).catch(function () {
            hideDropdown();
          });
        } else {
          renderResults((result.results || result.items || result).slice(0, maxResults));
        }
        return;
      }

      if (typeof AngaziaAPI !== 'undefined' && AngaziaAPI.search) {
        currentRequest = AngaziaAPI.search(term);
        if (currentRequest && typeof currentRequest.then === 'function') {
          currentRequest.then(function (data) {
            renderResults((data.results || data.items || data).slice(0, maxResults));
          }).catch(function () {
            hideDropdown();
          });
        }
        return;
      }

      var xhr = new XMLHttpRequest();
      currentRequest = xhr;
      xhr.open('GET', endpoint + '?q=' + encodeURIComponent(term), true);
      xhr.setRequestHeader('Accept', 'application/json');
      var token = localStorage.getItem('access_token');
      if (token) xhr.setRequestHeader('Authorization', 'Bearer ' + token);
      xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
          try {
            var data = JSON.parse(xhr.responseText);
            renderResults((data.results || data.items || data).slice(0, maxResults));
          } catch (e) {
            hideDropdown();
          }
        }
      };
      xhr.onerror = function () { hideDropdown(); };
      xhr.send();
    }

    input.addEventListener('input', function () {
      var val = input.value.trim();
      if (debounceTimer) clearTimeout(debounceTimer);
      debounceTimer = setTimeout(function () { doSearch(val); }, debounceMs);
    });

    input.addEventListener('keydown', function (e) {
      if (!dropdown || dropdown.style.display === 'none') return;
      var items = dropdown.querySelectorAll('.angazia-search-item');

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        var next = activeIndex < items.length - 1 ? activeIndex + 1 : 0;
        setActive(next);
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        var prev = activeIndex > 0 ? activeIndex - 1 : Math.max(0, items.length - 1);
        setActive(prev);
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (activeIndex >= 0 && activeIndex < results.length) {
          selectItem(results[activeIndex]);
        } else {
          selectItem({ query: input.value.trim() });
        }
      } else if (e.key === 'Escape') {
        hideDropdown();
      }
    });

    input.addEventListener('blur', function () {
      setTimeout(function () {
        if (input.dataset.searchFocus !== 'dropdown') {
          hideDropdown();
        }
      }, 200);
    });

    input.addEventListener('focus', function () {
      input.dataset.searchFocus = 'input';
      if (input.value.trim().length >= minChars) {
        doSearch(input.value.trim());
      }
    });

    function destroy() {
      if (debounceTimer) clearTimeout(debounceTimer);
      if (currentRequest && typeof currentRequest.abort === 'function') currentRequest.abort();
      if (dropdown && dropdown.parentNode) dropdown.parentNode.removeChild(dropdown);
      dropdown = null;
    }

    return { destroy: destroy, hide: hideDropdown, search: doSearch };
  }

  return { init: init };
})();
