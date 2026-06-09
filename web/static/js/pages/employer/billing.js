(function () {
  'use strict';

  var loadingEl = document.getElementById('billing-loading');
  var errorEl = document.getElementById('billing-error');
  var contentEl = document.getElementById('billing-content');
  var currentPlanEl = document.getElementById('current-plan-content');
  var plansEl = document.getElementById('plans-content');
  var invoicesTbody = document.getElementById('invoices-tbody');

  if (!contentEl) return;

  function showError(msg) {
    if (loadingEl) loadingEl.style.display = 'none';
    if (contentEl) contentEl.style.display = 'none';
    if (errorEl) {
      errorEl.textContent = msg;
      errorEl.style.display = '';
    }
  }

  function formatCurrency(amount, currency) {
    currency = currency || 'KES';
    if (currency === 'KES') {
      return 'KSh ' + Number(amount).toLocaleString();
    }
    return currency + ' ' + Number(amount).toLocaleString();
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    var d = new Date(dateStr);
    return d.toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function getPlanIcon(planId) {
    if (!planId) return '\u{1F3F0}';
    if (planId === 'enterprise') return '\u{1F451}';
    if (planId.indexOf('pro') === 0) return '\u{1F525}';
    if (planId.indexOf('basic') === 0 || planId === 'basic') return '\u{1F4A0}';
    return '\u{1F3F0}';
  }

  function getStatusBadge(status) {
    var cls = 'emp-badge-inactive';
    var label = status || 'unknown';
    if (status === 'active' || status === 'trialing') {
      cls = 'emp-badge-active';
      label = status === 'trialing' ? 'Trial' : 'Active';
    } else if (status === 'past_due') {
      cls = 'emp-badge-warning';
      label = 'Past Due';
    } else if (status === 'canceled') {
      label = 'Canceled';
    } else if (status === 'expired') {
      label = 'Expired';
    }
    return '<span class="emp-badge ' + cls + '">' + label + '</span>';
  }

  function getInvoiceStatusBadge(status) {
    var cls = 'emp-badge-inactive';
    var label = status || 'unknown';
    if (status === 'paid') { cls = 'emp-badge-active'; label = 'Paid'; }
    else if (status === 'pending') { cls = 'emp-badge-warning'; label = 'Pending'; }
    else if (status === 'overdue') { cls = 'emp-badge-warning'; label = 'Overdue'; }
    else if (status === 'failed') { label = 'Failed'; }
    return '<span class="emp-badge ' + cls + '">' + label + '</span>';
  }

  function renderCurrentPlan(sub, plans) {
    if (!sub) {
      currentPlanEl.innerHTML =
        '<div class="emp-current-plan-icon">\u{1F3F0}</div>' +
        '<h3 class="emp-current-plan-name">Free</h3>' +
        '<p style="font-size:13px;color:var(--muted);margin-bottom:16px">You are on the Free plan. Upgrade to unlock more features.</p>';
      return;
    }

    var plan = null;
    if (plans && plans.length) {
      for (var i = 0; i < plans.length; i++) {
        if (plans[i].plan_id === sub.plan_id) { plan = plans[i]; break; }
      }
    }

    var icon = getPlanIcon(sub.plan_id);
    var name = sub.plan_name || sub.plan_id || 'Unknown';
    var badge = getStatusBadge(sub.status);
    var limit = sub.job_post_limit || (plan ? plan.job_post_limit : 3);
    var features = sub.features || (plan ? plan.features : null);
    var endDate = sub.current_period_end || sub.end_date;

    var html = '';
    html += '<div class="emp-current-plan-icon">' + icon + '</div>';
    html += '<h3 class="emp-current-plan-name">' + name + '</h3>';
    html += badge;

    html += '<div class="emp-usage-bar">' +
      '<div class="emp-usage-head">' +
      '<span class="emp-usage-label">Job Limit</span>' +
      '<span class="emp-usage-count">' + limit + ' active jobs</span>' +
      '</div></div>';

    if (features && features.length) {
      html += '<ul class="emp-plan-features">';
      for (var f = 0; f < features.length; f++) {
        html += '<li class="emp-plan-feature"><span class="emp-plan-feature-icon">\u2705</span><span>' + features[f] + '</span></li>';
      }
      html += '</ul>';
    }

    if (endDate && sub.plan_id !== 'free') {
      var endYear = new Date(endDate).getFullYear();
      if (endYear < 2100) {
        html += '<p class="emp-renewal-date">\u{1F504} ' + (sub.status === 'canceled' ? 'Ends' : 'Renews') + ' on ' + formatDate(endDate) + '</p>';
      } else {
        html += '<p class="emp-renewal-date" style="color:var(--muted2)">\u{1F504} Never expires</p>';
      }
    }

    html += '<div class="emp-plan-actions">';
    if (sub.status === 'active' || sub.status === 'trialing') {
      html += '<button class="emp-btn emp-btn-outline emp-btn-sm" id="btn-cancel-subscription">Cancel Subscription</button>';
    }
    if (sub.status === 'canceled') {
      html += '<button class="emp-btn emp-btn-primary emp-btn-sm" id="btn-reactivate-subscription">Reactivate</button>';
    }
    html += '</div>';

    currentPlanEl.innerHTML = html;

    var cancelBtn = document.getElementById('btn-cancel-subscription');
    if (cancelBtn) {
      cancelBtn.addEventListener('click', function () {
        if (!confirm('Are you sure you want to cancel your subscription? You will lose access to premium features at the end of your billing period.')) return;
        cancelBtn.disabled = true;
        cancelBtn.textContent = 'Canceling...';
        AngaziaAPI.subscriptions.cancel({ subscription_id: sub.id })
          .then(function () {
            AngaziaApp && AngaziaApp.showToast ? AngaziaApp.showToast('Subscription canceled', 'success') : alert('Subscription canceled');
            setTimeout(function () { window.location.reload(); }, 1500);
          })
          .catch(function (err) {
            cancelBtn.disabled = false;
            cancelBtn.textContent = 'Cancel Subscription';
            showError(err.message || 'Failed to cancel subscription');
          });
      });
    }

    var reactBtn = document.getElementById('btn-reactivate-subscription');
    if (reactBtn) {
      reactBtn.addEventListener('click', function () {
        reactBtn.disabled = true;
        reactBtn.textContent = 'Reactivating...';
        AngaziaAPI.subscriptions.reactivate({ subscription_id: sub.id })
          .then(function () {
            AngaziaApp && AngaziaApp.showToast ? AngaziaApp.showToast('Subscription reactivated', 'success') : alert('Subscription reactivated');
            setTimeout(function () { window.location.reload(); }, 1500);
          })
          .catch(function (err) {
            reactBtn.disabled = false;
            reactBtn.textContent = 'Reactivate';
            showError(err.message || 'Failed to reactivate subscription');
          });
      });
    }
  }

  function renderPlans(plans, currentSub) {
    if (!plans || !plans.length) {
      plansEl.innerHTML = '<p style="font-size:13px;color:var(--muted);text-align:center;padding:20px 0">No plans available at this time.</p>';
      return;
    }

    var currentPlanId = currentSub ? currentSub.plan_id : 'free';
    var html = '';

    for (var i = 0; i < plans.length; i++) {
      var p = plans[i];
      var isCurrent = (p.plan_id === currentPlanId);
      var featured = p.is_popular && !isCurrent;
      var cls = 'emp-plan-option';
      if (featured) cls += ' emp-plan-option-featured';
      if (isCurrent) cls += ' emp-plan-current';

      html += '<div class="' + cls + '">';
      if (featured) html += '<span class="emp-plan-popular">Most Popular</span>';

      html += '<h4 class="emp-plan-option-name">' + (p.name || p.plan_id) + '</h4>';

      var priceDisplay;
      if (p.plan_id === 'enterprise') {
        priceDisplay = 'Custom';
      } else if (p.price > 0) {
        priceDisplay = formatCurrency(p.price, p.currency);
        if (p.interval === 'month') priceDisplay += '<span>/mo</span>';
        else if (p.interval === 'year') priceDisplay += '<span>/yr</span>';
      } else {
        priceDisplay = 'Free';
        if (p.interval === 'month' || p.interval === 'year') priceDisplay += '<span> forever</span>';
      }
      html += '<p class="emp-plan-price">' + priceDisplay + '</p>';

      if (p.features && p.features.length) {
        html += '<ul class="emp-plan-bullets">';
        for (var f = 0; f < p.features.length; f++) {
          html += '<li>' + p.features[f] + '</li>';
        }
        html += '</ul>';
      }

      if (isCurrent) {
        html += '<span class="emp-plan-current-tag">Current</span>';
      } else if (p.plan_id === 'enterprise') {
        html += '<a href="/employer/billing/upgrade/' + encodeURIComponent(p.plan_id) + '" class="emp-btn emp-btn-outline emp-btn-full">Contact Sales</a>';
      } else {
        html += '<a href="/employer/billing/upgrade/' + encodeURIComponent(p.plan_id) + '" class="emp-btn ' + (featured ? 'emp-btn-primary' : 'emp-btn-outline') + ' emp-btn-full">Upgrade</a>';
      }

      html += '</div>';
    }

    plansEl.innerHTML = html;
  }

  function renderInvoices(data) {
    if (!invoicesTbody) return;
    var invoices = data && data.invoices;
    if (!invoices || !invoices.length) {
      invoicesTbody.innerHTML = '<tr><td colspan="6" class="emp-table-empty">No invoices yet</td></tr>';
      return;
    }
    var html = '';
    for (var i = 0; i < invoices.length; i++) {
      var inv = invoices[i];
      var date = formatDate(inv.created_at || inv.due_date);
      var badge = getInvoiceStatusBadge(inv.status);
      html += '<tr>' +
        '<td><strong>' + (inv.invoice_number || inv.id.slice(0, 8)) + '</strong></td>' +
        '<td>' + date + '</td>' +
        '<td>' + (inv.plan_name || '-') + '</td>' +
        '<td>' + formatCurrency(inv.total || inv.amount, inv.currency) + '</td>' +
        '<td>' + badge + '</td>' +
        '<td>' + (inv.pdf_url ? '<a href="' + inv.pdf_url + '" class="emp-btn emp-btn-outline emp-btn-sm" target="_blank">Download</a>' : '') + '</td>' +
        '</tr>';
    }
    invoicesTbody.innerHTML = html;
  }

  function init() {
    AngaziaAPI.subscriptions.plans()
      .then(function (plans) {
        return AngaziaAPI.subscriptions.current()
          .then(function (sub) { return { plans: plans, sub: sub }; })
          .catch(function () {
            // No subscription record found — user is on the free plan
            return { plans: plans, sub: null };
          });
      })
      .then(function (result) {
        if (loadingEl) loadingEl.style.display = 'none';
        if (contentEl) contentEl.style.display = '';

        renderCurrentPlan(result.sub, result.plans);
        renderPlans(result.plans, result.sub);

        return AngaziaAPI.subscriptions.invoices()
          .then(function (invData) { renderInvoices(invData); })
          .catch(function () { renderInvoices(null); });
      })
      .catch(function (err) {
        // If both plans and subscription failed, show error
        showError(err.message || 'Failed to load billing information. Please try again.');
      });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
