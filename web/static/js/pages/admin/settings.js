'use strict';

(function () {
  function qs(sel, ctx) { return (ctx || document).querySelector(sel); }

  function showTfaError(el, m) {
    if (el) { el.textContent = m; el.style.display = ''; }
  }
  function hideTfaError(el) {
    if (el) el.style.display = 'none';
  }

  document.addEventListener('DOMContentLoaded', function () {
    var loadingEl = document.getElementById('tfa-status-loading');
    var notEnabled = document.getElementById('tfa-not-enabled');
    var enabledDiv = document.getElementById('tfa-enabled');
    var tfaError = document.getElementById('tfa-error');

    AngaziaAPI.auth.twoFA.status().then(function (data) {
      if (loadingEl) loadingEl.style.display = 'none';
      if (data && data.enabled) {
        if (notEnabled) notEnabled.style.display = 'none';
        if (enabledDiv) enabledDiv.style.display = '';
      } else {
        if (notEnabled) notEnabled.style.display = '';
        if (enabledDiv) enabledDiv.style.display = 'none';
      }
    }).catch(function () {
      if (loadingEl) loadingEl.style.display = 'none';
      if (notEnabled) notEnabled.style.display = '';
      if (enabledDiv) enabledDiv.style.display = 'none';
    });

    var enableBtn = document.getElementById('tfa-enable-btn');
    if (enableBtn) {
      enableBtn.addEventListener('click', function () {
        window.open('/auth/2fa/setup', '_blank');
      });
    }

    var disableBtn = document.getElementById('tfa-disable-btn');
    if (disableBtn) {
      disableBtn.addEventListener('click', function () {
        if (!confirm('Are you sure you want to disable two-factor authentication?')) return;
        var code = prompt('Enter your current TOTP code (or a backup code):');
        if (!code) return;
        var password = prompt('Confirm your password:');
        if (!password) return;
        hideTfaError(tfaError);
        disableBtn.disabled = true;
        disableBtn.textContent = 'Disabling\u2026';
        AngaziaAPI.auth.twoFA.disable({ code: code, password: password }).then(function () {
          window.location.reload();
        }).catch(function (err) {
          disableBtn.disabled = false;
          disableBtn.textContent = 'Disable 2FA';
          showTfaError(tfaError, (err.body && err.body.message) || err.message || 'Failed to disable 2FA');
        });
      });
    }

    var viewBackupBtn = document.getElementById('tfa-view-backup-btn');
    if (viewBackupBtn) {
      viewBackupBtn.addEventListener('click', function () {
        hideTfaError(tfaError);
        if (!confirm('Generate new backup codes? Your existing codes will be invalidated.')) return;
        viewBackupBtn.disabled = true;
        viewBackupBtn.textContent = 'Generating\u2026';
        AngaziaAPI.auth.twoFA.generateBackupCodes().then(function (data) {
          viewBackupBtn.disabled = false;
          viewBackupBtn.textContent = 'Generate New Backup Codes';
          if (data && data.backup_codes && data.backup_codes.length) {
            var msg = 'Your New Backup Codes:\n\n' + data.backup_codes.join('\n') + '\n\nStore these securely! You will not be able to view them again.';
            alert(msg);
          } else {
            showTfaError(tfaError, 'Failed to generate backup codes');
          }
        }).catch(function (err) {
          viewBackupBtn.disabled = false;
          viewBackupBtn.textContent = 'Generate New Backup Codes';
          showTfaError(tfaError, (err.body && err.body.message) || err.message || 'Failed to generate backup codes');
        });
      });
    }
  });
})();
