(function () {
  'use strict';

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  var state = {
    users: [],
    selected: [],
    page: 1,
    limit: 20,
    total: 0,
    totalPages: 0,
    search: '',
    role: '',
    status: '',
    verified: '',
    loading: false,
    pendingActions: {},
    chartPeriod: 30,
  };

  var chartInstance = null;

  var els = {};
  var modalCallback = null;

  function init() {
    cacheElements();
    bindEvents();
    fetchUsers();
    fetchChart();
    fetchStats();
  }

  function cacheElements() {
    var ids = [
      'au-search', 'au-filter-role', 'au-filter-status', 'au-filter-verified',
      'au-filter-btn', 'au-tbody', 'au-loading', 'au-error', 'au-error-text',
      'au-empty', 'au-empty-hint', 'au-table-wrap', 'au-pagination',
      'au-pagi-info', 'au-pagi-btns', 'au-bulk', 'au-bulk-count',
      'au-select-all', 'au-modal', 'au-modal-title', 'au-modal-msg',
      'au-modal-confirm', 'au-modal-cancel', 'au-modal-close',
      'au-view-modal', 'au-view-body', 'au-view-loading',
      'au-growth-chart', 'au-stat-total', 'au-stat-active', 'au-stat-new', 'au-stat-suspended', 'au-stat-unverified',
    ];
    ids.forEach(function (id) { els[id] = document.getElementById(id); });
  }

  function bindEvents() {
    els['au-filter-btn'].addEventListener('click', function () {
      state.page = 1;
      fetchUsers();
    });

    els['au-search'].addEventListener('keydown', function (e) {
      if (e.key === 'Enter') { state.page = 1; fetchUsers(); }
    });

    els['au-select-all'].addEventListener('change', function () {
      state.selected = els['au-select-all'].checked ? state.users.map(function (u) { return u.id; }) : [];
      updateBulkBar();
      updateRowCheckboxes();
    });

    document.getElementById('au-bulk-suspend').addEventListener('click', function () {
      if (!state.selected.length) return;
      showConfirmModal('Suspend Users', 'Are you sure you want to suspend <strong>' + state.selected.length + '</strong> user(s)?', 'danger', function () {
        bulkAction('suspend');
      });
    });

    document.getElementById('au-bulk-activate').addEventListener('click', function () {
      if (!state.selected.length) return;
      showConfirmModal('Activate Users', 'Are you sure you want to activate <strong>' + state.selected.length + '</strong> user(s)?', 'success', function () {
        bulkAction('activate');
      });
    });

    document.getElementById('au-bulk-delete').addEventListener('click', function () {
      if (!state.selected.length) return;
      showConfirmModal('Delete Users', 'Are you sure you want to delete <strong>' + state.selected.length + '</strong> user(s)?<br><br><em>This action cannot be undone.</em>', 'danger', function () {
        bulkAction('delete');
      });
    });

    document.getElementById('au-bulk-clear').addEventListener('click', function () {
      state.selected = [];
      els['au-select-all'].checked = false;
      updateBulkBar();
      updateRowCheckboxes();
    });

    var reloadBtn = document.querySelector('[data-action="reload"]');
    if (reloadBtn) {
      reloadBtn.addEventListener('click', function (e) { e.preventDefault(); state.page = 1; fetchUsers(); });
    }

    var retryBtn = document.querySelector('[data-action="retry"]');
    if (retryBtn) {
      retryBtn.addEventListener('click', function () { fetchUsers(); });
    }

    document.querySelectorAll('.au-chart-period').forEach(function (btn) {
      btn.addEventListener('click', function () {
        document.querySelectorAll('.au-chart-period').forEach(function (b) { b.classList.remove('active'); });
        btn.classList.add('active');
        state.chartPeriod = parseInt(btn.getAttribute('data-period'), 10) || 30;
        fetchChart();
      });
    });

    els['au-modal-confirm'].addEventListener('click', function () {
      if (typeof modalCallback === 'function') {
        var cb = modalCallback;
        modalCallback = null;
        cb();
      }
      hideModal();
    });

    els['au-modal-cancel'].addEventListener('click', hideModal);
    els['au-modal-close'].addEventListener('click', hideModal);

    document.getElementById('au-view-close').addEventListener('click', function () { hideViewModal(); });
    document.getElementById('au-view-close-btn').addEventListener('click', function () { hideViewModal(); });

    document.addEventListener('click', function (e) {
      if (e.target === els['au-modal']) hideModal();
      if (e.target === els['au-view-modal']) hideViewModal();
      var actionBtn = e.target.closest('[data-user-action]');
      if (actionBtn) handleActionClick(actionBtn);
    });

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') {
        if (els['au-view-modal'].style.display === 'flex') hideViewModal();
        else if (els['au-modal'].style.display === 'flex') hideModal();
      }
    });
  }

  function handleActionClick(btn) {
    var action = btn.getAttribute('data-user-action');
    var userId = btn.getAttribute('data-user-id');
    if (!action || !userId) return;

    var user = getUserById(userId);
    var userName = user && (user.full_name || user.email || 'this user');

    switch (action) {
      case 'view':
        viewUser(userId);
        break;
      case 'suspend':
        showConfirmModal('Suspend User', 'Are you sure you want to suspend <strong>' + escapeHtml(userName) + '</strong>?<br><br>They will be unable to access the platform.', 'danger', function () {
          executeAction(userId, 'suspend');
        });
        break;
      case 'activate':
        showConfirmModal('Activate User', 'Are you sure you want to activate <strong>' + escapeHtml(userName) + '</strong>?', 'success', function () {
          executeAction(userId, 'activate');
        });
        break;
      case 'verify':
        showConfirmModal('Verify User', 'Are you sure you want to verify <strong>' + escapeHtml(userName) + '</strong>?', 'primary', function () {
          executeAction(userId, 'verify');
        });
        break;
      case 'delete':
        showConfirmModal('Delete User', 'Are you sure you want to delete <strong>' + escapeHtml(userName) + '</strong>?<br><br><em>This action cannot be undone.</em>', 'danger', function () {
          executeAction(userId, 'delete');
        });
        break;
    }
  }

  function getUserById(id) {
    for (var i = 0; i < state.users.length; i++) {
      if (state.users[i].id === id) return state.users[i];
    }
    return null;
  }

  function executeAction(userId, action) {
    state.pendingActions[userId] = true;
    updateActionButton(userId, action, true);

    var promise;
    switch (action) {
      case 'suspend': promise = AngaziaAPI.admin.suspendUser(userId); break;
      case 'activate': promise = AngaziaAPI.admin.activateUser(userId); break;
      case 'verify': promise = AngaziaAPI.admin.verifyUser(userId); break;
      case 'delete': promise = AngaziaAPI.admin.deleteUser(userId); break;
    }

    promise
      .then(function () {
        showToast('User ' + action + 'd successfully', 'success');
        fetchUsers();
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to ' + action + ' user', 'error');
        state.pendingActions[userId] = false;
        updateActionButton(userId, action, false);
      });
  }

  function updateActionButton(userId, action, loading) {
    var row = els['au-tbody'].querySelector('tr[data-id="' + userId + '"]');
    if (!row) return;
    var btn = row.querySelector('[data-user-action="' + action + '"]');
    if (!btn) return;
    btn.disabled = loading;
    btn.textContent = loading ? 'Processing...' : btn.getAttribute('data-original-text') || action.charAt(0).toUpperCase() + action.slice(1);
  }

  function fetchUsers() {
    if (state.loading) return;
    state.loading = true;

    state.search = els['au-search'].value.trim();
    state.role = els['au-filter-role'].value;
    state.status = els['au-filter-status'].value;
    state.verified = els['au-filter-verified'].value;

    showView('loading');

    var params = {
      page: state.page,
      limit: state.limit,
    };
    if (state.search) params.search = state.search;
    if (state.role) params.role = state.role;
    if (state.status) params.is_active = state.status === 'active';
    if (state.verified) params.is_verified = state.verified === 'verified';

    AngaziaAPI.admin.users(params)
      .then(function (data) {
        state.loading = false;
        state.users = data.users || [];
        state.total = data.total || 0;
        state.page = data.page || 1;
        state.limit = data.limit || 20;
        state.totalPages = data.total_pages || Math.ceil(state.total / state.limit) || 0;
        state.selected = [];
        els['au-select-all'].checked = false;
        updateBulkBar();
        renderUsers();
      })
      .catch(function (err) {
        state.loading = false;
        showView('error', err.message || 'Network error');
      });
  }

  function showView(view, msg) {
    els['au-loading'].style.display = view === 'loading' ? 'flex' : 'none';
    els['au-error'].style.display = view === 'error' ? 'flex' : 'none';
    els['au-empty'].style.display = 'none';
    els['au-table-wrap'].style.display = 'none';
    els['au-pagination'].style.display = 'none';
    if (view === 'error' && msg) els['au-error-text'].textContent = msg;
  }

  function renderUsers() {
    els['au-loading'].style.display = 'none';
    els['au-error'].style.display = 'none';

    if (!state.users.length) {
      showEmpty();
      return;
    }

    els['au-empty'].style.display = 'none';
    els['au-table-wrap'].style.display = 'block';
    els['au-pagination'].style.display = state.totalPages > 1 ? 'flex' : 'none';

    var html = '';
    for (var i = 0; i < state.users.length; i++) {
      var u = state.users[i];
      var uid = u.id;
      var initials = getInitials(u.full_name || u.email);
      var isSelected = state.selected.indexOf(uid) !== -1;
      var roleClass = (u.role || 'employee').toLowerCase();
      var statusClass = u.is_active ? 'active' : 'inactive';
      var verifiedClass = u.is_verified ? 'verified' : 'unverified';
      var createdDate = u.created_at ? formatDate(u.created_at) : '-';

      html += '<tr' + (isSelected ? ' class="au-row-selected"' : '') + ' data-id="' + uid + '">';
      html += '<td><input type="checkbox" class="au-row-checkbox" data-id="' + uid + '"' + (isSelected ? ' checked' : '') + '></td>';
      html += '<td><div class="au-user-cell">';
      if (u.avatar_url) {
        html += '<img src="' + escapeHtml(u.avatar_url) + '" alt="" class="au-user-avatar">';
      } else {
        html += '<span class="au-user-avatar-text">' + escapeHtml(initials) + '</span>';
      }
      html += '<div>';
      html += '<a href="/admin/users/' + uid + '" class="au-user-name">' + escapeHtml(u.full_name || u.email) + '</a>';
      if (u.company_name) html += '<span class="au-user-email">' + escapeHtml(u.company_name) + '</span>';
      html += '</div></div></td>';
      html += '<td><span class="au-user-email">' + escapeHtml(u.email) + '</span></td>';
      html += '<td><span class="au-role-badge ' + roleClass + '">' + escapeHtml(roleClass) + '</span></td>';
      html += '<td><span class="au-status-badge ' + statusClass + '">' + (u.is_active ? 'Active' : 'Inactive') + '</span></td>';
      html += '<td><span class="au-verified-badge ' + verifiedClass + '">' + (u.is_verified ? 'Verified' : 'Unverified') + '</span></td>';
      html += '<td><span class="au-date-text">' + createdDate + '</span></td>';
      html += '<td><div class="au-actions-cell">';
      html += '<button class="au-action-btn au-action-view" data-user-action="view" data-user-id="' + uid + '" data-original-text="View">View</button>';
      if (u.is_active) {
        html += '<button class="au-action-btn au-action-suspend" data-user-action="suspend" data-user-id="' + uid + '" data-original-text="Suspend"' + (state.pendingActions[uid] ? ' disabled' : '') + '>' + (state.pendingActions[uid] ? 'Processing...' : 'Suspend') + '</button>';
      } else {
        html += '<button class="au-action-btn au-action-activate" data-user-action="activate" data-user-id="' + uid + '" data-original-text="Activate"' + (state.pendingActions[uid] ? ' disabled' : '') + '>' + (state.pendingActions[uid] ? 'Processing...' : 'Activate') + '</button>';
      }
      if (!u.is_verified) {
        html += '<button class="au-action-btn au-action-verify" data-user-action="verify" data-user-id="' + uid + '" data-original-text="Verify"' + (state.pendingActions[uid] ? ' disabled' : '') + '>' + (state.pendingActions[uid] ? 'Processing...' : 'Verify') + '</button>';
      }
      html += '<button class="au-action-btn au-action-delete" data-user-action="delete" data-user-id="' + uid + '" data-original-text="Delete"' + (state.pendingActions[uid] ? ' disabled' : '') + '>' + (state.pendingActions[uid] ? 'Processing...' : 'Delete') + '</button>';
      html += '</div></td></tr>';
    }

    els['au-tbody'].innerHTML = html;

    var checkboxes = els['au-tbody'].querySelectorAll('.au-row-checkbox');
    for (var j = 0; j < checkboxes.length; j++) {
      checkboxes[j].addEventListener('change', handleRowCheckboxChange);
    }

    renderPagination();
  }

  function handleRowCheckboxChange() {
    var id = this.getAttribute('data-id');
    var idx = state.selected.indexOf(id);
    if (this.checked) {
      if (idx === -1) state.selected.push(id);
    } else {
      if (idx !== -1) state.selected.splice(idx, 1);
    }
    els['au-select-all'].checked = state.selected.length === state.users.length;
    updateBulkBar();
  }

  function updateRowCheckboxes() {
    var checkboxes = els['au-tbody'].querySelectorAll('.au-row-checkbox');
    for (var i = 0; i < checkboxes.length; i++) {
      checkboxes[i].checked = state.selected.indexOf(checkboxes[i].getAttribute('data-id')) !== -1;
    }
    for (var i = 0; i < state.users.length; i++) {
      var row = els['au-tbody'].querySelector('tr[data-id="' + state.users[i].id + '"]');
      if (row) {
        row.className = state.selected.indexOf(state.users[i].id) !== -1 ? 'au-row-selected' : '';
      }
    }
  }

  function showEmpty() {
    els['au-loading'].style.display = 'none';
    els['au-error'].style.display = 'none';
    els['au-empty'].style.display = 'flex';
    els['au-table-wrap'].style.display = 'none';
    els['au-pagination'].style.display = 'none';
  }

  function renderPagination() {
    if (state.totalPages <= 1) {
      els['au-pagi-btns'].innerHTML = '';
      els['au-pagi-info'].textContent = 'Showing ' + state.users.length + ' of ' + state.total + ' users';
      return;
    }

    els['au-pagi-info'].textContent = 'Page ' + state.page + ' of ' + state.totalPages + ' (' + state.total + ' users)';

    var html = '';
    html += '<button class="au-pagi-btn" data-page="1"' + (state.page <= 1 ? ' disabled' : '') + '>&#xAB;</button>';
    html += '<button class="au-pagi-btn" data-page="' + (state.page - 1) + '"' + (state.page <= 1 ? ' disabled' : '') + '>&#x2039;</button>';

    var start = Math.max(1, state.page - 2);
    var end = Math.min(state.totalPages, state.page + 2);

    if (start > 1) {
      html += '<button class="au-pagi-btn" data-page="1">1</button>';
      if (start > 2) html += '<span class="au-pagi-btn" style="cursor:default;border:none;">...</span>';
    }

    for (var p = start; p <= end; p++) {
      html += '<button class="au-pagi-btn' + (p === state.page ? ' active' : '') + '" data-page="' + p + '">' + p + '</button>';
    }

    if (end < state.totalPages) {
      if (end < state.totalPages - 1) html += '<span class="au-pagi-btn" style="cursor:default;border:none;">...</span>';
      html += '<button class="au-pagi-btn" data-page="' + state.totalPages + '">' + state.totalPages + '</button>';
    }

    html += '<button class="au-pagi-btn" data-page="' + (state.page + 1) + '"' + (state.page >= state.totalPages ? ' disabled' : '') + '>&#x203A;</button>';
    html += '<button class="au-pagi-btn" data-page="' + state.totalPages + '"' + (state.page >= state.totalPages ? ' disabled' : '') + '>&#xBB;</button>';

    els['au-pagi-btns'].innerHTML = html;

    var pageBtns = els['au-pagi-btns'].querySelectorAll('.au-pagi-btn[data-page]');
    for (var i = 0; i < pageBtns.length; i++) {
      pageBtns[i].addEventListener('click', function () {
        var p = parseInt(this.getAttribute('data-page'), 10);
        if (p >= 1 && p <= state.totalPages && p !== state.page) {
          state.page = p;
          fetchUsers();
        }
      });
    }
  }

  function updateBulkBar() {
    var count = state.selected.length;
    els['au-bulk'].style.display = count > 0 ? 'flex' : 'none';
    els['au-bulk-count'].textContent = count > 0 ? count + ' selected' : '';
  }

  function showConfirmModal(title, html, btnClass, callback) {
    els['au-modal-title'].textContent = title;
    els['au-modal-msg'].innerHTML = html;
    els['au-modal-confirm'].className = 'au-btn ' + (btnClass === 'danger' ? 'au-btn-danger' : btnClass === 'success' ? 'au-btn-success' : 'au-btn-primary');
    els['au-modal'].style.display = 'flex';
    modalCallback = callback;
  }

  function hideModal() {
    els['au-modal'].style.display = 'none';
    modalCallback = null;
  }

  function hideViewModal() {
    els['au-view-modal'].style.display = 'none';
  }

  function viewUser(id) {
    els['au-view-modal'].style.display = 'flex';
    els['au-view-loading'].style.display = 'flex';
    els['au-view-body'].innerHTML = '<div class="au-loading" style="display:flex"><div class="au-spinner"></div><p class="au-loading-text">Loading user details...</p></div>';

    AngaziaAPI.admin.userDetail(id)
      .then(function (data) {
        els['au-view-loading'].style.display = 'none';
        if (!data) {
          els['au-view-body'].innerHTML = '<p style="color:var(--danger)">Failed to load user details.</p>';
          return;
        }
        var u = data;
        var initials = getInitials(u.full_name || u.email);
        var createdDate = u.created_at ? formatDateTime(u.created_at) : '-';
        var lastLogin = u.last_login_at ? formatDateTime(u.last_login_at) : 'Never';

        var avatarSrc = u.avatar_url || u.company_logo || '';

        var html = '<div class="au-view-detail">';
        html += '<div style="display:flex;align-items:center;gap:12px;margin-bottom:16px;">';
        if (avatarSrc) {
          html += '<img src="' + escapeHtml(avatarSrc) + '" alt="" style="width:40px;height:40px;border-radius:50%;object-fit:cover;">';
        } else {
          html += '<span class="au-user-avatar-text" style="width:40px;height:40px;font-size:16px;">' + escapeHtml(initials) + '</span>';
        }
        html += '<div><div style="font-weight:600;font-size:14px;color:var(--text)">' + escapeHtml(u.full_name || u.email) + '</div>';
        html += '<div style="font-size:10px;color:var(--muted)">' + escapeHtml(u.email) + '</div></div></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Role</span><span class="au-view-value"><span class="au-role-badge ' + (u.role || '').toLowerCase() + '">' + escapeHtml(u.role || '-') + '</span></span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Status</span><span class="au-view-value"><span class="au-status-badge ' + (u.is_active ? 'active' : 'inactive') + '">' + (u.is_active ? 'Active' : 'Inactive') + '</span></span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Verified</span><span class="au-view-value"><span class="au-verified-badge ' + (u.is_verified ? 'verified' : 'unverified') + '">' + (u.is_verified ? 'Yes' : 'No') + '</span></span></div>';
        if (u.verification_status) html += '<div class="au-view-row"><span class="au-view-label">Verification</span><span class="au-view-value">' + escapeHtml(u.verification_status) + '</span></div>';
        if (u.company_name) html += '<div class="au-view-row"><span class="au-view-label">Company</span><span class="au-view-value">' + escapeHtml(u.company_name) + '</span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Created</span><span class="au-view-value">' + createdDate + '</span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Last Login</span><span class="au-view-value">' + lastLogin + '</span></div>';
        if (u.job_count !== undefined) html += '<div class="au-view-row"><span class="au-view-label">Jobs Posted</span><span class="au-view-value">' + u.job_count + '</span></div>';
        if (u.application_count !== undefined) html += '<div class="au-view-row"><span class="au-view-label">Applications</span><span class="au-view-value">' + u.application_count + '</span></div>';
        if (u.reports_count !== undefined && u.reports_count > 0) html += '<div class="au-view-row"><span class="au-view-label">Reports</span><span class="au-view-value" style="color:var(--danger)">' + u.reports_count + '</span></div>';
        if (u.document_count !== undefined && u.document_count > 0) html += '<div class="au-view-row"><span class="au-view-label">Documents</span><span class="au-view-value">' + u.document_count + '</span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">ID</span><span class="au-view-value" style="font-family:monospace;font-size:10px;">' + escapeHtml(u.id) + '</span></div>';
        html += '</div>';
        els['au-view-body'].innerHTML = html;
      })
      .catch(function (err) {
        els['au-view-loading'].style.display = 'none';
        els['au-view-body'].innerHTML = '<p style="color:var(--danger)">' + escapeHtml(err.message || 'Network error') + '</p>';
      });
  }

  function bulkAction(action) {
    var ids = state.selected.slice();
    var promises = ids.map(function (id) {
      switch (action) {
        case 'suspend': return AngaziaAPI.admin.suspendUser(id);
        case 'activate': return AngaziaAPI.admin.activateUser(id);
        case 'delete': return AngaziaAPI.admin.deleteUser(id);
        default: return Promise.reject(new Error('Unknown action'));
      }
    });

    var actionLabel = action.charAt(0).toUpperCase() + action.slice(1);
    Promise.all(promises)
      .then(function () {
        showToast(actionLabel + 'd ' + ids.length + ' user(s)', 'success');
        state.selected = [];
        els['au-select-all'].checked = false;
        updateBulkBar();
        fetchUsers();
      })
      .catch(function () {
        showToast('Failed to ' + action + ' some users', 'error');
      });
  }

  function fetchChart() {
    AngaziaAPI.admin.chartData({ period: state.chartPeriod })
      .then(function (data) {
        var points = data && data.user_growth ? data.user_growth : [];
        renderChart(points);
      })
      .catch(function () {});
  }

  function renderChart(points) {
    if (chartInstance) {
      chartInstance.destroy();
      chartInstance = null;
    }

    var canvas = els['au-growth-chart'];
    if (!canvas) return;
    if (!points || points.length === 0) {
      return;
    }

    var labels = points.map(function (p) { return p.date; });
    var values = points.map(function (p) { return p.count; });

    chartInstance = new Chart(canvas.getContext('2d'), {
      type: 'line',
      data: {
        labels: labels,
        datasets: [{
          label: 'New Users',
          data: values,
          borderColor: '#1a8fff',
          backgroundColor: 'rgba(26,143,255,0.08)',
          borderWidth: 2,
          fill: true,
          tension: 0.3,
          pointRadius: 2,
          pointBackgroundColor: '#1a8fff',
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            backgroundColor: '#1a1a2e',
            titleFont: { size: 10 },
            bodyFont: { size: 11 },
            padding: 8,
            cornerRadius: 6,
          }
        },
        scales: {
          x: {
            ticks: { font: { size: 9 }, color: '#6b7280', maxTicksLimit: 10 },
            grid: { display: false },
          },
          y: {
            beginAtZero: true,
            ticks: { font: { size: 9 }, color: '#6b7280', precision: 0 },
            grid: { color: 'rgba(255,255,255,0.04)' },
          }
        }
      }
    });
  }

  function fetchStats() {
    AngaziaAPI.admin.platformStats()
      .then(function (data) {
        if (!data) return;
        if (els['au-stat-total']) els['au-stat-total'].textContent = data.total_users || 0;
        if (els['au-stat-active']) els['au-stat-active'].textContent = data.active_users_30_days || 0;
        if (els['au-stat-new']) els['au-stat-new'].textContent = data.new_users_30_days || 0;
        if (els['au-stat-suspended']) els['au-stat-suspended'].textContent = data.suspended_users || 0;
        if (els['au-stat-unverified']) els['au-stat-unverified'].textContent = data.unverified_users || 0;
      })
      .catch(function () {});
  }

  function getInitials(str) {
    if (!str) return '?';
    var parts = str.split(/\s+/);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return str.substring(0, 2).toUpperCase();
  }

  function formatDate(str) {
    if (!str) return '-';
    var d = new Date(str);
    if (isNaN(d.getTime())) return str;
    return (d.getMonth() + 1).toString().padStart(2, '0') + '/' +
      d.getDate().toString().padStart(2, '0') + '/' + d.getFullYear();
  }

  function formatDateTime(str) {
    if (!str) return '-';
    var d = new Date(str);
    if (isNaN(d.getTime())) return str;
    return (d.getMonth() + 1).toString().padStart(2, '0') + '/' +
      d.getDate().toString().padStart(2, '0') + '/' + d.getFullYear() + ' ' +
      d.getHours().toString().padStart(2, '0') + ':' + d.getMinutes().toString().padStart(2, '0');
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

  window.auLoadUsers = fetchUsers;
  window.auReload = function () { state.page = 1; fetchUsers(); };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();