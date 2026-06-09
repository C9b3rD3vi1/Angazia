'use strict';

(function () {
  var pathParts = window.location.pathname.split('/');
  var companyId = pathParts[pathParts.length - 1];

  if (!companyId || companyId === 'companies') return;

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
      return;
    }
    alert((type === 'error' ? 'Error: ' : '') + msg);
  }

  function closeModal() {
    var modal = document.getElementById('acd-reject-modal');
    if (modal) modal.style.display = 'none';
  }

  function handleApprove() {
    var btn = document.querySelector('[data-action="approve"]');
    if (!btn) return;

    var originalText = btn.textContent;
    btn.disabled = true;
    btn.textContent = 'Processing...';

    AngaziaAPI.admin.verifyCompany(companyId)
      .then(function () {
        showToast('Verification approved', 'success');
        setTimeout(function () {
          location.reload();
        }, 1500);
      })
      .catch(function (err) {
        var errorMsg = err.message || 'Failed to approve verification';
        if (err.body && err.body.message) errorMsg = err.body.message;
        showToast(errorMsg, 'error');
        btn.disabled = false;
        btn.textContent = originalText;
      });
  }

  function handleReject() {
    var modal = document.getElementById('acd-reject-modal');
    if (modal) modal.style.display = 'flex';
  }

  function submitReject() {
    var reason = document.getElementById('acd-reject-reason').value.trim();
    var confirmBtn = document.getElementById('acd-modal-confirm');

    if (!reason) {
      showToast('Please provide a reason for rejection', 'warning');
      return;
    }

    var originalText = confirmBtn.textContent;
    confirmBtn.disabled = true;
    confirmBtn.textContent = 'Processing...';

    AngaziaAPI.admin.rejectCompany(companyId, { reason: reason })
      .then(function () {
        showToast('Verification rejected', 'success');
        closeModal();
        setTimeout(function () {
          location.reload();
        }, 1500);
      })
      .catch(function (err) {
        var errorMsg = err.message || 'Failed to reject verification';
        if (err.body && err.body.message) errorMsg = err.body.message;
        showToast(errorMsg, 'error');
        confirmBtn.disabled = false;
        confirmBtn.textContent = originalText;
      });
  }

  document.addEventListener('DOMContentLoaded', function () {
    var approveBtn = document.querySelector('[data-action="approve"]');
    var rejectBtn = document.querySelector('[data-action="reject"]');
    var confirmBtn = document.getElementById('acd-modal-confirm');
    var closeBtn = document.getElementById('acd-modal-close');
    var cancelBtn = document.getElementById('acd-modal-cancel');
    var modalOverlay = document.getElementById('acd-reject-modal');

    if (approveBtn) approveBtn.addEventListener('click', handleApprove);
    if (rejectBtn) rejectBtn.addEventListener('click', handleReject);
    if (confirmBtn) confirmBtn.addEventListener('click', submitReject);
    if (closeBtn) closeBtn.addEventListener('click', closeModal);
    if (cancelBtn) cancelBtn.addEventListener('click', closeModal);
    if (modalOverlay) {
      modalOverlay.addEventListener('click', function (e) {
        if (e.target === this) closeModal();
      });
    }
  });
})();
