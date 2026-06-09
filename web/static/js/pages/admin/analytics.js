(function () {
  'use strict';
  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  function $(id) { return document.getElementById(id); }
  function qs(sel, ctx) { return (ctx || document).querySelector(sel); }
  function qsa(sel, ctx) { return (ctx || document).querySelectorAll(sel); }

  function setCardValue(parentId, stat, value, trend) {
    var cards = qsa('[data-stat="' + stat + '"]', $(parentId) || document);
    cards.forEach(function (card) {
      var valEl = card.querySelector('.aan-card-value');
      if (valEl) valEl.textContent = value;
      var trendEl = card.querySelector('.aan-trend');
      if (trendEl && trend !== undefined) {
        var isPos = trend >= 0;
        trendEl.textContent = (isPos ? '+' : '') + trend + '%';
        trendEl.className = 'aan-trend ' + (isPos ? 'up' : 'down');
      }
    });
  }

  function setGrowthRate(rate, card) {
    var valEl = card.querySelector('.aan-card-value');
    var trendEl = card.querySelector('.aan-trend');
    if (valEl) {
      valEl.textContent = (rate >= 0 ? '+' : '') + rate.toFixed(1) + '%';
    }
    if (trendEl) {
      trendEl.textContent = (rate >= 0 ? '+' : '') + rate.toFixed(1) + '%';
      trendEl.className = 'aan-trend ' + (rate >= 0 ? 'up' : 'down');
    }
  }

  function formatCurrency(amount) {
    return 'KES ' + Number(amount).toLocaleString('en-KE', {minimumFractionDigits: 0, maximumFractionDigits: 0});
  }

  function loadPlatformStats() {
    return AngaziaAPI.get('/admin/stats/platform').then(function (data) {
      var d = data;
      if ($('aan-row-primary')) {
        setCardValue('aan-row-primary', 'total_users', Number(d.total_users || 0).toLocaleString());
        setCardValue('aan-row-primary', 'total_jobs', Number(d.total_jobs || 0).toLocaleString());
        setCardValue('aan-row-primary', 'total_applications', Number(d.total_applications || 0).toLocaleString());
        setCardValue('aan-row-primary', 'total_revenue', formatCurrency(d.total_revenue || 0));
        setCardValue('aan-row-primary', 'mrr', formatCurrency(d.mrr || 0));
        setCardValue('aan-row-primary', 'average_match_score', (d.average_match_score || 0) + '%');
      }
      if ($('aan-row-activity')) {
        setCardValue('aan-row-activity', 'jobs_posted_7_days', Number(d.jobs_posted_7_days || 0).toLocaleString());
        setCardValue('aan-row-activity', 'jobs_posted_30_days', Number(d.jobs_posted_30_days || 0).toLocaleString());
        setCardValue('aan-row-activity', 'total_profile_views', Number(d.total_profile_views || 0).toLocaleString());
        setCardValue('aan-row-activity', 'total_job_views', Number(d.total_job_views || 0).toLocaleString());
        setCardValue('aan-row-activity', 'verified_employers', Number(d.verified_employers || 0).toLocaleString());
        setCardValue('aan-row-activity', 'total_candidates', Number(d.total_candidates || 0).toLocaleString());
      }
      if ($('aan-row-primary')) {
        var cards = qsa('[data-stat="active_users_30_days"]');
        cards.forEach(function (c) {
          var v = c.querySelector('.aan-card-value');
          if (v) v.textContent = Number(d.active_users_30_days || 0).toLocaleString();
        });
        cards = qsa('[data-stat="new_users_7_days"]');
        cards.forEach(function (c) {
          var v = c.querySelector('.aan-card-value');
          if (v) v.textContent = Number(d.new_users_7_days || 0).toLocaleString();
        });
        cards = qsa('[data-stat="new_users_30_days"]');
        cards.forEach(function (c) {
          var v = c.querySelector('.aan-card-value');
          if (v) v.textContent = Number(d.new_users_30_days || 0).toLocaleString();
        });
      }
      return d;
    }).catch(function (err) {
      showToast(err.message || 'Failed to load platform stats', 'error');
    });
  }

  function loadUserStats() {
    return AngaziaAPI.get('/admin/stats/users').then(function (data) {
      var d = data;
      if ($('aan-row-users')) {
        setCardValue('aan-row-users', 'total', Number(d.total || 0).toLocaleString());
        setCardValue('aan-row-users', 'candidates', Number(d.employee || d.candidates || 0).toLocaleString());
        setCardValue('aan-row-users', 'employers', Number(d.employer || 0).toLocaleString());
        setCardValue('aan-row-users', 'admins', Number(d.admin || 0).toLocaleString());
        setCardValue('aan-row-users', 'active', Number(d.active || 0).toLocaleString());
        setCardValue('aan-row-users', 'suspended', Number(d.suspended || 0).toLocaleString());
      }
      return d;
    }).catch(function (err) {
      showToast(err.message || 'Failed to load user stats', 'error');
    });
  }

  function loadJobStats() {
    return AngaziaAPI.get('/admin/stats/jobs').then(function (data) {
      var d = data;
      if ($('aan-row-jobs')) {
        setCardValue('aan-row-jobs', 'total', Number(d.total || 0).toLocaleString());
        setCardValue('aan-row-jobs', 'active', Number(d.active || 0).toLocaleString());
        setCardValue('aan-row-jobs', 'closed', Number(d.closed || 0).toLocaleString());
        setCardValue('aan-row-jobs', 'draft', Number(d.draft || 0).toLocaleString());
        setCardValue('aan-row-jobs', 'pending', Number(d.pending || 0).toLocaleString());
      }
      return d;
    }).catch(function (err) {
      showToast(err.message || 'Failed to load job stats', 'error');
    });
  }

  function loadGrowthRates() {
    return AngaziaAPI.get('/admin/stats/platform').then(function (data) {
      var d = data;
      if ($('aan-row-growth')) {
        var ug = parseFloat(d.user_growth_rate) || 0;
        var jg = parseFloat(d.job_growth_rate) || 0;
        var ag = parseFloat(d.application_growth_rate) || 0;
        var cards = qsa('[data-stat="user_growth_rate"]', $('aan-row-growth'));
        cards.forEach(function (c) { setGrowthRate(ug, c); });
        cards = qsa('[data-stat="job_growth_rate"]', $('aan-row-growth'));
        cards.forEach(function (c) { setGrowthRate(jg, c); });
        cards = qsa('[data-stat="application_growth_rate"]', $('aan-row-growth'));
        cards.forEach(function (c) { setGrowthRate(ag, c); });
      }
      return d;
    }).catch(function (err) {
      showToast(err.message || 'Failed to load growth data', 'error');
    });
  }

  function loadEngagement() {
    return AngaziaAPI.get('/admin/stats/engagement').then(function (data) {
      var d = data;
      Object.keys(d).forEach(function (key) {
        var cards = qsa('[data-stat="' + key + '"]');
        cards.forEach(function (c) {
          var v = c.querySelector('.aan-card-value');
          if (v) v.textContent = Number(d[key] || 0).toLocaleString();
        });
      });
    }).catch(function () {});
  }

  window.aanLoadAll = function () {
    return Promise.all([
      loadPlatformStats(),
      loadUserStats(),
      loadJobStats(),
      loadGrowthRates(),
      loadEngagement()
    ]);
  };

  document.addEventListener('DOMContentLoaded', function () {
    var refreshBtn = document.querySelector('[data-action="load-all"]');
    if (refreshBtn) {
      refreshBtn.addEventListener('click', function (e) {
        e.preventDefault();
        aanLoadAll();
      });
    }
    aanLoadAll();
  });
})();
