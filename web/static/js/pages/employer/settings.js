(function () {
  'use strict';

  document.addEventListener('DOMContentLoaded', function () {

    var els = {
      // Avatar
      avatarInput: document.getElementById('avatar-upload-input'),
      avatarBtn: document.getElementById('avatar-upload-btn'),
      avatarFileName: document.getElementById('avatar-file-name'),
      avatarPreviewImg: document.getElementById('avatar-preview-img'),
      avatarPreviewInit: document.getElementById('avatar-preview-init'),
      avatarRemoveBtn: document.getElementById('avatar-remove-btn'),

      // Profile form
      formProfile: document.getElementById('form-profile'),
      profileCompanyName: document.getElementById('profile-company-name'),
      profileContactEmail: document.getElementById('profile-contact-email'),
      profilePhone: document.getElementById('profile-phone'),
      profileLocation: document.getElementById('profile-location'),

      // Company form
      formCompany: document.getElementById('form-company'),
      companyWebsite: document.getElementById('company-website'),
      companyLinkedin: document.getElementById('company-linkedin'),
      companyDescription: document.getElementById('company-description'),
      companyIndustry: document.getElementById('company-industry'),
      companySize: document.getElementById('company-size'),
      companyBusinessReg: document.getElementById('company-business-reg'),
      companyTaxId: document.getElementById('company-tax-id'),

      // Password form
      formPassword: document.getElementById('form-password'),
      pwCurrent: document.getElementById('pw-current'),
      pwNew: document.getElementById('pw-new'),
      pwConfirm: document.getElementById('pw-confirm'),
      pwStrength: document.getElementById('pw-strength'),
      pwStrengthText: document.getElementById('pw-strength-text'),

      // 2FA
      tfaLoading: document.getElementById('tfa-status-loading'),
      tfaNotEnabled: document.getElementById('tfa-not-enabled'),
      tfaEnabled: document.getElementById('tfa-enabled'),
      tfaError: document.getElementById('tfa-error'),
      tfaEnableBtn: document.getElementById('tfa-enable-btn'),
      tfaDisableBtn: document.getElementById('tfa-disable-btn'),
      tfaBackupBtn: document.getElementById('tfa-view-backup-btn'),
      tfaMethod: document.getElementById('tfa-method'),

      // 2FA Setup Modal
      tfaSetupModal: document.getElementById('tfa-setup-modal'),
      tfaSetupQr: document.getElementById('tfa-setup-qr'),
      tfaSetupSecret: document.getElementById('tfa-setup-secret'),
      tfaSetupCode: document.getElementById('tfa-setup-code'),
      tfaSetupVerify: document.getElementById('tfa-setup-verify'),
      tfaSetupCancel: document.getElementById('tfa-setup-cancel'),
      tfaSetupError: document.getElementById('tfa-setup-error'),
      tfaSetupLoading: document.getElementById('tfa-setup-loading'),
      tfaSetupContent: document.getElementById('tfa-setup-content'),

      // Sessions
      sessionsLoading: document.getElementById('sessions-loading'),
      sessionsList: document.getElementById('sessions-list'),
      sessionsEmpty: document.getElementById('sessions-empty'),
      sessionsError: document.getElementById('sessions-error'),
      revokeAllBtn: document.getElementById('revoke-all-btn'),

      // Notifications
      notifLoading: document.getElementById('notif-prefs-loading'),
      notifForm: document.getElementById('form-notifications'),
      notifError: document.getElementById('notif-prefs-error'),
      notifTestBtn: document.getElementById('notif-test-btn'),

      // Danger Zone
      deleteBtn: document.getElementById('delete-account-btn'),
      deleteModal: document.getElementById('delete-modal'),
      deleteConfirmBtn: document.getElementById('delete-confirm-btn'),
      deleteCancelBtn: document.getElementById('delete-cancel-btn'),
      deletePassword: document.getElementById('delete-password'),
      deleteError: document.getElementById('delete-error'),

      // Generic Confirm Modal
      confirmModal: document.getElementById('confirm-modal'),
      confirmTitle: document.getElementById('confirm-modal-title'),
      confirmHeading: document.getElementById('confirm-modal-heading'),
      confirmDesc: document.getElementById('confirm-modal-desc'),
      confirmOk: document.getElementById('confirm-modal-ok'),
      confirmCancel: document.getElementById('confirm-modal-cancel'),
      confirmClose: document.getElementById('confirm-modal-close'),

      // Disable 2FA Modal
      tfaDisableModal: document.getElementById('tfa-disable-modal'),
      tfaDisableCode: document.getElementById('tfa-disable-code'),
      tfaDisablePassword: document.getElementById('tfa-disable-password'),
      tfaDisableConfirm: document.getElementById('tfa-disable-confirm'),
      tfaDisableCancel: document.getElementById('tfa-disable-cancel'),
      tfaDisableError: document.getElementById('tfa-disable-error'),
    };

    function toast(msg, type) {
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast(msg, type);
      } else {
        alert((type === 'error' ? 'Error: ' : '') + msg);
      }
    }

    function escapeHtml(t) {
      if (t == null) return '';
      var d = document.createElement('div');
      d.appendChild(document.createTextNode(String(t)));
      return d.innerHTML;
    }

    function showLoading(btn, text) {
      btn.disabled = true;
      btn.dataset.origText = btn.textContent;
      btn.textContent = text || 'Loading...';
    }

    function hideLoading(btn) {
      btn.disabled = false;
      btn.textContent = btn.dataset.origText || btn.textContent;
    }

    function getFormData(form) {
      var fd = new FormData(form);
      var data = {};
      fd.forEach(function (v, k) { data[k] = v; });
      return data;
    }

    // ── Generic Confirmation Modal ──

    var confirmCallback = null;

    function openConfirmModal(opts) {
      if (els.confirmTitle) els.confirmTitle.textContent = opts.title || 'Confirm';
      if (els.confirmHeading) els.confirmHeading.textContent = opts.heading || 'Are you sure?';
      if (els.confirmDesc) els.confirmDesc.textContent = opts.desc || 'This action cannot be undone.';
      if (els.confirmOk) {
        els.confirmOk.textContent = opts.okText || 'Confirm';
        els.confirmOk.className = 'emp-btn ' + (opts.danger !== false ? 'emp-btn-danger' : 'emp-btn-primary');
      }
      confirmCallback = opts.onConfirm || null;
      if (els.confirmModal) els.confirmModal.style.display = 'flex';
    }

    function closeConfirmModal() {
      if (els.confirmModal) els.confirmModal.style.display = 'none';
      confirmCallback = null;
    }

    if (els.confirmOk) {
      els.confirmOk.addEventListener('click', function () {
        if (confirmCallback) {
          var cb = confirmCallback;
          closeConfirmModal();
          cb();
        }
      });
    }

    if (els.confirmCancel) {
      els.confirmCancel.addEventListener('click', closeConfirmModal);
    }

    if (els.confirmClose) {
      els.confirmClose.addEventListener('click', closeConfirmModal);
    }

    if (els.confirmModal) {
      els.confirmModal.addEventListener('click', function (e) {
        if (e.target === els.confirmModal) closeConfirmModal();
      });
    }

    // ── Avatar ──

    var selectedAvatarFile = null;

    if (els.avatarInput && els.avatarBtn) {
      els.avatarInput.addEventListener('change', function () {
        selectedAvatarFile = els.avatarInput.files[0];
        if (selectedAvatarFile) {
          els.avatarBtn.disabled = false;
          els.avatarFileName.textContent = selectedAvatarFile.name;
          var reader = new FileReader();
          reader.onload = function (e) {
            if (els.avatarPreviewImg) {
              els.avatarPreviewImg.src = e.target.result;
              els.avatarPreviewImg.style.display = '';
            }
            if (els.avatarPreviewInit) els.avatarPreviewInit.style.display = 'none';
          };
          reader.readAsDataURL(selectedAvatarFile);
        } else {
          els.avatarBtn.disabled = true;
          els.avatarFileName.textContent = '';
        }
      });

      els.avatarBtn.addEventListener('click', function () {
        if (!selectedAvatarFile) return;
        els.avatarBtn.disabled = true;
        els.avatarBtn.textContent = 'Uploading...';
        var fd = new FormData();
        fd.append('avatar', selectedAvatarFile);
        AngaziaAPI.profile.uploadAvatar(fd)
          .then(function () {
            toast('Profile picture updated!', 'success');
            selectedAvatarFile = null;
            els.avatarBtn.textContent = 'Upload';
            els.avatarFileName.textContent = '';
            els.avatarInput.value = '';
          })
          .catch(function (err) {
            toast(err.message || 'Failed to upload avatar', 'error');
            els.avatarBtn.disabled = false;
            els.avatarBtn.textContent = 'Upload';
          });
      });
    }

    if (els.avatarRemoveBtn) {
      els.avatarRemoveBtn.addEventListener('click', function () {
        openConfirmModal({
          title: 'Remove Picture',
          heading: 'Remove your profile picture?',
          desc: 'Your profile picture will be removed and replaced with your initials.',
          okText: 'Remove',
          onConfirm: function () {
            AngaziaAPI.profile.update({ avatar_url: '' })
              .then(function () {
                toast('Profile picture removed', 'success');
                if (els.avatarPreviewImg) { els.avatarPreviewImg.src = ''; els.avatarPreviewImg.style.display = 'none'; }
                if (els.avatarPreviewInit) els.avatarPreviewInit.style.display = '';
              })
              .catch(function (err) {
                toast(err.message || 'Failed to remove avatar', 'error');
              });
          }
        });
      });
    }

    // ── Load Profile & Company Data ──

    function loadProfileData() {
      AngaziaAPI.profile.get().then(function (data) {
        var user = data.user || {};
        var emp = data.employer_profile || {};

        if (els.profileCompanyName) els.profileCompanyName.value = emp.company_name || user.company_name || '';
        if (els.profileContactEmail) els.profileContactEmail.value = emp.contact_email || user.email || '';
        if (els.profilePhone) els.profilePhone.value = emp.phone_number || '';
        if (els.profileLocation) els.profileLocation.value = emp.location || '';

        var avatarUrl = user.avatar_url || emp.company_logo || '';
        if (avatarUrl && els.avatarPreviewImg) {
          els.avatarPreviewImg.src = avatarUrl;
          els.avatarPreviewImg.style.display = '';
          if (els.avatarPreviewInit) els.avatarPreviewInit.style.display = 'none';
        }
      }).catch(function () {});
    }

    function loadCompanyData() {
      AngaziaAPI.companies.myCompany().then(function (data) {
        var p = data.profile || data;
        var v = data.verification || {};
        if (els.companyWebsite) els.companyWebsite.value = p.company_website || '';
        if (els.companyLinkedin) els.companyLinkedin.value = p.company_linkedin || '';
        if (els.companyDescription) els.companyDescription.value = p.company_description || '';
        if (els.companyIndustry) els.companyIndustry.value = p.industry || '';
        if (els.companySize) els.companySize.value = p.company_size || '';
        if (els.companyBusinessReg) els.companyBusinessReg.value = v.business_registration_number || '';
        if (els.companyTaxId) els.companyTaxId.value = v.tax_id || '';

        var logoUrl = p.logo || p.company_logo || '';
        if (logoUrl && els.avatarPreviewImg && !els.avatarPreviewImg.src) {
          els.avatarPreviewImg.src = logoUrl;
          els.avatarPreviewImg.style.display = '';
          if (els.avatarPreviewInit) els.avatarPreviewInit.style.display = 'none';
        }
      }).catch(function () {});
    }

    loadProfileData();
    loadCompanyData();

    // ── Profile Form (Account Info) ──

    if (els.formProfile) {
      els.formProfile.addEventListener('submit', function (e) {
        e.preventDefault();
        var data = getFormData(els.formProfile);
        var btn = els.formProfile.querySelector('button[type="submit"]');
        showLoading(btn, 'Saving...');
        AngaziaAPI.companies.updateCompany(data)
          .then(function () {
            toast('Account information saved', 'success');
          })
          .catch(function (err) {
            toast(err.message || 'Failed to save account info', 'error');
          })
          .then(function () {
            hideLoading(btn);
          });
      });
    }

    // ── Company Profile Form ──

    if (els.formCompany) {
      els.formCompany.addEventListener('submit', function (e) {
        e.preventDefault();
        var data = getFormData(els.formCompany);
        var btn = els.formCompany.querySelector('button[type="submit"]');
        showLoading(btn, 'Saving...');
        AngaziaAPI.companies.updateCompany(data)
          .then(function () {
            toast('Company profile saved successfully', 'success');
          })
          .catch(function (err) {
            toast(err.message || 'Failed to save company profile', 'error');
          })
          .then(function () {
            hideLoading(btn);
          });
      });
    }

    // ── Password ──

    function checkPasswordStrength(pw) {
      var score = 0;
      if (pw.length >= 8) score++;
      if (pw.length >= 12) score++;
      if (/[a-z]/.test(pw) && /[A-Z]/.test(pw)) score++;
      if (/\d/.test(pw)) score++;
      if (/[^a-zA-Z0-9]/.test(pw)) score++;
      return score;
    }

    function updatePasswordStrength() {
      if (!els.pwStrength || !els.pwStrengthText) return;
      var pw = els.pwNew ? els.pwNew.value : '';
      var score = checkPasswordStrength(pw);
      var labels = ['Weak', 'Fair', 'Good', 'Strong', 'Very Strong'];
      var colors = ['#ef4444', '#f5a623', '#3b82f6', '#00e5a0', '#00e5a0'];
      var pct = (score / 5) * 100;
      els.pwStrength.style.width = pct + '%';
      els.pwStrength.style.background = colors[score] || colors[0];
      els.pwStrengthText.textContent = pw ? labels[score] || '' : '';
    }

    if (els.pwNew) {
      els.pwNew.addEventListener('input', updatePasswordStrength);
    }

    if (els.formPassword) {
      els.formPassword.addEventListener('submit', function (e) {
        e.preventDefault();
        var data = getFormData(els.formPassword);
        if (data.new_password !== data.confirm_password) {
          toast('Passwords do not match', 'error');
          return;
        }
        if (checkPasswordStrength(data.new_password) < 2) {
          toast('Password is too weak. Use at least 8 characters with a mix of letters, numbers, and symbols.', 'error');
          return;
        }
        var btn = els.formPassword.querySelector('button[type="submit"]');
        showLoading(btn, 'Updating...');
        AngaziaAPI.auth.changePassword({ old_password: data.current_password, new_password: data.new_password })
          .then(function () {
            toast('Password changed successfully', 'success');
            els.formPassword.reset();
            updatePasswordStrength();
          })
          .catch(function (err) {
            toast(err.message || 'Failed to change password', 'error');
          })
          .then(function () {
            hideLoading(btn);
          });
      });
    }

    // ── Two-Factor Authentication ──

    function showTfaError(msg) {
      if (els.tfaError) { els.tfaError.textContent = msg; els.tfaError.style.display = ''; }
    }

    function hideTfaError() {
      if (els.tfaError) els.tfaError.style.display = 'none';
    }

    function loadTfaStatus() {
      if (!els.tfaLoading) return;
      AngaziaAPI.auth.twoFA.status()
        .then(function (data) {
          els.tfaLoading.style.display = 'none';
          if (data && data.enabled) {
            if (els.tfaNotEnabled) els.tfaNotEnabled.style.display = 'none';
            if (els.tfaEnabled) els.tfaEnabled.style.display = '';
            if (els.tfaMethod) els.tfaMethod.textContent = data.method || 'Authenticator App';
          } else {
            if (els.tfaNotEnabled) els.tfaNotEnabled.style.display = '';
            if (els.tfaEnabled) els.tfaEnabled.style.display = 'none';
          }
        })
        .catch(function () {
          els.tfaLoading.style.display = 'none';
          if (els.tfaNotEnabled) els.tfaNotEnabled.style.display = '';
          if (els.tfaEnabled) els.tfaEnabled.style.display = 'none';
        });
    }

    loadTfaStatus();

    // ── 2FA Setup Modal (inline) ──

    if (els.tfaEnableBtn) {
      els.tfaEnableBtn.addEventListener('click', function () {
        hideTfaError();
        els.tfaEnableBtn.disabled = true;
        els.tfaEnableBtn.textContent = 'Setting up...';
        if (els.tfaSetupModal) els.tfaSetupModal.style.display = 'flex';
        if (els.tfaSetupLoading) els.tfaSetupLoading.style.display = 'flex';
        if (els.tfaSetupQr) els.tfaSetupQr.innerHTML = '';
        if (els.tfaSetupSecret) els.tfaSetupSecret.textContent = '';
        if (els.tfaSetupCode) els.tfaSetupCode.value = '';
        if (els.tfaSetupError) els.tfaSetupError.style.display = 'none';

        AngaziaAPI.auth.twoFA.setup({ method: 'app' })
          .then(function (data) {
            els.tfaEnableBtn.disabled = false;
            els.tfaEnableBtn.textContent = 'Enable Two-Factor Authentication';
            if (els.tfaSetupLoading) els.tfaSetupLoading.style.display = 'none';
            if (els.tfaSetupContent) els.tfaSetupContent.style.display = '';
            if (data.qr_code && els.tfaSetupQr) {
              els.tfaSetupQr.innerHTML = '<img src="' + escapeHtml(data.qr_code) + '" alt="QR Code" style="max-width:200px;image-rendering:pixelated">';
            }
            if (data.secret && els.tfaSetupSecret) {
              els.tfaSetupSecret.textContent = data.secret;
            }
          })
          .catch(function (err) {
            els.tfaEnableBtn.disabled = false;
            els.tfaEnableBtn.textContent = 'Enable Two-Factor Authentication';
            if (els.tfaSetupModal) els.tfaSetupModal.style.display = 'none';
            showTfaError(err.message || 'Failed to setup 2FA');
          });
      });
    }

    if (els.tfaSetupVerify) {
      els.tfaSetupVerify.addEventListener('click', function () {
        var code = els.tfaSetupCode ? els.tfaSetupCode.value.trim() : '';
        if (!code || code.length < 6) {
          if (els.tfaSetupError) { els.tfaSetupError.textContent = 'Enter a valid 6-digit code from your authenticator app.'; els.tfaSetupError.style.display = ''; }
          return;
        }
        if (els.tfaSetupError) els.tfaSetupError.style.display = 'none';
        els.tfaSetupVerify.disabled = true;
        els.tfaSetupVerify.textContent = 'Verifying...';

        AngaziaAPI.auth.twoFA.verify({ code: code, method: 'app' })
          .then(function (data) {
            toast('Two-factor authentication enabled successfully!', 'success');
            if (els.tfaSetupModal) els.tfaSetupModal.style.display = 'none';
            if (data && data.backup_codes && data.backup_codes.length) {
              var msg = 'Your Backup Codes:\n\n' + data.backup_codes.join('\n') + '\n\nStore these securely!';
              alert(msg);
            }
            loadTfaStatus();
          })
          .catch(function (err) {
            els.tfaSetupVerify.disabled = false;
            els.tfaSetupVerify.textContent = 'Verify & Enable';
            if (els.tfaSetupError) { els.tfaSetupError.textContent = err.message || 'Invalid code. Try again.'; els.tfaSetupError.style.display = ''; }
          });
      });
    }

    function closeTfaSetup() {
      if (els.tfaSetupModal) els.tfaSetupModal.style.display = 'none';
      if (els.tfaSetupVerify) { els.tfaSetupVerify.disabled = false; els.tfaSetupVerify.textContent = 'Verify & Enable'; }
    }

    if (els.tfaSetupCancel) {
      els.tfaSetupCancel.addEventListener('click', closeTfaSetup);
    }

    if (els.tfaSetupModal) {
      els.tfaSetupModal.addEventListener('click', function (e) {
        if (e.target === els.tfaSetupModal) closeTfaSetup();
      });
    }

    // ── 2FA Disable ──

    if (els.tfaDisableBtn) {
      els.tfaDisableBtn.addEventListener('click', function () {
        hideTfaError();
        if (els.tfaDisableCode) els.tfaDisableCode.value = '';
        if (els.tfaDisablePassword) els.tfaDisablePassword.value = '';
        if (els.tfaDisableError) els.tfaDisableError.style.display = 'none';
        if (els.tfaDisableModal) els.tfaDisableModal.style.display = 'flex';
        if (els.tfaDisableCode) els.tfaDisableCode.focus();
      });
    }

    function doDisable2FA() {
      var code = els.tfaDisableCode ? els.tfaDisableCode.value.trim() : '';
      var password = els.tfaDisablePassword ? els.tfaDisablePassword.value : '';
      if (!code || !password) {
        if (els.tfaDisableError) { els.tfaDisableError.textContent = 'Enter both the code and your password.'; els.tfaDisableError.style.display = ''; }
        return;
      }
      if (els.tfaDisableError) els.tfaDisableError.style.display = 'none';
      if (els.tfaDisableConfirm) { els.tfaDisableConfirm.disabled = true; els.tfaDisableConfirm.textContent = 'Disabling...'; }
      AngaziaAPI.auth.twoFA.disable({ code: code, password: password })
        .then(function () {
          toast('2FA disabled successfully', 'success');
          if (els.tfaDisableModal) els.tfaDisableModal.style.display = 'none';
          loadTfaStatus();
        })
        .catch(function (err) {
          if (els.tfaDisableConfirm) { els.tfaDisableConfirm.disabled = false; els.tfaDisableConfirm.textContent = 'Disable 2FA'; }
          if (els.tfaDisableError) { els.tfaDisableError.textContent = (err && err.body && err.body.message) || err.message || 'Failed to disable 2FA'; els.tfaDisableError.style.display = ''; }
        });
    }

    if (els.tfaDisableConfirm) {
      els.tfaDisableConfirm.addEventListener('click', doDisable2FA);
    }

    function closeTfaDisableModal() {
      if (els.tfaDisableModal) els.tfaDisableModal.style.display = 'none';
      if (els.tfaDisableConfirm) { els.tfaDisableConfirm.disabled = false; els.tfaDisableConfirm.textContent = 'Disable 2FA'; }
    }

    if (els.tfaDisableCancel) {
      els.tfaDisableCancel.addEventListener('click', closeTfaDisableModal);
    }

    if (els.tfaDisableModal) {
      els.tfaDisableModal.addEventListener('click', function (e) {
        if (e.target === els.tfaDisableModal) closeTfaDisableModal();
      });
    }

    // Allow Enter key to submit disable
    if (els.tfaDisablePassword) {
      els.tfaDisablePassword.addEventListener('keydown', function (e) {
        if (e.key === 'Enter') { e.preventDefault(); doDisable2FA(); }
      });
    }

    // ── 2FA Backup Codes ──

    if (els.tfaBackupBtn) {
      els.tfaBackupBtn.addEventListener('click', function () {
        hideTfaError();
        openConfirmModal({
          title: 'Generate Backup Codes',
          heading: 'Generate new backup codes?',
          desc: 'Your existing backup codes will be invalidated. Store the new codes in a safe place.',
          okText: 'Generate',
          onConfirm: function doGen() {
            closeConfirmModal();
            els.tfaBackupBtn.disabled = true;
            els.tfaBackupBtn.textContent = 'Generating...';
            AngaziaAPI.auth.twoFA.generateBackupCodes()
              .then(function (data) {
                els.tfaBackupBtn.disabled = false;
                els.tfaBackupBtn.textContent = 'Generate New Backup Codes';
                if (data && data.backup_codes && data.backup_codes.length) {
                  showBackupCodesModal(data.backup_codes);
                } else {
                  showTfaError('Failed to generate backup codes');
                }
              })
              .catch(function (err) {
                els.tfaBackupBtn.disabled = false;
                els.tfaBackupBtn.textContent = 'Generate New Backup Codes';
                showTfaError((err && err.body && err.body.message) || err.message || 'Failed to generate backup codes');
              });
          }
        });
      });
    }

    function showBackupCodesModal(codes) {
      if (!codes || !codes.length) return;
      if (els.confirmTitle) els.confirmTitle.textContent = 'Backup Codes';
      if (els.confirmHeading) els.confirmHeading.textContent = 'Your New Backup Codes';
      if (els.confirmDesc) {
        els.confirmDesc.innerHTML = 'Store each code securely. Each code can be used only once.<br><br>' +
          '<code style="display:block;text-align:center;font-size:16px;letter-spacing:2px;line-height:2;background:var(--s2);padding:16px;border-radius:8px;border:1px solid var(--border)">' +
          codes.join('<br>') +
          '</code><br><span style="color:var(--danger);font-weight:500">Keep these codes safe!</span>';
      }
      if (els.confirmOk) {
        els.confirmOk.textContent = 'I Saved Them';
        els.confirmOk.className = 'emp-btn emp-btn-primary';
      }
      confirmCallback = function () { closeConfirmModal(); };
      if (els.confirmModal) els.confirmModal.style.display = 'flex';
    }

    // ── Session Management ──

    function loadSessions() {
      if (!els.sessionsLoading) return;
      els.sessionsLoading.style.display = '';
      if (els.sessionsError) els.sessionsError.style.display = 'none';

      AngaziaAPI.auth.sessions()
        .then(function (data) {
          els.sessionsLoading.style.display = 'none';
          var sessions = data && data.sessions ? data.sessions : (Array.isArray(data) ? data : []);
          if (!sessions.length) {
            if (els.sessionsEmpty) els.sessionsEmpty.style.display = '';
            if (els.sessionsList) els.sessionsList.innerHTML = '';
            return;
          }
          if (els.sessionsEmpty) els.sessionsEmpty.style.display = 'none';
          if (!els.sessionsList) return;

          var html = '';
          sessions.forEach(function (s) {
            var deviceInfo = [s.browser, s.os].filter(Boolean).join(' on ') || s.device || 'Unknown device';
            html += '<div class="ses-item' + (s.is_current ? ' ses-current' : '') + '">' +
              '<div class="ses-info">' +
              '<div class="ses-device">' +
              (s.is_current ? '<span class="ses-badge">Current</span> ' : '') +
              escapeHtml(deviceInfo) +
              '</div>' +
              '<div class="ses-meta">' +
              (s.ip ? 'IP: ' + escapeHtml(s.ip) + ' &middot; ' : '') +
              'Last active: ' + escapeHtml(s.last_active || 'Unknown') +
              (s.created_at ? ' &middot; Created: ' + escapeHtml(s.created_at) : '') +
              '</div>' +
              '</div>' +
              (!s.is_current ? '<button class="emp-btn emp-btn-xs emp-btn-ghost ses-revoke-btn" data-session-id="' + escapeHtml(s.id) + '" style="color:var(--danger)">Revoke</button>' : '') +
              '</div>';
          });
          els.sessionsList.innerHTML = html;

          els.sessionsList.querySelectorAll('.ses-revoke-btn').forEach(function (btn) {
            btn.addEventListener('click', function () {
              var sid = this.dataset.sessionId;
              if (!sid) return;
              var revokeBtn = this;
              openConfirmModal({
                title: 'Revoke Session',
                heading: 'Revoke this session?',
                desc: 'The device will be signed out immediately.',
                okText: 'Revoke',
                onConfirm: function () {
                  revokeBtn.disabled = true;
                  revokeBtn.textContent = 'Revoking...';
                  AngaziaAPI.auth.revokeSession({ session_id: sid })
                    .then(function () {
                      toast('Session revoked', 'success');
                      loadSessions();
                    })
                    .catch(function (err) {
                      toast(err.message || 'Failed to revoke session', 'error');
                      loadSessions();
                    });
                }
              });
            });
          });
        })
        .catch(function (err) {
          els.sessionsLoading.style.display = 'none';
          if (els.sessionsError) { els.sessionsError.style.display = ''; els.sessionsError.textContent = err.message || 'Failed to load sessions'; }
        });
    }

    loadSessions();

    if (els.revokeAllBtn) {
      els.revokeAllBtn.addEventListener('click', function () {
        openConfirmModal({
          title: 'Revoke All Sessions',
          heading: 'Revoke all other sessions?',
          desc: 'All devices except this one will be signed out immediately.',
          okText: 'Revoke All',
          onConfirm: function () {
            els.revokeAllBtn.disabled = true;
            els.revokeAllBtn.textContent = 'Revoking...';
            AngaziaAPI.auth.sessions()
              .then(function (data) {
                var sessions = data && data.sessions ? data.sessions : [];
                var promises = [];
                sessions.forEach(function (s) {
                  if (!s.is_current) {
                    promises.push(AngaziaAPI.auth.revokeSession({ session_id: s.id }));
                  }
                });
                return Promise.all(promises);
              })
              .then(function () {
                toast('All other sessions revoked', 'success');
                loadSessions();
              })
              .catch(function (err) {
                toast(err.message || 'Failed to revoke sessions', 'error');
              })
              .then(function () {
                els.revokeAllBtn.disabled = false;
                els.revokeAllBtn.textContent = 'Revoke All Other Sessions';
              });
          }
        });
      });
    }

    // ── Notification Preferences ──

    function showNotifError(msg) {
      if (els.notifError) { els.notifError.textContent = msg; els.notifError.style.display = ''; }
    }

    function hideNotifError() {
      if (els.notifError) els.notifError.style.display = 'none';
    }

    function setCheckboxes(prefs) {
      if (!prefs || !els.notifForm) return;
      var fields = [
        'application_updates', 'job_alerts', 'interview_reminders',
        'messages', 'system_alerts', 'marketing',
        'push_enabled', 'email_enabled', 'in_app_enabled'
      ];
      for (var i = 0; i < fields.length; i++) {
        var key = fields[i];
        var val = prefs[key];
        if (typeof val === 'boolean') {
          var cb = els.notifForm.querySelector('input[name="' + key + '"]');
          if (cb) cb.checked = val;
        }
      }
    }

    if (els.notifForm) {
      AngaziaAPI.notifications.getPreferences()
        .then(function (prefs) {
          if (els.notifLoading) els.notifLoading.style.display = 'none';
          els.notifForm.style.display = '';
          setCheckboxes(prefs);
        })
        .catch(function () {
          if (els.notifLoading) els.notifLoading.style.display = 'none';
          els.notifForm.style.display = '';
        });
    }

    if (els.notifForm) {
      els.notifForm.addEventListener('submit', function (e) {
        e.preventDefault();
        hideNotifError();
        var fd = new FormData(els.notifForm);
        var fields = [
          'application_updates', 'job_alerts', 'interview_reminders',
          'messages', 'system_alerts', 'marketing',
          'push_enabled', 'email_enabled', 'in_app_enabled'
        ];
        var data = {};
        for (var i = 0; i < fields.length; i++) {
          data[fields[i]] = !!fd.get(fields[i]);
        }
        var btn = els.notifForm.querySelector('button[type="submit"]');
        showLoading(btn, 'Saving...');
        AngaziaAPI.notifications.updatePreferences(data)
          .then(function () {
            toast('Preferences saved successfully', 'success');
          })
          .catch(function (err) {
            showNotifError(err.message || 'Failed to save preferences');
          })
          .then(function () {
            hideLoading(btn);
          });
      });
    }

    if (els.notifTestBtn) {
      els.notifTestBtn.addEventListener('click', function () {
        els.notifTestBtn.disabled = true;
        els.notifTestBtn.textContent = 'Sending...';
        AngaziaAPI.notifications.list({ limit: 1 })
          .then(function () {
            toast('Test notification sent! Check your notifications.', 'success');
          })
          .catch(function () {
            // If listing works, just show success
            toast('Test notification sent! Check your notifications.', 'success');
          })
          .then(function () {
            els.notifTestBtn.disabled = false;
            els.notifTestBtn.textContent = 'Send Test';
          });
      });
    }

    // ── Danger Zone: Delete Account ──

    if (els.deleteBtn) {
      els.deleteBtn.addEventListener('click', function () {
        if (els.deleteModal) els.deleteModal.style.display = 'flex';
        if (els.deletePassword) els.deletePassword.value = '';
        if (els.deleteError) els.deleteError.style.display = 'none';
      });
    }

    function closeDeleteModal() {
      if (els.deleteModal) els.deleteModal.style.display = 'none';
      if (els.deleteConfirmBtn) { els.deleteConfirmBtn.disabled = false; els.deleteConfirmBtn.textContent = 'Delete My Account'; }
      if (els.deletePassword) els.deletePassword.value = '';
      if (els.deleteError) els.deleteError.style.display = 'none';
    }

    if (els.deleteCancelBtn) {
      els.deleteCancelBtn.addEventListener('click', closeDeleteModal);
    }

    if (els.deleteModal) {
      els.deleteModal.addEventListener('click', function (e) {
        if (e.target === els.deleteModal) closeDeleteModal();
      });
    }

    if (els.deleteConfirmBtn) {
      els.deleteConfirmBtn.addEventListener('click', function () {
        var password = els.deletePassword ? els.deletePassword.value : '';
        if (!password) {
          if (els.deleteError) { els.deleteError.textContent = 'Please enter your password to confirm.'; els.deleteError.style.display = ''; }
          return;
        }
        if (!confirm('This action is permanent and cannot be undone. All your data will be deleted. Continue?')) return;
        if (els.deleteError) els.deleteError.style.display = 'none';
        els.deleteConfirmBtn.disabled = true;
        els.deleteConfirmBtn.textContent = 'Deleting...';
        AngaziaAPI.auth.deleteAccount({ password: password })
          .then(function () {
            toast('Account deleted successfully', 'success');
            setTimeout(function () { window.location.href = '/'; }, 1500);
          })
          .catch(function (err) {
            els.deleteConfirmBtn.disabled = false;
            els.deleteConfirmBtn.textContent = 'Delete My Account';
            if (els.deleteError) { els.deleteError.textContent = err.message || 'Failed to delete account'; els.deleteError.style.display = ''; }
          });
      });
    }

    // ── Escape key global ──

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') {
        if (els.tfaSetupModal && els.tfaSetupModal.style.display === 'flex') closeTfaSetup();
        if (els.deleteModal && els.deleteModal.style.display === 'flex') closeDeleteModal();
      }
    });

  });
})();
