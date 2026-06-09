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
  function val(id) { return $(id).value; }
  function setVal(id, v) { $(id).value = v; }

  function escapeHtml(str) {
    if (!str) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(str));
    return d.innerHTML;
  }

  // ========== PLAN MANAGEMENT (existing) ==========

  function renderPlanCard(plan) {
    var features = plan.features || [];
    var featuresHtml = '';

    if (features && features.length) {
      featuresHtml = '<div class="asub-features"><span class="asub-features-label">Features</span><ul class="asub-features-list">' +
        features.map(function (f) {
          return '<li class="asub-feature-item"><span class="asub-feature-check">v</span><span class="asub-feature-name">' + escapeHtml(f) + '</span></li>';
        }).join('') + '</ul></div>';
    }

    var priceHtml = plan.price === 0 ? 'Free' : plan.currency + ' ' + Number(plan.price).toLocaleString();
    var intervalHtml = plan.price !== 0 ? '<span class="asub-price-interval">/' + plan.interval + '</span>' : '';
    var trialHtml = plan.trial_days ? '<span class="asub-trial-badge">' + plan.trial_days + '-day trial</span>' : '';
    var descHtml = plan.description ? '<p class="asub-card-desc">' + escapeHtml(plan.description) + '</p>' : '';

    return '<div class="asub-card" data-plan-id="' + plan.id + '">' +
      '<div class="asub-card-head">' +
        '<div class="asub-card-title-row">' +
          '<h3 class="asub-card-name">' + escapeHtml(plan.name) + '</h3>' +
          '<label class="asub-toggle">' +
            '<input type="checkbox" ' + (plan.is_active ? 'checked' : '') + ' data-action="asubToggleActive" data-id="' + plan.id + '">' +
            '<span class="asub-toggle-slider"></span>' +
          '</label>' +
        '</div>' +
        '<div class="asub-card-price"><span class="asub-price-amount">' + priceHtml + '</span>' + intervalHtml + '</div>' +
        '<div class="asub-card-code"><span class="asub-code-badge">' + escapeHtml(plan.plan_id) + '</span>' + trialHtml + '</div>' +
        descHtml +
      '</div>' +
      '<div class="asub-card-body">' +
        '<div class="asub-meta-row">' +
          '<span class="asub-meta-item">' + (plan.job_post_limit || 0) + ' jobs</span>' +
          '<span class="asub-meta-item">' + (plan.is_popular ? 'Popular' : 'Standard') + '</span>' +
        '</div>' +
        featuresHtml +
      '</div>' +
      '<div class="asub-card-foot">' +
        '<button class="asub-btn asub-btn-ghost" data-action="asubOpenEdit" data-id="' + plan.id + '">Edit</button>' +
        '<button class="asub-btn asub-btn-danger" data-action="asubDelete" data-id="' + plan.id + '">Delete</button>' +
      '</div>' +
    '</div>';
  }

  function reloadPlans() {
    AngaziaAPI.plans.adminList().then(function (data) {
      var grid = $('asub-grid');
      if (!grid) return;

      grid.innerHTML = '';
      var plans = Array.isArray(data) ? data : (data.plans || data.data || []);

      if (plans && plans.length) {
        plans.forEach(function (plan) {
          grid.insertAdjacentHTML('beforeend', renderPlanCard(plan));
        });
      } else {
        grid.innerHTML = '<div class="asub-empty">No subscription plans found. Click "New Plan" to create one.</div>';
      }
    }).catch(function (err) {
      console.error('Failed to reload plans:', err);
      showToast(err.message || 'Failed to reload plans', 'error');
    });
  }

  function getFormData() {
    var featuresText = val('asub-f-features') || '';
    var features = featuresText.split('\n').filter(function(l) { return l.trim(); });

    return {
      plan_id: val('asub-f-code'),
      name: val('asub-f-name'),
      description: val('asub-f-description'),
      price: parseFloat(val('asub-f-price')) || 0,
      currency: val('asub-f-currency'),
      interval: val('asub-f-interval'),
      interval_count: 1,
      trial_days: parseInt(val('asub-f-trial')) || 0,
      job_post_limit: parseInt(val('asub-f-jobs')) || 3,
      sort_order: parseInt(val('asub-f-sort')) || 0,
      is_active: $('asub-f-active').checked,
      is_popular: false,
      features: features,
      feature_flags: {}
    };
  }

  function setFormData(plan) {
    setVal('asub-f-name', plan.name || '');
    setVal('asub-f-code', plan.plan_id || '');
    setVal('asub-f-description', plan.description || '');
    setVal('asub-f-price', plan.price || 0);
    setVal('asub-f-currency', plan.currency || 'KES');
    setVal('asub-f-interval', plan.interval || 'month');
    setVal('asub-f-trial', plan.trial_days || 0);
    setVal('asub-f-jobs', plan.job_post_limit || 3);
    setVal('asub-f-unlocks', 0);
    setVal('asub-f-featured', 0);
    setVal('asub-f-sort', plan.sort_order || 0);
    $('asub-f-active').checked = plan.is_active !== false;

    var featuresText = (plan.features || []).join('\n');
    setVal('asub-f-features', featuresText);
    setVal('asub-edit-id', plan.id || '');
  }

  function resetForm() {
    setVal('asub-f-name', '');
    setVal('asub-f-code', '');
    setVal('asub-f-description', '');
    setVal('asub-f-price', '');
    setVal('asub-f-currency', 'KES');
    setVal('asub-f-interval', 'month');
    setVal('asub-f-trial', '');
    setVal('asub-f-jobs', '3');
    setVal('asub-f-unlocks', '');
    setVal('asub-f-featured', '');
    setVal('asub-f-sort', '');
    $('asub-f-active').checked = true;
    setVal('asub-f-features', '');
    setVal('asub-edit-id', '');
  }

  window.asubOpenCreate = function () {
    resetForm();
    $('asub-modal-title').textContent = 'New Plan';
    $('asub-save-btn').textContent = 'Create Plan';
    $('asub-modal').style.display = 'flex';
  };

  window.asubOpenEdit = function (id) {
    AngaziaAPI.plans.adminGet(id).then(function (plan) {
      setFormData(plan);
      $('asub-modal-title').textContent = 'Edit Plan';
      $('asub-save-btn').textContent = 'Save Changes';
      $('asub-modal').style.display = 'flex';
    }).catch(function (err) {
      console.error('Failed to load plan:', err);
      showToast(err.message || 'Failed to load plan details', 'error');
    });
  };

  window.asubCloseModal = function () {
    $('asub-modal').style.display = 'none';
  };

  window.asubSave = function () {
    var formData = getFormData();
    var editId = val('asub-edit-id');
    var btn = $('asub-save-btn');

    btn.disabled = true;
    btn.textContent = 'Saving...';

    var promise = editId
      ? AngaziaAPI.plans.adminUpdate(editId, formData)
      : AngaziaAPI.plans.adminCreate(formData);

    promise.then(function () {
      showToast('Plan ' + (editId ? 'updated' : 'created') + ' successfully', 'success');
      asubCloseModal();
      reloadPlans();
    }).catch(function (err) {
      console.error('Failed to save plan:', err);
      showToast(err.message || 'Failed to save plan', 'error');
    }).then(function () {
      btn.disabled = false;
      btn.textContent = editId ? 'Save Changes' : 'Create Plan';
    });
  };

  window.asubDelete = function (id) {
    setVal('asub-delete-id', id);
    $('asub-delete-modal').style.display = 'flex';
  };

  window.asubCloseDelete = function () {
    $('asub-delete-modal').style.display = 'none';
  };

  window.asubConfirmDelete = function () {
    var id = val('asub-delete-id');
    var btn = $('asub-delete-confirm-btn');
    btn.disabled = true;
    btn.textContent = 'Deleting...';

    AngaziaAPI.plans.adminDelete(id).then(function () {
      showToast('Plan deleted successfully', 'success');
      asubCloseDelete();
      reloadPlans();
    }).catch(function (err) {
      console.error('Failed to delete plan:', err);
      showToast(err.message || 'Failed to delete plan', 'error');
    }).then(function () {
      btn.disabled = false;
      btn.textContent = 'Delete';
    });
  };

  window.asubToggleActive = function (id, isActive) {
    AngaziaAPI.plans.adminToggle(id, { is_active: isActive }).then(function () {
      showToast('Plan ' + (isActive ? 'activated' : 'deactivated') + ' successfully', 'success');
      reloadPlans();
    }).catch(function (err) {
      console.error('Failed to toggle plan:', err);
      showToast(err.message || 'Failed to toggle plan', 'error');
      reloadPlans();
    });
  };

  // ========== SUBSCRIPTION MANAGEMENT (new) ==========

  var assubPage = 1;
  var assubLimit = 20;
  var assubTotalPages = 1;

  function formatDate(dateStr) {
    if (!dateStr) return '';
    var d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: '2-digit', year: 'numeric' });
  }

  function renderSubRow(sub) {
    var statusClass = 'assub-status-' + (sub.status || 'unknown');
    var actionsHtml = '';

    actionsHtml += '<button class="asub-btn asub-btn-ghost" data-action="assubChangePlan" data-id="' + sub.id + '">Change Plan</button>';

    if (sub.status === 'active') {
      actionsHtml += '<button class="asub-btn asub-btn-danger" data-action="assubCancelSub" data-id="' + sub.id + '">Cancel</button>';
    }

    if (sub.status === 'cancelled' || sub.status === 'expired') {
      actionsHtml += '<button class="asub-btn asub-btn-primary" data-action="assubReactivateSub" data-id="' + sub.id + '">Reactivate</button>';
    }

    var userEmail = (sub.user && sub.user.email) ? sub.user.email : sub.user_id;
    var amount = sub.currency + ' ' + Number(sub.amount || 0).toFixed(2);

    return '<tr data-sub-id="' + sub.id + '">' +
      '<td class="assub-cell-user"><span class="assub-user-email">' + escapeHtml(userEmail) + '</span><span class="assub-user-id">' + sub.user_id + '</span></td>' +
      '<td><span class="assub-plan-name">' + escapeHtml(sub.plan_name) + '</span></td>' +
      '<td><span class="assub-status-badge ' + statusClass + '">' + (sub.status || 'unknown') + '</span></td>' +
      '<td>' + escapeHtml(amount) + '</td>' +
      '<td>' + formatDate(sub.start_date) + '</td>' +
      '<td>' + formatDate(sub.end_date) + '</td>' +
      '<td>' + (sub.auto_renew ? 'Yes' : 'No') + '</td>' +
      '<td class="assub-cell-actions">' + actionsHtml + '</td>' +
    '</tr>';
  }

  function renderSubTable(data) {
    var tbody = $('assub-tbody');
    var subs = (data && data.subscriptions) ? data.subscriptions : [];
    var total = (data && data.total) ? data.total : 0;
    assubTotalPages = (data && data.total_pages) ? data.total_pages : 1;

    if (subs.length === 0) {
      tbody.innerHTML = '<tr><td colspan="8" class="asub-table-empty">No subscriptions found.</td></tr>';
    } else {
      tbody.innerHTML = subs.map(renderSubRow).join('');
    }

    var pageInfo = $('assub-page-info');
    if (pageInfo) pageInfo.textContent = 'Total: ' + total;

    var currentPage = $('assub-current-page');
    if (currentPage) currentPage.textContent = 'Page ' + assubPage + ' of ' + assubTotalPages;

    var prevBtn = $('assub-prev-btn');
    var nextBtn = $('assub-next-btn');
    if (prevBtn) prevBtn.disabled = assubPage <= 1;
    if (nextBtn) nextBtn.disabled = assubPage >= assubTotalPages;
  }

  function assubLoadSubscriptions() {
    var params = {
      page: assubPage,
      limit: assubLimit
    };

    var userId = $('assub-filter-user') ? $('assub-filter-user').value.trim() : '';
    var planId = $('assub-filter-plan') ? $('assub-filter-plan').value : '';
    var status = $('assub-filter-status') ? $('assub-filter-status').value : '';

    if (userId) params.user_id = userId;
    if (planId) params.plan_id = planId;
    if (status) params.status = status;

    AngaziaAPI.admin.listSubscriptions(params).then(function (resp) {
      if (resp && resp.data) {
        renderSubTable(resp.data);
      } else if (resp && resp.subscriptions) {
        renderSubTable(resp);
      }
    }).catch(function (err) {
      console.error('Failed to load subscriptions:', err);
      showToast(err.message || 'Failed to load subscriptions', 'error');
    });
  }

  // Assign Subscription
  window.assubOpenAssign = function () {
    setVal('assub-assign-user', '');
    $('assub-assign-modal').style.display = 'flex';
  };

  window.assubCloseAssign = function () {
    $('assub-assign-modal').style.display = 'none';
  };

  window.assubConfirmAssign = function () {
    var userId = val('assub-assign-user').trim();
    var planId = val('assub-assign-plan');

    if (!userId) {
      showToast('Please enter a User ID', 'warning');
      return;
    }

    var btn = $('assub-assign-btn');
    btn.disabled = true;
    btn.textContent = 'Assigning...';

    AngaziaAPI.admin.assignSubscription({ user_id: userId, plan_id: planId })
      .then(function () {
        showToast('Subscription assigned successfully', 'success');
        assubCloseAssign();
        assubPage = 1;
        assubLoadSubscriptions();
      })
      .catch(function (err) {
        console.error('Failed to assign subscription:', err);
        showToast(err.message || 'Failed to assign subscription', 'error');
      })
      .then(function () {
        btn.disabled = false;
        btn.textContent = 'Assign';
      });
  };

  // Change Plan
  window.assubChangePlan = function (id) {
    setVal('assub-change-id', id);
    $('assub-change-modal').style.display = 'flex';
  };

  window.assubCloseChange = function () {
    $('assub-change-modal').style.display = 'none';
  };

  window.assubConfirmChange = function () {
    var id = val('assub-change-id');
    var planId = val('assub-change-plan');
    var btn = $('assub-change-btn');
    btn.disabled = true;
    btn.textContent = 'Changing...';

    AngaziaAPI.admin.changeSubscriptionPlan(id, { plan_id: planId })
      .then(function () {
        showToast('Plan changed successfully', 'success');
        assubCloseChange();
        assubLoadSubscriptions();
      })
      .catch(function (err) {
        console.error('Failed to change plan:', err);
        showToast(err.message || 'Failed to change plan', 'error');
      })
      .then(function () {
        btn.disabled = false;
        btn.textContent = 'Change Plan';
      });
  };

  // Cancel Subscription
  window.assubCancelSub = function (id) {
    setVal('assub-cancel-id', id);
    setVal('assub-cancel-reason', '');
    $('assub-cancel-modal').style.display = 'flex';
  };

  window.assubCloseCancel = function () {
    $('assub-cancel-modal').style.display = 'none';
  };

  window.assubConfirmCancel = function () {
    var id = val('assub-cancel-id');
    var reason = val('assub-cancel-reason').trim() || 'Cancelled by admin';
    var btn = $('assub-cancel-btn');
    btn.disabled = true;
    btn.textContent = 'Cancelling...';

    AngaziaAPI.admin.cancelSubscription(id, { reason: reason })
      .then(function () {
        showToast('Subscription cancelled', 'success');
        assubCloseCancel();
        assubLoadSubscriptions();
      })
      .catch(function (err) {
        console.error('Failed to cancel subscription:', err);
        showToast(err.message || 'Failed to cancel subscription', 'error');
      })
      .then(function () {
        btn.disabled = false;
        btn.textContent = 'Confirm Cancel';
      });
  };

  // Reactivate Subscription
  window.assubReactivateSub = function (id) {
    AngaziaAPI.admin.reactivateSubscription(id)
      .then(function () {
        showToast('Subscription reactivated', 'success');
        assubLoadSubscriptions();
      })
      .catch(function (err) {
        console.error('Failed to reactivate subscription:', err);
        showToast(err.message || 'Failed to reactivate subscription', 'error');
      });
  };

  // Pagination
  window.assubPrevPage = function () {
    if (assubPage > 1) {
      assubPage--;
      assubLoadSubscriptions();
    }
  };

  window.assubNextPage = function () {
    if (assubPage < assubTotalPages) {
      assubPage++;
      assubLoadSubscriptions();
    }
  };

  window.assubApplyFilters = function () {
    assubPage = 1;
    assubLoadSubscriptions();
  };

  // Event listeners
  document.addEventListener('click', function (e) {
    var el = e.target.closest('[data-action]');
    if (!el) return;
    var action = el.getAttribute('data-action');

    switch (action) {
      // Plan actions
      case 'asubOpenCreate':
        e.preventDefault();
        window.asubOpenCreate();
        break;
      case 'asubOpenEdit':
        e.preventDefault();
        window.asubOpenEdit(el.getAttribute('data-id'));
        break;
      case 'asubDelete':
        e.preventDefault();
        window.asubDelete(el.getAttribute('data-id'));
        break;
      case 'asubCloseModal':
        if (el.classList && el.classList.contains('asub-modal-overlay') && e.target !== el) break;
        window.asubCloseModal();
        break;
      case 'asubCloseDelete':
        if (el.classList && el.classList.contains('asub-modal-overlay') && e.target !== el) break;
        window.asubCloseDelete();
        break;
      case 'asubSave':
        e.preventDefault();
        window.asubSave();
        break;
      case 'asubConfirmDelete':
        e.preventDefault();
        window.asubConfirmDelete();
        break;

      // Subscription actions
      case 'assubOpenAssign':
        e.preventDefault();
        window.assubOpenAssign();
        break;
      case 'assubCloseAssign':
        if (el.classList && el.classList.contains('asub-modal-overlay') && e.target !== el) break;
        window.assubCloseAssign();
        break;
      case 'assubConfirmAssign':
        e.preventDefault();
        window.assubConfirmAssign();
        break;
      case 'assubChangePlan':
        e.preventDefault();
        window.assubChangePlan(el.getAttribute('data-id'));
        break;
      case 'assubCloseChange':
        if (el.classList && el.classList.contains('asub-modal-overlay') && e.target !== el) break;
        window.assubCloseChange();
        break;
      case 'assubConfirmChange':
        e.preventDefault();
        window.assubConfirmChange();
        break;
      case 'assubCancelSub':
        e.preventDefault();
        window.assubCancelSub(el.getAttribute('data-id'));
        break;
      case 'assubCloseCancel':
        if (el.classList && el.classList.contains('asub-modal-overlay') && e.target !== el) break;
        window.assubCloseCancel();
        break;
      case 'assubConfirmCancel':
        e.preventDefault();
        window.assubConfirmCancel();
        break;
      case 'assubReactivateSub':
        e.preventDefault();
        window.assubReactivateSub(el.getAttribute('data-id'));
        break;
      case 'assubPrevPage':
        e.preventDefault();
        window.assubPrevPage();
        break;
      case 'assubNextPage':
        e.preventDefault();
        window.assubNextPage();
        break;
      case 'assubApplyFilters':
        e.preventDefault();
        window.assubApplyFilters();
        break;
    }
  });

  document.addEventListener('change', function (e) {
    var el = e.target.closest('[data-action]');
    if (!el) return;
    var action = el.getAttribute('data-action');

    if (action === 'asubToggleActive') {
      window.asubToggleActive(el.getAttribute('data-id'), el.checked);
    }
  });

  // Initial load
  document.addEventListener('DOMContentLoaded', function () {
    reloadPlans();
    assubLoadSubscriptions();
  });
})();
