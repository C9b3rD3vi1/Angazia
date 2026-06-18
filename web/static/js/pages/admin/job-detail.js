'use strict';

function showToast(msg, type) {
  if (window.AngaziaApp && AngaziaApp.showToast) {
    AngaziaApp.showToast(msg, type);
  } else {
    alert((type === 'error' ? 'Error: ' : '') + msg);
  }
}

(function () {
  var jobIdEl = document.querySelector('[data-job-id]');
  var jobId = jobIdEl ? jobIdEl.getAttribute('data-job-id') : null;
  if (!jobId) return;

  function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function getInitials(name) {
    if (!name) return '?';
    var parts = name.split(/\s+/);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return name.substring(0, 2).toUpperCase();
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    try {
      var d = new Date(dateStr);
      return d.toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
    } catch (e) {
      return dateStr;
    }
  }

  function statusBadgeClass(status) {
    switch (status) {
      case 'pending': return 'pending';
      case 'shortlisted': return 'shortlisted';
      case 'rejected': return 'rejected';
      case 'hired': return 'hired';
      case 'withdrawn': return 'withdrawn';
      default: return 'pending';
    }
  }

  function loadApplications() {
    var loadingEl = document.getElementById('ajd-app-loading');
    var errorEl = document.getElementById('ajd-app-error');
    var listEl = document.getElementById('ajd-app-list');
    var emptyEl = document.getElementById('ajd-app-empty');
    var countEl = document.getElementById('ajd-app-count');

    if (loadingEl) loadingEl.style.display = 'flex';
    if (errorEl) errorEl.style.display = 'none';
    if (listEl) listEl.innerHTML = '';
    if (emptyEl) emptyEl.style.display = 'none';

    AngaziaAPI.get('/admin/jobs/' + jobId + '/applications', { page: 1, limit: 10 })
      .then(function (res) {
        var data = res && res.data ? res.data : res;
        var apps = (data && data.applications) || [];
        var total = data ? data.total : 0;

        if (loadingEl) loadingEl.style.display = 'none';
        if (countEl) countEl.textContent = total;

        if (!apps.length) {
          if (emptyEl) emptyEl.style.display = 'flex';
          return;
        }

        var html = '';
        for (var i = 0; i < apps.length; i++) {
          var a = apps[i];
          var employee = a.employee || {};
          var name = employee.full_name || employee.FullName || 'Unknown';
          var avatar = employee.avatar_url || '';
          var status = a.status || 'pending';
          var appliedAt = formatDate(a.applied_at || a.appliedAt || a.AppliedAt);
          var matchScore = a.match_score || a.matchScore || 0;

          html += '<div class="ajd-app-item">';
          if (avatar) {
            html += '<img src="' + escapeHtml(avatar) + '" alt="" class="ajd-app-avatar">';
          } else {
            html += '<span class="ajd-app-avatar ajd-app-avatar-text">' + getInitials(name) + '</span>';
          }
          html += '<div class="ajd-app-info">';
          html += '<a href="/admin/users/' + escapeHtml(a.employee_id || a.employeeId || employee.user_id) + '" class="ajd-app-name">' + escapeHtml(name) + '</a>';
          html += '<span class="ajd-app-date">Applied ' + appliedAt + '</span>';
          html += '</div>';
          if (matchScore > 0) {
            html += '<span class="ajd-app-match">' + matchScore + '% match</span>';
          }
          html += '<span class="ajd-app-status ' + statusBadgeClass(status) + '">' + status + '</span>';
          html += '</div>';
        }

        if (listEl) listEl.innerHTML = html;
      })
      .catch(function (err) {
        if (loadingEl) loadingEl.style.display = 'none';
        if (errorEl) {
          errorEl.style.display = 'flex';
          var errorText = document.getElementById('ajd-app-error-text');
          if (errorText) errorText.textContent = err.message || 'Failed to load applications';
        }
      });
  }

  function closeModal() {
    document.getElementById('ajd-reject-modal').style.display = 'none';
  }

  function handleApprove() {
    var btn = document.querySelector('[data-action="approve"]');
    if (btn) btn.disabled = true;
    AngaziaAPI.admin.approveContent(jobId)
      .then(function () {
        showToast('Job listing approved!', 'success');
        location.reload();
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to approve job', 'error');
      })
      .then(function () { if (btn) btn.disabled = false; });
  }

  function handleReject() {
    document.getElementById('ajd-reject-modal').style.display = 'flex';
  }

  function submitReject() {
    var reason = document.getElementById('ajd-reject-reason').value.trim();
    var confirmBtn = document.getElementById('ajd-modal-confirm');
    confirmBtn.disabled = true;
    AngaziaAPI.admin.rejectContent(jobId, { reason: reason })
      .then(function () {
        showToast('Job listing rejected', 'success');
        closeModal();
        location.reload();
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to reject job', 'error');
      })
      .then(function () { confirmBtn.disabled = false; });
  }

  document.addEventListener('DOMContentLoaded', function () {
    loadApplications();

    var retryBtn = document.querySelector('[data-action="retry-apps"]');
    if (retryBtn) {
      retryBtn.addEventListener('click', function () {
        loadApplications();
      });
    }

    var approveBtn = document.querySelector('[data-action="approve"]');
    var rejectBtn = document.querySelector('[data-action="reject"]');
    if (approveBtn) approveBtn.addEventListener('click', handleApprove);
    if (rejectBtn) rejectBtn.addEventListener('click', handleReject);
    document.getElementById('ajd-modal-confirm').addEventListener('click', submitReject);
    document.getElementById('ajd-modal-close').addEventListener('click', closeModal);
    document.getElementById('ajd-modal-cancel').addEventListener('click', closeModal);
    document.getElementById('ajd-reject-modal').addEventListener('click', function (e) {
      if (e.target === this) closeModal();
    });
  });
})();
