(function () {
  'use strict';

  function init() {
    initFilter();
    initStatusToggle();
    initDeleteButtons();
    initSearch();
  }

  function initFilter() {
    var sel = document.getElementById('job-status-filter');
    if (!sel) return;
    sel.addEventListener('change', function () {
      var val = this.value;
      var rows = document.querySelectorAll('.emp-job-row');
      rows.forEach(function (row) {
        if (!val || row.dataset.status === val) {
          row.style.display = '';
        } else {
          row.style.display = 'none';
        }
      });
    });
  }

  function initStatusToggle() {
    document.querySelectorAll('.emp-toggle-status').forEach(function (btn) {
      btn.addEventListener('click', function (e) {
        e.preventDefault();
        var jobId = this.dataset.jobId;
        var action = this.dataset.action;
        if (!jobId || typeof AngaziaAPI === 'undefined') return;

        if (action === 'close') {
          AngaziaAPI.jobs.close(jobId)
            .then(function () {
              showToast('Job closed successfully', 'success');
              location.reload();
            })
            .catch(function (err) {
              showToast(err.message || 'Failed to close job', 'error');
            });
        }
      });
    });
  }

  function initDeleteButtons() {
    document.querySelectorAll('.emp-delete-job').forEach(function (btn) {
      btn.addEventListener('click', function (e) {
        e.preventDefault();
        var jobId = this.dataset.jobId;
        var title = this.dataset.jobTitle || 'this job';
        if (!jobId) return;

        if (typeof AngaziaApp !== 'undefined' && AngaziaApp.confirmDialog) {
          AngaziaApp.confirmDialog('Are you sure you want to delete "' + title + '"? This action cannot be undone.')
            .then(function (confirmed) {
              if (!confirmed || typeof AngaziaAPI === 'undefined') return;
              AngaziaAPI.jobs.delete(jobId)
                .then(function () {
                  showToast('Job deleted', 'success');
                  var row = btn.closest('.emp-job-row');
                  if (row) row.remove();
                })
                .catch(function (err) {
                  showToast(err.message || 'Failed to delete job', 'error');
                });
            });
        }
      });
    });
  }

  function initSearch() {
    var input = document.getElementById('job-search');
    if (!input) return;
    if (typeof AngaziaSearch !== 'undefined') {
      AngaziaSearch.init(input, {
        endpoint: '/api/v1/search/jobs',
        minChars: 2,
        debounceMs: 300,
        onSelect: function (item) {
          if (item.id) window.location.href = '/employer/job-applications/' + item.id;
        },
        renderItem: function (item) {
          return '<span style="font-weight:500;">' + escapeHtml(item.title || '') + '</span>'
            + '<span style="font-size:11px;color:var(--muted,#6b7280);margin-left:8px;">' + escapeHtml(item.status || '') + '</span>';
        },
      });
    }
  }

  function showToast(msg, type) {
    if (typeof AngaziaApp !== 'undefined' && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
      return;
    }
    var container = document.getElementById('toast-container');
    if (!container) {
      container = document.createElement('div');
      container.id = 'toast-container';
      container.style.cssText = 'position:fixed;bottom:16px;right:16px;z-index:9999;display:flex;flex-direction:column;gap:8px;';
      document.body.appendChild(container);
    }
    var toast = document.createElement('div');
    var bg = type === 'success' ? '#00e5a0' : type === 'error' ? '#ef4444' : '#3b82f6';
    toast.style.cssText = 'background:' + bg + ';color:#fff;padding:12px 20px;border-radius:10px;font-size:13px;font-family:var(--fm,sans-serif);box-shadow:0 4px 16px rgba(0,0,0,0.15);animation:fadeIn 0.2s;';
    toast.textContent = msg;
    container.appendChild(toast);
    setTimeout(function () { toast.style.opacity = '0'; setTimeout(function () { toast.remove(); }, 200); }, 3500);
  }

  function escapeHtml(t) {
    if (!t) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(t));
    return d.innerHTML;
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
