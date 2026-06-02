(function () {
  'use strict';

  function init() {
    initLogoUpload();
    initFormSave();
    initVerificationRequest();
  }

  function initLogoUpload() {
    var input = document.getElementById('logo-upload');
    var preview = document.getElementById('logo-preview');
    if (!input) return;
    input.addEventListener('change', function () {
      var file = this.files[0];
      if (!file) return;
      if (!file.type.match('image.*')) {
        showToast('Please select an image file', 'error');
        return;
      }
      if (file.size > 2 * 1024 * 1024) {
        showToast('Image must be under 2MB', 'error');
        return;
      }
      var reader = new FileReader();
      reader.onload = function (e) {
        if (preview) {
          preview.src = e.target.result;
          preview.style.display = '';
        }
        if (typeof AngaziaAPI !== 'undefined') {
          var fd = new FormData();
          fd.append('logo', file);
          AngaziaAPI.companies.uploadLogo(fd, function (pct) {
            var bar = document.getElementById('upload-progress');
            if (bar) { bar.style.width = pct + '%'; bar.style.display = ''; }
          })
            .then(function () {
              showToast('Logo uploaded successfully', 'success');
            })
            .catch(function (err) {
              showToast(err.body && err.body.error ? err.body.error : 'Upload failed', 'error');
            });
        }
      };
      reader.readAsDataURL(file);
    });
  }

  function initFormSave() {
    var form = document.getElementById('company-form');
    if (!form) return;
    form.addEventListener('submit', function (e) {
      e.preventDefault();
      if (typeof AngaziaAPI === 'undefined') return;
      var data = {};
      var fd = new FormData(form);
      fd.forEach(function (v, k) { data[k] = v; });

      var btn = form.querySelector('[type="submit"]');
      if (btn) { btn.disabled = true; btn.textContent = 'Saving...'; }

      AngaziaAPI.companies.updateCompany(data)
        .then(function () {
          showToast('Company profile updated', 'success');
          if (btn) { btn.disabled = false; btn.textContent = 'Save Changes'; }
        })
        .catch(function (err) {
          showToast(err.body && err.body.error ? err.body.error : 'Save failed', 'error');
          if (btn) { btn.disabled = false; btn.textContent = 'Save Changes'; }
          if (typeof AngaziaApp !== 'undefined' && err.body && err.body.errors) {
            AngaziaApp.handleFormErrors(form, err.body.errors);
          }
        });
    });
  }

  function initVerificationRequest() {
    var btn = document.getElementById('request-verify');
    if (!btn) return;
    btn.addEventListener('click', function () {
      if (typeof AngaziaAPI === 'undefined') return;
      btn.disabled = true;
      btn.textContent = 'Requesting...';
      AngaziaAPI.companies.verify()
        .then(function () {
          showToast('Verification request submitted', 'success');
          btn.textContent = 'Verification Requested';
        })
        .catch(function (err) {
          showToast(err.body && err.body.error ? err.body.error : 'Request failed', 'error');
          btn.disabled = false;
          btn.textContent = 'Request Verification';
        });
    });
  }

  function showToast(msg, type) {
    if (typeof AngaziaApp !== 'undefined' && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
      return;
    }
    var c = document.getElementById('toast-container');
    if (!c) { c = document.createElement('div'); c.id = 'toast-container'; c.style.cssText = 'position:fixed;bottom:16px;right:16px;z-index:9999;display:flex;flex-direction:column;gap:8px;'; document.body.appendChild(c); }
    var t = document.createElement('div');
    var bg = type === 'success' ? '#00e5a0' : type === 'error' ? '#ef4444' : '#3b82f6';
    t.style.cssText = 'background:' + bg + ';color:#fff;padding:12px 20px;border-radius:10px;font-size:13px;font-family:var(--fm,sans-serif);box-shadow:0 4px 16px rgba(0,0,0,0.15);';
    t.textContent = msg;
    c.appendChild(t);
    setTimeout(function () { t.style.opacity = '0'; setTimeout(function () { t.remove(); }, 200); }, 3500);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
