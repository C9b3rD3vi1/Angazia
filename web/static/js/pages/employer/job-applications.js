(function () {
  'use strict';

  var jobId = null;
  var currentPage = 1;
  var totalPages = 1;
  var selectedIds = {};
  var elements = {};
  var pendingBulkAction = null;

  function getJobId() {
    var parts = window.location.pathname.split('/');
    return parts[parts.length - 1];
  }

  function qs(id) { return document.getElementById(id); }

  function initElements() {
    elements = {
      loading: qs('ja-loading'),
      error: qs('ja-error'),
      errorMsg: qs('ja-error-msg'),
      content: qs('ja-content'),
      title: qs('ja-title'),
      subtitle: qs('ja-subtitle'),
      tbody: qs('ja-tbody'),
      count: qs('ja-count'),
      statusFilter: qs('ja-status-filter'),
      sort: qs('ja-sort'),
      pagination: qs('ja-pagination'),
      selectAll: qs('ja-select-all'),
      bulkBar: qs('ja-bulk-bar'),
      selectedCount: qs('ja-selected-count'),
      bulkShortlist: qs('ja-bulk-shortlist'),
      bulkReject: qs('ja-bulk-reject'),
      clearSelection: qs('ja-clear-selection'),
      confirmModal: qs('ja-confirm-modal'),
      confirmTitle: qs('ja-confirm-title'),
      confirmIcon: qs('ja-confirm-icon'),
      confirmHeading: qs('ja-confirm-heading'),
      confirmDesc: qs('ja-confirm-desc'),
      confirmYes: qs('ja-confirm-yes'),
      confirmYesLabel: qs('ja-confirm-yes-label'),
      confirmNo: qs('ja-confirm-no'),
      confirmClose: qs('ja-confirm-close'),
    };
  }

  function showLoading() {
    if (elements.loading) elements.loading.style.display = 'block';
    if (elements.error) elements.error.style.display = 'none';
    if (elements.content) elements.content.style.display = 'none';
  }

  function showError(msg) {
    if (elements.loading) elements.loading.style.display = 'none';
    if (elements.error) {
      if (elements.errorMsg) elements.errorMsg.textContent = msg;
      elements.error.style.display = 'block';
    }
  }

  function showContent() {
    if (elements.loading) elements.loading.style.display = 'none';
    if (elements.content) elements.content.style.display = 'block';
  }

  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(function (n) { return n[0]; }).join('').toUpperCase().slice(0, 2);
  }

  function renderAvatar(name, url) {
    if (url) return '<img src="' + url + '" alt="' + escapeHtml(name) + '" style="width:100%;height:100%;object-fit:cover;border-radius:50%">';
    return getInitials(name);
  }

  function escapeHtml(text) {
    if (!text) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(text));
    return d.innerHTML;
  }

  function timeAgo(dateStr) {
    if (!dateStr) return '';
    var date = new Date(dateStr);
    var now = new Date();
    var diffMs = now - date;
    var diffSec = Math.floor(diffMs / 1000);
    if (diffSec < 60) return 'just now';
    var diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return diffMin + 'm ago';
    var diffHour = Math.floor(diffMin / 60);
    if (diffHour < 24) return diffHour + 'h ago';
    var diffDay = Math.floor(diffHour / 24);
    if (diffDay < 7) return diffDay + 'd ago';
    return date.toLocaleDateString();
  }

  function toast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      var toast = document.createElement('div');
      var bg = type === 'success' ? '#10b981' : type === 'error' ? '#ef4444' : '#3b82f6';
      toast.style.cssText = 'position:fixed;bottom:20px;right:20px;background:' + bg + ';color:white;padding:12px 20px;border-radius:8px;font-size:13px;z-index:9999';
      toast.textContent = msg;
      document.body.appendChild(toast);
      setTimeout(function () {
        toast.style.opacity = '0';
        setTimeout(function () { toast.remove(); }, 300);
      }, 3000);
    }
  }

  /* ── Modal helpers ── */
  function setJaConfirmLoading(loading) {
    if (!elements.confirmYes) return;
    elements.confirmYes.disabled = loading;
    elements.confirmYes.classList.toggle('emp-btn-loading', loading);
  }

  function showJaConfirmModal(title, icon, iconCls, heading, desc, btnClass, btnLabel, action) {
    if (elements.confirmTitle) elements.confirmTitle.textContent = title;
    if (elements.confirmIcon) {
      elements.confirmIcon.textContent = icon;
      elements.confirmIcon.className = 'emp-modal-icon ' + iconCls;
    }
    if (elements.confirmHeading) elements.confirmHeading.textContent = heading;
    if (elements.confirmDesc) elements.confirmDesc.textContent = desc;
    if (elements.confirmYesLabel) elements.confirmYesLabel.textContent = btnLabel;
    elements.confirmYes.className = 'emp-btn ' + btnClass;
    setJaConfirmLoading(false);
    pendingBulkAction = action;
    if (elements.confirmModal) elements.confirmModal.style.display = 'flex';
  }

  function hideJaConfirmModal() {
    if (elements.confirmModal) elements.confirmModal.style.display = 'none';
    setJaConfirmLoading(false);
    pendingBulkAction = null;
  }

  function executeJaConfirmAction() {
    if (!pendingBulkAction) return;
    setJaConfirmLoading(true);

    if (pendingBulkAction === 'shortlist') {
      var ids = Object.keys(selectedIds);
      AngaziaAPI.applications.bulkShortlist({ application_ids: ids }).then(function () {
        hideJaConfirmModal();
        toast(ids.length + ' application(s) shortlisted', 'success');
        selectedIds = {};
        updateBulkBar();
        loadApplications();
      }).catch(function (err) {
        toast(err.message || 'Bulk shortlist failed', 'error');
        setJaConfirmLoading(false);
      });
    } else if (pendingBulkAction === 'reject') {
      var ids = Object.keys(selectedIds);
      AngaziaAPI.applications.bulkReject({ application_ids: ids }).then(function () {
        hideJaConfirmModal();
        toast(ids.length + ' application(s) rejected', 'success');
        selectedIds = {};
        updateBulkBar();
        loadApplications();
      }).catch(function (err) {
        toast(err.message || 'Bulk reject failed', 'error');
        setJaConfirmLoading(false);
      });
    } else {
      hideJaConfirmModal();
    }
  }

  function loadApplications() {
    showLoading();

    var params = { page: currentPage, limit: 20 };
    var status = elements.statusFilter ? elements.statusFilter.value : '';
    if (status) params.status = status;
    var sort = elements.sort ? elements.sort.value : 'newest';

    AngaziaAPI.applications.jobApplications(jobId, params).then(function (resp) {
      var apps = [];
      var total = 0;

      if (resp && resp.data) {
        if (Array.isArray(resp.data)) {
          apps = resp.data;
        } else if (resp.data.applications) {
          apps = resp.data.applications;
          total = resp.data.total || apps.length;
          totalPages = resp.data.total_pages || 1;
        }
      } else if (Array.isArray(resp)) {
        apps = resp;
      } else if (resp && resp.applications) {
        apps = resp.applications;
        total = resp.total || apps.length;
        totalPages = resp.total_pages || 1;
      }

      total = total || apps.length;
      totalPages = totalPages || Math.ceil(total / 20) || 1;

      if (elements.count) elements.count.textContent = total;
      if (elements.subtitle) elements.subtitle.textContent = total + ' application' + (total !== 1 ? 's' : '');

      if (sort === 'match') {
        apps.sort(function (a, b) { return (b.match_score || 0) - (a.match_score || 0); });
      }

      renderApplications(apps);
      renderPagination();
      showContent();
    }).catch(function (err) {
      showError(err.message || 'Failed to load applications');
    });
  }

  function renderApplications(apps) {
    if (!elements.tbody) return;

    if (!apps || apps.length === 0) {
      elements.tbody.innerHTML = '<tr><td colspan="6" class="emp-loading-cell">No applications yet</td></tr>';
      return;
    }

    elements.tbody.innerHTML = apps.map(function (app) {
      var id = app.id;
      var checked = selectedIds[id] ? 'checked' : '';
      var emp = app.employee || {};
      var candidateName = emp.full_name || app.candidate_name || app.name || 'Candidate';
      var email = (emp.user && emp.user.email) || app.candidate_email || '';
      var avatarUrl = (emp.user && emp.user.avatar_url) || '';
      var status = app.status || 'pending';
      var match = app.match_score || 0;
      var applied = app.applied_at || app.created_at;

      return '<tr>' +
        '<td style="width:40px;"><input type="checkbox" class="ja-select-item" data-id="' + id + '" ' + checked + '></td>' +
        '<td>' +
          '<div style="display:flex;align-items:center;gap:10px;">' +
            '<div class="emp-avatar">' + renderAvatar(candidateName, avatarUrl) + '</div>' +
            '<div>' +
              '<div style="font-weight:500;">' + escapeHtml(candidateName) + '</div>' +
              '<div style="font-size:11px;color:var(--muted);">' + escapeHtml(email) + '</div>' +
            '</div>' +
          '</div>' +
        '</td>' +
        '<td>' + timeAgo(applied) + '</td>' +
        '<td><span class="emp-match-score">' + match + '%</span></td>' +
        '<td><span class="emp-status-badge ' + status + '">' + status + '</span></td>' +
        '<td><a href="/employer/applications/' + id + '" class="emp-link">Review</a></td>' +
      '</tr>';
    }).join('');

    elements.tbody.querySelectorAll('.ja-select-item').forEach(function (cb) {
      cb.addEventListener('change', function () {
        if (this.checked) {
          selectedIds[this.getAttribute('data-id')] = true;
        } else {
          delete selectedIds[this.getAttribute('data-id')];
        }
        updateBulkBar();
      });
    });
  }

  function renderPagination() {
    if (!elements.pagination) return;
    if (totalPages <= 1) {
      elements.pagination.innerHTML = '';
      return;
    }

    var html = '';
    html += '<button class="emp-page-btn" id="ja-prev" ' + (currentPage <= 1 ? 'disabled' : '') + '>Previous</button>';

    for (var i = 1; i <= totalPages; i++) {
      if (i === 1 || i === totalPages || (i >= currentPage - 2 && i <= currentPage + 2)) {
        html += '<button class="emp-page-btn ' + (i === currentPage ? 'active' : '') + '" data-page="' + i + '">' + i + '</button>';
      } else if (i === currentPage - 3 || i === currentPage + 3) {
        html += '<span class="emp-page-info">...</span>';
      }
    }

    html += '<button class="emp-page-btn" id="ja-next" ' + (currentPage >= totalPages ? 'disabled' : '') + '>Next</button>';
    elements.pagination.innerHTML = html;

    var prev = qs('ja-prev');
    var next = qs('ja-next');

    if (prev) prev.addEventListener('click', function () {
      if (currentPage > 1) { currentPage--; loadApplications(); }
    });
    if (next) next.addEventListener('click', function () {
      if (currentPage < totalPages) { currentPage++; loadApplications(); }
    });

    elements.pagination.querySelectorAll('[data-page]').forEach(function (btn) {
      btn.addEventListener('click', function () {
        currentPage = parseInt(this.getAttribute('data-page'));
        loadApplications();
      });
    });
  }

  function updateBulkBar() {
    var count = Object.keys(selectedIds).length;
    if (elements.bulkBar) {
      elements.bulkBar.style.display = count > 0 ? 'flex' : 'none';
    }
    if (elements.selectedCount) {
      elements.selectedCount.textContent = count + ' selected';
    }
  }

  function bulkAction(action) {
    var ids = Object.keys(selectedIds);
    if (ids.length === 0) return;

    if (action === 'shortlist') {
      showJaConfirmModal(
        'Bulk Shortlist',
        '\u2B50',
        'icon-info',
        'Shortlist ' + ids.length + ' ' + (ids.length === 1 ? 'Candidate' : 'Candidates') + '?',
        ids.length + ' application(s) will be moved to shortlisted status.',
        'emp-btn-primary',
        'Yes, Shortlist',
        'shortlist'
      );
    } else {
      showJaConfirmModal(
        'Bulk Reject',
        '\u2715',
        'icon-danger',
        'Reject ' + ids.length + ' ' + (ids.length === 1 ? 'Candidate' : 'Candidates') + '?',
        ids.length + ' application(s) will be rejected and moved to rejected status.',
        'emp-btn-danger',
        'Yes, Reject',
        'reject'
      );
    }
  }

  function attachEvents() {
    if (elements.statusFilter) {
      elements.statusFilter.addEventListener('change', function () {
        currentPage = 1;
        loadApplications();
      });
    }
    if (elements.sort) {
      elements.sort.addEventListener('change', function () {
        currentPage = 1;
        loadApplications();
      });
    }
    if (elements.selectAll) {
      elements.selectAll.addEventListener('change', function () {
        var rows = elements.tbody ? elements.tbody.querySelectorAll('.ja-select-item') : [];
        rows.forEach(function (cb) {
          cb.checked = elements.selectAll.checked;
          var id = cb.getAttribute('data-id');
          if (elements.selectAll.checked) {
            selectedIds[id] = true;
          } else {
            delete selectedIds[id];
          }
        });
        updateBulkBar();
      });
    }
    if (elements.bulkShortlist) {
      elements.bulkShortlist.addEventListener('click', function () { bulkAction('shortlist'); });
    }
    if (elements.bulkReject) {
      elements.bulkReject.addEventListener('click', function () { bulkAction('reject'); });
    }
    if (elements.clearSelection) {
      elements.clearSelection.addEventListener('click', function () {
        selectedIds = {};
        updateBulkBar();
        if (elements.selectAll) elements.selectAll.checked = false;
        if (elements.tbody) {
          elements.tbody.querySelectorAll('.ja-select-item').forEach(function (cb) { cb.checked = false; });
        }
      });
    }
    if (elements.confirmYes) {
      elements.confirmYes.addEventListener('click', executeJaConfirmAction);
    }
    if (elements.confirmNo) {
      elements.confirmNo.addEventListener('click', hideJaConfirmModal);
    }
    if (elements.confirmClose) {
      elements.confirmClose.addEventListener('click', hideJaConfirmModal);
    }
    if (elements.confirmModal) {
      elements.confirmModal.addEventListener('click', function (e) {
        if (e.target === elements.confirmModal) hideJaConfirmModal();
      });
    }
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && elements.confirmModal && elements.confirmModal.style.display === 'flex') {
        hideJaConfirmModal();
      }
    });
  }

  function init() {
    jobId = getJobId();
    if (!jobId || jobId === 'job-applications') {
      window.location.href = '/employer/applications';
      return;
    }
    initElements();
    attachEvents();
    loadApplications();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
