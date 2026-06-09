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

  function closeModal() {
    document.getElementById('ajd-reject-modal').style.display = 'none';
  }

  function handleApprove() {
    var btn = document.querySelector('[data-action="approve"]');
    if (btn) btn.disabled = true;
    AngaziaAPI.post('/admin/moderation/' + jobId + '/approve')
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
    AngaziaAPI.post('/admin/moderation/' + jobId + '/reject', { reason: reason })
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
