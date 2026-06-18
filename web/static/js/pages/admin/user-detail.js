(function () {
  'use strict';

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  var userId = window.location.pathname.split('/').pop();
  var userData = null;
  var modalCallback = null;

  var els = {};

  function init() {
    els.loading = document.getElementById('aud-loading');
    els.error = document.getElementById('aud-error');
    els.errorText = document.getElementById('aud-error-text');
    els.content = document.getElementById('aud-content');
    els.title = document.getElementById('aud-title');
    els.avatar = document.getElementById('aud-avatar');
    els.name = document.getElementById('aud-name');
    els.email = document.getElementById('aud-email');
    els.role = document.getElementById('aud-role');
    els.status = document.getElementById('aud-status');
    els.verified = document.getElementById('aud-verified');
    els.profileActions = document.getElementById('aud-profile-actions');
    els.quickActions = document.getElementById('aud-quick-actions');
    els.infoId = document.getElementById('aud-info-id');
    els.infoName = document.getElementById('aud-info-name');
    els.infoEmail = document.getElementById('aud-info-email');
    els.infoRole = document.getElementById('aud-info-role');
    els.infoCompany = document.getElementById('aud-info-company');
    els.infoStatus = document.getElementById('aud-info-status');
    els.infoVerified = document.getElementById('aud-info-verified');
    els.infoCreated = document.getElementById('aud-info-created');
    els.infoLastLogin = document.getElementById('aud-info-lastlogin');
    els.statJobs = document.getElementById('aud-stat-jobs');
    els.statApps = document.getElementById('aud-stat-apps');
    els.statReports = document.getElementById('aud-stat-reports');
    els.timeline = document.getElementById('aud-timeline');
    els.activityLoading = document.getElementById('aud-activity-loading');
    els.activityEmpty = document.getElementById('aud-activity-empty');

    els.modal = document.getElementById('aud-modal');
    els.modalTitle = document.getElementById('aud-modal-title');
    els.modalMsg = document.getElementById('aud-modal-msg');
    els.modalConfirm = document.getElementById('aud-modal-confirm');
    els.modalCancel = document.getElementById('aud-modal-cancel');
    els.modalClose = document.getElementById('aud-modal-close');

    els.modalConfirm.addEventListener('click', function () {
      if (typeof modalCallback === 'function') {
        var cb = modalCallback;
        modalCallback = null;
        cb();
      }
      hideModal();
    });
    els.modalCancel.addEventListener('click', hideModal);
    els.modalClose.addEventListener('click', hideModal);
    window.addEventListener('click', function (e) {
      if (e.target === els.modal) hideModal();
    });

    audLoad();
  }

  function audLoad() {
    showLoading();
    AngaziaAPI.admin.userDetail(userId)
      .then(function (data) {
        if (!data) {
          showError('Failed to load user details');
          return;
        }
        userData = data;
        renderUser();
      })
      .catch(function (err) {
        showError(err.message || 'Network error');
      });
  }

  function audReload() {
    audLoad();
  }

  function showLoading() {
    els.loading.style.display = 'flex';
    els.error.style.display = 'none';
    els.content.style.display = 'none';
  }

  function showError(msg) {
    els.loading.style.display = 'none';
    els.error.style.display = 'flex';
    els.errorText.textContent = msg;
    els.content.style.display = 'none';
  }

  function renderUser() {
    var u = userData;
    if (!u) return;

    els.loading.style.display = 'none';
    els.error.style.display = 'none';
    els.content.style.display = 'block';

    var initials = (u.full_name || u.email || '?').charAt(0).toUpperCase();
    var roleClass = (u.role || 'employee').toLowerCase();
    var statusClass = u.is_active ? 'active' : 'inactive';
    var verifiedClass = u.is_verified ? 'verified' : 'unverified';
    var createdDate = u.created_at ? formatDateTime(u.created_at) : '-';
    var lastLogin = u.last_login_at ? formatDateTime(u.last_login_at) : 'Never';

    els.title.textContent = u.full_name || u.email || 'User Details';
    if (u.avatar_url) {
      els.avatar.innerHTML = '<img src="' + escapeHtml(u.avatar_url) + '" alt="" class="aud-avatar-img">';
    } else {
      els.avatar.textContent = initials;
    }
    els.name.textContent = u.full_name || '-';
    els.email.textContent = u.email || '-';
    els.role.textContent = u.role || '-';
    els.role.className = 'aud-role-badge ' + roleClass;
    els.status.textContent = u.is_active ? 'Active' : 'Inactive';
    els.status.className = 'aud-status-badge ' + statusClass;
    els.verified.textContent = u.is_verified ? 'Verified' : 'Unverified';
    els.verified.className = 'aud-verified-badge ' + verifiedClass;

    els.infoId.textContent = u.id || '-';
    els.infoName.textContent = u.full_name || '-';
    els.infoEmail.textContent = u.email || '-';
    els.infoRole.textContent = u.role || '-';
    els.infoCompany.textContent = u.company_name || '-';
    els.infoStatus.innerHTML = '<span class="aud-status-badge ' + statusClass + '">' + (u.is_active ? 'Active' : 'Inactive') + '</span>';
    els.infoVerified.innerHTML = '<span class="aud-verified-badge ' + verifiedClass + '">' + (u.is_verified ? 'Yes' : 'No') + '</span>';
    els.infoCreated.textContent = createdDate;
    els.infoLastLogin.textContent = lastLogin;

    els.statJobs.textContent = u.job_count !== undefined ? u.job_count : '-';
    els.statApps.textContent = u.application_count !== undefined ? u.application_count : '-';
    els.statReports.textContent = u.reports_count !== undefined ? u.reports_count : '-';

    renderActions(u);
    renderQuickActions(u);
    renderActivity(u);
  }

  function renderActions(u) {
    var html = '';
    if (u.is_active) {
      html += '<button class="aud-btn aud-btn-warn aud-btn-sm" data-action="audSuspend">&#x1F6AB; Suspend</button>';
    } else {
      html += '<button class="aud-btn aud-btn-success aud-btn-sm" data-action="audActivate">&#x2705; Activate</button>';
    }
    if (!u.is_verified) {
      html += '<button class="aud-btn aud-btn-success aud-btn-sm" data-action="audVerify">&#x2705; Verify</button>';
    }
    html += '<button class="aud-btn aud-btn-danger aud-btn-sm" data-action="audDelete">&#x1F5D1; Delete</button>';
    els.profileActions.innerHTML = html;
  }

  function renderQuickActions(u) {
    var html = '';
    html += '<div class="aud-action-block"><div><div class="aud-action-label">View Jobs</div><div class="aud-action-desc">View jobs posted by this user</div></div><button class="aud-btn aud-btn-ghost aud-btn-sm" data-action="audViewJobs">Go</button></div>';
    html += '<div class="aud-action-block"><div><div class="aud-action-label">View Applications</div><div class="aud-action-desc">View applications submitted by this user</div></div><button class="aud-btn aud-btn-ghost aud-btn-sm" data-action="audViewApps">Go</button></div>';
    html += '<div class="aud-action-block"><div><div class="aud-action-label">View Reports</div><div class="aud-action-desc">View reports involving this user</div></div><button class="aud-btn aud-btn-ghost aud-btn-sm" data-action="audViewReports">Go</button></div>';
    if (u.role === 'employer' && u.company_id) {
      html += '<div class="aud-action-block"><div><div class="aud-action-label">View Company</div><div class="aud-action-desc">View company profile</div></div><button class="aud-btn aud-btn-ghost aud-btn-sm" data-action="audViewCompany" data-id="' + u.company_id + '">Go</button></div>';
    }
    els.quickActions.innerHTML = html;
  }

  function renderActivity(u) {
    els.activityLoading.style.display = 'none';
    els.timeline.style.display = 'none';
    els.activityEmpty.style.display = 'none';

    var activities = [];

    if (u.created_at) {
      activities.push({ action: 'Account Created', detail: 'User registered on the platform', time: u.created_at });
    }
    if (u.last_login_at) {
      activities.push({ action: 'Last Login', detail: 'User logged into the platform', time: u.last_login_at });
    }
    if (u.is_verified) {
      activities.push({ action: 'Account Verified', detail: 'User email/identity was verified', time: null });
    }
    if (u.is_active === false) {
      activities.push({ action: 'Account Suspended', detail: 'User account is currently suspended', time: null });
    } else {
      activities.push({ action: 'Account Active', detail: 'User account is active', time: null });
    }

    if (activities.length === 0) {
      els.activityEmpty.style.display = 'block';
      return;
    }

    var html = '';
    for (var i = 0; i < activities.length; i++) {
      var a = activities[i];
      html += '<div class="aud-timeline-item">';
      html += '<div class="aud-timeline-dot"></div>';
      html += '<div class="aud-timeline-body">';
      html += '<div class="aud-timeline-action">' + escapeHtml(a.action) + '</div>';
      if (a.detail) html += '<div class="aud-timeline-detail">' + escapeHtml(a.detail) + '</div>';
      html += '</div>';
      if (a.time) html += '<div class="aud-timeline-time">' + formatDateTime(a.time) + '</div>';
      html += '</div>';
    }
    els.timeline.innerHTML = html;
    els.timeline.style.display = 'flex';
  }

  function showModal(title, msg, callback) {
    els.modalTitle.textContent = title;
    els.modalMsg.textContent = msg;
    els.modal.style.display = 'flex';
    modalCallback = callback;
  }

  function hideModal() {
    els.modal.style.display = 'none';
    modalCallback = null;
  }

  function audSuspend() {
    showModal('Suspend User', 'Are you sure you want to suspend this user? They will be unable to access the platform.', function () {
      AngaziaAPI.admin.suspendUser(userId)
        .then(function () {
          showToast('User suspended successfully', 'success');
          audLoad();
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
        });
    });
  }

  function audActivate() {
    showModal('Activate User', 'Are you sure you want to activate this user?', function () {
      AngaziaAPI.admin.activateUser(userId)
        .then(function () {
          showToast('User activated successfully', 'success');
          audLoad();
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
        });
    });
  }

  function audVerify() {
    showModal('Verify User', 'Are you sure you want to verify this user?', function () {
      AngaziaAPI.admin.verifyUser(userId)
        .then(function () {
          showToast('User verified successfully', 'success');
          audLoad();
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
        });
    });
  }

  function audDelete() {
    showModal('Delete User', 'Are you sure you want to delete this user? This action cannot be undone.', function () {
      AngaziaAPI.admin.deleteUser(userId)
        .then(function () {
          showToast('User deleted successfully', 'success');
          window.location.href = '/admin/users';
        })
        .catch(function (err) {
          showToast(err.message || 'Network error', 'error');
        });
    });
  }

  function audViewJobs() {
    window.location.href = '/admin/jobs?user_id=' + userId;
  }

  function audViewApps() {
    window.location.href = '/admin/jobs?applicant_id=' + userId;
  }

  function audViewReports() {
    window.location.href = '/admin/reports?user_id=' + userId;
  }

  function audViewCompany(companyId) {
    window.location.href = '/admin/companies/' + companyId;
  }

  function formatDateTime(str) {
    if (!str) return '-';
    var d = new Date(str);
    if (isNaN(d.getTime())) return str;
    var month = String(d.getMonth() + 1).padStart(2, '0');
    var day = String(d.getDate()).padStart(2, '0');
    var year = d.getFullYear();
    var hours = String(d.getHours()).padStart(2, '0');
    var mins = String(d.getMinutes()).padStart(2, '0');
    return month + '/' + day + '/' + year + ' ' + hours + ':' + mins;
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

  window.audLoad = audLoad;
  window.audReload = audReload;
  window.audSuspend = audSuspend;
  window.audActivate = audActivate;
  window.audVerify = audVerify;
  window.audDelete = audDelete;
  window.audViewJobs = audViewJobs;
  window.audViewApps = audViewApps;
  window.audViewReports = audViewReports;
  window.audViewCompany = audViewCompany;

  document.addEventListener('click', function (e) {
    var el = e.target.closest('[data-action]');
    if (!el) return;
    var action = el.getAttribute('data-action');
    switch (action) {
      case 'audReload':
        window.audReload();
        break;
      case 'audLoad':
        window.audLoad();
        break;
      case 'audSuspend':
        window.audSuspend();
        break;
      case 'audActivate':
        window.audActivate();
        break;
      case 'audVerify':
        window.audVerify();
        break;
      case 'audDelete':
        window.audDelete();
        break;
      case 'audViewJobs':
        window.audViewJobs();
        break;
      case 'audViewApps':
        window.audViewApps();
        break;
      case 'audViewReports':
        window.audViewReports();
        break;
      case 'audViewCompany':
        window.audViewCompany(el.getAttribute('data-id'));
        break;
    }
  });

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
