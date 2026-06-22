(function () {
  'use strict';

  var btn = document.getElementById('btn-confirm-upgrade');
  var errDiv = document.getElementById('upgrade-error');
  var successDiv = document.getElementById('upgrade-success');
  var loadingEl = document.getElementById('upgrade-loading');
  var contentEl = document.getElementById('upgrade-content');

  if (!btn) return;

  var planId = btn.getAttribute('data-plan');
  var pollTimer = null;

  function showError(msg) {
    if (errDiv) {
      errDiv.textContent = msg;
      errDiv.style.display = '';
    }
  }

  function hideError() {
    if (errDiv) errDiv.style.display = 'none';
  }

  function loadPlanAndSub() {
    if (loadingEl) loadingEl.style.display = '';
    if (contentEl) contentEl.style.display = 'none';

    Promise.all([
      AngaziaAPI.subscriptions.plans(),
      AngaziaAPI.subscriptions.current()
    ])
      .then(function (results) {
        var plans = results[0];
        var sub = results[1];

        if (loadingEl) loadingEl.style.display = 'none';
        if (contentEl) contentEl.style.display = '';

        if (!sub || !sub.id) {
          showError('No active subscription found. Please subscribe to a plan first.');
          btn.disabled = true;
          return;
        }

        var plan = null;
        if (plans && plans.length) {
          for (var i = 0; i < plans.length; i++) {
            if (plans[i].plan_id === planId) { plan = plans[i]; break; }
          }
        }

        if (!plan) {
          showError('Plan "' + planId + '" not found.');
          btn.disabled = true;
          return;
        }

        if (sub.plan_id === planId) {
          showError('You are already on this plan.');
          btn.disabled = true;
          return;
        }

        btn.setAttribute('data-subscription-id', sub.id);
        btn.disabled = false;

        var nameEl = document.getElementById('upgrade-plan-name');
        var priceEl = document.getElementById('upgrade-plan-price');
        var featuresEl = document.getElementById('upgrade-plan-features');
        var confirmDesc = document.getElementById('upgrade-confirm-desc');

        if (nameEl) nameEl.textContent = plan.name || plan.plan_id;
        if (priceEl) {
          if (plan.price > 0) {
            priceEl.innerHTML = (plan.currency === 'KES' ? 'KSh ' : plan.currency + ' ') + Number(plan.price).toLocaleString() + (plan.interval === 'month' ? '<span>/month</span>' : plan.interval === 'year' ? '<span>/year</span>' : '');
          } else {
            priceEl.innerHTML = 'Free';
          }
        }
        if (featuresEl && plan.features && plan.features.length) {
          featuresEl.innerHTML = plan.features.map(function (f) { return '<li>\u2705 ' + f + '</li>'; }).join('');
        }
        if (confirmDesc) {
          var priceStr = plan.price > 0 ? (plan.currency === 'KES' ? 'KSh ' : plan.currency + ' ') + Number(plan.price).toLocaleString() : 'Free';
          confirmDesc.innerHTML = 'You are about to upgrade to the <strong>' + (plan.name || plan.plan_id) + '</strong> plan.' +
            (plan.price > 0 ? ' You will be charged <strong>' + priceStr + '</strong> per ' + (plan.interval || 'month') + '.' : '');
        }
      })
      .catch(function (err) {
        if (loadingEl) loadingEl.style.display = 'none';
        if (contentEl) contentEl.style.display = '';
        showError(err.message || 'Failed to load plan details. Please try again.');
        btn.disabled = true;
      });
  }

  btn.addEventListener('click', function () {
    var subId = btn.getAttribute('data-subscription-id');
    if (!subId) {
      showError('Subscription not loaded. Please refresh the page.');
      return;
    }

    hideError();
    if (successDiv) successDiv.style.display = 'none';
    btn.disabled = true;
    btn.textContent = 'Processing...';

    AngaziaAPI.subscriptions.upgrade({
      subscription_id: subId,
      new_plan_id: planId
    })
      .then(function (data) {
        var charge = data && data.charge;
        if (charge && charge.status === 'pending') {
          if (successDiv) {
            successDiv.textContent = 'Payment request sent to your phone. Check M-Pesa and enter PIN to complete upgrade.';
            successDiv.style.display = '';
          }
          btn.textContent = 'Waiting for payment...';
          pollTimer = setInterval(function () {
            AngaziaAPI.subscriptions.current().then(function (sub) {
              if (sub && sub.plan_id === planId && sub.status === 'active') {
                clearInterval(pollTimer);
                if (successDiv) successDiv.textContent = 'Upgrade successful!';
                btn.textContent = 'Upgraded!';
                setTimeout(function () { window.location.href = '/employer/billing'; }, 1500);
              }
            });
          }, 3000);
          setTimeout(function () { clearInterval(pollTimer); pollTimer = null; btn.disabled = false; btn.textContent = 'Confirm Upgrade'; }, 120000);
        } else {
          if (successDiv) {
            successDiv.textContent = 'Successfully upgraded to ' + (planId.charAt(0).toUpperCase() + planId.slice(1)) + '!';
            successDiv.style.display = '';
          }
          btn.textContent = 'Upgraded!';
          setTimeout(function () { window.location.href = '/employer/billing'; }, 2000);
        }
      })
      .catch(function (err) {
        btn.disabled = false;
        btn.textContent = 'Confirm Upgrade';
        showError(err.message || 'Upgrade failed. Please try again.');
      });
  });

  loadPlanAndSub();

  window.addEventListener('beforeunload', function () {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  });
})();
