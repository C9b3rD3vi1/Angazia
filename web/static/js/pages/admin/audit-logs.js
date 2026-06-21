(function () {
  'use strict';

  var state = {
    page: 1,
    totalPages: 1,
    limit: 20,
    total: 0,
    logs: [],
    autoRefreshInterval: null,
    autoRefreshMs: 30000,
    allExpanded: false
  };

  function $(id) { return document.getElementById(id); }
  function qs(sel, ctx) { return (ctx || document).querySelector(sel); }
  function qsa(sel, ctx) { return (ctx || document).querySelectorAll(sel); }

  function escapeHtml(str) {
    if (!str && str !== 0) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(String(str)));
    return div.innerHTML;
  }

  function formatDateTime(str) {
    if (!str) return '-';
    var d = new Date(str);
    if (isNaN(d.getTime())) return str;
    return d.toLocaleDateString('en-US', {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  }

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  function showLoading() {
    var el = $('aal-loading');
    if (el) el.style.display = 'flex';
  }

  function hideLoading() {
    var el = $('aal-loading');
    if (el) el.style.display = 'none';
  }

  function showError(msg) {
    var el = $('aal-error');
    var txt = $('aal-error-text');
    if (txt) txt.textContent = msg;
    if (el) el.style.display = 'flex';
    var tw = $('aal-table-wrap');
    if (tw) tw.style.display = 'none';
    var em = $('aal-empty');
    if (em) em.style.display = 'none';
    var sm = $('aal-summary');
    if (sm) sm.style.display = 'none';
  }

  function hideError() {
    var el = $('aal-error');
    if (el) el.style.display = 'none';
  }

  function getActiveFilterCount() {
    var count = 0;
    var actionEl = $('aal-filter-action');
    var entityEl = $('aal-filter-entity');
    var dateFromEl = $('aal-filter-date-from');
    var dateToEl = $('aal-filter-date-to');
    if (actionEl && actionEl.value) count++;
    if (entityEl && entityEl.value) count++;
    if (dateFromEl && dateFromEl.value) count++;
    if (dateToEl && dateToEl.value) count++;
    return count;
  }

  function updateFilterUI() {
    var count = getActiveFilterCount();
    var badge = $('aal-filter-badge');
    var clearBtn = $('aal-clear-filters');
    if (badge) {
      if (count > 0) {
        badge.style.display = 'inline-flex';
        badge.textContent = count + ' filter' + (count !== 1 ? 's' : '') + ' active';
      } else {
        badge.style.display = 'none';
      }
    }
    if (clearBtn) {
      clearBtn.style.display = count > 0 ? 'inline-flex' : 'none';
    }
  }

  function resolveAdminName(log) {
    if (!log || !log.admin) return 'System';
    var admin = log.admin;
    if (admin.name) return admin.name;
    if (admin.email) {
      var parts = admin.email.split('@');
      return parts[0];
    }
    return 'System';
  }

  function fetchLogs(page) {
    showLoading();
    hideError();

    var p = page || 1;
    state.page = p;

    var params = { page: p, limit: state.limit };
    var actionEl = $('aal-filter-action');
    var entityEl = $('aal-filter-entity');
    var dateFromEl = $('aal-filter-date-from');
    var dateToEl = $('aal-filter-date-to');

    var action = actionEl ? actionEl.value : '';
    var entity = entityEl ? entityEl.value : '';
    var dateFrom = dateFromEl ? dateFromEl.value : '';
    var dateTo = dateToEl ? dateToEl.value : '';

    if (dateFrom && dateTo && dateFrom > dateTo) {
      hideLoading();
      showToast('"From" date cannot be after "To" date', 'error');
      return;
    }

    if (action) params.action = action;
    if (entity) params.entity_type = entity;
    if (dateFrom) params.date_from = dateFrom;
    if (dateTo) params.date_to = dateTo;

    updateFilterUI();

    AngaziaAPI.admin.auditLogs(params).then(function (data) {
      hideLoading();
      state.logs = data.logs || [];
      state.total = data.total || 0;
      state.page = data.page || 1;
      state.totalPages = data.total_pages || 1;
      state.limit = data.limit || 20;
      state.allExpanded = false;
      renderLogs();
      renderPagination();
      renderSummary();
    }).catch(function (err) {
      hideLoading();
      var msg = err.message || 'Failed to load audit logs';
      showError(msg);
      showToast(msg, 'error');
    });
  }

  function renderLogs() {
    var tbody = $('aal-tbody');
    var container = $('aal-table-wrap');
    var empty = $('aal-empty');

    if (!state.logs || state.logs.length === 0) {
      if (container) container.style.display = 'none';
      if (empty) {
        var count = getActiveFilterCount();
        var emptyText = empty.querySelector('.aal-empty-text') || $('aal-empty-text');
        if (emptyText) {
          emptyText.textContent = count > 0
            ? 'No audit log entries match your filters.'
            : 'No audit log entries found.';
        }
        empty.style.display = 'flex';
      }
      var expandBtn = $('aal-expand-all');
      if (expandBtn) expandBtn.style.display = 'none';
      return;
    }

    empty.style.display = 'none';
    if (container) container.style.display = 'block';
    var expandBtn = $('aal-expand-all');
    if (expandBtn) expandBtn.style.display = 'inline-flex';

    var html = '';
    for (var i = 0; i < state.logs.length; i++) {
      var log = state.logs[i];
      var adminName = resolveAdminName(log);
      var admin = log.admin || {};
      var adminAvatar = admin.avatar_url || '';
      var adminInitial = adminName ? adminName.charAt(0).toUpperCase() : '?';
      var adminAvatarHtml = adminAvatar
        ? '<img src="' + escapeHtml(adminAvatar) + '" alt="" class="aal-admin-avatar">'
        : '<span class="aal-admin-avatar aal-admin-avatar-text">' + escapeHtml(adminInitial) + '</span>';

      var oldVal = log.old_value;
      var newVal = log.new_value;
      if (oldVal && typeof oldVal === 'object') oldVal = JSON.stringify(oldVal, null, 2);
      if (newVal && typeof newVal === 'object') newVal = JSON.stringify(newVal, null, 2);
      oldVal = oldVal ? escapeHtml(String(oldVal)) : '<em class="aal-detail-none">none</em>';
      newVal = newVal ? escapeHtml(String(newVal)) : '<em class="aal-detail-none">none</em>';

      html += '<tr class="aal-row" data-log-id="' + log.id + '">' +
        '<td class="aal-col-expand"><button class="aal-expand-btn" data-expand="' + log.id + '" aria-label="Expand details">&#x25B6;</button></td>' +
        '<td><div class="aal-admin-cell">' + adminAvatarHtml + '<span class="aal-admin-name">' + escapeHtml(adminName) + '</span></div></td>' +
        '<td><span class="aal-action-badge aal-action-' + escapeHtml(log.action || '') + '">' + escapeHtml(log.action || '') + '</span></td>' +
        '<td><span class="aal-entity-badge aal-entity-' + escapeHtml(log.entity_type || '') + '">' + escapeHtml(log.entity_type || '') + '</span></td>' +
        '<td><code class="aal-entity-id">' + escapeHtml(String(log.entity_name || log.entity_id || '')) + '</code></td>' +
        '<td class="aal-timestamp">' + formatDateTime(log.created_at) + '</td>' +
        '<td><code class="aal-ip">' + escapeHtml(log.ip_address || '') + '</code></td>' +
        '</tr>' +
        '<tr class="aal-detail-row" data-detail-id="' + log.id + '" style="display:none">' +
        '<td colspan="7"><div class="aal-detail-content">' +
        '<div class="aal-detail-col"><div class="aal-detail-label">Old Value</div><pre class="aal-detail-json">' + oldVal + '</pre></div>' +
        '<div class="aal-detail-col"><div class="aal-detail-label">New Value</div><pre class="aal-detail-json">' + newVal + '</pre></div>' +
        '</div></td></tr>';
    }

    if (tbody) {
      tbody.innerHTML = html;
    } else if (container) {
      var tbl = container.querySelector('table');
      if (tbl) {
        var tb = tbl.querySelector('tbody');
        if (tb) tb.innerHTML = html;
      }
    }
  }

  function renderPagination() {
    var pag = $('aal-pagination');
    if (!pag) return;

    if (!state.totalPages || state.totalPages <= 1) {
      pag.innerHTML = '';
      return;
    }

    var html = '';

    if (state.page > 1) {
      html += '<button class="aal-page-btn" data-page="prev">&larr; Previous</button>';
    }

    var startPage = Math.max(1, state.page - 2);
    var endPage = Math.min(state.totalPages, state.page + 2);

    if (startPage > 1) {
      html += '<button class="aal-page-btn" data-page="1">1</button>';
      if (startPage > 2) html += '<span class="aal-page-info">...</span>';
    }

    for (var p = startPage; p <= endPage; p++) {
      var active = p === state.page ? ' active' : '';
      html += '<button class="aal-page-btn' + active + '" data-page="' + p + '">' + p + '</button>';
    }

    if (endPage < state.totalPages) {
      if (endPage < state.totalPages - 1) html += '<span class="aal-page-info">...</span>';
      html += '<button class="aal-page-btn" data-page="' + state.totalPages + '">' + state.totalPages + '</button>';
    }

    if (state.page < state.totalPages) {
      html += '<button class="aal-page-btn" data-page="next">Next &rarr;</button>';
    }

    html += '<span class="aal-page-info">Page ' + state.page + ' of ' + state.totalPages + '</span>';

    html += '<input type="number" class="aal-page-input" id="aal-page-input" min="1" max="' + state.totalPages + '" placeholder="Go">';
    html += '<button class="aal-page-btn" id="aal-page-go" data-page="go">Go</button>';

    pag.innerHTML = html;
  }

  function renderSummary() {
    var summary = $('aal-summary');
    var totalText = $('aal-total-text');
    if (!summary || !totalText) return;

    var count = state.logs.length;
    var total = state.total;
    var showing = count > 0
      ? 'Showing ' + count + ' of ' + total + ' entr' + (total !== 1 ? 'ies' : 'y')
      : '0 entries';

    var filtersActive = getActiveFilterCount() > 0;
    if (filtersActive) {
      var actionEl = $('aal-filter-action');
      var entityEl = $('aal-filter-entity');
      var labels = [];
      if (actionEl && actionEl.value) labels.push('action: ' + actionEl.value);
      if (entityEl && entityEl.value) labels.push('type: ' + entityEl.value);
      if (labels.length) showing += ' (' + labels.join(', ') + ')';
    }

    totalText.textContent = showing;
    summary.style.display = 'flex';
  }

  function toggleAllDetails(expand) {
    state.allExpanded = expand;
    var rows = qsa('.aal-detail-row');
    var btns = qsa('.aal-expand-btn');
    for (var i = 0; i < rows.length; i++) {
      rows[i].style.display = expand ? 'table-row' : 'none';
    }
    for (var j = 0; j < btns.length; j++) {
      btns[j].classList.toggle('expanded', expand);
    }
    var expandBtn = $('aal-expand-all');
    if (expandBtn) {
      expandBtn.innerHTML = expand ? '&#x25B2; Collapse All' : '&#x25BC; Expand All';
    }
  }

  function clearFilters() {
    var actionEl = $('aal-filter-action');
    var entityEl = $('aal-filter-entity');
    var dateFromEl = $('aal-filter-date-from');
    var dateToEl = $('aal-filter-date-to');
    if (actionEl) actionEl.value = '';
    if (entityEl) entityEl.value = '';
    if (dateFromEl) dateFromEl.value = '';
    if (dateToEl) dateToEl.value = '';
    updateFilterUI();
    fetchLogs(1);
  }

  function startAutoRefresh() {
    if (state.autoRefreshInterval) clearInterval(state.autoRefreshInterval);
    state.autoRefreshInterval = setInterval(function () {
      fetchLogs(state.page);
    }, state.autoRefreshMs);
  }

  function stopAutoRefresh() {
    if (state.autoRefreshInterval) {
      clearInterval(state.autoRefreshInterval);
      state.autoRefreshInterval = null;
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    var pageEl = qs('.aal-page');
    if (pageEl) {
      state.page = parseInt(pageEl.getAttribute('data-page'), 10) || 1;
      state.totalPages = parseInt(pageEl.getAttribute('data-total'), 10) || 1;
      state.limit = parseInt(pageEl.getAttribute('data-limit'), 10) || 20;
    }

    updateFilterUI();
    fetchLogs(state.page);

    /* Refresh button */
    var refreshBtn = qs('[data-action="refresh"]');
    if (refreshBtn) {
      refreshBtn.addEventListener('click', function (e) {
        e.preventDefault();
        fetchLogs(state.page);
      });
    }

    /* Clear filters */
    var clearBtn = $('aal-clear-filters');
    if (clearBtn) {
      clearBtn.addEventListener('click', clearFilters);
    }

    /* Expand all / collapse all */
    var expandAllBtn = $('aal-expand-all');
    if (expandAllBtn) {
      expandAllBtn.addEventListener('click', function () {
        toggleAllDetails(!state.allExpanded);
      });
    }

    /* Retry on error */
    var retryBtn = $('aal-error-retry');
    if (retryBtn) {
      retryBtn.addEventListener('click', function () {
        fetchLogs(1);
      });
    }

    /* Global click delegation */
    document.addEventListener('click', function (e) {
      /* Expand button */
      var expandBtn = e.target.closest('.aal-expand-btn');
      if (expandBtn) {
        var logId = expandBtn.getAttribute('data-expand');
        var detailRow = qs('.aal-detail-row[data-detail-id="' + logId + '"]');
        if (detailRow) {
          var isHidden = detailRow.style.display === 'none' || !detailRow.style.display;
          detailRow.style.display = isHidden ? 'table-row' : 'none';
          expandBtn.classList.toggle('expanded', isHidden);
        }
        return;
      }

      /* Pagination buttons */
      var pageBtn = e.target.closest('.aal-page-btn');
      if (pageBtn && !pageBtn.disabled) {
        var dir = pageBtn.getAttribute('data-page');
        if (dir === 'prev') { fetchLogs(state.page - 1); }
        else if (dir === 'next') { fetchLogs(state.page + 1); }
        else if (dir === 'go') { doGoToPage(); }
        else if (dir !== 'prev' && dir !== 'next' && dir !== 'go') {
          var pNum = parseInt(dir, 10);
          if (!isNaN(pNum) && pNum >= 1 && pNum <= state.totalPages) {
            fetchLogs(pNum);
          }
        }
        return;
      }
    });

    function doGoToPage() {
      var input = $('aal-page-input');
      if (!input) return;
      var val = parseInt(input.value, 10);
      if (isNaN(val) || val < 1 || val > state.totalPages) {
        showToast('Enter a valid page number (1-' + state.totalPages + ')', 'error');
        return;
      }
      fetchLogs(val);
    }

    /* Enter key on page input */
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') {
        var pageInput = e.target.closest('#aal-page-input');
        if (pageInput) {
          e.preventDefault();
          doGoToPage();
          return;
        }
      }
    });

    /* Apply filter button */
    var filterBtn = $('aal-filter-btn');
    if (filterBtn) {
      filterBtn.addEventListener('click', function () {
        fetchLogs(1);
      });
    }

    /* Enter key on date inputs */
    var dateInputs = qsa('#aal-filter-date-from, #aal-filter-date-to');
    for (var i = 0; i < dateInputs.length; i++) {
      dateInputs[i].addEventListener('keydown', function (e) {
        if (e.key === 'Enter') {
          e.preventDefault();
          fetchLogs(1);
        }
      });
    }

    /* Auto-refresh toggle */
    var autoRefreshToggle = $('aal-auto-refresh');
    if (autoRefreshToggle) {
      autoRefreshToggle.addEventListener('change', function () {
        if (this.checked) {
          startAutoRefresh();
        } else {
          stopAutoRefresh();
        }
      });
    }
  });
})();
