'use strict';

(function () {
  function $(id) { return document.getElementById(id); }
  function qs(sel, ctx) { return (ctx || document).querySelector(sel); }
  function qsa(sel, ctx) { return (ctx || document).querySelectorAll(sel); }

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      var t = $('as-save-toast');
      if (t) {
        t.textContent = msg;
        t.style.display = 'inline-block';
        t.className = 'as-toast as-toast-' + (type || 'success');
        clearTimeout(t._hide);
        t._hide = setTimeout(function () { t.style.display = 'none'; }, 3000);
      }
    }
  }

  function showTfaError(el, m) {
    if (el) { el.textContent = m; el.style.display = ''; }
  }
  function hideTfaError(el) {
    if (el) el.style.display = 'none';
  }

  var saveTimers = {};

  function saveSetting(key, value, el) {
    var modEl = $('modified-' + key);
    return AngaziaAPI.admin.updateSetting(key, { value: value }).then(function () {
      if (el) el.setAttribute('data-original', value);
      if (modEl) modEl.classList.remove('active');
      showToast('"' + key + '" saved', 'success');
    }).catch(function (err) {
      var msg = (err && (err.body && err.body.message || err.message)) || 'Failed to save setting';
      showToast(msg, 'error');
      if (modEl) modEl.classList.remove('active');
      throw err;
    });
  }

  function markModified(key) {
    var modEl = $('modified-' + key);
    if (modEl) modEl.classList.add('active');
  }

  function isDirty(el) {
    if (!el) return false;
    if (el.type === 'checkbox') {
      return String(el.checked) !== el.getAttribute('data-original');
    }
    return el.value !== el.getAttribute('data-original');
  }

  document.addEventListener('DOMContentLoaded', function () {
    var loadingEl = $('as-loading');
    var errorEl = $('as-error');
    var contentEl = $('as-content');

    function hideLoading() {
      if (loadingEl) loadingEl.classList.remove('active');
      if (contentEl) contentEl.style.display = '';
    }

    function showError(msg) {
      if (loadingEl) loadingEl.classList.remove('active');
      if (contentEl) contentEl.style.display = 'none';
      if (errorEl) errorEl.classList.add('active');
      var et = $('as-error-text');
      if (et) et.textContent = msg || 'Failed to load settings.';
    }

    AngaziaAPI.admin.settings().then(function () {
      hideLoading();
    }).catch(function (err) {
      showError((err && err.message) || 'Failed to load settings');
    });

    function attachSaveHandlers() {
      qsa('[data-action="save-setting"]').forEach(function (el) {
        el.addEventListener('change', function () {
          if (this.type === 'checkbox') {
            var setting = this.closest('.as-setting');
            if (!setting) return;
            var key = setting.getAttribute('data-key');
            var val = String(this.checked);
            markModified(key);
            saveSetting(key, val, this);
          }
        });
        el.addEventListener('blur', function () {
          if (this.type !== 'checkbox' && isDirty(this)) {
            var setting = this.closest('.as-setting');
            if (!setting) return;
            var key = setting.getAttribute('data-key');
            var val = this.value;
            markModified(key);
            saveSetting(key, val, this);
          }
        });
        el.addEventListener('input', function () {
          if (this.type !== 'checkbox') {
            var setting = this.closest('.as-setting');
            if (!setting) return;
            var key = setting.getAttribute('data-key');
            if (isDirty(this)) {
              markModified(key);
            }
          }
        });
        el.addEventListener('keydown', function (e) {
          if (e.key === 'Enter' && this.type !== 'checkbox' && this.tagName !== 'TEXTAREA') {
            e.preventDefault();
            this.blur();
          }
        });
      });

      qsa('[data-action="save-json"]').forEach(function (btn) {
        btn.addEventListener('click', function () {
          var setting = this.closest('.as-setting');
          if (!setting) return;
          var key = setting.getAttribute('data-key');
          var textarea = setting.querySelector('textarea');
          if (textarea) {
            markModified(key);
            saveSetting(key, textarea.value, textarea);
          }
        });
      });
    }

    attachSaveHandlers();

    var retryBtn = $('as-retry-btn');
    if (retryBtn) {
      retryBtn.addEventListener('click', function () {
        if (errorEl) errorEl.classList.remove('active');
        if (loadingEl) loadingEl.classList.add('active');
        window.location.reload();
      });
    }

    var activeCategory = 'all';
    var searchQuery = '';

    function filterSettings() {
      var settings = qsa('.as-setting');
      var visibleCount = 0;
      settings.forEach(function (s) {
        var cat = s.getAttribute('data-category') || 'general';
        var key = s.getAttribute('data-key') || '';
        var desc = (s.querySelector('.as-setting-desc') || {}).textContent || '';
        var matchesCategory = activeCategory === 'all' || cat === activeCategory;
        var matchesSearch = !searchQuery ||
          key.toLowerCase().indexOf(searchQuery) !== -1 ||
          desc.toLowerCase().indexOf(searchQuery) !== -1;
        var show = matchesCategory && matchesSearch;
        s.style.display = show ? '' : 'none';
        if (show) visibleCount++;
      });

      var emptyMsg = qs('.as-empty');
      if (settings.length === 0) return;
      if (visibleCount === 0) {
        if (!emptyMsg) {
          var list = $('as-settings-list');
          if (list) {
            var div = document.createElement('div');
            div.className = 'as-empty';
            div.innerHTML = '<div class="as-empty-icon">S</div><p class="as-empty-desc">No settings match your filter.</p>';
            list.appendChild(div);
          }
        } else {
          emptyMsg.style.display = '';
        }
      } else {
        if (emptyMsg) emptyMsg.style.display = 'none';
      }

      var tabs = qsa('.as-category-tab');
      tabs.forEach(function (tab) {
        var tabCat = tab.getAttribute('data-category');
        var countEl = tab.querySelector('.as-count-badge');
        if (tabCat === 'all') {
          if (countEl) countEl.textContent = visibleCount + '/' + settings.length;
        } else {
          var catCount = settings.filter(function (s) {
            return (s.getAttribute('data-category') || 'general') === tabCat;
          }).length;
          if (countEl) countEl.textContent = catCount;
        }
      });
    }

    var searchInput = $('as-search');
    if (searchInput) {
      searchInput.addEventListener('input', function () {
        searchQuery = this.value.toLowerCase().trim();
        filterSettings();
      });
    }

    var tabContainer = $('as-category-tabs');
    if (tabContainer) {
      tabContainer.addEventListener('click', function (e) {
        var tab = e.target.closest('.as-category-tab');
        if (!tab) return;
        var cat = tab.getAttribute('data-category');
        if (cat === activeCategory) return;
        activeCategory = cat;
        qsa('.as-category-tab', tabContainer).forEach(function (t) {
          t.classList.toggle('active', t.getAttribute('data-category') === cat);
        });
        filterSettings();
      });
    }

    var enableBtn = $('tfa-enable-btn');
    if (enableBtn) {
      enableBtn.addEventListener('click', function () {
        window.open('/auth/2fa/setup', '_blank');
      });
    }

    var disableBtn = $('tfa-disable-btn');
    if (disableBtn) {
      disableBtn.addEventListener('click', function () {
        if (!confirm('Are you sure you want to disable two-factor authentication?')) return;
        var code = prompt('Enter your current TOTP code (or a backup code):');
        if (!code) return;
        var password = prompt('Confirm your password:');
        if (!password) return;
        var tfaError = $('tfa-error');
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

    var viewBackupBtn = $('tfa-view-backup-btn');
    if (viewBackupBtn) {
      viewBackupBtn.addEventListener('click', function () {
        var tfaError = $('tfa-error');
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

    var modalOverlay = $('as-modal-overlay');
    var addBtn = $('as-add-btn');
    var modalClose = $('as-modal-close');
    var modalCancel = $('as-modal-cancel');
    var modalSave = $('as-modal-save');
    var createError = $('as-create-error');

    function showModal() {
      if (modalOverlay) modalOverlay.style.display = 'flex';
      if (createError) createError.style.display = 'none';
    }
    function hideModal() {
      if (modalOverlay) modalOverlay.style.display = 'none';
    }

    if (addBtn) addBtn.addEventListener('click', showModal);
    if (modalClose) modalClose.addEventListener('click', hideModal);
    if (modalCancel) modalCancel.addEventListener('click', hideModal);
    if (modalOverlay) modalOverlay.addEventListener('click', function (e) {
      if (e.target === modalOverlay) hideModal();
    });

    if (modalSave) {
      modalSave.addEventListener('click', function () {
        var key = $('as-new-key');
        var value = $('as-new-value');
        var type = $('as-new-type');
        var category = $('as-new-category');
        var desc = $('as-new-desc');

        if (!key || !key.value.trim()) {
          if (createError) { createError.textContent = 'Key is required'; createError.style.display = ''; }
          return;
        }
        if (!value || !value.value.trim()) {
          if (createError) { createError.textContent = 'Value is required'; createError.style.display = ''; }
          return;
        }

        if (createError) createError.style.display = 'none';
        modalSave.disabled = true;
        modalSave.textContent = 'Creating...';

        AngaziaAPI.admin.createSetting({
          key: key.value.trim(),
          value: value.value.trim(),
          type: type ? type.value : 'string',
          category: category ? category.value.trim() : 'general',
          description: desc ? desc.value.trim() : ''
        }).then(function () {
          hideModal();
          window.location.reload();
        }).catch(function (err) {
          modalSave.disabled = false;
          modalSave.textContent = 'Create Setting';
          var msg = (err && (err.body && err.body.message || err.message)) || 'Failed to create setting';
          if (createError) { createError.textContent = msg; createError.style.display = ''; }
        });
      });
    }

    var loadingEl2 = $('tfa-status-loading');
    var notEnabled = $('tfa-not-enabled');
    var enabledDiv = $('tfa-enabled');
    AngaziaAPI.auth.twoFA.status().then(function (data) {
      if (loadingEl2) loadingEl2.style.display = 'none';
      if (data && data.enabled) {
        if (notEnabled) notEnabled.style.display = 'none';
        if (enabledDiv) enabledDiv.style.display = '';
      } else {
        if (notEnabled) notEnabled.style.display = '';
        if (enabledDiv) enabledDiv.style.display = 'none';
      }
    }).catch(function () {
      if (loadingEl2) loadingEl2.style.display = 'none';
      if (notEnabled) notEnabled.style.display = '';
      if (enabledDiv) enabledDiv.style.display = 'none';
    });
  });
})();
