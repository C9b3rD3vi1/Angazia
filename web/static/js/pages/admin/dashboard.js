(function () {
  'use strict';

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  var pollTimer = null;
  var POLL_MS = 30000;
  var chartsInitialized = false;

  function init() {
    countUpStats();
    initCharts();
    fetchPendingVerifications();
    startPolling();

    document.addEventListener('click', handleDelegatedClick);
    document.addEventListener('keydown', handleKeyboard);

    var refreshBtns = document.querySelectorAll('[data-action=refresh]');
    for (var i = 0; i < refreshBtns.length; i++) {
      refreshBtns[i].addEventListener('click', function (e) {
        e.preventDefault();
        refreshNow();
      });
    }

    var periodSelects = document.querySelectorAll('.ad-chart-period');
    for (var i = 0; i < periodSelects.length; i++) {
      periodSelects[i].addEventListener('change', function () {
        var chart = this.getAttribute('data-chart');
        if (chart) {
          initSingleChart(chart, this);
        }
      });
    }

    document.addEventListener('visibilitychange', function () {
      if (document.hidden) {
        stopPolling();
      } else {
        fetchStats();
        fetchPendingVerifications();
        startPolling();
      }
    });
  }

  /* ── Count-Up Animation ── */
  function countUpStats() {
    var cards = document.querySelectorAll('.ad-stat-card[data-value]');
    for (var i = 0; i < cards.length; i++) {
      var card = cards[i];
      var valEl = card.querySelector('.ad-stat-value');
      var target = parseFloat(card.getAttribute('data-value')) || 0;
      var isCurrency = card.getAttribute('data-stat') === 'mrr';
      animateCountUp(valEl, target, isCurrency);
    }
  }

  function animateCountUp(el, target, isCurrency) {
    var duration = 800;
    var startTime = Date.now();
    var startVal = 0;

    function tick() {
      var elapsed = Date.now() - startTime;
      var progress = Math.min(elapsed / duration, 1);
      var eased = 1 - Math.pow(1 - progress, 3);
      var current = Math.round(startVal + (target - startVal) * eased);

      if (isCurrency) {
        el.textContent = current.toLocaleString();
      } else {
        if (current >= 1000000) {
          el.textContent = (current / 1000000).toFixed(1) + 'M';
        } else if (current >= 1000) {
          el.textContent = (current / 1000).toFixed(1) + 'K';
        } else {
          el.textContent = current.toLocaleString();
        }
      }

      if (progress < 1) {
        requestAnimationFrame(tick);
      } else {
        el.classList.add('ad-count-flash');
        setTimeout(function () { el.classList.remove('ad-count-flash'); }, 600);
      }
    }

    requestAnimationFrame(tick);
  }

  /* ── Canvas Charts ── */
  function initCharts() {
    chartsInitialized = true;
    var canvases = document.querySelectorAll('.ad-chart-canvas');
    for (var i = 0; i < canvases.length; i++) {
      var canvas = canvases[i];
      var chartType = canvas.getAttribute('data-chart');
      var select = document.querySelector('.ad-chart-period[data-chart="' + chartType + '"]');
      drawChart(canvas, chartType, select ? parseInt(select.value, 10) : 30);
    }
  }

  function initSingleChart(chartType, selectEl) {
    var canvas = document.querySelector('.ad-chart-canvas[data-chart="' + chartType + '"]');
    if (canvas) {
      drawChart(canvas, chartType, parseInt(selectEl.value, 10));
    }
  }

  function generateDataPoints(count, base, variance) {
    var points = [];
    var val = base;
    for (var i = 0; i < count; i++) {
      val += (Math.random() - 0.48) * variance;
      if (val < 0) val = 0;
      points.push(Math.round(val));
    }
    return points;
  }

  function drawChart(canvas, type, days) {
    if (!canvas || !canvas.getContext) return;
    var ctx = canvas.getContext('2d');
    var rect = canvas.parentElement.getBoundingClientRect();
    var W = canvas.width = Math.min(canvas.width, rect.width || 800);
    var H = canvas.height = 220;
    var pad = { top: 16, bottom: 24, left: 40, right: 16 };
    var chartW = W - pad.left - pad.right;
    var chartH = H - pad.top - pad.bottom;

    ctx.clearRect(0, 0, W, H);

    var base, vari, color;
    if (type === 'user-growth') {
      base = 50;
      vari = 15;
      color = '#00e5a0';
    } else {
      base = 20;
      vari = 8;
      color = '#3d9be9';
    }

    var data = generateDataPoints(days, base, vari);
    var maxVal = Math.max.apply(null, data);
    var minVal = Math.min.apply(null, data);
    if (maxVal === minVal) maxVal = minVal + 1;
    var range = maxVal - minVal;

    function x(i) { return pad.left + (i / (data.length - 1)) * chartW; }
    function y(v) { return pad.top + chartH - ((v - minVal) / range) * chartH; }

    // Grid lines
    ctx.strokeStyle = 'rgba(90,122,114,0.15)';
    ctx.lineWidth = 1;
    var gridLines = 4;
    for (var g = 0; g <= gridLines; g++) {
      var gy = pad.top + (g / gridLines) * chartH;
      ctx.beginPath();
      ctx.moveTo(pad.left, gy);
      ctx.lineTo(W - pad.right, gy);
      ctx.stroke();

      // Y-axis labels
      var labelVal = Math.round(maxVal - (g / gridLines) * range);
      ctx.fillStyle = 'rgba(90,122,114,0.6)';
      ctx.font = '9px var(--fm, monospace)';
      ctx.textAlign = 'right';
      ctx.fillText(labelVal, pad.left - 6, gy + 3);
    }

    // Fill area
    ctx.beginPath();
    ctx.moveTo(x(0), y(data[0]));
    for (var i = 1; i < data.length; i++) {
      ctx.lineTo(x(i), y(data[i]));
    }
    ctx.lineTo(x(data.length - 1), H - pad.bottom);
    ctx.lineTo(x(0), H - pad.bottom);
    ctx.closePath();
    var grad = ctx.createLinearGradient(0, pad.top, 0, H - pad.bottom);
    grad.addColorStop(0, color + '40');
    grad.addColorStop(1, color + '05');
    ctx.fillStyle = grad;
    ctx.fill();

    // Line
    ctx.beginPath();
    ctx.moveTo(x(0), y(data[0]));
    for (var i = 1; i < data.length; i++) {
      ctx.lineTo(x(i), y(data[i]));
    }
    ctx.strokeStyle = color;
    ctx.lineWidth = 2;
    ctx.lineJoin = 'round';
    ctx.lineCap = 'round';
    ctx.stroke();

    // Dot at latest
    var lastIdx = data.length - 1;
    ctx.beginPath();
    ctx.arc(x(lastIdx), y(data[lastIdx]), 3, 0, Math.PI * 2);
    ctx.fillStyle = color;
    ctx.fill();
  }

  /* ── Polling & Updates ── */
  function refreshNow() {
    var btn = document.querySelector('[data-action=refresh]');
    if (btn) {
      btn.disabled = true;
      btn.textContent = 'Refreshing...';
    }
    Promise.all([
      fetchStats(),
      fetchPendingVerifications()
    ]).then(function () {
      if (btn) {
        btn.disabled = false;
        btn.textContent = 'Refresh';
      }
    });
  }

  function fetchStats() {
    return AngaziaAPI.get('/admin/stats/platform')
      .then(function (data) {
        updateStatCards(data);
        updateLastUpdated();
      })
      .catch(function () {});
  }

  function fetchPendingVerifications() {
    AngaziaAPI.admin.pendingVerifications({ limit: 10 })
      .then(function (res) {
        var data = res && res.data ? res.data : res;
        renderPendingVerifications(data);
      })
      .catch(function () {});
  }

  function renderPendingVerifications(data) {
    var listEl = document.getElementById('ad-verif-list');
    var loadingEl = document.getElementById('ad-verif-loading');
    var emptyEl = document.getElementById('ad-verif-empty');
    if (!listEl) return;

    var verifications = (data && data.verifications) || [];

    if (loadingEl) loadingEl.style.display = 'none';

    if (verifications.length === 0) {
      listEl.innerHTML = '';
      if (emptyEl) emptyEl.style.display = 'block';
      return;
    }

    if (emptyEl) emptyEl.style.display = 'none';

    var html = '';
    for (var i = 0; i < verifications.length; i++) {
      var v = verifications[i];
      var company = v.company || {};
      var user = company.user || {};
      var companyName = company.name || user.name || 'Unknown Company';
      var submittedAt = v.submitted_at || v.created_at || '';
      var dateStr = '';
      if (submittedAt) {
        var d = new Date(submittedAt);
        dateStr = d.toLocaleDateString();
      }

      html += '<div class="ad-verif-item">';
      html += '  <div class="ad-verif-info">';
      html += '    <strong class="ad-verif-company">' + escapeHtml(companyName) + '</strong>';
      html += '    <span class="ad-verif-meta">Submitted ' + dateStr + '</span>';
      html += '  </div>';
      html += '  <div class="ad-verif-actions">';
      html += '    <button class="ad-btn ad-btn-sm ad-btn-success" data-action="approve" data-company-id="' + v.company_id + '">Approve</button>';
      html += '    <button class="ad-btn ad-btn-sm ad-btn-danger" data-action="reject" data-company-id="' + v.company_id + '">Reject</button>';
      html += '  </div>';
      html += '</div>';
    }

    listEl.innerHTML = html;
  }

  function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function updateStatCards(data) {
    var cards = document.querySelectorAll('.ad-stat-card');
    if (!cards.length) return;

    var fields = [
      'total_users', 'total_companies', 'active_jobs', 'applications',
      'mrr', 'active_users', 'pending_verifications', 'reports'
    ];

    for (var i = 0; i < cards.length && i < fields.length; i++) {
      var card = cards[i];
      var valEl = card.querySelector('.ad-stat-value');
      var statAttr = card.getAttribute('data-stat');
      var apiKey = statAttr;
      if (data[apiKey] === undefined) continue;

      var newVal = formatStatValue(data[apiKey]);
      var isCurrency = statAttr === 'mrr';
      if (isCurrency) {
        newVal = Number(data[apiKey]).toLocaleString();
      }
      if (valEl.textContent !== newVal) {
        animateValue(card);
        valEl.textContent = newVal;
      }
    }
  }

  function animateValue(card) {
    card.style.transition = 'background 0.2s';
    var origBg = card.style.background;
    card.style.background = 'var(--s2)';
    setTimeout(function () {
      card.style.background = origBg || '';
    }, 200);
  }

  function formatStatValue(n) {
    if (n === null || n === undefined) return '0';
    var num = Number(n);
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toLocaleString();
  }

  function updateLastUpdated() {
    var tsEl = document.querySelector('.ad-timestamp');
    if (tsEl) {
      var now = new Date();
      var timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
      tsEl.textContent = 'Last updated: ' + timeStr;
    }
  }

  function startPolling() {
    stopPolling();
    pollTimer = setInterval(function () {
      fetchStats();
      fetchPendingVerifications();
    }, POLL_MS);
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  function handleKeyboard(e) {
    if (e.key === 'r' || e.key === 'R') {
      if (!e.ctrlKey && !e.metaKey && !e.altKey) {
        var active = document.activeElement;
        if (active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA' || active.tagName === 'SELECT')) {
          return;
        }
        e.preventDefault();
        refreshNow();
      }
    }
  }

  function handleDelegatedClick(e) {
    var btn = e.target.closest('[data-action]');
    if (!btn) return;

    var action = btn.getAttribute('data-action');
    if (action === 'refresh') return;

    var companyId = btn.getAttribute('data-company-id');
    if (!companyId) return;

    if (action === 'approve') {
      e.preventDefault();
      approveCompany(btn, companyId);
    } else if (action === 'reject') {
      e.preventDefault();
      rejectCompany(btn, companyId);
    }
  }

  function approveCompany(btn, companyId) {
    btn.disabled = true;
    btn.textContent = 'Approving...';
    AngaziaAPI.post('/admin/companies/' + companyId + '/verify')
      .then(function () {
        showToast('Verification approved', 'success');
        removeVerificationItem(btn);
        fetchStats();
        fetchPendingVerifications();
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to approve verification', 'error');
        btn.disabled = false;
        btn.textContent = 'Approve';
      });
  }

  function rejectCompany(btn, companyId) {
    var reason = prompt('Reason for rejection (optional):');
    if (reason === null) return;

    btn.disabled = true;
    btn.textContent = 'Rejecting...';
    var body = {};
    if (reason && reason.trim()) {
      body.reason = reason.trim();
    }

    AngaziaAPI.post('/admin/companies/' + companyId + '/reject', body)
      .then(function () {
        showToast('Verification rejected', 'success');
        removeVerificationItem(btn);
        fetchStats();
        fetchPendingVerifications();
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to reject verification', 'error');
        btn.disabled = false;
        btn.textContent = 'Reject';
      });
  }

  function removeVerificationItem(btn) {
    var item = btn.closest('.ad-verif-item');
    if (!item) return;

    item.style.transition = 'opacity 0.25s ease, transform 0.25s ease';
    item.style.opacity = '0';
    item.style.transform = 'translateX(30px)';

    setTimeout(function () {
      if (item.parentNode) {
        item.parentNode.removeChild(item);
      }
    }, 250);
  }

  var animStyle = document.createElement('style');
  animStyle.textContent = '\
@keyframes adDeltaPulse { 0% { transform: scale(1); } 50% { transform: scale(1.3); } 100% { transform: scale(1); } }\
@keyframes adFadeIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }\
';
  if (!document.getElementById('ad-dash-dynamic-styles')) {
    animStyle.id = 'ad-dash-dynamic-styles';
    document.head.appendChild(animStyle);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
