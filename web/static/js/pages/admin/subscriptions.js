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

  // ========== SUBSCRIPTION STATS ==========

  function assubLoadStats() {
    AngaziaAPI.admin.getSubscriptionStats().then(function (stats) {
      if (!stats) return;
      $('assub-stat-total').textContent = stats.total_subscriptions || 0;
      $('assub-stat-active').textContent = stats.status_breakdown ? (stats.status_breakdown.active || 0) : '-';
      $('assub-stat-pending').textContent = stats.status_breakdown ? (stats.status_breakdown.pending || 0) : '-';
      $('assub-stat-cancelled').textContent = stats.status_breakdown ? (stats.status_breakdown.cancelled || 0) : '-';
      $('assub-stat-revenue').textContent = stats.currency + ' ' + Number(stats.revenue_this_month || 0).toLocaleString();
      $('assub-stat-stale').textContent = stats.stale_pending_count || 0;
    }).catch(function (err) {
      console.error('Failed to load subscription stats:', err);
    });
  }

  // ========== SUBSCRIPTION MANAGEMENT ==========

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

    actionsHtml += '<button class="asub-btn asub-btn-ghost" data-action="assubViewDetail" data-id="' + sub.id + '">View</button>';
    actionsHtml += '<button class="asub-btn asub-btn-ghost" data-action="assubOpenEditSub" data-id="' + sub.id + '">Edit</button>';
    actionsHtml += '<button class="asub-btn asub-btn-ghost" data-action="assubChangePlan" data-id="' + sub.id + '">Plan</button>';

    if (sub.pending_plan_id) {
      actionsHtml += '<button class="asub-btn asub-btn-primary" data-action="assubCompletePending" data-id="' + sub.id + '">Apply Pending</button>';
      actionsHtml += '<button class="asub-btn asub-btn-danger" data-action="assubCancelPending" data-id="' + sub.id + '">Clear Pending</button>';
    }

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
    var sel = $('assub-assign-user');
    sel.innerHTML = '<option value="">Loading users...</option>';
    sel.disabled = true;

    AngaziaAPI.admin.users({ limit: 500 })
      .then(function (data) {
        var users = data.users || [];
        sel.innerHTML = '<option value="">Select a user...</option>';
        users.forEach(function (u) {
          var opt = document.createElement('option');
          opt.value = u.id;
          opt.textContent = u.email + (u.full_name ? ' (' + u.full_name + ')' : '');
          sel.appendChild(opt);
        });
        sel.disabled = false;
      })
      .catch(function () {
        sel.innerHTML = '<option value="">Failed to load users</option>';
        sel.disabled = false;
      });

    $('assub-assign-modal').style.display = 'flex';
  };

  window.assubCloseAssign = function () {
    $('assub-assign-modal').style.display = 'none';
  };

  window.assubConfirmAssign = function () {
    var userId = val('assub-assign-user');
    var planId = val('assub-assign-plan');

    if (!userId) {
      showToast('Please select a user', 'warning');
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

  // Reconcile Pending Subscriptions
  window.assubReconcile = function () {
    var btn = $('assub-reconcile-btn');
    btn.disabled = true;
    btn.textContent = 'Reconciling...';

    AngaziaAPI.admin.reconcileSubscriptions().then(function (result) {
      var msg = result ? 'Reconciled: ' + (result.reconciled || 0) + ' subscriptions' : 'Reconciliation complete';
      showToast(msg, 'success');
      assubLoadSubscriptions();
      assubLoadStats();
    }).catch(function (err) {
      console.error('Failed to reconcile:', err);
      showToast(err.message || 'Reconciliation failed', 'error');
    }).then(function () {
      btn.disabled = false;
      btn.textContent = 'Reconcile Pending';
    });
  };

  // ========== EDIT SUBSCRIPTION ==========

  window.assubOpenEditSub = function (id) {
    AngaziaAPI.admin.getSubscription(id).then(function (sub) {
      setVal('assub-edit-id', sub.id || '');
      setVal('assub-edit-status', sub.status || 'active');
      setVal('assub-edit-plan-name', sub.plan_name || '');
      setVal('assub-edit-plan-id', sub.plan_id || '');
      setVal('assub-edit-amount', sub.amount || 0);
      setVal('assub-edit-currency', sub.currency || 'KES');
      setVal('assub-edit-interval', sub.interval || 'month');
      setVal('assub-edit-job-limit', sub.job_post_limit || 0);
      $('assub-edit-auto-renew').checked = sub.auto_renew !== false;
      $('assub-edit-modal').style.display = 'flex';
    }).catch(function (err) {
      showToast(err.message || 'Failed to load subscription', 'error');
    });
  };

  window.assubCloseEdit = function () {
    $('assub-edit-modal').style.display = 'none';
  };

  window.assubConfirmEdit = function () {
    var id = val('assub-edit-id');
    var data = {};
    var status = val('assub-edit-status');
    var planName = val('assub-edit-plan-name').trim();
    var planId = val('assub-edit-plan-id').trim();
    var amount = parseFloat(val('assub-edit-amount'));
    var currency = val('assub-edit-currency').trim();
    var interval = val('assub-edit-interval');
    var jobLimit = parseInt(val('assub-edit-job-limit'));
    var autoRenew = $('assub-edit-auto-renew').checked;

    if (status) data.status = status;
    if (planName) data.plan_name = planName;
    if (planId) data.plan_id = planId;
    if (!isNaN(amount)) data.amount = amount;
    if (currency) data.currency = currency;
    if (interval) data.interval = interval;
    if (!isNaN(jobLimit)) data.job_post_limit = jobLimit;
    data.auto_renew = autoRenew;

    var btn = $('assub-edit-save-btn');
    btn.disabled = true;
    btn.textContent = 'Saving...';

    AngaziaAPI.admin.updateSubscription(id, data).then(function () {
      showToast('Subscription updated', 'success');
      assubCloseEdit();
      assubLoadSubscriptions();
    }).catch(function (err) {
      showToast(err.message || 'Failed to update subscription', 'error');
    }).then(function () {
      btn.disabled = false;
      btn.textContent = 'Save Changes';
    });
  };

  // ========== PENDING PLAN MANAGEMENT ==========

  window.assubCompletePending = function (id) {
    if (!confirm('Apply the pending plan change to this subscription?')) return;
    AngaziaAPI.admin.completePendingUpgrade(id).then(function () {
      showToast('Pending plan change applied', 'success');
      assubLoadSubscriptions();
    }).catch(function (err) {
      showToast(err.message || 'Failed to apply pending change', 'error');
    });
  };

  window.assubCancelPending = function (id) {
    if (!confirm('Clear the pending plan change for this subscription?')) return;
    AngaziaAPI.admin.cancelPendingUpgrade(id).then(function () {
      showToast('Pending plan change cleared', 'success');
      assubLoadSubscriptions();
    }).catch(function (err) {
      showToast(err.message || 'Failed to clear pending change', 'error');
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

  // ========== SUBSCRIPTION DETAIL VIEW ==========

  var assubDetailData = null;

  window.assubViewDetail = function (id) {
    var modal = $('assub-detail-modal');
    var loading = $('assub-detail-loading');
    var content = $('assub-detail-content');

    modal.style.display = 'flex';
    loading.style.display = 'block';
    content.style.display = 'none';

    // Reset tabs to info
    document.querySelectorAll('.assub-detail-tab').forEach(function (t) { t.classList.remove('active'); });
    document.querySelector('.assub-detail-tab[data-tab="info"]').classList.add('active');
    document.querySelectorAll('.assub-detail-panel').forEach(function (p) { p.classList.remove('active'); });
    $('assub-detail-panel-info').classList.add('active');

    AngaziaAPI.admin.getSubscriptionDetail(id).then(function (data) {
      assubDetailData = data;
      renderAssubDetail(data);
      loading.style.display = 'none';
      content.style.display = 'block';
    }).catch(function (err) {
      console.error('Failed to load subscription detail:', err);
      loading.textContent = 'Failed to load details: ' + (err.message || 'Unknown error');
      showToast(err.message || 'Failed to load subscription details', 'error');
    });
  };

  function renderAssubDetail(data) {
    var sub = data.subscription || {};
    var user = sub.user || {};

    $('assub-dd-user').innerHTML = '<span class="assub-user-email">' + escapeHtml(user.email || '') + '</span><span class="assub-user-id">' + escapeHtml(sub.user_id || '') + '</span>';
    $('assub-dd-plan').textContent = sub.plan_name || '';
    $('assub-dd-status').innerHTML = '<span class="assub-status-badge assub-status-' + (sub.status || 'unknown') + '">' + (sub.status || 'unknown') + '</span>';
    $('assub-dd-amount').textContent = (sub.currency || '') + ' ' + Number(sub.amount || 0).toFixed(2);
    $('assub-dd-interval').textContent = sub.interval || '';
    $('assub-dd-autorenew').textContent = sub.auto_renew ? 'Yes' : 'No';
    $('assub-dd-start').textContent = formatDate(sub.start_date);
    $('assub-dd-end').textContent = formatDate(sub.end_date);
    $('assub-dd-joblimit').textContent = sub.job_post_limit != null ? sub.job_post_limit : 'N/A';
    $('assub-dd-pendingplan').textContent = sub.pending_plan_id || 'None';

    // Usage
    var usage = data.usage || [];
    var usageBody = $('assub-dd-usage-body');
    var usageTable = $('assub-dd-usage-table');
    var usageEmpty = $('assub-dd-usage-empty');

    if (usage.length) {
      usageTable.style.display = '';
      usageEmpty.style.display = 'none';
      usageBody.innerHTML = usage.map(function (u) {
        return '<tr>' +
          '<td>' + escapeHtml(u.metric_key || '') + '</td>' +
          '<td>' + (u.used || 0) + '</td>' +
          '<td>' + (u.limit || 0) + '</td>' +
          '<td>' + formatDate(u.period_start) + '</td>' +
          '<td>' + formatDate(u.period_end) + '</td>' +
        '</tr>';
      }).join('');
    } else {
      usageTable.style.display = 'none';
      usageEmpty.style.display = '';
    }

    // Payments
    var payments = data.payments || [];
    var paymentsBody = $('assub-dd-payments-body');
    var paymentsTable = $('assub-dd-payments-table');
    var paymentsEmpty = $('assub-dd-payments-empty');

    if (payments.length) {
      paymentsTable.style.display = '';
      paymentsEmpty.style.display = 'none';
      paymentsBody.innerHTML = payments.map(function (p) {
        return '<tr>' +
          '<td>' + escapeHtml(p.reference || p.id) + '</td>' +
          '<td>' + escapeHtml(p.currency || '') + ' ' + Number(p.amount || 0).toFixed(2) + '</td>' +
          '<td><span class="assub-status-badge assub-status-' + (p.status || 'unknown') + '">' + (p.status || 'unknown') + '</span></td>' +
          '<td>' + escapeHtml(p.payment_method || '-') + '</td>' +
          '<td>' + formatDate(p.created_at || p.paid_at) + '</td>' +
        '</tr>';
      }).join('');
    } else {
      paymentsTable.style.display = 'none';
      paymentsEmpty.style.display = '';
    }

    // Timeline
    var history = data.history || [];
    var historyList = $('assub-dd-history-list');
    var historyEmpty = $('assub-dd-history-empty');

    if (history.length) {
      historyList.style.display = '';
      historyEmpty.style.display = 'none';
      historyList.innerHTML = history.map(function (h) {
        var label = h.action || 'unknown';
        var detail = '';
        if (h.old_plan_id && h.new_plan_id) {
          detail = escapeHtml(h.old_plan_id) + ' \u2192 ' + escapeHtml(h.new_plan_id);
        } else if (h.new_plan_id) {
          detail = 'Plan: ' + escapeHtml(h.new_plan_id);
        }
        if (h.old_amount && h.new_amount) {
          detail += (detail ? ' | ' : '') + Number(h.old_amount).toFixed(2) + ' \u2192 ' + Number(h.new_amount).toFixed(2);
        }
        return '<div class="assub-timeline-item">' +
          '<div class="assub-timeline-dot assub-timeline-dot-' + label + '"></div>' +
          '<div class="assub-timeline-body">' +
            '<div class="assub-timeline-action">' + escapeHtml(label) + '</div>' +
            (detail ? '<div class="assub-timeline-detail">' + detail + '</div>' : '') +
            '<div class="assub-timeline-date">' + formatDate(h.created_at) + '</div>' +
          '</div>' +
        '</div>';
      }).join('');
    } else {
      historyList.style.display = 'none';
      historyEmpty.style.display = '';
    }
  }

  // Tab switching within detail modal
  document.addEventListener('click', function (e) {
    var tab = e.target.closest('.assub-detail-tab');
    if (!tab) return;

    var tabName = tab.getAttribute('data-tab');
    document.querySelectorAll('.assub-detail-tab').forEach(function (t) { t.classList.remove('active'); });
    tab.classList.add('active');
    document.querySelectorAll('.assub-detail-panel').forEach(function (p) { p.classList.remove('active'); });
    var panel = $('assub-detail-panel-' + tabName);
    if (panel) panel.classList.add('active');
  });

  window.assubCloseDetail = function () {
    $('assub-detail-modal').style.display = 'none';
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
      case 'assubViewDetail':
        e.preventDefault();
        window.assubViewDetail(el.getAttribute('data-id'));
        break;
      case 'assubCloseDetail':
        if (el.classList && el.classList.contains('asub-modal-overlay') && e.target !== el) break;
        window.assubCloseDetail();
        break;
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
      case 'assubReconcile':
        e.preventDefault();
        window.assubReconcile();
        break;
      case 'assubOpenEditSub':
        e.preventDefault();
        window.assubOpenEditSub(el.getAttribute('data-id'));
        break;
      case 'assubCloseEdit':
        if (el.classList && el.classList.contains('asub-modal-overlay') && e.target !== el) break;
        window.assubCloseEdit();
        break;
      case 'assubConfirmEdit':
        e.preventDefault();
        window.assubConfirmEdit();
        break;
      case 'assubCompletePending':
        e.preventDefault();
        window.assubCompletePending(el.getAttribute('data-id'));
        break;
      case 'assubCancelPending':
        e.preventDefault();
        window.assubCancelPending(el.getAttribute('data-id'));
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
    assubLoadStats();
  });
})();
