(function () {
  'use strict';
  function showToast(msg, type) {    if (window.AngaziaApp && AngaziaApp.showToast) {      AngaziaApp.showToast(msg, type);    } else {      alert((type === 'error' ? 'Error: ' : '') + msg);    }  }

  var state = {
    items: [],
    page: 1,
    limit: 20,
    total: 0,
    totalPages: 0,
    status: '',
    entityType: '',
    loading: false
  };

  var els = {};
  var rejectTargetId = null;
  var pollTimer = null;
  var POLL_MS = 15000;

  function init() {
    cacheElements();
    bindEvents();
    fetchItems();
  }

  function cacheElements() {
    els.loading = document.getElementById('am-loading');
    els.error = document.getElementById('am-error');
    els.errorText = document.getElementById('am-error-text');
    els.empty = document.getElementById('am-empty');
    els.list = document.getElementById('am-list');
    els.pagination = document.getElementById('am-pagination');
    els.pagiInfo = document.getElementById('am-pagi-info');
    els.pagiBtns = document.getElementById('am-pagi-btns');
    els.tabs = document.querySelectorAll('.am-tab');
    els.entityType = document.getElementById('am-entity-type');
    els.modal = document.getElementById('am-modal');
    els.rejectReason = document.getElementById('am-reject-reason');
    els.modalConfirm = document.getElementById('am-modal-confirm');
    els.modalCancel = document.getElementById('am-modal-cancel');
    els.modalClose = document.getElementById('am-modal-close');
    els.autoRefresh = document.getElementById('am-auto-refresh');
    els.refreshBtn = document.getElementById('am-refresh-btn');
    els.countBadge = document.getElementById('am-count');
  }

  function bindEvents() {
    for (var i = 0; i < els.tabs.length; i++) {
      els.tabs[i].addEventListener('click', function () {
        var status = this.getAttribute('data-status') || '';
        setActiveTab(this);
        state.status = status;
        state.page = 1;
        fetchItems();
      });
    }

    if (els.entityType) {
      els.entityType.addEventListener('change', function () {
        state.entityType = this.value;
        state.page = 1;
        fetchItems();
      });
    }

    if (els.modalConfirm) {
      els.modalConfirm.addEventListener('click', function () {
        submitRejection();
      });
    }

    if (els.modalCancel) {
      els.modalCancel.addEventListener('click', function () {
        hideRejectModal();
      });
    }

    if (els.modalClose) {
      els.modalClose.addEventListener('click', function () {
        hideRejectModal();
      });
    }

    if (els.autoRefresh) {
      els.autoRefresh.addEventListener('change', function () {
        if (this.checked) {
          startPolling();
        } else {
          stopPolling();
        }
      });
    }

    if (els.refreshBtn) {
      els.refreshBtn.addEventListener('click', function (e) {
        e.preventDefault();
        fetchItems();
      });
    }

    if (els.modal) {
      window.addEventListener('click', function (e) {
        if (e.target === els.modal) hideRejectModal();
      });
    }

    if (els.list) {
      els.list.addEventListener('click', function (e) {
        var btn = e.target.closest('[data-am-action]');
        if (!btn) return;
        var action = btn.getAttribute('data-am-action');
        var id = btn.getAttribute('data-am-id');
        if (!id) return;

        if (action === 'approve') {
          approveItem(btn, id);
        } else if (action === 'reject') {
          openRejectModal(id);
        }
      });
    }
  }

  function setActiveTab(el) {
    for (var i = 0; i < els.tabs.length; i++) {
      els.tabs[i].classList.remove('active');
    }
    el.classList.add('active');
  }

  function fetchItems() {
    if (state.loading) return;
    state.loading = true;
    showLoading();

    var params = [];
    params.push('page=' + state.page);
    params.push('limit=' + state.limit);
    if (state.status) params.push('status=' + encodeURIComponent(state.status));
    if (state.entityType) params.push('entity_type=' + encodeURIComponent(state.entityType));

    AngaziaAPI.get('/admin/moderation?' + params.join('&'))
      .then(function (data) {
        state.loading = false;
        state.items = data.items || data.moderation_items || [];
        state.total = data.total || 0;
        state.page = data.page || 1;
        state.totalPages = data.total_pages || Math.ceil(state.total / state.limit) || 0;
        if (els.countBadge) {
          els.countBadge.textContent = state.total;
        }
        renderItems();
      })
      .catch(function (err) {
        state.loading = false;
        showError(err.message || 'Network error');
      });
  }

  function showLoading() {
    if (els.loading) els.loading.style.display = 'flex';
    if (els.error) els.error.style.display = 'none';
    if (els.empty) els.empty.style.display = 'none';
    if (els.list) els.list.style.display = 'none';
    if (els.pagination) els.pagination.style.display = 'none';
  }

  function showError(msg) {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) {
      els.error.style.display = 'flex';
      if (els.errorText) els.errorText.textContent = msg || 'Failed to load items.';
    }
    if (els.empty) els.empty.style.display = 'none';
    if (els.list) els.list.style.display = 'none';
    if (els.pagination) els.pagination.style.display = 'none';
  }

  function showEmpty(hint) {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) els.error.style.display = 'none';
    if (els.empty) {
      els.empty.style.display = 'flex';
      var hintEl = els.empty.querySelector('.am-empty-hint');
      if (hintEl) hintEl.textContent = hint || 'No moderation items found.';
    }
    if (els.list) els.list.style.display = 'none';
    if (els.pagination) els.pagination.style.display = 'none';
  }

  function renderItems() {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) els.error.style.display = 'none';

    if (!state.items.length) {
      showEmpty(state.status ? 'No ' + state.status + ' items.' : 'No moderation items found.');
      return;
    }

    if (els.empty) els.empty.style.display = 'none';
    if (els.list) els.list.style.display = 'block';
    if (els.pagination) els.pagination.style.display = 'flex';

    var html = '';
    for (var i = 0; i < state.items.length; i++) {
      var item = state.items[i];
      html += renderItem(item);
    }

    if (els.list) els.list.innerHTML = html;
    renderPagination();
  }

  function renderItem(item) {
    var id = item.id;
    var status = item.status || 'pending';
    var entityType = item.entity_type || 'unknown';
    var title = item.title || item.name || entityType + ' #' + id;
    var submittedBy = item.submitted_by_name || item.submitted_by || 'Unknown';
    var submittedAt = item.created_at ? formatDateTime(item.created_at) : '-';
    var reason = item.rejection_reason || item.reason || '';

    var statusBadge = '';
    var statusColor = '';
    if (status === 'pending') { statusColor = 'warn'; statusBadge = 'Pending'; }
    else if (status === 'approved') { statusColor = 'accent'; statusBadge = 'Approved'; }
    else if (status === 'rejected') { statusColor = 'danger'; statusBadge = 'Rejected'; }
    else { statusColor = 'muted'; statusBadge = status; }

    var html = '';
    html += '<div class="am-item" data-id="' + id + '" data-status="' + status + '">';
    html += '<div class="am-item-head">';
    html += '<div class="am-item-info">';
    html += '<span class="am-item-type">' + escapeHtml(entityType) + '</span>';
    html += '<a href="' + (item.url || '#') + '" class="am-item-title">' + escapeHtml(title) + '</a>';
    html += '</div>';
    html += '<span class="am-status-badge" style="color:var(--' + statusColor + ');background:rgba(var(--' + statusColor + '-rgb),0.1);">' + statusBadge + '</span>';
    html += '</div>';

    html += '<div class="am-item-meta">';
    html += '<span>Submitted by ' + escapeHtml(submittedBy) + '</span>';
    html += '<span>' + submittedAt + '</span>';
    if (item.description) {
      html += '<p class="am-item-desc">' + escapeHtml(item.description) + '</p>';
    }
    if (reason && status === 'rejected') {
      html += '<div class="am-item-reason"><strong>Reason:</strong> ' + escapeHtml(reason) + '</div>';
    }
    html += '</div>';

    if (status === 'pending') {
      html += '<div class="am-item-actions">';
      html += '<button class="am-btn am-btn-sm am-btn-approve" data-am-action="approve" data-am-id="' + id + '">&#x2705; Approve</button>';
      html += '<button class="am-btn am-btn-sm am-btn-reject" data-am-action="reject" data-am-id="' + id + '">&#x274C; Reject</button>';
      html += '</div>';
    }

    html += '</div>';
    return html;
  }

  function renderPagination() {
    if (!els.pagiInfo || !els.pagiBtns) return;

    if (state.totalPages <= 1) {
      els.pagiBtns.innerHTML = '';
      els.pagiInfo.textContent = 'Showing ' + state.items.length + ' of ' + state.total + ' items';
      return;
    }

    els.pagiInfo.textContent = 'Page ' + state.page + ' of ' + state.totalPages + ' (' + state.total + ' items)';

    var html = '';
    html += '<button class="am-pagi-btn" onclick="amGoPage(1)"' + (state.page <= 1 ? ' disabled' : '') + '>&#xAB;</button>';
    html += '<button class="am-pagi-btn" onclick="amGoPage(' + (state.page - 1) + ')"' + (state.page <= 1 ? ' disabled' : '') + '>&#x2039;</button>';

    var start = Math.max(1, state.page - 2);
    var end = Math.min(state.totalPages, state.page + 2);

    if (start > 1) {
      html += '<button class="am-pagi-btn" onclick="amGoPage(1)">1</button>';
      if (start > 2) html += '<span class="am-pagi-btn" style="cursor:default;border:none;">...</span>';
    }

    for (var p = start; p <= end; p++) {
      html += '<button class="am-pagi-btn' + (p === state.page ? ' active' : '') + '" onclick="amGoPage(' + p + ')">' + p + '</button>';
    }

    if (end < state.totalPages) {
      if (end < state.totalPages - 1) html += '<span class="am-pagi-btn" style="cursor:default;border:none;">...</span>';
      html += '<button class="am-pagi-btn" onclick="amGoPage(' + state.totalPages + ')">' + state.totalPages + '</button>';
    }

    html += '<button class="am-pagi-btn" onclick="amGoPage(' + (state.page + 1) + ')"' + (state.page >= state.totalPages ? ' disabled' : '') + '>&#x203A;</button>';
    html += '<button class="am-pagi-btn" onclick="amGoPage(' + state.totalPages + ')"' + (state.page >= state.totalPages ? ' disabled' : '') + '>&#xBB;</button>';

    els.pagiBtns.innerHTML = html;
  }

  function approveItem(btn, id) {
    btn.disabled = true;
    AngaziaAPI.post('/admin/moderation/' + id + '/approve')
      .then(function () {
        showToast('Item approved', 'success');
        removeItem(id);
        updateCount(-1);
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to approve item', 'error');
      })
      .then(function () {
        btn.disabled = false;
      });
  }

  function openRejectModal(id) {
    rejectTargetId = id;
    if (els.modal) els.modal.style.display = 'flex';
    if (els.rejectReason) els.rejectReason.value = '';
    if (els.rejectReason) els.rejectReason.focus();
  }

  function hideRejectModal() {
    if (els.modal) els.modal.style.display = 'none';
    rejectTargetId = null;
    if (els.rejectReason) els.rejectReason.value = '';
  }

  function submitRejection() {
    var id = rejectTargetId;
    if (!id) return;
    var reason = els.rejectReason ? els.rejectReason.value.trim() : '';
    var btn = els.modalConfirm;
    if (btn) btn.disabled = true;

    AngaziaAPI.post('/admin/moderation/' + id + '/reject', { reason: reason })
      .then(function () {
        showToast('Item rejected', 'success');
        removeItem(id);
        updateCount(-1);
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to reject item', 'error');
      })
      .then(function () {
        if (btn) btn.disabled = false;
        hideRejectModal();
      });
  }

  function removeItem(id) {
    var el = document.querySelector('.am-item[data-id="' + id + '"]');
    if (el) {
      el.style.transition = 'opacity 0.2s, transform 0.2s';
      el.style.opacity = '0';
      el.style.transform = 'translateX(20px)';
      setTimeout(function () {
        if (el.parentNode) el.parentNode.removeChild(el);
        var remaining = document.querySelectorAll('.am-item');
        if (!remaining.length) {
          fetchItems();
        }
      }, 200);
    }
    state.items = state.items.filter(function (i) { return i.id !== id; });
    state.total = Math.max(0, state.total - 1);
  }

  function updateCount(delta) {
    if (els.countBadge) {
      var current = parseInt(els.countBadge.textContent, 10) || 0;
      els.countBadge.textContent = Math.max(0, current + delta);
    }
  }

  function startPolling() {
    stopPolling();
    pollTimer = setInterval(function () {
      if (els.autoRefresh && els.autoRefresh.checked) {
        var prevPage = state.page;
        fetchItems();
      }
    }, POLL_MS);
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  function amGoPage(p) {
    if (p < 1 || p > state.totalPages || p === state.page) return;
    state.page = p;
    fetchItems();
  }

  function amRefresh() {
    state.page = 1;
    fetchItems();
  }

  function formatDateTime(str) {
    if (!str) return '-';
    var d = new Date(str);
    if (isNaN(d.getTime())) return str;
    return (d.getMonth() + 1).toString().padStart(2, '0') + '/' +
      d.getDate().toString().padStart(2, '0') + '/' +
      d.getFullYear() + ' ' +
      d.getHours().toString().padStart(2, '0') + ':' +
      d.getMinutes().toString().padStart(2, '0');
  }

  function escapeHtml(str) {
    if (str === null || str === undefined) return '';
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  window.amGoPage = amGoPage;
  window.amRefresh = amRefresh;

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }

  if (els.autoRefresh && els.autoRefresh.checked !== false) {
    setTimeout(startPolling, POLL_MS);
  }
})();
