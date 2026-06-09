(function () {
  'use strict';
  function showToast(msg, type) {    if (window.AngaziaApp && AngaziaApp.showToast) {      AngaziaApp.showToast(msg, type);    } else {      alert((type === 'error' ? 'Error: ' : '') + msg);    }  }

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
    loading: false
  };

  var els = {};
  var modalCallback = null;

  function init() {
    cacheElements();
    bindEvents();
    fetchUsers();
  }

  function cacheElements() {
    els.search = document.getElementById('au-search');
    els.filterRole = document.getElementById('au-filter-role');
    els.filterStatus = document.getElementById('au-filter-status');
    els.filterVerified = document.getElementById('au-filter-verified');
    els.filterBtn = document.getElementById('au-filter-btn');
    els.tbody = document.getElementById('au-tbody');
    els.loading = document.getElementById('au-loading');
    els.error = document.getElementById('au-error');
    els.errorText = document.getElementById('au-error-text');
    els.empty = document.getElementById('au-empty');
    els.emptyHint = document.getElementById('au-empty-hint');
    els.tableWrap = document.getElementById('au-table-wrap');
    els.pagination = document.getElementById('au-pagination');
    els.pagiInfo = document.getElementById('au-pagi-info');
    els.pagiBtns = document.getElementById('au-pagi-btns');
    els.bulk = document.getElementById('au-bulk');
    els.bulkCount = document.getElementById('au-bulk-count');
    els.selectAll = document.getElementById('au-select-all');
    els.modal = document.getElementById('au-modal');
    els.modalTitle = document.getElementById('au-modal-title');
    els.modalMsg = document.getElementById('au-modal-msg');
    els.modalConfirm = document.getElementById('au-modal-confirm');
    els.modalCancel = document.getElementById('au-modal-cancel');
    els.modalClose = document.getElementById('au-modal-close');
    els.viewModal = document.getElementById('au-view-modal');
    els.viewBody = document.getElementById('au-view-body');
    els.viewLoading = document.getElementById('au-view-loading');
  }

  function bindEvents() {
    els.filterBtn.addEventListener('click', function () {
      state.page = 1;
      fetchUsers();
    });

    els.search.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') {
        state.page = 1;
        fetchUsers();
      }
    });

    els.selectAll.addEventListener('change', function () {
      var checked = els.selectAll.checked;
      state.selected = [];
      if (checked) {
        state.selected = state.users.map(function (u) { return u.id; });
      }
      updateBulkBar();
      renderUsers();
    });

    document.getElementById('au-bulk-suspend').addEventListener('click', function () {
      if (!state.selected.length) return;
      showModal('Suspend Users', 'Are you sure you want to suspend ' + state.selected.length + ' user(s)?', function () {
        bulkAction('suspend');
      });
    });

    document.getElementById('au-bulk-activate').addEventListener('click', function () {
      if (!state.selected.length) return;
      showModal('Activate Users', 'Are you sure you want to activate ' + state.selected.length + ' user(s)?', function () {
        bulkAction('activate');
      });
    });

    document.getElementById('au-bulk-delete').addEventListener('click', function () {
      if (!state.selected.length) return;
      showModal('Delete Users', 'Are you sure you want to delete ' + state.selected.length + ' user(s)? This action cannot be undone.', function () {
        bulkAction('delete');
      });
    });

    document.getElementById('au-bulk-clear').addEventListener('click', function () {
      state.selected = [];
      els.selectAll.checked = false;
      updateBulkBar();
      renderUsers();
    });

    els.modalConfirm.addEventListener('click', function () {
      hideModal();
      if (typeof modalCallback === 'function') {
        var cb = modalCallback;
        modalCallback = null;
        cb();
      }
    });

    els.modalCancel.addEventListener('click', hideModal);
    els.modalClose.addEventListener('click', hideModal);

    document.getElementById('au-view-close').addEventListener('click', function () { hideModal(els.viewModal); });
    document.getElementById('au-view-close-btn').addEventListener('click', function () { hideModal(els.viewModal); });

    window.addEventListener('click', function (e) {
      if (e.target === els.modal) hideModal();
      if (e.target === els.viewModal) hideModal(els.viewModal);
    });
  }

  function fetchUsers() {
    if (state.loading) return;
    state.loading = true;

    state.search = els.search.value.trim();
    state.role = els.filterRole.value;
    state.status = els.filterStatus.value;
    state.verified = els.filterVerified.value;

    showLoading();

    var params = [];
    params.push('page=' + state.page);
    params.push('limit=' + state.limit);
    if (state.search) params.push('search=' + encodeURIComponent(state.search));
    if (state.role) params.push('role=' + encodeURIComponent(state.role));
    if (state.status) params.push('is_active=' + (state.status === 'active' ? 'true' : 'false'));
    if (state.verified) params.push('is_verified=' + (state.verified === 'verified' ? 'true' : 'false'));

    AngaziaAPI.get('/admin/users?' + params.join('&'))
      .then(function (data) {
        state.loading = false;
        state.users = data.users || [];
        state.total = data.total || 0;
        state.page = data.page || 1;
        state.limit = data.limit || 20;
        state.totalPages = data.total_pages || Math.ceil(state.total / state.limit) || 0;
        state.selected = [];
        els.selectAll.checked = false;
        updateBulkBar();
        renderUsers();
      })
      .catch(function (err) {
        state.loading = false;
        showError(err.message || 'Network error');
      });
  }

  function showLoading() {
    els.loading.style.display = 'flex';
    els.error.style.display = 'none';
    els.empty.style.display = 'none';
    els.tableWrap.style.display = 'none';
    els.pagination.style.display = 'none';
  }

  function showError(msg) {
    els.loading.style.display = 'none';
    els.error.style.display = 'flex';
    els.errorText.textContent = msg || 'Failed to load users.';
    els.empty.style.display = 'none';
    els.tableWrap.style.display = 'none';
    els.pagination.style.display = 'none';
  }

  function showEmpty(hint) {
    els.loading.style.display = 'none';
    els.error.style.display = 'none';
    els.empty.style.display = 'flex';
    els.emptyHint.textContent = hint || 'No users found matching your criteria.';
    els.tableWrap.style.display = 'none';
    els.pagination.style.display = 'none';
  }

  function renderUsers() {
    els.loading.style.display = 'none';
    els.error.style.display = 'none';

    if (!state.users.length) {
      showEmpty('Try adjusting your filters or search terms.');
      return;
    }

    els.empty.style.display = 'none';
    els.tableWrap.style.display = 'block';
    els.pagination.style.display = 'flex';

    var html = '';
    for (var i = 0; i < state.users.length; i++) {
      var u = state.users[i];
      var uid = u.id;
      var initials = (u.full_name || u.email || '?').charAt(0).toUpperCase();
      var isSelected = state.selected.indexOf(uid) !== -1;
      var roleClass = (u.role || 'employee').toLowerCase();
      var statusClass = u.is_active ? 'active' : 'inactive';
      var verifiedClass = u.is_verified ? 'verified' : 'unverified';
      var createdDate = u.created_at ? formatDate(u.created_at) : '-';

      html += '<tr' + (isSelected ? ' class="au-row-selected"' : '') + ' data-id="' + uid + '">';
      html += '<td><input type="checkbox" class="au-row-checkbox" data-id="' + uid + '"' + (isSelected ? ' checked' : '') + '></td>';
      html += '<td><div class="au-user-cell">';
      html += '<span class="au-user-avatar-text">' + escapeHtml(initials) + '</span>';
      html += '<div>';
      html += '<a href="/admin/users/' + uid + '" class="au-user-name">' + escapeHtml(u.full_name || u.email) + '</a>';
      if (u.company_name) {
        html += '<span class="au-user-email">' + escapeHtml(u.company_name) + '</span>';
      }
      html += '</div></div></td>';
      html += '<td><span class="au-user-email">' + escapeHtml(u.email) + '</span></td>';
      html += '<td><span class="au-role-badge ' + roleClass + '">' + escapeHtml(roleClass) + '</span></td>';
      html += '<td><span class="au-status-badge ' + statusClass + '">' + (u.is_active ? 'Active' : 'Inactive') + '</span></td>';
      html += '<td><span class="au-verified-badge ' + verifiedClass + '">' + (u.is_verified ? 'Verified' : 'Unverified') + '</span></td>';
      html += '<td><span class="au-date-text">' + createdDate + '</span></td>';
      html += '<td><div class="au-actions-cell">';
      html += '<button class="au-action-btn au-action-view" onclick="auViewUser(\'' + uid + '\')">View</button>';
      if (u.is_active) {
        html += '<button class="au-action-btn au-action-suspend" onclick="auSuspendUser(\'' + uid + '\')">Suspend</button>';
      } else {
        html += '<button class="au-action-btn au-action-activate" onclick="auActivateUser(\'' + uid + '\')">Activate</button>';
      }
      if (!u.is_verified) {
        html += '<button class="au-action-btn au-action-verify" onclick="auVerifyUser(\'' + uid + '\')">Verify</button>';
      }
      html += '<button class="au-action-btn au-action-delete" onclick="auDeleteUser(\'' + uid + '\')">Delete</button>';
      html += '</div></td></tr>';
    }

    els.tbody.innerHTML = html;

    var checkboxes = els.tbody.querySelectorAll('.au-row-checkbox');
    for (var j = 0; j < checkboxes.length; j++) {
      checkboxes[j].addEventListener('change', function () {
        var id = this.getAttribute('data-id');
        var idx = state.selected.indexOf(id);
        if (this.checked) {
          if (idx === -1) state.selected.push(id);
        } else {
          if (idx !== -1) state.selected.splice(idx, 1);
        }
        els.selectAll.checked = state.selected.length === state.users.length;
        updateBulkBar();
      });
    }

    renderPagination();
  }

  function renderPagination() {
    if (state.totalPages <= 1) {
      els.pagiBtns.innerHTML = '';
      els.pagiInfo.textContent = 'Showing ' + state.users.length + ' of ' + state.total + ' users';
      return;
    }

    els.pagiInfo.textContent = 'Page ' + state.page + ' of ' + state.totalPages + ' (' + state.total + ' users)';

    var html = '';
    html += '<button class="au-pagi-btn" onclick="auGoPage(1)"' + (state.page <= 1 ? ' disabled' : '') + '>&#xAB;</button>';
    html += '<button class="au-pagi-btn" onclick="auGoPage(' + (state.page - 1) + ')"' + (state.page <= 1 ? ' disabled' : '') + '>&#x2039;</button>';

    var start = Math.max(1, state.page - 2);
    var end = Math.min(state.totalPages, state.page + 2);

    if (start > 1) {
      html += '<button class="au-pagi-btn" onclick="auGoPage(1)">1</button>';
      if (start > 2) html += '<span class="au-pagi-btn" style="cursor:default;border:none;">...</span>';
    }

    for (var p = start; p <= end; p++) {
      html += '<button class="au-pagi-btn' + (p === state.page ? ' active' : '') + '" onclick="auGoPage(' + p + ')">' + p + '</button>';
    }

    if (end < state.totalPages) {
      if (end < state.totalPages - 1) html += '<span class="au-pagi-btn" style="cursor:default;border:none;">...</span>';
      html += '<button class="au-pagi-btn" onclick="auGoPage(' + state.totalPages + ')">' + state.totalPages + '</button>';
    }

    html += '<button class="au-pagi-btn" onclick="auGoPage(' + (state.page + 1) + ')"' + (state.page >= state.totalPages ? ' disabled' : '') + '>&#x203A;</button>';
    html += '<button class="au-pagi-btn" onclick="auGoPage(' + state.totalPages + ')"' + (state.page >= state.totalPages ? ' disabled' : '') + '>&#xBB;</button>';

    els.pagiBtns.innerHTML = html;
  }

  function updateBulkBar() {
    var count = state.selected.length;
    if (count > 0) {
      els.bulk.style.display = 'flex';
      els.bulkCount.textContent = count + ' selected';
    } else {
      els.bulk.style.display = 'none';
    }
  }

  function showModal(title, msg, callback) {
    els.modalTitle.textContent = title;
    els.modalMsg.textContent = msg;
    els.modal.style.display = 'flex';
    modalCallback = callback;
  }

  function hideModal(modalEl) {
    var m = modalEl || els.modal;
    m.style.display = 'none';
    modalCallback = null;
  }

  function findActionBtn(id, action) {
    var tr = els.tbody.querySelector('tr[data-id="' + id + '"]');
    if (!tr) return null;
    var btns = tr.querySelectorAll('.au-action-btn');
    for (var i = 0; i < btns.length; i++) {
      if (btns[i].textContent.trim().toLowerCase().indexOf(action) !== -1) return btns[i];
    }
    return null;
  }

  function auReload() {
    state.page = 1;
    fetchUsers();
  }

  function auGoPage(p) {
    if (p < 1 || p > state.totalPages || p === state.page) return;
    state.page = p;
    fetchUsers();
  }

  function auViewUser(id) {
    els.viewModal.style.display = 'flex';
    els.viewLoading.style.display = 'flex';
    els.viewBody.innerHTML = '<div class="au-loading" style="display:flex"><div class="au-spinner"></div><p class="au-loading-text">Loading user details...</p></div>';

    AngaziaAPI.get('/admin/users/' + id)
      .then(function (data) {
        els.viewLoading.style.display = 'none';
        if (!data) {
          els.viewBody.innerHTML = '<p style="color:var(--danger)">Failed to load user details.</p>';
          return;
        }
        var u = data;
        var initials = (u.full_name || u.email || '?').charAt(0).toUpperCase();
        var createdDate = u.created_at ? formatDateTime(u.created_at) : '-';
        var lastLogin = u.last_login_at ? formatDateTime(u.last_login_at) : 'Never';

        var html = '<div class="au-view-detail">';
        html += '<div style="display:flex;align-items:center;gap:12px;margin-bottom:8px;">';
        html += '<span class="au-user-avatar-text" style="width:40px;height:40px;font-size:16px;">' + escapeHtml(initials) + '</span>';
        html += '<div><div style="font-weight:600;font-size:14px;color:var(--text)">' + escapeHtml(u.full_name || u.email) + '</div>';
        html += '<div style="font-size:10px;color:var(--muted)">' + escapeHtml(u.email) + '</div></div></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Role</span><span class="au-view-value"><span class="au-role-badge ' + (u.role || '').toLowerCase() + '">' + escapeHtml(u.role || '-') + '</span></span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Status</span><span class="au-view-value"><span class="au-status-badge ' + (u.is_active ? 'active' : 'inactive') + '">' + (u.is_active ? 'Active' : 'Inactive') + '</span></span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Verified</span><span class="au-view-value"><span class="au-verified-badge ' + (u.is_verified ? 'verified' : 'unverified') + '">' + (u.is_verified ? 'Yes' : 'No') + '</span></span></div>';
        if (u.company_name) html += '<div class="au-view-row"><span class="au-view-label">Company</span><span class="au-view-value">' + escapeHtml(u.company_name) + '</span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Created</span><span class="au-view-value">' + createdDate + '</span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">Last Login</span><span class="au-view-value">' + lastLogin + '</span></div>';
        if (u.job_count !== undefined) html += '<div class="au-view-row"><span class="au-view-label">Jobs</span><span class="au-view-value">' + u.job_count + '</span></div>';
        if (u.application_count !== undefined) html += '<div class="au-view-row"><span class="au-view-label">Applications</span><span class="au-view-value">' + u.application_count + '</span></div>';
        html += '<div class="au-view-row"><span class="au-view-label">ID</span><span class="au-view-value" style="font-family:monospace;font-size:10px;">' + escapeHtml(u.id) + '</span></div>';
        html += '</div>';
        els.viewBody.innerHTML = html;
      })
      .catch(function (err) {
        els.viewLoading.style.display = 'none';
        els.viewBody.innerHTML = '<p style="color:var(--danger)">' + escapeHtml(err.message || 'Network error') + '</p>';
      });
  }

  function auSuspendUser(id) {
    showModal('Suspend User', 'Are you sure you want to suspend this user? They will be unable to access the platform.', function () {
      var btn = findActionBtn(id, 'suspend');
      if (btn) btn.disabled = true;
      AngaziaAPI.post('/admin/users/' + id + '/suspend')
        .then(function () {
          showToast('User suspended successfully', 'success');
          fetchUsers();
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
          if (btn) btn.disabled = false;
        });
    });
  }

  function auActivateUser(id) {
    showModal('Activate User', 'Are you sure you want to activate this user?', function () {
      var btn = findActionBtn(id, 'activate');
      if (btn) btn.disabled = true;
      AngaziaAPI.post('/admin/users/' + id + '/activate')
        .then(function () {
          showToast('User activated successfully', 'success');
          fetchUsers();
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
          if (btn) btn.disabled = false;
        });
    });
  }

  function auVerifyUser(id) {
    showModal('Verify User', 'Are you sure you want to verify this user?', function () {
      var btn = findActionBtn(id, 'verify');
      if (btn) btn.disabled = true;
      AngaziaAPI.post('/admin/users/' + id + '/verify')
        .then(function () {
          showToast('User verified successfully', 'success');
          fetchUsers();
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
          if (btn) btn.disabled = false;
        });
    });
  }

  function auDeleteUser(id) {
    showModal('Delete User', 'Are you sure you want to delete this user? This action cannot be undone.', function () {
      var btn = findActionBtn(id, 'delete');
      if (btn) btn.disabled = true;
      AngaziaAPI.del('/admin/users/' + id)
        .then(function () {
          showToast('User deleted successfully', 'success');
          fetchUsers();
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
          if (btn) btn.disabled = false;
        });
    });
  }

  function bulkAction(action) {
    var ids = state.selected.slice();
    var promises = ids.map(function (id) {
      if (action === 'suspend') return AngaziaAPI.post('/admin/users/' + id + '/suspend');
      if (action === 'activate') return AngaziaAPI.post('/admin/users/' + id + '/activate');
      if (action === 'delete') return AngaziaAPI.del('/admin/users/' + id);
    });

    var actionLabel = action.charAt(0).toUpperCase() + action.slice(1);
    Promise.all(promises)
      .then(function () {
        showToast(actionLabel + 'd ' + ids.length + ' user(s)', 'success');
        state.selected = [];
        updateBulkBar();
        fetchUsers();
      })
      .catch(function () {
        showToast('Failed to ' + action + ' some users', 'error');
      });
  }

  function formatDate(str) {
    if (!str) return '-';
    var d = new Date(str);
    if (isNaN(d.getTime())) return str;
    return (d.getMonth() + 1).toString().padStart(2, '0') + '/' +
      d.getDate().toString().padStart(2, '0') + '/' +
      d.getFullYear();
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

  window.auLoadUsers = fetchUsers;
  window.auReload = auReload;
  window.auGoPage = auGoPage;
  window.auViewUser = auViewUser;
  window.auSuspendUser = auSuspendUser;
  window.auActivateUser = auActivateUser;
  window.auVerifyUser = auVerifyUser;
  window.auDeleteUser = auDeleteUser;

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
