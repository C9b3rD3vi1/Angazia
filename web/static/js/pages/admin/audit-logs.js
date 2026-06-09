(function () {
  'use strict';
  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  var pageEl = document.querySelector('.aal-page');
  var currentPage = parseInt(pageEl ? pageEl.getAttribute('data-page') : '1', 10) || 1;
  var currentTotalPages = parseInt(pageEl ? pageEl.getAttribute('data-total') : '1', 10) || 1;
  var currentLimit = parseInt(pageEl ? pageEl.getAttribute('data-limit') : '20', 10) || 20;
  var autoRefreshInterval = null;
  var autoRefreshMs = 30000;

  function aalShowLoading() {
    var el = document.getElementById('aal-loading');
    if (el) el.style.display = 'flex';
  }

  function aalHideLoading() {
    var el = document.getElementById('aal-loading');
    if (el) el.style.display = 'none';
  }

  function aalBuildParams(page) {
    var params = [];
    var action = document.getElementById('aal-filter-action').value;
    var entity = document.getElementById('aal-filter-entity').value;
    var dateFrom = document.getElementById('aal-filter-date-from').value;
    var dateTo = document.getElementById('aal-filter-date-to').value;
    if (action) params.push('action=' + encodeURIComponent(action));
    if (entity) params.push('entity_type=' + encodeURIComponent(entity));
    if (dateFrom) params.push('date_from=' + encodeURIComponent(dateFrom));
    if (dateTo) params.push('date_to=' + encodeURIComponent(dateTo));
    if (page && page > 1) params.push('page=' + page);
    if (currentLimit) params.push('limit=' + currentLimit);
    return params.join('&');
  }

  function aalFetchLogs(page) {
    aalShowLoading();
    var qs = aalBuildParams(page || 1);
    var url = '/admin/audit-logs?' + qs;

    AngaziaAPI.get(url).then(function (data) {
      aalHideLoading();
      aalRenderLogs(data);
    }).catch(function (err) {
      aalHideLoading();
      showToast(err.message || 'Failed to load audit logs', 'error');
    });
  }

  function aalRenderLogs(data) {
    currentPage = data.page || 1;
    currentTotalPages = data.total_pages || 1;
    currentLimit = data.limit || 20;

    var tbody = document.getElementById('aal-tbody');
    var container = document.querySelector('.aal-table-wrap');
    var empty = document.querySelector('.aal-empty');
    var pagination = document.querySelector('.aal-pagination');
    var pageContainer = document.querySelector('.aal-page');

    if (!data.logs || data.logs.length === 0) {
      if (container) container.style.display = 'none';
      if (pagination) pagination.style.display = 'none';
      if (empty) {
        empty.style.display = 'block';
      } else {
        var ref = container || document.querySelector('.aal-filters');
        if (ref) {
          ref.insertAdjacentHTML('afterend',
            '<div class="aal-empty">' +
            '<span class="aal-empty-icon">&#x1F50D;</span>' +
            '<p class="aal-empty-text">No audit log entries found.</p>' +
            '</div>');
        }
      }
      return;
    }

    if (empty) empty.style.display = 'none';
    if (container) container.style.display = 'block';

    var html = '';
    for (var i = 0; i < data.logs.length; i++) {
      var log = data.logs[i];
      var adminName = (log.admin && log.admin.name) ? escapeHtml(log.admin.name) : 'System';
      var adminAvatar = (log.admin && log.admin.avatar) ? log.admin.avatar : '';
      var adminAvatarHtml = adminAvatar
        ? '<img src="' + escapeHtml(adminAvatar) + '" alt="" class="aal-admin-avatar">'
        : '<span class="aal-admin-avatar aal-admin-avatar-text">' + escapeHtml(adminName.charAt(0).toUpperCase()) + '</span>';

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
        '<td><code class="aal-entity-id">' + escapeHtml(String(log.entity_id || '')) + '</code></td>' +
        '<td class="aal-timestamp">' + escapeHtml(log.created_at || '') + '</td>' +
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
      container.querySelector('table tbody').innerHTML = html;
    }

    aalRenderPagination(data);
  }

  function aalRenderPagination(data) {
    var container = document.querySelector('.aal-page');
    if (!container) return;
    var oldPag = container.querySelector('.aal-pagination');
    if (oldPag) oldPag.remove();

    if (data.total_pages && data.total_pages > 1) {
      var pagHtml = '<div class="aal-pagination" data-page="' + data.page + '" data-total="' + data.total_pages + '" data-limit="' + (data.limit || 20) + '">';
      if (data.page > 1) {
        pagHtml += '<button class="aal-page-btn" data-page="prev">&larr; Previous</button>';
      }
      pagHtml += '<span class="aal-page-info">Page ' + data.page + ' of ' + data.total_pages + '</span>';
      if (data.page < data.total_pages) {
        pagHtml += '<button class="aal-page-btn" data-page="next">Next &rarr;</button>';
      }
      pagHtml += '</div>';
      var insertAfter = container.querySelector('.aal-table-wrap') || container.querySelector('.aal-empty');
      if (insertAfter) {
        insertAfter.insertAdjacentHTML('afterend', pagHtml);
      } else {
        container.insertAdjacentHTML('beforeend', pagHtml);
      }
    }
  }

  function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  window.aalRefresh = function () {
    aalFetchLogs(currentPage);
  };

  window.aalGoToPage = function (page) {
    if (page < 1 || page > currentTotalPages) return;
    aalFetchLogs(page);
  };

  function aalStartAutoRefresh() {
    if (autoRefreshInterval) clearInterval(autoRefreshInterval);
    autoRefreshInterval = setInterval(function () {
      aalFetchLogs(currentPage);
    }, autoRefreshMs);
  }

  function aalStopAutoRefresh() {
    if (autoRefreshInterval) {
      clearInterval(autoRefreshInterval);
      autoRefreshInterval = null;
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    var refreshBtn = document.querySelector('[data-action="refresh"]');
    if (refreshBtn) {
      refreshBtn.addEventListener('click', function (e) {
        e.preventDefault();
        aalRefresh();
      });
    }

    document.addEventListener('click', function (e) {
      var expandBtn = e.target.closest('.aal-expand-btn');
      if (expandBtn) {
        var logId = expandBtn.getAttribute('data-expand');
        var detailRow = document.querySelector('.aal-detail-row[data-detail-id="' + logId + '"]');
        if (detailRow) {
          var isHidden = detailRow.style.display === 'none' || !detailRow.style.display;
          detailRow.style.display = isHidden ? 'table-row' : 'none';
          expandBtn.classList.toggle('expanded', isHidden);
        }
      }

      var pageBtn = e.target.closest('.aal-page-btn');
      if (pageBtn && !pageBtn.disabled) {
        var dir = pageBtn.getAttribute('data-page');
        if (dir === 'prev') { aalGoToPage(currentPage - 1); }
        else if (dir === 'next') { aalGoToPage(currentPage + 1); }
      }
    });

    var filterBtn = document.getElementById('aal-filter-btn');
    if (filterBtn) {
      filterBtn.addEventListener('click', function () {
        aalFetchLogs(1);
      });
    }

    var toggle = document.getElementById('aal-auto-refresh');
    if (toggle) {
      toggle.addEventListener('change', function () {
        if (this.checked) {
          aalStartAutoRefresh();
        } else {
          aalStopAutoRefresh();
        }
      });
    }

    var dateInputs = document.querySelectorAll('#aal-filter-date-from, #aal-filter-date-to');
    for (var i = 0; i < dateInputs.length; i++) {
      dateInputs[i].addEventListener('keydown', function (e) {
        if (e.key === 'Enter') {
          e.preventDefault();
          aalFetchLogs(1);
        }
      });
    }
  });
})();
