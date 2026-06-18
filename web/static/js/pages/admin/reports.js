'use strict';

function showToast(msg, type) {
  if (window.AngaziaApp && AngaziaApp.showToast) {
    AngaziaApp.showToast(msg, type);
  } else {
    alert((type === 'error' ? 'Error: ' : '') + msg);
  }
}

(function () {
  var TAB_STATUS_MAP = {
    reported: 'pending',
    resolved: 'approved',
    dismissed: 'rejected'
  };

  var state = {
    tab: '',
    entityType: '',
    reasonFilter: '',
    searchQuery: '',
    dateFrom: '',
    dateTo: '',
    page: 1,
    totalPages: 1,
    total: 0,
    limit: 20,
    items: [],
    allItems: [],
    stats: { total: 0, pending: 0, approved: 0, rejected: 0 },
    resolveTargetId: null,
    dismissTargetId: null,
    entityChart: null,
    statusChart: null,
    editReasonId: null
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

  function formatDateTime(str) {
    if (!str) return '-';
    var d = new Date(str);
    if (isNaN(d.getTime())) return str;
    return d.toLocaleDateString('en-US', {
      year: 'numeric', month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit'
    });
  }

  function showLoading() {
    $('ar-loading').classList.add('active');
    $('ar-loading').style.display = 'flex';
    $('ar-error').style.display = 'none';
    $('ar-list-wrap').style.display = 'none';
  }

  function hideLoading() {
    $('ar-loading').classList.remove('active');
    $('ar-loading').style.display = 'none';
    $('ar-list-wrap').style.display = 'block';
  }

  function showError(msg) {
    $('ar-loading').style.display = 'none';
    $('ar-list-wrap').style.display = 'none';
    $('ar-error').style.display = 'flex';
    $('ar-error-text').textContent = msg || 'Failed to load reports';
  }

  function fetchStats() {
    AngaziaAPI.admin.moderation({ status: 'pending', limit: 1 }).then(function (d) {
      state.stats.pending = d.total || 0;
      renderStats();
      renderStatusChart();
    }).catch(function () {});
    AngaziaAPI.admin.moderation({ status: 'approved', limit: 1 }).then(function (d) {
      state.stats.approved = d.total || 0;
      renderStats();
      renderStatusChart();
    }).catch(function () {});
    AngaziaAPI.admin.moderation({ status: 'rejected', limit: 1 }).then(function (d) {
      state.stats.rejected = d.total || 0;
      renderStats();
      renderStatusChart();
    }).catch(function () {});
  }

  function fetchReports(page) {
    showLoading();
    var p = page || 1;
    state.page = p;
    var params = { page: p, limit: state.limit };
    var apiStatus = TAB_STATUS_MAP[state.tab];
    if (apiStatus) params.status = apiStatus;
    if (state.entityType) params.entity_type = state.entityType;
    if (state.dateFrom) params.date_from = state.dateFrom;
    if (state.dateTo) params.date_to = state.dateTo;

    AngaziaAPI.admin.moderation(params).then(function (data) {
      hideLoading();
      state.items = data.items || [];
      state.allItems = state.items.slice();
      state.total = data.total || 0;
      state.page = data.page || 1;
      state.totalPages = data.total_pages || 1;
      state.limit = data.limit || 20;
      applyClientFilters();
      renderStats();
      renderPagination();
      renderCharts();
    }).catch(function (err) {
      showError(err.message || 'Failed to load reports');
    });
  }

  function fetchReportReasons() {
    AngaziaAPI.admin.reportReasons().then(function (data) {
      if (data) {
        reasonOptions = data;
      }
      populateReasonFilter();
      renderReasonsList();
    }).catch(function () {
      reasonOptions = [];
      populateReasonFilter();
      renderReasonsList();
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
        var label = r.name || r.reason || r.label || '';
        var val = r.id || r.value || r.name || '';
        var opt = document.createElement('option');
        opt.value = val;
        opt.textContent = label;
        sel.appendChild(opt);
      }
    }
    sel.value = current;
  }

  function renderStats() {
    var s = state.stats;
    var total = s.pending + s.approved + s.rejected;
    state.stats.total = total;
    $('ar-stat-total').querySelector('.ar-stat-value').textContent = total;
    $('ar-stat-reported').querySelector('.ar-stat-value').textContent = s.pending;
    $('ar-stat-resolved').querySelector('.ar-stat-value').textContent = s.approved;
    $('ar-stat-dismissed').querySelector('.ar-stat-value').textContent = s.rejected;
  }

  function applyClientFilters() {
    var query = state.searchQuery.toLowerCase().trim();
    var reason = state.reasonFilter;
    if (!query && !reason) {
      state.items = state.allItems;
      renderReports(state.items);
      return;
    }
    state.items = state.allItems.filter(function (item) {
      if (query) {
        var sub = resolveSubmitter(item);
        var searchable = ((item.reason || '') + ' ' + (sub.name || '') + ' ' + (item.entity_type || '')).toLowerCase();
        if (searchable.indexOf(query) === -1) return false;
      }
      if (reason) {
        var itemReason = (item.reason || '').toLowerCase();
        var reasonText = '';
        for (var i = 0; i < reasonOptions.length; i++) {
          if ((reasonOptions[i].id || reasonOptions[i].name) === reason) {
            reasonText = (reasonOptions[i].name || '').toLowerCase();
            break;
          }
        }
        if (itemReason.indexOf(reasonText) === -1) return false;
      }
      return true;
    });
    renderReports(state.items);
  }

  function resolveSubmitter(item) {
    if (item.submitter && typeof item.submitter === 'object') return item.submitter;
    if (item.submitted_by && typeof item.submitted_by === 'object') return item.submitted_by;
    return { name: 'Unknown', avatar: '' };
  }

  function resolveReviewer(item) {
    if (item.reviewer && typeof item.reviewer === 'object') return item.reviewer;
    if (item.reviewed_by && typeof item.reviewed_by === 'object') return item.reviewed_by;
    return { name: '', avatar: '' };
  }

  function renderReports(items) {
    var list = $('ar-list');
    var empty = $('ar-empty');
    if (!list) return;

    if (!items || items.length === 0) {
      list.innerHTML = '';
      list.style.display = 'none';
      empty.style.display = 'block';
      $('ar-empty-text').textContent = state.searchQuery || state.reasonFilter
        ? 'No reports match your filters.'
        : 'No reports found.';
      return;
    }

    empty.style.display = 'none';
    list.style.display = 'flex';

    var html = '';
    for (var i = 0; i < items.length; i++) {
      var item = items[i];
      var type = (item.entity_type || 'job').toLowerCase();
      var status = (item.status || 'pending').toLowerCase();
      var entityBadge = 'ar-badge-' + type;
      var statusClass = 'ar-status-badge ' + status;
      var submitted = resolveSubmitter(item);
      var submittedName = submitted.name || 'Unknown';
      var submittedAvatar = submitted.avatar || '';
      var submittedInitial = submittedName.charAt(0).toUpperCase();
      var avatarHtml = submittedAvatar
        ? '<img src="' + escapeHtml(submittedAvatar) + '" alt="" class="ar-meta-user-avatar">'
        : '<span class="ar-meta-user-avatar-text">' + escapeHtml(submittedInitial) + '</span>';
      var reason = item.reason || 'No reason provided';
      var createdAt = formatDateTime(item.created_at);
      var entityId = item.entity_id || '';
      var actionsHtml = '';
      if (status === 'pending') {
        actionsHtml = '<div class="ar-item-actions">' +
          '<button class="ar-btn ar-btn-resolve" data-action="resolve" data-id="' + item.id + '">&#x2705; Resolve</button>' +
          '<button class="ar-btn ar-btn-dismiss" data-action="dismiss" data-id="' + item.id + '">&#x274C; Dismiss</button>' +
          '</div>';
      } else {
        actionsHtml = '<div class="ar-item-actions">' +
          '<button class="ar-btn ar-btn-view" data-action="view" data-id="' + item.id + '">&#x1F50D; View Details</button>' +
          '</div>';
      }
      var reviewer = resolveReviewer(item);
      var reviewerHtml = reviewer.name
        ? '<span class="ar-meta-reviewer">Reviewed by: ' + escapeHtml(reviewer.name) + '</span>'
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
        '<span class="ar-meta-date">&#x1F4C5; ' + createdAt + '</span>' +
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

  function renderCharts() {
    var entityCounts = {};
    var entityLabels = [];
    var entityData = [];
    for (var i = 0; i < state.items.length; i++) {
      var et = state.items[i].entity_type || 'unknown';
      entityCounts[et] = (entityCounts[et] || 0) + 1;
    }
    entityLabels = Object.keys(entityCounts);
    entityData = entityLabels.map(function (k) { return entityCounts[k]; });

    if (entityLabels.length === 0) {
      $('ar-charts').style.display = 'none';
      return;
    }
    $('ar-charts').style.display = 'grid';

    var colors = ['#00e5a0', '#ff4f4f', '#ffa94d', '#4dabf7', '#845ef7', '#ff6b6b', '#ffd43b', '#69db7c'];
    var bgColors = entityLabels.map(function (_, i) { return colors[i % colors.length]; });

    if (state.entityChart) { state.entityChart.destroy(); }
    var eCtx = $('ar-entity-chart').getContext('2d');
    state.entityChart = new Chart(eCtx, {
      type: 'doughnut',
      data: {
        labels: entityLabels,
        datasets: [{
          data: entityData,
          backgroundColor: bgColors,
          borderWidth: 0
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: true,
        plugins: {
          legend: {
            position: 'bottom',
            labels: { color: '#8b9bb5', font: { size: 10 }, padding: 12, usePointStyle: true }
          }
        }
      }
    });

    renderStatusChart();
  }

  function renderStatusChart() {
    if (!state.statusChart && !$('ar-charts')) return;
    var s = state.stats;
    var labels = ['Pending', 'Approved', 'Rejected'];
    var data = [s.pending, s.approved, s.rejected];
    var bgColors = ['#ffa94d', '#00e5a0', '#ff4f4f'];

    if (state.statusChart) { state.statusChart.destroy(); state.statusChart = null; }
    var sCtx = $('ar-status-chart');
    if (!sCtx) return;
    state.statusChart = new Chart(sCtx.getContext('2d'), {
      type: 'doughnut',
      data: {
        labels: labels,
        datasets: [{
          data: data,
          backgroundColor: bgColors,
          borderWidth: 0
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: true,
        plugins: {
          legend: {
            position: 'bottom',
            labels: { color: '#8b9bb5', font: { size: 10 }, padding: 12, usePointStyle: true }
          }
        }
      }
    });
  }

  function resolveReport(id) {
    state.resolveTargetId = id;
    $('ar-resolve-notes').value = '';
    $('ar-resolve-modal').style.display = 'flex';
    setTimeout(function () { $('ar-resolve-notes').focus(); }, 100);
  }

  function confirmResolve() {
    var id = state.resolveTargetId;
    if (!id) return;
    var btn = $('ar-modal-confirm-resolve');
    btn.disabled = true;
    btn.textContent = 'Resolving...';

    AngaziaAPI.admin.approveContent(id).then(function () {
      showToast('Report resolved successfully', 'success');
      removeItem(id, 'resolve');
      closeResolveModal();
    }).catch(function (err) {
      showToast(err.message || 'Failed to resolve report', 'error');
    }).then(function () {
      btn.disabled = false;
      btn.textContent = 'Resolve';
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
    setTimeout(function () { $('ar-dismiss-reason').focus(); }, 100);
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
    btn.textContent = 'Dismissing...';

    AngaziaAPI.admin.rejectContent(id, { reason: reason }).then(function () {
      showToast('Report dismissed successfully', 'success');
      removeItem(id, 'dismiss');
      closeDismissModal();
    }).catch(function (err) {
      showToast(err.message || 'Failed to dismiss report', 'error');
    }).then(function () {
      btn.disabled = false;
      btn.textContent = 'Dismiss';
      state.dismissTargetId = null;
    });
  }

  function closeDismissModal() {
    $('ar-dismiss-modal').style.display = 'none';
    state.dismissTargetId = null;
  }

  function removeItem(id, action) {
    var el = qs('.ar-item[data-id="' + id + '"]');
    if (el) el.remove();
    state.items = state.items.filter(function (i) { return String(i.id) !== String(id); });
    state.allItems = state.allItems.filter(function (i) { return String(i.id) !== String(id); });
    if (action === 'resolve') {
      state.stats.pending = Math.max(0, state.stats.pending - 1);
      state.stats.approved += 1;
    } else if (action === 'dismiss') {
      state.stats.pending = Math.max(0, state.stats.pending - 1);
      state.stats.rejected += 1;
    }
    renderStats();
    renderStatusChart();
    var remaining = qsa('.ar-item');
    if (!remaining || remaining.length === 0) {
      var list = $('ar-list');
      if (list) list.innerHTML = '';
      $('ar-empty').style.display = 'block';
      $('ar-empty-text').textContent = 'No reports found.';
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
      for (var j = 0; j < state.allItems.length; j++) {
        if (String(state.allItems[j].id) === String(id)) {
          item = state.allItems[j];
          break;
        }
      }
    }
    if (!item) {
      showToast('Report not found', 'error');
      return;
    }
    var submitted = resolveSubmitter(item);
    var reviewed = resolveReviewer(item);

    function setField(id, val) {
      var el = $(id);
      if (el) el.textContent = val || '-';
    }
    setField('ar-detail-id', item.id);
    setField('ar-detail-entity-type', item.entity_type);
    setField('ar-detail-entity-id', item.entity_id);
    setField('ar-detail-status', item.status);
    setField('ar-detail-reason', item.reason);
    setField('ar-detail-submitted', submitted.name || 'Unknown');
    setField('ar-detail-created', formatDateTime(item.created_at));
    setField('ar-detail-reviewed', reviewed.name || 'Not yet reviewed');
    var notesEl = $('ar-detail-notes');
    if (notesEl) notesEl.textContent = item.notes || item.admin_notes || 'No notes';
    $('ar-detail-modal').style.display = 'flex';
  }

  function closeDetailModal() {
    $('ar-detail-modal').style.display = 'none';
  }

  /* ── Report Reasons CRUD ── */

  function renderReasonsList() {
    var list = $('ar-reasons-list');
    var loading = $('ar-reasons-loading');
    if (!list) return;
    loading.style.display = 'none';
    if (!reasonOptions || reasonOptions.length === 0) {
      list.innerHTML = '<div class="ar-empty" style="display:block"><p class="ar-empty-text">No report reasons configured.</p></div>';
      return;
    }
    var html = '';
    for (var i = 0; i < reasonOptions.length; i++) {
      var r = reasonOptions[i];
      var active = r.is_active !== false;
      html += '<div class="ar-reason-item" data-id="' + r.id + '">' +
        '<div class="ar-reason-info">' +
        '<span class="ar-reason-name">' + escapeHtml(r.name) + '</span>' +
        '<span class="ar-reason-type">' + escapeHtml(r.entity_type || 'all') + '</span>' +
        (r.description ? '<p class="ar-reason-desc">' + escapeHtml(r.description) + '</p>' : '') +
        '</div>' +
        '<div class="ar-reason-controls">' +
        '<span class="ar-reason-badge ' + (active ? 'ar-reason-active' : 'ar-reason-inactive') + '">' + (active ? 'Active' : 'Inactive') + '</span>' +
        '<button class="ar-btn ar-btn-xs" data-action="edit-reason" data-id="' + r.id + '">&#x270F;&#xFE0F;</button>' +
        '<button class="ar-btn ar-btn-xs ar-btn-danger-xs" data-action="delete-reason" data-id="' + r.id + '">&#x1F5D1;&#xFE0F;</button>' +
        '</div>' +
        '</div>';
    }
    list.innerHTML = html;
  }

  function openAddReasonModal() {
    state.editReasonId = null;
    $('ar-reason-modal-title').textContent = 'Add Report Reason';
    $('ar-reason-name').value = '';
    $('ar-reason-desc').value = '';
    $('ar-reason-entity').value = '';
    $('ar-reason-sort').value = '0';
    $('ar-reason-modal').style.display = 'flex';
  }

  function openEditReasonModal(id) {
    var reason = null;
    for (var i = 0; i < reasonOptions.length; i++) {
      if (reasonOptions[i].id === id) { reason = reasonOptions[i]; break; }
    }
    if (!reason) { showToast('Reason not found', 'error'); return; }
    state.editReasonId = id;
    $('ar-reason-modal-title').textContent = 'Edit Report Reason';
    $('ar-reason-name').value = reason.name || '';
    $('ar-reason-desc').value = reason.description || '';
    $('ar-reason-entity').value = reason.entity_type || '';
    $('ar-reason-sort').value = reason.sort_order || 0;
    $('ar-reason-modal').style.display = 'flex';
  }

  function saveReason() {
    var name = $('ar-reason-name').value.trim();
    if (!name) { showToast('Reason name is required', 'error'); return; }
    var data = {
      name: name,
      description: $('ar-reason-desc').value.trim(),
      entity_type: $('ar-reason-entity').value,
      sort_order: parseInt($('ar-reason-sort').value, 10) || 0
    };
    var btn = $('ar-reason-modal-save');
    btn.disabled = true;
    btn.textContent = 'Saving...';

    var promise;
    if (state.editReasonId) {
      promise = AngaziaAPI.admin.updateReportReason(state.editReasonId, data);
    } else {
      promise = AngaziaAPI.admin.createReportReason(data);
    }

    promise.then(function () {
      showToast(state.editReasonId ? 'Reason updated' : 'Reason created', 'success');
      closeReasonModal();
      fetchReportReasons();
    }).catch(function (err) {
      showToast(err.message || 'Failed to save reason', 'error');
    }).then(function () {
      btn.disabled = false;
      btn.textContent = 'Save';
    });
  }

  function closeReasonModal() {
    $('ar-reason-modal').style.display = 'none';
    state.editReasonId = null;
  }

  function deleteReason(id) {
    if (!confirm('Delete this report reason?')) return;
    AngaziaAPI.admin.deleteReportReason(id).then(function () {
      showToast('Reason deleted', 'success');
      fetchReportReasons();
    }).catch(function (err) {
      showToast(err.message || 'Failed to delete reason', 'error');
    });
  }

  /* ── Init ── */

  document.addEventListener('DOMContentLoaded', function () {
    fetchReports(1);
    fetchReportReasons();
    fetchStats();

    var refreshBtn = document.querySelector('[data-action="refresh"]');
    if (refreshBtn) {
      refreshBtn.addEventListener('click', function () {
        fetchReports(state.page);
        fetchReportReasons();
        fetchStats();
      });
    }

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

    var entityFilter = $('ar-entity-filter');
    if (entityFilter) {
      entityFilter.addEventListener('change', function () {
        state.entityType = this.value;
        state.page = 1;
        fetchReports(1);
        fetchReportReasons();
      });
    }

    var reasonFilter = $('ar-reason-filter');
    if (reasonFilter) {
      reasonFilter.addEventListener('change', function () {
        state.reasonFilter = this.value;
        applyClientFilters();
      });
    }

    var searchInput = $('ar-search');
    if (searchInput) {
      var searchTimer = null;
      searchInput.addEventListener('input', function () {
        if (searchTimer) clearTimeout(searchTimer);
        searchTimer = setTimeout(function () {
          state.searchQuery = searchInput.value;
          applyClientFilters();
        }, 300);
      });
    }

    var dateFrom = $('ar-date-from');
    var dateTo = $('ar-date-to');
    if (dateFrom && dateTo) {
      var applyBtn = document.querySelector('[data-action="apply-filters"]');
      if (applyBtn) {
        applyBtn.addEventListener('click', function () {
          state.dateFrom = dateFrom.value;
          state.dateTo = dateTo.value;
          state.page = 1;
          fetchReports(1);
        });
      }
    }

    var toggleReasons = document.querySelector('[data-action="toggle-reasons"]');
    if (toggleReasons) {
      toggleReasons.addEventListener('click', function () {
        var body = $('ar-reasons-body');
        var toggle = $('ar-reasons-toggle');
        if (body.style.display === 'none') {
          body.style.display = 'block';
          toggle.textContent = '\u25B2';
        } else {
          body.style.display = 'none';
          toggle.textContent = '\u25BC';
        }
      });
    }

    document.addEventListener('click', function (e) {
      var actionBtn = e.target.closest('[data-action]');
      if (actionBtn) {
        var action = actionBtn.getAttribute('data-action');
        var id = actionBtn.getAttribute('data-id');
        if (id) {
          if (action === 'resolve') { e.stopPropagation(); resolveReport(id); return; }
          if (action === 'dismiss') { e.stopPropagation(); dismissReport(id); return; }
          if (action === 'view') { e.stopPropagation(); viewDetail(id); return; }
          if (action === 'edit-reason') { e.stopPropagation(); openEditReasonModal(id); return; }
          if (action === 'delete-reason') { e.stopPropagation(); deleteReason(id); return; }
        }
        if (action === 'add-reason') { openAddReasonModal(); return; }
        if (action === 'retry') { fetchReports(state.page); return; }
        return;
      }

      var itemEl = e.target.closest('.ar-item');
      if (itemEl && !e.target.closest('[data-action]') && !e.target.closest('button')) {
        var itemId = itemEl.getAttribute('data-id');
        if (itemId) viewDetail(itemId);
        return;
      }

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

    $('ar-modal-confirm-resolve').addEventListener('click', confirmResolve);
    $('ar-modal-close').addEventListener('click', closeResolveModal);
    $('ar-modal-cancel').addEventListener('click', closeResolveModal);
    $('ar-resolve-modal').addEventListener('click', function (e) {
      if (e.target === this) closeResolveModal();
    });
    $('ar-resolve-notes').addEventListener('keydown', function (e) {
      if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); confirmResolve(); }
    });

    $('ar-modal-confirm-dismiss').addEventListener('click', confirmDismiss);
    $('ar-dismiss-modal-close').addEventListener('click', closeDismissModal);
    $('ar-dismiss-modal-cancel').addEventListener('click', closeDismissModal);
    $('ar-dismiss-modal').addEventListener('click', function (e) {
      if (e.target === this) closeDismissModal();
    });
    $('ar-dismiss-reason').addEventListener('keydown', function (e) {
      if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); confirmDismiss(); }
    });

    $('ar-detail-modal-close').addEventListener('click', closeDetailModal);
    $('ar-detail-modal-close-btn').addEventListener('click', closeDetailModal);
    $('ar-detail-modal').addEventListener('click', function (e) {
      if (e.target === this) closeDetailModal();
    });

    $('ar-reason-modal-save').addEventListener('click', saveReason);
    $('ar-reason-modal-close').addEventListener('click', closeReasonModal);
    $('ar-reason-modal-cancel').addEventListener('click', closeReasonModal);
    $('ar-reason-modal').addEventListener('click', function (e) {
      if (e.target === this) closeReasonModal();
    });
    $('ar-reason-name').addEventListener('keydown', function (e) {
      if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); saveReason(); }
    });
  });
})();
