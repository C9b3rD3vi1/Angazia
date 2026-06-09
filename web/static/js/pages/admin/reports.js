'use strict';

function showToast(msg, type) {
  if (window.AngaziaApp && AngaziaApp.showToast) {
    AngaziaApp.showToast(msg, type);
  } else {
    alert((type === 'error' ? 'Error: ' : '') + msg);
  }
}

(function () {
  var state = {
    tab: 'reported',
    entityType: '',
    reasonFilter: '',
    searchQuery: '',
    page: 1,
    totalPages: 1,
    total: 0,
    limit: 20,
    items: [],
    resolveTargetId: null,
    dismissTargetId: null
  };

  var reasonOptions = [];

  function $(id) { return document.getElementById(id); }
  function qs(sel, ctx) { return (ctx || document).querySelector(sel); }
  function qsa(sel, ctx) { return (ctx || document).querySelectorAll(sel); }

  function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function showLoading() {
    $('ar-loading').classList.add('active');
    $('ar-list-wrap').style.display = 'none';
  }

  function hideLoading() {
    $('ar-loading').classList.remove('active');
    $('ar-list-wrap').style.display = 'block';
  }

  function buildParams(page) {
    var params = [];
    if (state.tab) params.push('status=' + encodeURIComponent(state.tab));
    if (state.entityType) params.push('entity_type=' + encodeURIComponent(state.entityType));
    if (page && page > 1) params.push('page=' + page);
    params.push('limit=' + state.limit);
    return params.join('&');
  }

  function fetchReports(page) {
    showLoading();
    var p = page || 1;
    state.page = p;
    var qs = buildParams(p);
    var url = '/admin/moderation?' + qs;

    AngaziaAPI.get(url).then(function (data) {
      hideLoading();
      state.items = data.items || [];
      state.total = data.total || 0;
      state.page = data.page || 1;
      state.totalPages = data.total_pages || 1;
      state.limit = data.limit || 20;
      renderReports(state.items);
      renderStats(state.items);
      renderPagination();
    }).catch(function (err) {
      hideLoading();
      showToast(err.message || 'Failed to load reports', 'error');
    });
  }

  function fetchReportReasons() {
    var url = '/admin/report-reasons';
    if (state.entityType) url += '?entity_type=' + encodeURIComponent(state.entityType);
    AngaziaAPI.get(url).then(function (data) {
      if (data) {
        reasonOptions = data;
      }
      populateReasonFilter();
    }).catch(function () {
      reasonOptions = [];
      populateReasonFilter();
    });
  }

  function populateReasonFilter() {
    var sel = $('ar-reason-filter');
    if (!sel) return;
    var current = sel.value;
    sel.innerHTML = '<option value="">All Reasons</option>';
    if (reasonOptions && reasonOptions.length) {
      for (var i = 0; i < reasonOptions.length; i++) {
        var r = reasonOptions[i];
        var label = r.label || r.reason || r.name || r;
        var val = r.value || r.reason || r.name || r;
        var opt = document.createElement('option');
        opt.value = val;
        opt.textContent = label;
        sel.appendChild(opt);
      }
    }
    sel.value = current;
  }

  function renderStats(items) {
    var total = items.length;
    var reported = 0;
    var resolved = 0;
    var dismissed = 0;
    for (var i = 0; i < items.length; i++) {
      var st = (items[i].status || '').toLowerCase();
      if (st === 'reported') reported++;
      else if (st === 'resolved') resolved++;
      else if (st === 'dismissed') dismissed++;
    }
    $('ar-stat-total').querySelector('.ar-stat-value').textContent = total;
    $('ar-stat-reported').querySelector('.ar-stat-value').textContent = reported;
    $('ar-stat-resolved').querySelector('.ar-stat-value').textContent = resolved;
    $('ar-stat-dismissed').querySelector('.ar-stat-value').textContent = dismissed;
  }

  function applyClientFilters(items) {
    var query = state.searchQuery.toLowerCase().trim();
    var reason = state.reasonFilter;
    if (!query && !reason) return items;
    return items.filter(function (item) {
      if (query) {
        var searchable = ((item.reason || '') + ' ' + ((item.submitted_by && item.submitted_by.name) || '') + ' ' + (item.entity_type || '')).toLowerCase();
        if (searchable.indexOf(query) === -1) return false;
      }
      if (reason) {
        var itemReason = (item.reason || '').toLowerCase();
        if (itemReason.indexOf(reason.toLowerCase()) === -1) return false;
      }
      return true;
    });
  }

  function renderReports(items) {
    var list = $('ar-list');
    var empty = $('ar-empty');
    if (!list) return;

    var filtered = applyClientFilters(items);

    if (!filtered || filtered.length === 0) {
      list.innerHTML = '';
      list.style.display = 'none';
      empty.style.display = 'block';
      return;
    }

    empty.style.display = 'none';
    list.style.display = 'flex';

    var html = '';
    for (var i = 0; i < filtered.length; i++) {
      var item = filtered[i];
      var type = (item.entity_type || 'job').toLowerCase();
      var status = (item.status || 'reported').toLowerCase();
      var entityBadge = 'ar-badge-' + type;
      var statusClass = 'ar-status-badge ' + status;
      var submitted = item.submitted_by || {};
      var submittedName = submitted.name || 'Unknown';
      var submittedAvatar = submitted.avatar || '';
      var submittedInitial = submittedName.charAt(0).toUpperCase();
      var avatarHtml = submittedAvatar
        ? '<img src="' + escapeHtml(submittedAvatar) + '" alt="" class="ar-meta-user-avatar">'
        : '<span class="ar-meta-user-avatar-text">' + escapeHtml(submittedInitial) + '</span>';
      var reason = item.reason || 'No reason provided';
      var createdAt = item.created_at || '';
      var entityId = item.entity_id || '';
      var actionsHtml = '';
      if (status === 'reported') {
        actionsHtml = '<div class="ar-item-actions">' +
          '<button class="ar-btn ar-btn-resolve" data-action="resolve" data-id="' + item.id + '">&#x2705; Resolve</button>' +
          '<button class="ar-btn ar-btn-dismiss" data-action="dismiss" data-id="' + item.id + '">&#x274C; Dismiss</button>' +
          '</div>';
      } else {
        actionsHtml = '<div class="ar-item-actions">' +
          '<button class="ar-btn ar-btn-view" data-action="view" data-id="' + item.id + '">&#x1F50D; View Details</button>' +
          '</div>';
      }
      var reviewerName = (item.reviewed_by && item.reviewed_by.name) ? item.reviewed_by.name : '';
      var reviewerHtml = reviewerName
        ? '<span class="ar-meta-reviewer">Reviewed by: ' + escapeHtml(reviewerName) + '</span>'
        : '';

      html += '<div class="ar-item" data-id="' + item.id + '">' +
        '<div class="ar-item-top">' +
        '<div class="ar-item-badges">' +
        '<span class="ar-badge ' + entityBadge + '">' + escapeHtml(type) + '</span>' +
        '<span class="' + statusClass + '">' + escapeHtml(status) + '</span>' +
        '</div>' +
        '<p class="ar-item-reason">' + escapeHtml(reason) + '</p>' +
        '</div>' +
        '<div class="ar-item-meta">' +
        '<span class="ar-meta-user">' + avatarHtml + '<strong>' + escapeHtml(submittedName) + '</strong></span>' +
        '<span class="ar-meta-date">&#x1F4C5; ' + escapeHtml(createdAt) + '</span>' +
        reviewerHtml +
        '<span class="ar-meta-id">#' + escapeHtml(String(entityId)) + '</span>' +
        '</div>' +
        actionsHtml +
        '</div>';
    }
    list.innerHTML = html;
  }

  function renderPagination() {
    var pag = $('ar-pagination');
    if (!pag) return;
    if (!state.totalPages || state.totalPages <= 1) {
      pag.classList.add('hidden');
      return;
    }
    pag.classList.remove('hidden');

    var html = '';
    if (state.page > 1) {
      html += '<button class="ar-page-btn" data-page="prev">&larr; Previous</button>';
    }
    var startPage = Math.max(1, state.page - 2);
    var endPage = Math.min(state.totalPages, state.page + 2);
    if (startPage > 1) {
      html += '<button class="ar-page-btn" data-page="1">1</button>';
      if (startPage > 2) html += '<span class="ar-page-info">...</span>';
    }
    for (var p = startPage; p <= endPage; p++) {
      var active = p === state.page ? ' active' : '';
      html += '<button class="ar-page-btn' + active + '" data-page="' + p + '">' + p + '</button>';
    }
    if (endPage < state.totalPages) {
      if (endPage < state.totalPages - 1) html += '<span class="ar-page-info">...</span>';
      html += '<button class="ar-page-btn" data-page="' + state.totalPages + '">' + state.totalPages + '</button>';
    }
    if (state.page < state.totalPages) {
      html += '<button class="ar-page-btn" data-page="next">Next &rarr;</button>';
    }
    html += '<span class="ar-page-info">Page ' + state.page + ' of ' + state.totalPages + '</span>';
    pag.innerHTML = html;
  }

  function resolveReport(id) {
    state.resolveTargetId = id;
    $('ar-resolve-notes').value = '';
    $('ar-resolve-modal').style.display = 'flex';
    $('ar-resolve-notes').focus();
  }

  function confirmResolve() {
    var id = state.resolveTargetId;
    if (!id) return;
    var notes = $('ar-resolve-notes').value.trim();
    var btn = $('ar-modal-confirm-resolve');
    btn.disabled = true;

    AngaziaAPI.post('/admin/moderation/' + id + '/approve', { notes: notes }).then(function () {
      showToast('Report resolved successfully', 'success');
      removeItem(id);
      closeResolveModal();
    }).catch(function (err) {
      showToast(err.message || 'Failed to resolve report', 'error');
    }).then(function () {
      btn.disabled = false;
      state.resolveTargetId = null;
    });
  }

  function closeResolveModal() {
    $('ar-resolve-modal').style.display = 'none';
    state.resolveTargetId = null;
  }

  function dismissReport(id) {
    state.dismissTargetId = id;
    $('ar-dismiss-reason').value = '';
    $('ar-dismiss-modal').style.display = 'flex';
    $('ar-dismiss-reason').focus();
  }

  function confirmDismiss() {
    var id = state.dismissTargetId;
    if (!id) return;
    var reason = $('ar-dismiss-reason').value.trim();
    if (!reason) {
      showToast('Please provide a reason for dismissal', 'error');
      return;
    }
    var btn = $('ar-modal-confirm-dismiss');
    btn.disabled = true;

    AngaziaAPI.post('/admin/moderation/' + id + '/reject', { reason: reason }).then(function () {
      showToast('Report dismissed', 'success');
      removeItem(id);
      closeDismissModal();
    }).catch(function (err) {
      showToast(err.message || 'Failed to dismiss report', 'error');
    }).then(function () {
      btn.disabled = false;
      state.dismissTargetId = null;
    });
  }

  function closeDismissModal() {
    $('ar-dismiss-modal').style.display = 'none';
    state.dismissTargetId = null;
  }

  function removeItem(id) {
    var el = qs('.ar-item[data-id="' + id + '"]');
    if (el) el.remove();
    state.items = state.items.filter(function (i) { return String(i.id) !== String(id); });
    renderStats(state.items);
    var remaining = qsa('.ar-item');
    if (!remaining || remaining.length === 0) {
      var list = $('ar-list');
      if (list) list.innerHTML = '';
      $('ar-empty').style.display = 'block';
    }
  }

  function viewDetail(id) {
    var item = null;
    for (var i = 0; i < state.items.length; i++) {
      if (String(state.items[i].id) === String(id)) {
        item = state.items[i];
        break;
      }
    }
    if (!item) {
      showToast('Report not found', 'error');
      return;
    }
    var submitted = item.submitted_by || {};
    var reviewed = item.reviewed_by || {};
    $('ar-detail-id').textContent = item.id;
    $('ar-detail-entity-type').textContent = item.entity_type || '-';
    $('ar-detail-entity-id').textContent = item.entity_id || '-';
    $('ar-detail-status').textContent = item.status || '-';
    $('ar-detail-reason').textContent = item.reason || '-';
    $('ar-detail-submitted').textContent = submitted.name || 'Unknown';
    $('ar-detail-created').textContent = item.created_at || '-';
    $('ar-detail-reviewed').textContent = reviewed.name || 'Not yet reviewed';
    $('ar-detail-notes').textContent = item.notes || item.admin_notes || 'No notes';
    $('ar-detail-modal').style.display = 'flex';
  }

  function closeDetailModal() {
    $('ar-detail-modal').style.display = 'none';
  }

  document.addEventListener('DOMContentLoaded', function () {
    fetchReports(1);
    fetchReportReasons();

    /* Refresh button */
    var refreshBtn = document.querySelector('[data-action="refresh"]');
    if (refreshBtn) {
      refreshBtn.addEventListener('click', function () {
        fetchReports(state.page);
        fetchReportReasons();
      });
    }

    /* Tabs */
    var tabs = qsa('.ar-tab');
    for (var i = 0; i < tabs.length; i++) {
      tabs[i].addEventListener('click', function () {
        var tab = this.getAttribute('data-tab');
        if (tab === state.tab) return;
        state.tab = tab;
        var allTabs = qsa('.ar-tab');
        for (var j = 0; j < allTabs.length; j++) {
          allTabs[j].classList.remove('active');
        }
        this.classList.add('active');
        state.page = 1;
        fetchReports(1);
      });
    }

    /* Entity filter */
    var entityFilter = $('ar-entity-filter');
    if (entityFilter) {
      entityFilter.addEventListener('change', function () {
        state.entityType = this.value;
        state.page = 1;
        fetchReports(1);
        fetchReportReasons();
      });
    }

    /* Reason filter */
    var reasonFilter = $('ar-reason-filter');
    if (reasonFilter) {
      reasonFilter.addEventListener('change', function () {
        state.reasonFilter = this.value;
        renderReports(state.items);
      });
    }

    /* Search with debounce */
    var searchInput = $('ar-search');
    if (searchInput) {
      var searchTimer = null;
      searchInput.addEventListener('input', function () {
        if (searchTimer) clearTimeout(searchTimer);
        searchTimer = setTimeout(function () {
          state.searchQuery = searchInput.value;
          renderReports(state.items);
        }, 300);
      });
    }

    /* Global click delegation */
    document.addEventListener('click', function (e) {
      /* Action buttons */
      var actionBtn = e.target.closest('[data-action]');
      if (actionBtn) {
        var action = actionBtn.getAttribute('data-action');
        var id = actionBtn.getAttribute('data-id');
        if (id) {
          if (action === 'resolve') {
            e.stopPropagation();
            resolveReport(id);
          } else if (action === 'dismiss') {
            e.stopPropagation();
            dismissReport(id);
          } else if (action === 'view') {
            e.stopPropagation();
            viewDetail(id);
          }
        }
        return;
      }

      /* Item click for detail */
      var itemEl = e.target.closest('.ar-item');
      if (itemEl && !e.target.closest('[data-action]') && !e.target.closest('button')) {
        var itemId = itemEl.getAttribute('data-id');
        if (itemId) viewDetail(itemId);
        return;
      }

      /* Pagination */
      var pageBtn = e.target.closest('.ar-page-btn');
      if (pageBtn && !pageBtn.disabled) {
        var dir = pageBtn.getAttribute('data-page');
        if (dir === 'prev' && state.page > 1) {
          fetchReports(state.page - 1);
        } else if (dir === 'next' && state.page < state.totalPages) {
          fetchReports(state.page + 1);
        } else if (dir !== 'prev' && dir !== 'next') {
          var pNum = parseInt(dir, 10);
          if (!isNaN(pNum) && pNum >= 1 && pNum <= state.totalPages) {
            fetchReports(pNum);
          }
        }
      }
    });

    /* Resolve modal */
    $('ar-modal-confirm-resolve').addEventListener('click', confirmResolve);
    $('ar-modal-close').addEventListener('click', closeResolveModal);
    $('ar-modal-cancel').addEventListener('click', closeResolveModal);
    $('ar-resolve-modal').addEventListener('click', function (e) {
      if (e.target === this) closeResolveModal();
    });
    $('ar-resolve-notes').addEventListener('keydown', function (e) {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        confirmResolve();
      }
    });

    /* Dismiss modal */
    $('ar-modal-confirm-dismiss').addEventListener('click', confirmDismiss);
    $('ar-dismiss-modal-close').addEventListener('click', closeDismissModal);
    $('ar-dismiss-modal-cancel').addEventListener('click', closeDismissModal);
    $('ar-dismiss-modal').addEventListener('click', function (e) {
      if (e.target === this) closeDismissModal();
    });
    $('ar-dismiss-reason').addEventListener('keydown', function (e) {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        confirmDismiss();
      }
    });

    /* Detail modal */
    $('ar-detail-modal-close').addEventListener('click', closeDetailModal);
    $('ar-detail-modal-close-btn').addEventListener('click', closeDetailModal);
    $('ar-detail-modal').addEventListener('click', function (e) {
      if (e.target === this) closeDetailModal();
    });
  });
})();
