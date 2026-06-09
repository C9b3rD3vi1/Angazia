(function () {
  'use strict';

  var trendsChart = null;
  var sourcesChart = null;

  var els = {};

  var chartColors = {
    purple: '#7c3aed',
    indigo: '#6366f1',
    blue: '#3b82f6',
    amber: '#f59e0b',
    green: '#10b981',
    pink: '#ec4899',
    red: '#ef4444',
    teal: '#14b8a6'
  };

  function cacheEls() {
    els = {
      loading: document.getElementById('an-loading'),
      error: document.getElementById('an-error'),
      errorMsg: document.getElementById('an-error-msg'),
      content: document.getElementById('an-content'),
      retryBtn: document.getElementById('an-retry-btn'),
      refreshBtn: document.getElementById('an-refresh-btn'),
      period: document.getElementById('an-period'),

      statViews: document.getElementById('an-stat-views'),
      statApplications: document.getElementById('an-stat-applications'),
      statApplyRate: document.getElementById('an-stat-apply-rate'),
      statResponseRate: document.getElementById('an-stat-response-rate'),
      statShortlisted: document.getElementById('an-stat-shortlisted'),
      statHired: document.getElementById('an-stat-hired'),

      trendViews: document.getElementById('an-trend-views'),
      trendApplications: document.getElementById('an-trend-applications'),
      trendApplyRate: document.getElementById('an-trend-apply-rate'),
      trendResponseRate: document.getElementById('an-trend-response-rate'),
      trendShortlisted: document.getElementById('an-trend-shortlisted'),
      trendHired: document.getElementById('an-trend-hired'),

      funnelViewed: document.getElementById('an-funnel-viewed'),
      funnelApplied: document.getElementById('an-funnel-applied'),
      funnelShortlisted: document.getElementById('an-funnel-shortlisted'),
      funnelInterviewed: document.getElementById('an-funnel-interviewed'),
      funnelHired: document.getElementById('an-funnel-hired'),
      funnelBarViewed: document.getElementById('an-funnel-bar-viewed'),
      funnelBarApplied: document.getElementById('an-funnel-bar-applied'),
      funnelBarShortlisted: document.getElementById('an-funnel-bar-shortlisted'),
      funnelBarInterviewed: document.getElementById('an-funnel-bar-interviewed'),
      funnelBarHired: document.getElementById('an-funnel-bar-hired'),
      funnelOverall: document.getElementById('an-funnel-overall'),

      topJobs: document.getElementById('an-top-jobs'),
      quality: document.getElementById('an-quality'),
      timeToHire: document.getElementById('an-time-to-hire'),
      recentApps: document.getElementById('an-recent-apps'),

      trendsCanvas: document.getElementById('an-trends-chart'),
      sourcesCanvas: document.getElementById('an-sources-chart')
    };
  }

  function showLoading() {
    if (els.loading) els.loading.style.display = '';
    if (els.error) els.error.style.display = 'none';
    if (els.content) els.content.style.display = 'none';
  }

  function showError(msg) {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) {
      els.error.style.display = '';
      if (els.errorMsg) els.errorMsg.textContent = msg || 'An unexpected error occurred.';
    }
    if (els.content) els.content.style.display = 'none';
  }

  function showContent() {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) els.error.style.display = 'none';
    if (els.content) els.content.style.display = '';
  }

  function showToast(msg, type) {
    if (window.AngaziaApp && window.AngaziaApp.showToast) {
      window.AngaziaApp.showToast(msg, type);
      return;
    }
    var toast = document.createElement('div');
    toast.style.cssText = 'position:fixed;bottom:20px;right:20px;background:' +
      (type === 'success' ? '#10b981' : type === 'error' ? '#ef4444' : '#3b82f6') +
      ';color:#fff;padding:12px 20px;border-radius:8px;font-size:13px;z-index:9999;animation:slideIn 0.3s ease;';
    toast.textContent = msg;
    document.body.appendChild(toast);
    setTimeout(function () {
      toast.style.opacity = '0';
      setTimeout(function () { toast.remove(); }, 300);
    }, 3000);
  }

  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(function (w) { return w[0]; }).join('').toUpperCase().slice(0, 2);
  }

  function formatDate(dateStr) {
    if (!dateStr) return '';
    try {
      return new Date(dateStr).toLocaleDateString('en-KE', { month: 'short', day: 'numeric', year: 'numeric' });
    } catch (e) { return dateStr; }
  }

  function timeAgo(dateStr) {
    if (!dateStr) return '';
    var diff = Date.now() - new Date(dateStr).getTime();
    var sec = Math.floor(diff / 1000);
    if (sec < 60) return 'just now';
    var min = Math.floor(sec / 60);
    if (min < 60) return min + 'm ago';
    var hr = Math.floor(min / 60);
    if (hr < 24) return hr + 'h ago';
    var d = Math.floor(hr / 24);
    if (d < 7) return d + 'd ago';
    return new Date(dateStr).toLocaleDateString();
  }

  function escapeHtml(text) {
    if (!text) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(text));
    return d.innerHTML;
  }

  function setStatValue(el, val) {
    if (el) el.textContent = val || 0;
  }

  function animateCount(el, target, suffix) {
    if (!el) return;
    suffix = suffix || '';
    var start = 0;
    var duration = 600;
    var startTime = null;
    function step(timestamp) {
      if (!startTime) startTime = timestamp;
      var progress = Math.min((timestamp - startTime) / duration, 1);
      var eased = 1 - Math.pow(1 - progress, 3);
      var current = Math.round(start + (target - start) * eased);
      el.textContent = current + suffix;
      if (progress < 1) requestAnimationFrame(step);
    }
    requestAnimationFrame(step);
  }

  function animateFunnelBars() {
    var bars = document.querySelectorAll('.emp-funnel-bar');
    bars.forEach(function (b, i) {
      var w = b.getAttribute('data-width');
      if (w) {
        setTimeout(function () {
          b.style.width = w + '%';
        }, 100 + i * 60);
      }
    });
  }

  function animatePerfBars() {
    var bars = document.querySelectorAll('.emp-perf-bar span');
    bars.forEach(function (b, i) {
      var w = b.getAttribute('data-width');
      if (w) {
        setTimeout(function () {
          b.style.width = w + '%';
        }, 150 + i * 50);
      }
    });
  }

  function staggerIn(selector, delay) {
    var els = document.querySelectorAll(selector);
    els.forEach(function (el, i) {
      el.style.opacity = '0';
      el.style.transform = 'translateY(10px)';
      setTimeout(function () {
        el.style.transition = 'opacity 0.4s ease, transform 0.4s ease';
        el.style.opacity = '1';
        el.style.transform = 'translateY(0)';
      }, delay || 60 * i);
    });
  }

  function setTrend(el, value, isPositiveGood) {
    if (!el) return;
    if (value === null || value === undefined) {
      el.style.display = 'none';
      return;
    }
    el.style.display = '';
    var prefix = value > 0 ? '+' : '';
    el.textContent = prefix + value.toFixed(1) + '%';
    if (value > 0) {
      el.className = 'emp-stat-trend ' + (isPositiveGood !== false ? 'up' : 'down');
    } else if (value < 0) {
      el.className = 'emp-stat-trend ' + (isPositiveGood !== false ? 'down' : 'up');
    } else {
      el.className = 'emp-stat-trend flat';
    }
  }

  function renderNumber(n) {
    if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
    if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
    return n;
  }

  // ====== LOADERS ======

  async function loadStats() {
    try {
      var data = await AngaziaAPI.analytics.employerDashboard();

      animateCount(els.statViews, data.profile_views || 0);
      animateCount(els.statApplications, data.total_applicants || 0);
      animateCount(els.statShortlisted, data.shortlisted_count || 0);
      animateCount(els.statHired, data.hired_count || 0);

      var appRate = data.profile_views > 0
        ? ((data.total_applicants / data.profile_views) * 100).toFixed(1)
        : 0;
      animateCount(els.statApplyRate, parseFloat(appRate), '%');
      setStatValue(els.statResponseRate, '—');

      return data;
    } catch (err) {
      console.error('loadStats failed:', err);
      return null;
    }
  }

  async function loadTrends() {
    try {
      var days = parseInt(els.period ? els.period.value : '30', 10);
      var data = await AngaziaAPI.analytics.employerTrends({ period: 'daily', duration: days });

      var daily = data.daily || [];
      if (daily.length === 0) {
        if (trendsChart) { trendsChart.destroy(); trendsChart = null; }
        return;
      }
      renderTrendsChart(daily);
    } catch (err) {
      console.error('loadTrends failed:', err);
    }
  }

  function renderTrendsChart(daily) {
    var canvas = els.trendsCanvas;
    if (!canvas) return;
    var ctx = canvas.getContext('2d');
    if (!ctx) return;

    if (trendsChart) { trendsChart.destroy(); trendsChart = null; }

    var labels = daily.map(function (d) { return d.date ? d.date.slice(5) : ''; });
    var totals = daily.map(function (d) { return d.total || 0; });

    trendsChart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: labels,
        datasets: [{
          label: 'Applications',
          data: totals,
          borderColor: chartColors.purple,
          backgroundColor: 'rgba(124, 58, 237, 0.08)',
          borderWidth: 2,
          fill: true,
          tension: 0.4,
          pointRadius: 2,
          pointHoverRadius: 5,
          pointBackgroundColor: chartColors.purple
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: true,
        plugins: {
          legend: { display: false },
          tooltip: {
            mode: 'index',
            intersect: false,
            backgroundColor: 'rgba(0,0,0,0.85)',
            titleColor: '#fff',
            bodyColor: '#ddd',
            callbacks: {
              label: function (ctx) { return 'Applications: ' + ctx.raw; }
            }
          }
        },
        scales: {
          y: { beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { stepSize: 1 } },
          x: { grid: { display: false }, ticks: { maxRotation: 45, minRotation: 0, maxTicksLimit: 15 } }
        }
      }
    });
  }

  async function loadFunnel() {
    try {
      var data = await AngaziaAPI.analytics.employerFunnel();
      var stages = data.stages || [];

      var findCount = function (name) {
        var s = stages.find(function (st) { return st.stage === name; });
        return s ? s.count : 0;
      };

      var viewed = findCount('viewed');
      var applied = findCount('applied');
      var shortlisted = findCount('shortlisted');
      var interviewed = findCount('interview');
      var hired = findCount('hired');

      var total = viewed || applied;
      var maxVal = Math.max(viewed, applied, shortlisted, interviewed, hired, 1);

      setStatValue(els.funnelViewed, viewed);
      setStatValue(els.funnelApplied, applied);
      setStatValue(els.funnelShortlisted, shortlisted);
      setStatValue(els.funnelInterviewed, interviewed);
      setStatValue(els.funnelHired, hired);

      var setBar = function (el, val) {
        if (el) {
          var pct = (val / maxVal * 100);
          el.setAttribute('data-width', pct.toFixed(1));
          el.style.width = '0';
        }
      };
      setBar(els.funnelBarViewed, viewed);
      setBar(els.funnelBarApplied, applied);
      setBar(els.funnelBarShortlisted, shortlisted);
      setBar(els.funnelBarInterviewed, interviewed);
      setBar(els.funnelBarHired, hired);
      animateFunnelBars();

      if (els.funnelOverall) {
        var overall = data.overall_rate || (total > 0 ? (hired / total * 100) : 0);
        els.funnelOverall.textContent = overall.toFixed(1) + '%';
      }

      // Update response rate stat from funnel data
      var respRate = applied > 0 ? ((viewed / applied) * 100) : 0;
      setStatValue(els.statResponseRate, respRate.toFixed(1) + '%');
    } catch (err) {
      console.error('loadFunnel failed:', err);
    }
  }

  async function loadTopJobs() {
    try {
      var data = await AngaziaAPI.analytics.employerJobs();
      var jobs = Array.isArray(data) ? data : (data.jobs || []);
      if (!els.topJobs) return;

      if (jobs.length === 0) {
        els.topJobs.innerHTML = '<div class="emp-empty-state"><div class="emp-empty-icon">📋</div><div class="emp-empty-title">No jobs yet</div><div class="emp-empty-desc">Post a job to start tracking performance.</div></div>';
        return;
      }

      var top = jobs.slice(0, 10).sort(function (a, b) {
        return (b.applications || 0) - (a.applications || 0);
      });
      var maxApps = Math.max.apply(null, top.map(function (j) { return j.applications || 0; })) || 1;

      els.topJobs.innerHTML = '<div class="emp-perf-list">' + top.map(function (j) {
        var pct = (j.applications || 0) / maxApps * 100;
        return '<div class="emp-perf-item">' +
          '<div class="emp-perf-info">' +
            '<span class="emp-perf-title">' + escapeHtml(j.title) + '</span>' +
            '<span class="emp-perf-meta">' + (j.views || 0) + ' views · ' + (j.applications || 0) + ' apps</span>' +
          '</div>' +
          '<div class="emp-perf-bar-wrap"><span class="emp-perf-bar"><span data-width="' + pct.toFixed(1) + '" style="width:0"></span></span></div>' +
        '</div>';
      }).join('') + '</div>';
      animatePerfBars();
    } catch (err) {
      console.error('loadTopJobs failed:', err);
      if (els.topJobs) els.topJobs.innerHTML = '<div class="emp-empty-state"><div class="emp-empty-title">Failed to load</div></div>';
    }
  }

  async function loadQuality() {
    try {
      var data = await AngaziaAPI.analytics.quality();
      if (!els.quality) return;

      var avgScore = data.average_match_score || 0;
      var high = data.high_quality_count || 0;
      var med = data.medium_quality_count || 0;
      var low = data.low_quality_count || 0;
      var respTime = data.average_response_time || 0;
      var totalQual = high + med + low || 1;

      els.quality.innerHTML =
        '<div class="emp-quality-grid">' +
          '<div class="emp-quality-item"><div class="emp-quality-value">' + avgScore.toFixed(0) + '</div><div class="emp-quality-label">Avg Match Score</div></div>' +
          '<div class="emp-quality-item"><div class="emp-quality-value">' + respTime.toFixed(0) + 'h</div><div class="emp-quality-label">Avg Response Time</div></div>' +
          '<div class="emp-quality-item high"><div class="emp-quality-value">' + high + '</div><div class="emp-quality-label">High Quality</div></div>' +
          '<div class="emp-quality-item medium"><div class="emp-quality-value">' + med + '</div><div class="emp-quality-label">Medium Quality</div></div>' +
          '<div class="emp-quality-item low"><div class="emp-quality-value">' + low + '</div><div class="emp-quality-label">Low Quality</div></div>' +
          '<div class="emp-quality-item"><div class="emp-quality-value">' + (high / totalQual * 100).toFixed(0) + '%</div><div class="emp-quality-label">High Quality Rate</div></div>' +
        '</div>';
    } catch (err) {
      console.error('loadQuality failed:', err);
      if (els.quality) els.quality.innerHTML = '<div class="emp-empty-state"><div class="emp-empty-title">No data</div></div>';
    }
  }

  async function loadSources() {
    try {
      var data = await AngaziaAPI.analytics.sources();
      var sources = Array.isArray(data) ? data : [];
      var canvas = els.sourcesCanvas;
      if (!canvas || sources.length === 0) {
        if (sourcesChart) { sourcesChart.destroy(); sourcesChart = null; }
        return;
      }
      renderSourcesChart(sources);
    } catch (err) {
      console.error('loadSources failed:', err);
    }
  }

  function renderSourcesChart(sources) {
    var canvas = els.sourcesCanvas;
    if (!canvas) return;
    var ctx = canvas.getContext('2d');
    if (!ctx) return;

    if (sourcesChart) { sourcesChart.destroy(); sourcesChart = null; }

    var palette = [chartColors.purple, chartColors.blue, chartColors.amber, chartColors.green, chartColors.pink, chartColors.teal];

    sourcesChart = new Chart(ctx, {
      type: 'doughnut',
      data: {
        labels: sources.map(function (s) { return s.source || 'Unknown'; }),
        datasets: [{
          data: sources.map(function (s) { return s.count || 0; }),
          backgroundColor: sources.map(function (_, i) { return palette[i % palette.length]; }),
          borderWidth: 0
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: true,
        plugins: {
          legend: {
            position: 'bottom',
            labels: { color: '#9ca3af', padding: 10, font: { size: 10 }, boxWidth: 10, boxHeight: 10 }
          },
          tooltip: {
            callbacks: {
              label: function (ctx) {
                var total = ctx.dataset.data.reduce(function (a, b) { return a + b; }, 0);
                var pct = total > 0 ? ((ctx.raw / total) * 100).toFixed(1) : 0;
                return ctx.label + ': ' + ctx.raw + ' (' + pct + '%)';
              }
            }
          }
        }
      }
    });
  }

  async function loadTimeToHire() {
    try {
      var data = await AngaziaAPI.analytics.timeToHire();
      if (!els.timeToHire) return;

      var avg = data.average_days || 0;
      var med = data.median_days || 0;
      var min = data.min_days || 0;
      var max = data.max_days || 0;
      var byTitle = data.by_job_title || {};

      var titleKeys = Object.keys(byTitle);
      var html =
        '<div class="emp-tth-grid">' +
          '<div class="emp-tth-item"><div class="emp-tth-value">' + avg + '</div><div class="emp-tth-label">Avg Days</div></div>' +
          '<div class="emp-tth-item"><div class="emp-tth-value">' + med + '</div><div class="emp-tth-label">Median Days</div></div>' +
          '<div class="emp-tth-item"><div class="emp-tth-value">' + min + '</div><div class="emp-tth-label">Min Days</div></div>' +
          '<div class="emp-tth-item"><div class="emp-tth-value">' + max + '</div><div class="emp-tth-label">Max Days</div></div>' +
        '</div>';

      if (titleKeys.length > 0) {
        html += '<div class="emp-tth-by-role">' +
          titleKeys.map(function (k) {
            return '<div class="emp-tth-role"><span class="emp-tth-role-name">' + escapeHtml(k) + '</span><span class="emp-tth-role-days">' + byTitle[k] + ' days</span></div>';
          }).join('') +
        '</div>';
      }

      els.timeToHire.innerHTML = html;
    } catch (err) {
      console.error('loadTimeToHire failed:', err);
      if (els.timeToHire) els.timeToHire.innerHTML = '<div class="emp-empty-state"><div class="emp-empty-title">No hires yet</div><div class="emp-empty-desc">Time-to-hire data appears once candidates are hired.</div></div>';
    }
  }

  async function loadRecentApps() {
    try {
      var data = await AngaziaAPI.applications.companyApplications({ limit: 8, sort: 'newest' });
      var apps = data.applications || data.data || [];
      if (!els.recentApps) return;

      if (apps.length === 0) {
        els.recentApps.innerHTML = '<div class="emp-empty-state"><div class="emp-empty-icon">📭</div><div class="emp-empty-title">No applications</div><div class="emp-empty-desc">Applications will appear here once candidates apply.</div></div>';
        return;
      }

      els.recentApps.innerHTML = '<table class="emp-recent-table">' +
        apps.map(function (a) {
          var name = a.candidate_name || a.name || 'Unknown';
          return '<tr onclick="window.location.href=\'/employer/applications/' + (a.id || a.ID || '') + '\'" style="cursor:pointer">' +
            '<td style="width:36px"><div class="emp-avatar-sm">' + getInitials(name) + '</div></td>' +
            '<td><div style="font-weight:500;font-size:12px">' + escapeHtml(name) + '</div></td>' +
            '<td><div style="font-size:11px;color:var(--muted)">' + escapeHtml(a.job_title || '') + '</div></td>' +
            '<td style="text-align:right"><span class="emp-status-badge ' + (a.status || 'pending') + '">' + (a.status || 'pending') + '</span></td>' +
            '<td style="text-align:right;color:var(--muted);font-size:11px">' + timeAgo(a.applied_at) + '</td>' +
          '</tr>';
        }).join('') +
      '</table>';
    } catch (err) {
      console.error('loadRecentApps failed:', err);
      if (els.recentApps) els.recentApps.innerHTML = '<div class="emp-empty-state"><div class="emp-empty-title">Failed to load</div></div>';
    }
  }

  async function loadAllData() {
    showLoading();
    try {
      var [statsData] = await Promise.all([
        loadStats(),
        loadTrends(),
        loadFunnel(),
        loadTopJobs(),
        loadQuality(),
        loadSources(),
        loadTimeToHire(),
        loadRecentApps()
      ]);

      showContent();
      setTimeout(function () {
        staggerIn('.emp-stat-card', 80);
        staggerIn('.emp-section', 100);
      }, 200);
    } catch (err) {
      console.error('loadAllData failed:', err);
      showError(err.message || 'Failed to load analytics data');
    }
  }

  function initEventListeners() {
    if (els.period) {
      els.period.addEventListener('change', function () {
        loadTrends();
        loadStats();
        loadFunnel();
      });
    }
    if (els.retryBtn) {
      els.retryBtn.addEventListener('click', loadAllData);
    }
    if (els.refreshBtn) {
      els.refreshBtn.addEventListener('click', function () {
        els.refreshBtn.classList.add('spin');
        setTimeout(function () { els.refreshBtn.classList.remove('spin'); }, 500);
        loadAllData();
        showToast('Analytics refreshed', 'success');
      });
    }
  }

  function init() {
    cacheEls();
    initEventListeners();
    loadAllData();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
