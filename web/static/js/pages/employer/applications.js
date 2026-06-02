(function () {
  'use strict';

  var selectedIds = [];
  var currentFilters = {};

  function init() {
    initFilterBar();
    initStatusActions();
    initBulkActions();
    initSelectAll();
    initIndividualSelect();
    initSearch();
  }

  function initFilterBar() {
    if (typeof AngaziaFilters === 'undefined') return;
    var container = document.getElementById('app-filters');
    if (!container) return;

    AngaziaFilters.create(container, [
      { key: 'status', label: 'Status', type: 'select', options: [
        { value: 'new', label: 'New' },
        { value: 'reviewed', label: 'Reviewed' },
        { value: 'shortlisted', label: 'Shortlisted' },
        { value: 'interview', label: 'Interview' },
        { value: 'rejected', label: 'Rejected' },
        { value: 'hired', label: 'Hired' },
      ]},
      { key: 'job_id', label: 'Job', type: 'select', options: getJobOptions() },
      { key: 'date_from', label: 'From Date', type: 'text', placeholder: 'YYYY-MM-DD' },
      { key: 'date_to', label: 'To Date', type: 'text', placeholder: 'YYYY-MM-DD' },
    ], function (vals) {
      currentFilters = vals;
      loadApplications(vals);
    });
  }

  function getJobOptions() {
    var opts = [];
    document.querySelectorAll('[data-job-option]').forEach(function (el) {
      opts.push({ value: el.dataset.jobOption, label: el.textContent });
    });
    return opts;
  }

  function loadApplications(filters) {
    if (typeof AngaziaAPI === 'undefined') return;
    var params = Object.assign({ limit: 50 }, filters);
    AngaziaAPI.applications.companyApplications(params)
      .then(function (data) {
        var list = data.applications || data.data || data || [];
        renderTable(list);
      })
      .catch(function () {});
  }

  function renderTable(applications) {
    var tbody = document.getElementById('app-table-body');
    if (!tbody) return;
    if (!applications.length) {
      tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--muted,#6b7280);font-size:13px;">No applications found</td></tr>';
      return;
    }
    tbody.innerHTML = applications.map(function (a) {
      var initials = (a.candidate_name || a.name || '?').substring(0, 2).toUpperCase();
      return '<tr class="emp-app-row" data-id="' + a.id + '">'
        + '<td><input type="checkbox" class="emp-app-cb" value="' + a.id + '"></td>'
        + '<td><div class="emp-app-candidate">'
        + '<span class="emp-app-avatar">' + initials + '</span>'
        + '<div><a href="/employer/applications/' + a.id + '" class="emp-app-name">' + escapeHtml(a.candidate_name || a.name || 'Unknown') + '</a>'
        + '<span class="emp-app-email">' + escapeHtml(a.candidate_email || '') + '</span></div></div></td>'
        + '<td><span class="emp-table-muted">' + escapeHtml(a.job_title || '') + '</span></td>'
        + '<td><span class="emp-status-badge ' + (a.status || 'new') + '">' + (a.status || 'new') + '</span></td>'
        + '<td><span class="emp-table-muted">' + (a.applied_at ? timeAgo(a.applied_at) : '') + '</span></td>'
        + '<td><div class="emp-app-actions">'
        + '<button class="emp-action-btn" data-action="shortlist" data-id="' + a.id + '" title="Shortlist">&#x2B50;</button>'
        + '<button class="emp-action-btn" data-action="reject" data-id="' + a.id + '" title="Reject">&#x274C;</button>'
        + '<button class="emp-action-btn" data-action="interview" data-id="' + a.id + '" title="Schedule Interview">&#x1F4C5;</button>'
        + '</div></td></tr>';
    }).join('');
    selectedIds = [];
    initStatusActions();
    initIndividualSelect();
  }

  function initStatusActions() {
    document.querySelectorAll('.emp-action-btn').forEach(function (btn) {
      btn.addEventListener('click', function (e) {
        e.stopPropagation();
        var id = this.dataset.id;
        var action = this.dataset.action;
        if (!id || typeof AngaziaAPI === 'undefined') return;

        var actions = {
          shortlist: function () { return AngaziaAPI.applications.shortlist(id); },
          reject: function () { return AngaziaAPI.applications.reject(id); },
          interview: function () {
            var date = prompt('Interview date (YYYY-MM-DD):');
            if (!date) return Promise.reject();
            return AngaziaAPI.applications.interview(id, { scheduled_at: date });
          },
        };

        var fn = actions[action];
        if (!fn) return;
        fn()
          .then(function () {
            showToast('Application ' + action + 'ed', 'success');
            var row = btn.closest('.emp-app-row');
            if (row) {
              var badge = row.querySelector('.emp-status-badge');
              if (badge) badge.textContent = action;
            }
          })
          .catch(function (err) {
            if (err) showToast(err.message || 'Action failed', 'error');
          });
      });
    });
  }

  function initBulkActions() {
    document.querySelectorAll('.emp-bulk-action').forEach(function (btn) {
      btn.addEventListener('click', function () {
        if (!selectedIds.length) {
          showToast('Select at least one application', 'warning');
          return;
        }
        var action = this.dataset.action;
        if (action === 'bulk-shortlist' && typeof AngaziaAPI !== 'undefined') {
          AngaziaAPI.applications.bulkShortlist({ ids: selectedIds })
            .then(function () {
              showToast(selectedIds.length + ' applications shortlisted', 'success');
              selectedIds.forEach(function (id) {
                var row = document.querySelector('.emp-app-row[data-id="' + id + '"]');
                if (row) {
                  var badge = row.querySelector('.emp-status-badge');
                  if (badge) badge.textContent = 'shortlisted';
                }
              });
            })
            .catch(function (err) {
              showToast(err.message || 'Bulk action failed', 'error');
            });
        }
      });
    });
  }

  function initSelectAll() {
    var cb = document.getElementById('select-all-apps');
    if (!cb) return;
    cb.addEventListener('change', function () {
      var checked = this.checked;
      document.querySelectorAll('.emp-app-cb').forEach(function (c) {
        c.checked = checked;
        if (checked) {
          if (selectedIds.indexOf(c.value) === -1) selectedIds.push(c.value);
        } else {
          selectedIds = [];
        }
      });
    });
  }

  function initIndividualSelect() {
    document.querySelectorAll('.emp-app-cb').forEach(function (cb) {
      cb.addEventListener('change', function () {
        var id = this.value;
        if (this.checked) {
          if (selectedIds.indexOf(id) === -1) selectedIds.push(id);
        } else {
          selectedIds = selectedIds.filter(function (s) { return s !== id; });
        }
      });
    });
  }

  function initSearch() {
    if (typeof AngaziaSearch === 'undefined') return;
    var input = document.getElementById('app-search');
    if (!input) return;
    AngaziaSearch.init(input, {
      endpoint: '/api/v1/search/candidates',
      minChars: 2,
      onSelect: function (item) {
        if (item.id) window.location.href = '/employer/applications?candidate=' + item.id;
      },
    });
  }

  function timeAgo(dateStr) {
    if (!dateStr) return '';
    var d = new Date(dateStr);
    var diff = Math.floor((Date.now() - d) / 1000);
    if (diff < 60) return 'just now';
    if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
    if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
    return Math.floor(diff / 86400) + 'd ago';
  }

  function showToast(msg, type) {
    if (typeof AngaziaApp !== 'undefined' && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
      return;
    }
    var c = document.getElementById('toast-container');
    if (!c) { c = document.createElement('div'); c.id = 'toast-container'; c.style.cssText = 'position:fixed;bottom:16px;right:16px;z-index:9999;display:flex;flex-direction:column;gap:8px;'; document.body.appendChild(c); }
    var t = document.createElement('div');
    var bg = type === 'success' ? '#00e5a0' : type === 'error' ? '#ef4444' : type === 'warning' ? '#f59e0b' : '#3b82f6';
    t.style.cssText = 'background:' + bg + ';color:#fff;padding:12px 20px;border-radius:10px;font-size:13px;font-family:var(--fm,sans-serif);box-shadow:0 4px 16px rgba(0,0,0,0.15);';
    t.textContent = msg;
    c.appendChild(t);
    setTimeout(function () { t.style.opacity = '0'; setTimeout(function () { t.remove(); }, 200); }, 3500);
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
