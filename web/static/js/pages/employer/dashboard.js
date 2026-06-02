(function () {
  'use strict';

  function init() {
    loadStats();
    loadChart();
    loadRecentApplicants();
    initPeriodSwitch();
    initActions();
  }

  function loadStats() {
    if (typeof AngaziaAPI === 'undefined') return;
    AngaziaAPI.analytics.employerDashboard()
      .then(function (data) {
        var stats = data.stats || data;
        var mapping = {
          'stat-active-jobs': stats.active_jobs,
          'stat-total-applicants': stats.total_applicants,
          'stat-new-applications': stats.new_applications,
          'stat-profile-views': stats.profile_views,
        };
        Object.keys(mapping).forEach(function (id) {
          var el = document.getElementById(id);
          if (el) el.textContent = mapping[id] != null ? mapping[id] : '0';
        });
      })
      .catch(function () {});
  }

  function loadChart() {
    if (typeof AngaziaAPI === 'undefined' || typeof AngaziaChart === 'undefined') return;
    AngaziaAPI.analytics.employerTrends()
      .then(function (data) {
        var points = data.trends || data.data || data || [];
        if (!points.length) {
          points = generateFallbackData();
        }
        var container = document.getElementById('perf-chart');
        if (!container) return;
        container.innerHTML = '';
        AngaziaChart.line(container, points, {
          height: 180,
          showGrid: true,
          curved: true,
          xKey: 'date',
          legend: false,
          showXLabels: true,
        });
      })
      .catch(function () {});
  }

  function generateFallbackData() {
    var data = [];
    for (var i = 29; i >= 0; i--) {
      var d = new Date();
      d.setDate(d.getDate() - i);
      data.push({
        date: d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        Applications: Math.floor(Math.random() * 20) + 5,
      });
    }
    return data;
  }

  function loadRecentApplicants() {
    if (typeof AngaziaAPI === 'undefined') return;
    AngaziaAPI.applications.companyApplications({ limit: 5, sort: 'newest' })
      .then(function (data) {
        var list = data.applications || data.data || data || [];
        var tbody = document.getElementById('recent-applicants-body');
        if (!tbody) return;
        if (!list.length) {
          tbody.innerHTML = '<tr><td colspan="4" class="emp-empty-cell">No applications yet</td></tr>';
          return;
        }
        tbody.innerHTML = list.map(function (a) {
          var initials = (a.candidate_name || a.name || '?').charAt(0).toUpperCase();
          return '<tr>'
            + '<td><a href="/employer/applications/' + a.id + '" class="emp-table-link">'
            + '<span class="emp-table-avatar emp-table-avatar-text">' + initials + '</span>'
            + '<span>' + escapeHtml(a.candidate_name || a.name || 'Unknown') + '</span></a></td>'
            + '<td><span class="emp-table-muted">' + escapeHtml(a.job_title || '') + '</span></td>'
            + '<td><span class="emp-table-muted">' + (a.applied_at ? timeAgo(a.applied_at) : '') + '</span></td>'
            + '<td><span class="emp-status-badge ' + (a.status || 'new') + '">' + (a.status || 'new') + '</span></td>'
            + '</tr>';
        }).join('');
      })
      .catch(function () {});
  }

  function initPeriodSwitch() {
    var sel = document.getElementById('chart-period');
    if (!sel) return;
    sel.addEventListener('change', function () { loadChart(); });
  }

  function initActions() {
    document.querySelectorAll('.emp-action-card').forEach(function (card) {
      card.addEventListener('click', function (e) {
        var href = this.getAttribute('href');
        if (href && href !== '#') return;
        e.preventDefault();
      });
    });
  }

  function timeAgo(dateStr) {
    if (!dateStr) return '';
    var date = new Date(dateStr);
    var now = new Date();
    var diff = Math.floor((now - date) / 1000);
    if (diff < 60) return 'just now';
    if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
    if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
    return Math.floor(diff / 86400) + 'd ago';
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
