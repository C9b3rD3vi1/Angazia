(function () {
  'use strict';

  var els = {};
  var applicationsChart = null;
  var refreshInterval = null;
  var currentPeriod = 30;
  var dashboardData = null;

  function qs(id) { return document.getElementById(id); }

  function cacheEls() {
    els.loading = qs('dash-loading');
    els.error = qs('dash-error');
    els.errorMsg = qs('dash-error-msg');
    els.retryBtn = qs('dash-retry-btn');
    els.content = qs('dash-content');
    els.greeting = qs('dash-greeting');
    els.subtitle = qs('dash-subtitle');
    els.exportBtn = qs('dash-export-btn');

    els.statActiveJobs = qs('stat-active-jobs');
    els.statActiveJobsChange = qs('stat-active-jobs-change');
    els.statActiveJobsTrend = qs('stat-active-jobs-trend');
    els.statTotalApplicants = qs('stat-total-applicants');
    els.statApplicantsChange = qs('stat-applicants-change');
    els.statApplicantsTrend = qs('stat-applicants-trend');
    els.statNewApps = qs('stat-new-applications');
    els.statNewAppsChange = qs('stat-new-applications-change');
    els.statNewAppsTrend = qs('stat-new-applications-trend');
    els.statViews = qs('stat-profile-views');
    els.statViewsChange = qs('stat-views-change');
    els.statViewsTrend = qs('stat-views-trend');
    els.statShortlisted = qs('stat-shortlisted');
    els.statShortlistedChange = qs('stat-shortlisted-change');
    els.statShortlistedTrend = qs('stat-shortlisted-trend');
    els.statHired = qs('stat-hired');
    els.statHiredChange = qs('stat-hired-change');
    els.statHiredTrend = qs('stat-hired-trend');

    els.insightsBanner = qs('insights-banner');
    els.insightsText = qs('insights-text');

    els.cardQuality = qs('card-quality');
    els.qualityAvgScore = qs('quality-avg-score');
    els.qualityBarHigh = qs('quality-bar-high');
    els.qualityBarMedium = qs('quality-bar-medium');
    els.qualityBarLow = qs('quality-bar-low');
    els.qualityCountHigh = qs('quality-count-high');
    els.qualityCountMedium = qs('quality-count-medium');
    els.qualityCountLow = qs('quality-count-low');
    els.qualityResponseTime = qs('quality-response-time');

    els.funnelApps = qs('funnel-applications');
    els.funnelShortlisted = qs('funnel-shortlisted');
    els.funnelInterviewed = qs('funnel-interviewed');
    els.funnelHired = qs('funnel-hired');
    els.funnelShortlistedBar = qs('funnel-shortlisted-bar');
    els.funnelInterviewedBar = qs('funnel-interviewed-bar');
    els.funnelHiredBar = qs('funnel-hired-bar');
    els.funnelMeta = qs('funnel-meta');

    els.cardSource = qs('card-source');
    els.sourceList = qs('source-list');

    els.cardTth = qs('card-tth');
    els.tthAvg = qs('tth-average');
    els.tthMin = qs('tth-min');
    els.tthMax = qs('tth-max');
    els.tthDetail = qs('tth-detail');

    els.subPlan = qs('subscription-plan');
    els.subPrice = qs('subscription-price');
    els.subUsage = qs('subscription-usage');
    els.subUsageFill = qs('subscription-usage-fill');
    els.subFeatures = qs('subscription-features');

    els.recentTbody = qs('recent-applications-tbody');
    els.jobPerfTbody = qs('job-performance-tbody');
    els.topJobsList = qs('top-jobs-list');

    els.chartPeriod = qs('chart-period');
    els.chartSubtitle = qs('chart-subtitle');
    els.jobPerfSort = qs('job-performance-sort');
    els.upgradeBtn = qs('upgrade-plan-btn');
  }

  function showLoading() {
    if (els.loading) els.loading.style.display = 'flex';
    if (els.error) els.error.style.display = 'none';
    if (els.content) els.content.style.display = 'none';
  }

  function showError(msg) {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) els.error.style.display = '';
    if (els.errorMsg) els.errorMsg.textContent = msg || 'An unexpected error occurred. Please try again.';
    if (els.content) els.content.style.display = 'none';
  }

  function showContent() {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) els.error.style.display = 'none';
    if (els.content) els.content.style.display = '';
  }

  function updateGreeting() {
    if (!els.greeting) return;
    var h = new Date().getHours();
    var g = h < 12 ? 'Good morning' : h < 17 ? 'Good afternoon' : 'Good evening';
    els.greeting.textContent = g + ', welcome back';
    if (els.subtitle) {
      els.subtitle.textContent = 'Here\'s what\'s happening with your job listings today.';
    }
  }

  function loadData() {
    showLoading();
    var period = els.chartPeriod ? parseInt(els.chartPeriod.value) : 30;
    currentPeriod = period;
    applicationsChart = { labels: [], values: [] };

    AngaziaAPI.dashboard.employer({ days: period }).then(function (data) {
      dashboardData = data;
      updateGreeting();
      showContent();
      renderStats(data.stats);
      renderTrends(data.trends);
      renderFunnel(data.funnel);
      renderJobs(data.jobs);
      renderRecentApps(data.recent_applications);
      renderSubscription(data.subscription);
      renderQuality(data.application_quality);
      renderTimeToHire(data.time_to_hire);
      renderSourceAnalytics(data.source_analytics);
      renderInsights(data);
      renderChart();
    }).catch(function (err) {
      showError(err.message || 'Failed to load dashboard data');
    });
  }

  function animateCounter(el, target, duration) {
    if (!el) return;
    var start = parseInt(el.textContent) || 0;
    if (start === target) return;
    var startTime = performance.now();
    function step(now) {
      var elapsed = now - startTime;
      var progress = Math.min(elapsed / duration, 1);
      var eased = 1 - Math.pow(1 - progress, 3);
      var current = Math.round(start + (target - start) * eased);
      el.textContent = current;
      if (progress < 1) requestAnimationFrame(step);
    }
    requestAnimationFrame(step);
  }

  function renderStats(stats) {
    if (!stats) return;
    var fields = [
      { el: els.statActiveJobs, val: stats.active_jobs || 0 },
      { el: els.statTotalApplicants, val: stats.total_applicants || 0 },
      { el: els.statNewApps, val: stats.new_applications || 0 },
      { el: els.statViews, val: stats.profile_views || 0 },
      { el: els.statShortlisted, val: stats.shortlisted_count || 0 },
      { el: els.statHired, val: stats.hired_count || 0 },
    ];
    fields.forEach(function (f) { animateCounter(f.el, f.val, 800); });

    var trendEls = [
      { el: els.statActiveJobsChange, trendEl: els.statActiveJobsTrend, label: 'active jobs' },
      { el: els.statApplicantsChange, trendEl: els.statApplicantsTrend, label: 'applicants' },
      { el: els.statNewAppsChange, trendEl: els.statNewAppsTrend, label: 'new apps' },
      { el: els.statViewsChange, trendEl: els.statViewsTrend, label: 'views' },
      { el: els.statShortlistedChange, trendEl: els.statShortlistedTrend, label: 'shortlisted' },
      { el: els.statHiredChange, trendEl: els.statHiredTrend, label: 'hired' },
    ];
    if (dashboardData && dashboardData.trends && dashboardData.trends.summary) {
      var growth = dashboardData.trends.summary.growth_rate || 0;
      var isUp = growth >= 0;
      var cls = isUp ? 'up' : 'down';
      var arrow = isUp ? '\u2191' : '\u2193';
      var pct = Math.abs(growth).toFixed(1);
      trendEls.forEach(function (t) {
        if (t.el) {
          t.el.textContent = (isUp ? '+' : '') + pct + '%';
          t.el.className = 'emp-stat-change ' + cls;
        }
        if (t.trendEl) {
          t.trendEl.textContent = arrow + ' ' + pct + '% vs previous period';
          t.trendEl.className = 'emp-stat-trend ' + cls;
        }
      });
    }
  }

  function renderTrends(trends) {
    if (trends && trends.daily && trends.daily.length > 0) {
      applicationsChart = {
        labels: trends.daily.map(function (d) { return d.date; }),
        values: trends.daily.map(function (d) { return d.total || 0; })
      };
    } else {
      applicationsChart = { labels: [], values: [] };
    }
  }

  function renderChart() {
    var canvas = qs('applications-chart');
    if (!canvas) return;
    var ctx = canvas.getContext('2d');
    if (!ctx) return;

    try {
      var prev = Chart.getChart(canvas);
      if (prev) prev.destroy();
    } catch (e) {}

    try {
      new Chart(ctx, {
        type: 'line',
        data: {
          labels: applicationsChart ? applicationsChart.labels : [],
          datasets: [{
            label: 'Applications',
            data: applicationsChart ? applicationsChart.values : [],
            borderColor: '#6366f1',
            backgroundColor: 'rgba(99,102,241,0.1)',
            borderWidth: 2,
            fill: true,
            tension: 0.4,
            pointRadius: 3,
            pointHoverRadius: 6,
            pointBackgroundColor: '#6366f1',
            pointBorderColor: '#fff',
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
              backgroundColor: 'rgba(0,0,0,0.8)',
              titleColor: '#fff',
              bodyColor: '#ddd',
              callbacks: {
                label: function (context) { return 'Applications: ' + context.raw; }
              }
            }
          },
          scales: {
            y: { beginAtZero: true, grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { stepSize: 1 } },
            x: { grid: { display: false }, ticks: { maxRotation: 45, minRotation: 45, maxTicksLimit: 15 } }
          }
        }
      });
    } catch (e) {
      console.error('Chart render error:', e);
    }
  }

  function renderFunnel(funnel) {
    if (!funnel || !funnel.stages || funnel.stages.length === 0) return;
    var stages = funnel.stages;
    var findStage = function (name) {
      var s = stages.find(function (st) { return st.stage === name; });
      return s ? s.count : 0;
    };
    var total = findStage('applied') || findStage('viewed');
    var shortlisted = findStage('shortlisted');
    var interviewed = findStage('interview');
    var hired = findStage('hired');

    if (els.funnelApps) els.funnelApps.textContent = total;
    if (els.funnelShortlisted) els.funnelShortlisted.textContent = shortlisted;
    if (els.funnelInterviewed) els.funnelInterviewed.textContent = interviewed;
    if (els.funnelHired) els.funnelHired.textContent = hired;

    var sp = total > 0 ? (shortlisted / total) * 100 : 0;
    var ip = shortlisted > 0 ? (interviewed / shortlisted) * 100 : 0;
    var hp = interviewed > 0 ? (hired / interviewed) * 100 : 0;

    if (els.funnelShortlistedBar) els.funnelShortlistedBar.style.width = sp + '%';
    if (els.funnelInterviewedBar) els.funnelInterviewedBar.style.width = ip + '%';
    if (els.funnelHiredBar) els.funnelHiredBar.style.width = hp + '%';

    if (els.funnelMeta) {
      var overall = funnel.overall_rate || 0;
      els.funnelMeta.textContent = 'Overall conversion: ' + overall.toFixed(1) + '% from application to hire';
    }
  }

  function renderQuality(quality) {
    if (!quality || !els.cardQuality) return;
    els.cardQuality.style.display = '';
    if (els.qualityAvgScore) {
      els.qualityAvgScore.textContent = (quality.average_match_score || 0).toFixed(1);
    }

    var high = quality.high_quality_count || 0;
    var medium = quality.medium_quality_count || 0;
    var low = quality.low_quality_count || 0;
    var total = high + medium + low;

    if (els.qualityBarHigh) els.qualityBarHigh.style.width = (total > 0 ? (high / total) * 100 : 0) + '%';
    if (els.qualityBarMedium) els.qualityBarMedium.style.width = (total > 0 ? (medium / total) * 100 : 0) + '%';
    if (els.qualityBarLow) els.qualityBarLow.style.width = (total > 0 ? (low / total) * 100 : 0) + '%';

    if (els.qualityCountHigh) els.qualityCountHigh.textContent = high;
    if (els.qualityCountMedium) els.qualityCountMedium.textContent = medium;
    if (els.qualityCountLow) els.qualityCountLow.textContent = low;

    if (els.qualityResponseTime) {
      var rt = quality.average_response_time || 0;
      if (rt < 1) {
        els.qualityResponseTime.textContent = (rt * 60).toFixed(0) + ' minutes';
      } else if (rt < 24) {
        els.qualityResponseTime.textContent = rt.toFixed(1) + ' hours';
      } else {
        els.qualityResponseTime.textContent = (rt / 24).toFixed(1) + ' days';
      }
    }
  }

  function renderTimeToHire(tth) {
    if (!tth || !els.cardTth) return;
    els.cardTth.style.display = '';

    if (els.tthAvg) els.tthAvg.textContent = tth.average_days || 0;
    if (els.tthMin) els.tthMin.textContent = tth.min_days || 0;
    if (els.tthMax) els.tthMax.textContent = tth.max_days || 0;

    if (els.tthDetail && tth.by_job_title) {
      var titles = Object.keys(tth.by_job_title);
      if (titles.length > 0) {
        els.tthDetail.innerHTML = '<div style="font-size:11px;color:var(--muted);margin-bottom:8px">Days by job:</div>' +
          titles.map(function (title) {
            return '<div class="emp-tth-job-row"><span class="emp-tth-job-name">' + escapeHtml(title) + '</span><span class="emp-tth-job-days">' + tth.by_job_title[title] + ' days</span></div>';
          }).join('');
      }
    }
  }

  function renderSourceAnalytics(sources) {
    if (!sources || !Array.isArray(sources) || sources.length === 0 || !els.cardSource) return;
    els.cardSource.style.display = '';

    var total = 0;
    sources.forEach(function (s) { total += s.count || 0; });

    if (els.sourceList) {
      els.sourceList.innerHTML = sources.map(function (src) {
        var pct = total > 0 ? ((src.count || 0) / total * 100).toFixed(1) : 0;
        return '<div class="emp-source-row">' +
          '<span class="emp-source-name">' + escapeHtml(src.source) + '</span>' +
          '<div class="emp-source-bar-track"><div class="emp-source-bar" style="width:' + pct + '%"></div></div>' +
          '<span class="emp-source-count">' + (src.count || 0) + '</span>' +
          '<span class="emp-source-pct">' + pct + '%</span>' +
          '</div>';
      }).join('');
    }
  }

  function renderInsights(data) {
    if (!els.insightsBanner || !els.insightsText) return;
    var insights = [];

    if (data.stats) {
      if (data.stats.total_applicants > 0) {
        insights.push('You have received ' + data.stats.total_applicants + ' total application' + (data.stats.total_applicants !== 1 ? 's' : '') + ' across ' + data.stats.active_jobs + ' active job' + (data.stats.active_jobs !== 1 ? 's' : '') + '.');
      }
      if (data.stats.new_applications > 0) {
        insights.push(data.stats.new_applications + ' new application' + (data.stats.new_applications !== 1 ? 's' : '') + ' in the last 30 days.');
      }
    }

    if (data.trends && data.trends.summary) {
      var growth = data.trends.summary.growth_rate;
      if (growth !== undefined && growth !== 0) {
        var direction = growth > 0 ? 'increasing' : 'decreasing';
        insights.push('Application volume is ' + direction + ' (' + Math.abs(growth).toFixed(1) + '% vs previous period).');
      }
      if (data.trends.summary.peak_day) {
        insights.push('Peak application day was ' + data.trends.summary.peak_day + ' with ' + data.trends.summary.peak_applications + ' applications.');
      }
    }

    if (data.funnel && data.funnel.overall_rate !== undefined && data.funnel.overall_rate > 0) {
      insights.push('Your application-to-hire conversion rate is ' + data.funnel.overall_rate.toFixed(1) + '%.');
    }

    if (data.application_quality) {
      var high = data.application_quality.high_quality_count || 0;
      var total = high + (data.application_quality.medium_quality_count || 0) + (data.application_quality.low_quality_count || 0);
      if (total > 0) {
        var highPct = (high / total * 100).toFixed(0);
        insights.push(highPct + '% of your applications are high-quality (match score \u226580).');
      }
    }

    if (data.time_to_hire && data.time_to_hire.average_days > 0) {
      insights.push('Average time-to-hire is ' + data.time_to_hire.average_days + ' days.');
    }

    if (data.jobs && Array.isArray(data.jobs)) {
      var noApps = data.jobs.filter(function (j) { return !j.applications || j.applications === 0; });
      if (noApps.length > 0) {
        insights.push(noApps.length + ' job' + (noApps.length !== 1 ? 's' : '') + ' have no applications yet.');
      }
    }

    if (insights.length > 0) {
      els.insightsBanner.style.display = 'flex';
      els.insightsText.textContent = insights.slice(0, 3).join(' ');
    } else {
      els.insightsBanner.style.display = 'none';
    }
  }

  function renderJobs(jobs) {
    if (!jobs || !Array.isArray(jobs)) {
      if (els.jobPerfTbody) els.jobPerfTbody.innerHTML = '<tr><td colspan="10" style="text-align:center;padding:40px;color:var(--muted)">No jobs posted yet</td></tr>';
      if (els.topJobsList) els.topJobsList.innerHTML = '<div class="emp-list-item" style="justify-content:center;color:var(--muted)">No jobs posted yet</div>';
      return;
    }

    var sortBy = els.jobPerfSort ? els.jobPerfSort.value : 'applications';
    var sorted = jobs.slice().sort(function (a, b) { return (b[sortBy] || 0) - (a[sortBy] || 0); });

    if (els.jobPerfTbody) {
      if (jobs.length === 0) {
        els.jobPerfTbody.innerHTML = '<tr><td colspan="10" style="text-align:center;padding:40px;color:var(--muted)">No jobs posted yet</td></tr>';
      } else {
        els.jobPerfTbody.innerHTML = sorted.map(function (job) {
          var vta = job.view_to_app_rate !== undefined ? job.view_to_app_rate.toFixed(1) : (job.views > 0 ? ((job.applications / job.views) * 100).toFixed(1) : '0.0');
          var ats = job.app_to_shortlist_rate !== undefined ? job.app_to_shortlist_rate.toFixed(1) : (job.applications > 0 ? ((job.shortlisted / job.applications) * 100).toFixed(1) : '0.0');
          return '<tr data-href="/employer/jobs/' + (job.job_id || '') + '">'
            + '<td style="font-weight:500">' + escapeHtml(job.title) + '</td>'
            + '<td>' + formatDate(job.posted_at) + '</td>'
            + '<td>' + (job.views || 0) + '</td>'
            + '<td>' + (job.applications || 0) + '</td>'
            + '<td>' + (job.shortlisted || 0) + '</td>'
            + '<td>' + (job.hired || 0) + '</td>'
            + '<td>' + vta + '%</td>'
            + '<td>' + ats + '%</td>'
            + '<td><span class="emp-status-badge ' + (job.is_active !== false ? 'active' : 'closed') + '">' + (job.is_active !== false ? 'Active' : 'Closed') + '</span></td>'
            + '<td><a href="/employer/jobs/' + (job.job_id || '') + '" class="emp-link">View &rarr;</a></td>'
            + '</tr>';
        }).join('');
        els.jobPerfTbody.querySelectorAll('tr[data-href]').forEach(function (row) {
          row.addEventListener('click', function () { window.location.href = this.getAttribute('data-href'); });
        });
      }
    }

    if (els.topJobsList) {
      var topJobs = sorted.slice(0, 5);
      if (topJobs.length === 0) {
        els.topJobsList.innerHTML = '<div class="emp-list-item" style="justify-content:center;color:var(--muted)">No jobs posted yet</div>';
      } else {
        els.topJobsList.innerHTML = topJobs.map(function (job) {
          return '<div class="emp-list-item" data-href="/employer/jobs/' + (job.job_id || '') + '">'
            + '<div class="emp-list-item-info">'
            + '<div class="emp-list-item-title">' + escapeHtml(job.title) + '</div>'
            + '<div class="emp-list-item-meta">' + (job.applications || 0) + ' applications &bull; ' + (job.views || 0) + ' views</div>'
            + '</div>'
            + '<div class="emp-list-item-value">' + (job.applications || 0) + '</div>'
            + '</div>';
        }).join('');
        els.topJobsList.querySelectorAll('.emp-list-item[data-href]').forEach(function (row) {
          row.addEventListener('click', function () { window.location.href = this.getAttribute('data-href'); });
        });
      }
    }
  }

  function renderRecentApps(apps) {
    if (!els.recentTbody) return;
    if (!apps || !Array.isArray(apps) || apps.length === 0) {
      els.recentTbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:40px;color:var(--muted)">No applications yet</td></tr>';
      return;
    }
    els.recentTbody.innerHTML = apps.map(function (app) {
      var avatarHtml = app.candidate_avatar ? '<img src="' + app.candidate_avatar + '" alt="' + escapeHtml(app.candidate_name) + '" style="width:100%;height:100%;object-fit:cover;border-radius:50%">' : getInitials(app.candidate_name);
      return '<tr data-href="/employer/applications/' + escapeHtml(app.id) + '">'
        + '<td><div style="display:flex;align-items:center;gap:10px;">'
        + '<div class="emp-avatar">' + avatarHtml + '</div>'
        + '<div><div style="font-weight:500">' + escapeHtml(app.candidate_name || 'Unknown') + '</div>'
        + '<div style="font-size:11px;color:var(--muted)">' + escapeHtml(app.candidate_email || '') + '</div></div></div></td>'
        + '<td>' + escapeHtml(app.job_title || 'Unknown') + '</td>'
        + '<td>' + timeAgo(app.applied_at) + '</td>'
        + '<td><span class="emp-match-score">' + (app.match_score || 0) + '%</span></td>'
        + '<td><span class="emp-status-badge ' + (app.status || 'pending') + '">' + (app.status || 'pending') + '</span></td>'
        + '<td><a href="/employer/applications/' + escapeHtml(app.id) + '" class="emp-link">Review &rarr;</a></td>'
        + '</tr>';
    }).join('');
    els.recentTbody.querySelectorAll('tr[data-href]').forEach(function (row) {
      row.addEventListener('click', function () { window.location.href = this.getAttribute('data-href'); });
    });
  }

  function renderSubscription(sub) {
    if (!sub) {
      if (els.subPlan) els.subPlan.textContent = 'Free Plan';
      if (els.subPrice) els.subPrice.textContent = 'KES 0/month';
      if (els.subUsage) els.subUsage.textContent = '0 / 3';
      if (els.subUsageFill) els.subUsageFill.style.width = '0%';
      return;
    }
    if (els.subPlan) els.subPlan.textContent = sub.plan_name || 'Free Plan';
    if (els.subPrice) {
      var p = sub.amount || 0;
      var c = sub.currency || 'KES';
      var i = sub.interval || 'month';
      els.subPrice.textContent = c + ' ' + p + '/' + i;
    }
    if (els.subUsage) {
      var used = sub.jobs_used || 0;
      var limit = sub.jobs_limit || 3;
      els.subUsage.textContent = used + ' / ' + limit;
      if (els.subUsageFill) {
        var pct = limit > 0 ? (used / limit) * 100 : 0;
        els.subUsageFill.style.width = Math.min(pct, 100) + '%';
      }
    }
    if (els.subFeatures && sub.features && Array.isArray(sub.features)) {
      els.subFeatures.innerHTML = sub.features.map(function (f) {
        return '<div class="emp-feature-item">&#x2713; ' + escapeHtml(f) + '</div>';
      }).join('');
    }
  }

  function setupListeners() {
    if (els.retryBtn) els.retryBtn.addEventListener('click', loadData);
    if (els.exportBtn) {
      els.exportBtn.addEventListener('click', function () {
        window.location.href = '/employer/analytics/export?format=csv';
      });
    }
    if (els.chartPeriod) {
      els.chartPeriod.addEventListener('change', function () {
        currentPeriod = parseInt(this.value);
        if (els.chartSubtitle) {
          els.chartSubtitle.textContent = 'Daily application volume (last ' + currentPeriod + ' days)';
        }
        loadData();
      });
    }
    if (els.jobPerfSort) {
      els.jobPerfSort.addEventListener('change', function () {
        if (dashboardData && dashboardData.jobs) {
          renderJobs(dashboardData.jobs);
        }
      });
    }
    if (els.upgradeBtn) {
      els.upgradeBtn.addEventListener('click', function () { window.location.href = '/employer/billing'; });
    }
  }

  function startAutoRefresh() {
    if (refreshInterval) clearInterval(refreshInterval);
    refreshInterval = setInterval(function () {
      if (dashboardData) {
        applicationsChart = { labels: [], values: [] };
        AngaziaAPI.dashboard.employer({ days: currentPeriod }).then(function (data) {
          dashboardData = data;
          renderStats(data.stats);
          renderTrends(data.trends);
          renderFunnel(data.funnel);
          renderJobs(data.jobs);
          renderRecentApps(data.recent_applications);
          renderSubscription(data.subscription);
          renderQuality(data.application_quality);
          renderTimeToHire(data.time_to_hire);
          renderSourceAnalytics(data.source_analytics);
          renderInsights(data);
          renderChart();
        }).catch(function () {});
      }
    }, 30000);
  }

  function getInitials(name) {
    if (!name) return '??';
    return name.split(' ').map(function (n) { return n[0]; }).join('').toUpperCase().slice(0, 2);
  }

  function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    try {
      return new Date(dateStr).toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
    } catch (e) { return 'N/A'; }
  }

  function timeAgo(dateStr) {
    if (!dateStr) return '';
    try {
      var date = new Date(dateStr);
      var now = new Date();
      var s = Math.floor((now - date) / 1000);
      if (s < 60) return 'just now';
      var m = Math.floor(s / 60); if (m < 60) return m + 'm ago';
      var h = Math.floor(m / 60); if (h < 24) return h + 'h ago';
      var d = Math.floor(h / 24); if (d < 7) return d + 'd ago';
      return date.toLocaleDateString();
    } catch (e) { return ''; }
  }

  function escapeHtml(text) {
    if (!text) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(text));
    return d.innerHTML;
  }

  function init() {
    if (!window.AngaziaAPI) {
      showError('Dashboard API not available. Please refresh the page.');
      return;
    }
    cacheEls();
    setupListeners();
    loadData();
    startAutoRefresh();

    window.addEventListener('beforeunload', function () {
      if (refreshInterval) {
        clearInterval(refreshInterval);
        refreshInterval = null;
      }
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
