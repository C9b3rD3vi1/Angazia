(function () {
  'use strict';

  var chartInstances = {};
  var currentPeriod = 30;

  function $(id) { return document.getElementById(id); }
  function qs(sel, ctx) { return (ctx || document).querySelector(sel); }
  function qsa(sel, ctx) { return (ctx || document).querySelectorAll(sel); }

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  function escapeHtml(str) {
    if (!str && str !== 0) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(String(str)));
    return div.innerHTML;
  }

  function showLoading() {
    var l = $('aan-loading');
    if (l) l.classList.add('active');
    var e = $('aan-error');
    if (e) e.classList.remove('active');
    var c = $('aan-content');
    if (c) c.classList.remove('active');
  }

  function hideLoading() {
    var l = $('aan-loading');
    if (l) l.classList.remove('active');
    var c = $('aan-content');
    if (c) c.classList.add('active');
  }

  function showError(msg) {
    var l = $('aan-loading');
    if (l) l.classList.remove('active');
    var c = $('aan-content');
    if (c) c.classList.remove('active');
    var e = $('aan-error');
    if (e) e.classList.add('active');
    var t = $('aan-error-text');
    if (t) t.textContent = msg;
  }

  function formatCurrency(amount) {
    return 'KES ' + Number(amount).toLocaleString('en-KE', { minimumFractionDigits: 0, maximumFractionDigits: 0 });
  }

  function setCardValue(parentId, stat, value) {
    var cards = qsa('[data-stat="' + stat + '"]', $(parentId) || document);
    for (var i = 0; i < cards.length; i++) {
      var valEl = cards[i].querySelector('.aan-card-value');
      if (valEl) valEl.textContent = value;
    }
  }

  function setGrowthRate(stat, rate) {
    var cards = qsa('[data-stat="' + stat + '"]', $('aan-row-growth') || document);
    for (var i = 0; i < cards.length; i++) {
      var valEl = cards[i].querySelector('.aan-card-value');
      var trendEl = cards[i].querySelector('.aan-trend');
      var formatted = (rate >= 0 ? '+' : '') + rate.toFixed(1) + '%';
      if (valEl) valEl.textContent = formatted;
      if (trendEl) {
        trendEl.textContent = formatted;
        trendEl.className = 'aan-trend ' + (rate >= 0 ? 'up' : 'down');
      }
    }
  }

  function destroyChart(key) {
    if (chartInstances[key]) {
      chartInstances[key].destroy();
      delete chartInstances[key];
    }
  }

  function hexToRgba(hex, alpha) {
    var r = parseInt(hex.slice(1, 3), 16);
    var g = parseInt(hex.slice(3, 5), 16);
    var b = parseInt(hex.slice(5, 7), 16);
    return 'rgba(' + r + ',' + g + ',' + b + ',' + alpha + ')';
  }

  var COLORS = {
    accent: '#00e5a0',
    info: '#3d9be9',
    danger: '#ff4f4f',
    warn: '#f5a623',
    purple: '#7c3aed',
    muted: '#6b7280',
    surface: '#1a1d23'
  };

  function loadPlatformStats() {
    return AngaziaAPI.admin.platformStats().then(function (d) {
      setCardValue('aan-row-primary', 'total_users', Number(d.total_users || 0).toLocaleString());
      setCardValue('aan-row-primary', 'total_jobs', Number(d.total_jobs || 0).toLocaleString());
      setCardValue('aan-row-primary', 'total_applications', Number(d.total_applications || 0).toLocaleString());
      setCardValue('aan-row-primary', 'total_revenue', formatCurrency(d.total_revenue || 0));
      setCardValue('aan-row-primary', 'mrr', formatCurrency(d.mrr || 0));
      setCardValue('aan-row-primary', 'average_match_score', (d.average_match_score || 0).toFixed(1) + '%');

      setCardValue('aan-row-growth', 'active_users_30_days', Number(d.active_users_30_days || 0).toLocaleString());
      setCardValue('aan-row-growth', 'new_users_7_days', Number(d.new_users_7_days || 0).toLocaleString());
      setCardValue('aan-row-growth', 'new_users_30_days', Number(d.new_users_30_days || 0).toLocaleString());

      setCardValue('aan-row-activity', 'jobs_posted_7_days', Number(d.jobs_posted_7_days || 0).toLocaleString());
      setCardValue('aan-row-activity', 'jobs_posted_30_days', Number(d.jobs_posted_30_days || 0).toLocaleString());
      setCardValue('aan-row-activity', 'total_profile_views', Number(d.total_profile_views || 0).toLocaleString());
      setCardValue('aan-row-activity', 'total_job_views', Number(d.total_job_views || 0).toLocaleString());
      setCardValue('aan-row-activity', 'verified_employers', Number(d.verified_employers || 0).toLocaleString());
      setCardValue('aan-row-activity', 'total_candidates', Number(d.total_candidates || 0).toLocaleString());

      setGrowthRate('user_growth_rate', parseFloat(d.user_growth_rate) || 0);
      setGrowthRate('job_growth_rate', parseFloat(d.job_growth_rate) || 0);
      setGrowthRate('application_growth_rate', parseFloat(d.application_growth_rate) || 0);
    }).catch(function (err) {
      showToast('Failed to load platform stats: ' + (err.message || 'Unknown error'), 'error');
    });
  }

  function loadUserStats() {
    return AngaziaAPI.admin.userStats().then(function (d) {
      var employee = parseInt(d.employee || d.candidates || d.role_employee || 0, 10);
      var employer = parseInt(d.employer || d.role_employer || 0, 10);
      var admin = parseInt(d.admin || d.role_admin || 0, 10);
      var active = parseInt(d.active || d.role_active || 0, 10);
      var suspended = parseInt(d.suspended || d.role_suspended || 0, 10);
      var total = employee + employer + admin;
      setCardValue('aan-row-users', 'total', total.toLocaleString());
      setCardValue('aan-row-users', 'employee', employee.toLocaleString());
      setCardValue('aan-row-users', 'employer', employer.toLocaleString());
      setCardValue('aan-row-users', 'admin', admin.toLocaleString());
      setCardValue('aan-row-users', 'active', active.toLocaleString());
      setCardValue('aan-row-users', 'suspended', suspended.toLocaleString());
      renderUserRoleChart({ employee: employee, employer: employer, admin: admin });
    }).catch(function (err) {
      showToast('Failed to load user stats: ' + (err.message || 'Unknown error'), 'error');
    });
  }

  function loadJobStats() {
    return AngaziaAPI.admin.jobStats().then(function (d) {
      var active = parseInt(d.status_active || d.active || 0, 10);
      var inactive = parseInt(d.status_inactive || d.inactive || 0, 10);
      var total = active + inactive;
      setCardValue('aan-row-jobs', 'status_active', active.toLocaleString());
      setCardValue('aan-row-jobs', 'status_inactive', inactive.toLocaleString());
      setCardValue('aan-row-jobs', 'total', total.toLocaleString());
      renderJobStatusChart({ status_active: active, status_inactive: inactive });
    }).catch(function (err) {
      showToast('Failed to load job stats: ' + (err.message || 'Unknown error'), 'error');
    });
  }

  function loadChartData(period) {
    return AngaziaAPI.admin.chartData({ period: period || currentPeriod }).then(function (d) {
      renderLineCharts(d);
    }).catch(function (err) {
      showToast('Failed to load chart data: ' + (err.message || 'Unknown error'), 'error');
    });
  }

  function renderLineCharts(data) {
    var commonOptions = function (label, color) {
      return {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          x: {
            display: true,
            ticks: { color: COLORS.muted, font: { size: 8 }, maxTicksLimit: 8 },
            grid: { display: false }
          },
          y: {
            display: true,
            beginAtZero: true,
            ticks: { color: COLORS.muted, font: { size: 8 }, precision: 0 },
            grid: { color: 'rgba(107,114,128,0.1)' }
          }
        }
      };
    };

    if (data.user_growth) {
      destroyChart('userGrowth');
      var ctx = $('aan-chart-users');
      if (ctx) {
        chartInstances.userGrowth = new Chart(ctx.getContext('2d'), {
          type: 'line',
          data: {
            labels: data.user_growth.map(function (p) { return p.date; }),
            datasets: [{
              label: 'Users',
              data: data.user_growth.map(function (p) { return p.count; }),
              borderColor: COLORS.accent,
              backgroundColor: hexToRgba(COLORS.accent, 0.1),
              fill: true,
              tension: 0.3,
              pointRadius: 2,
              pointHoverRadius: 4,
              borderWidth: 2
            }]
          },
          options: commonOptions('Users', COLORS.accent)
        });
      }
    }

    if (data.job_postings) {
      destroyChart('jobPostings');
      var ctx2 = $('aan-chart-jobs');
      if (ctx2) {
        chartInstances.jobPostings = new Chart(ctx2.getContext('2d'), {
          type: 'line',
          data: {
            labels: data.job_postings.map(function (p) { return p.date; }),
            datasets: [{
              label: 'Jobs',
              data: data.job_postings.map(function (p) { return p.count; }),
              borderColor: COLORS.info,
              backgroundColor: hexToRgba(COLORS.info, 0.1),
              fill: true,
              tension: 0.3,
              pointRadius: 2,
              pointHoverRadius: 4,
              borderWidth: 2
            }]
          },
          options: commonOptions('Jobs', COLORS.info)
        });
      }
    }

    if (data.applications) {
      destroyChart('applications');
      var ctx3 = $('aan-chart-apps');
      if (ctx3) {
        chartInstances.applications = new Chart(ctx3.getContext('2d'), {
          type: 'line',
          data: {
            labels: data.applications.map(function (p) { return p.date; }),
            datasets: [{
              label: 'Applications',
              data: data.applications.map(function (p) { return p.count; }),
              borderColor: COLORS.warn,
              backgroundColor: hexToRgba(COLORS.warn, 0.1),
              fill: true,
              tension: 0.3,
              pointRadius: 2,
              pointHoverRadius: 4,
              borderWidth: 2
            }]
          },
          options: commonOptions('Applications', COLORS.warn)
        });
      }
    }
  }

  function renderUserRoleChart(stats) {
    destroyChart('userRole');
    var ctx = $('aan-chart-user-role');
    if (!ctx) return;

    var labels = [];
    var values = [];
    var colors = [];

    var roleMap = {
      employee: { label: 'Candidates', color: COLORS.accent },
      employer: { label: 'Employers', color: COLORS.info },
      admin: { label: 'Admins', color: COLORS.purple }
    };

    for (var key in roleMap) {
      if (roleMap.hasOwnProperty(key)) {
        var val = parseInt(stats[key] || stats['role_' + key] || 0, 10);
        if (val > 0) {
          labels.push(roleMap[key].label);
          values.push(val);
          colors.push(roleMap[key].color);
        }
      }
    }

    if (values.length === 0) {
      ctx.parentElement.innerHTML = '<div class="aan-empty-chart">No data</div>';
      return;
    }

    chartInstances.userRole = new Chart(ctx.getContext('2d'), {
      type: 'doughnut',
      data: {
        labels: labels,
        datasets: [{
          data: values,
          backgroundColor: colors,
          borderColor: COLORS.surface,
          borderWidth: 2
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'bottom',
            labels: { color: COLORS.muted, font: { size: 9 }, padding: 8 }
          }
        }
      }
    });
  }

  function renderJobStatusChart(stats) {
    destroyChart('jobStatus');
    var ctx = $('aan-chart-job-status');
    if (!ctx) return;

    var labels = [];
    var values = [];
    var colors = [];

    var statusMap = {
      status_active: { label: 'Active', color: COLORS.accent },
      status_inactive: { label: 'Inactive', color: COLORS.danger }
    };

    for (var key in statusMap) {
      if (statusMap.hasOwnProperty(key)) {
        var val = parseInt(stats[key] || 0, 10);
        if (val > 0) {
          labels.push(statusMap[key].label);
          values.push(val);
          colors.push(statusMap[key].color);
        }
      }
    }

    if (values.length === 0) {
      ctx.parentElement.innerHTML = '<div class="aan-empty-chart">No data</div>';
      return;
    }

    chartInstances.jobStatus = new Chart(ctx.getContext('2d'), {
      type: 'doughnut',
      data: {
        labels: labels,
        datasets: [{
          data: values,
          backgroundColor: colors,
          borderColor: COLORS.surface,
          borderWidth: 2
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'bottom',
            labels: { color: COLORS.muted, font: { size: 9 }, padding: 8 }
          }
        }
      }
    });
  }

  function loadAll() {
    showLoading();
    Promise.all([
      loadPlatformStats(),
      loadUserStats(),
      loadJobStats(),
      loadChartData(currentPeriod)
    ]).then(function () {
      hideLoading();
    }).catch(function (err) {
      showError(err && err.message ? err.message : 'Failed to load analytics data');
    });
  }

  function setPeriod(period) {
    currentPeriod = period;
    var tabs = qsa('.aan-period-tab');
    for (var i = 0; i < tabs.length; i++) {
      var p = parseInt(tabs[i].getAttribute('data-period'), 10);
      tabs[i].classList.toggle('active', p === period);
    }
    loadChartData(period);
  }

  document.addEventListener('DOMContentLoaded', function () {
    loadAll();

    var refreshBtns = qsa('[data-action="aan-load"]');
    for (var i = 0; i < refreshBtns.length; i++) {
      refreshBtns[i].addEventListener('click', function (e) {
        e.preventDefault();
        loadAll();
      });
    }

    var periodTabs = qsa('.aan-period-tab');
    for (var j = 0; j < periodTabs.length; j++) {
      periodTabs[j].addEventListener('click', function () {
        var p = parseInt(this.getAttribute('data-period'), 10);
        if (p !== currentPeriod) setPeriod(p);
      });
    }
  });
})();
