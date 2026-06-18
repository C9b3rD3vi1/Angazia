(function () {
  'use strict';

  var state = {
    jobs: [],
    page: 1,
    perPage: 20,
    total: 0,
    totalPages: 0,
    selectedIds: [],
    pendingJobId: null,
    loading: false,
    searchQuery: '',
    statusFilter: '',
    employmentType: '',
    experienceLevel: '',
    activeTab: 'all',
  };

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  function escapeHtml(str) {
    if (!str) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function getInitials(name) {
    if (!name) return '?';
    var parts = name.split(/\s+/);
    if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
    return name.substring(0, 2).toUpperCase();
  }

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    try {
      var d = new Date(dateStr);
      return d.toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
    } catch (e) {
      return dateStr;
    }
  }

  function statusBadgeClass(status) {
    switch (status) {
      case 'active': return 'active';
      case 'pending': return 'pending';
      case 'closed': return 'closed';
      default: return 'draft';
    }
  }

  function fetchStats() {
    AngaziaAPI.admin.jobStats()
      .then(function (res) {
        var data = res && res.data ? res.data : res;
        var el = function (id) { return document.getElementById(id); };
        if (data) {
          if (el('aj-stat-total')) el('aj-stat-total').textContent = (data.status_active || 0) + (data.status_inactive || 0);
          if (el('aj-stat-active')) el('aj-stat-active').textContent = data.status_active || 0;
          if (el('aj-stat-pending')) el('aj-stat-pending').textContent = 0;
          if (el('aj-stat-closed')) el('aj-stat-closed').textContent = data.status_inactive || 0;
        }
      })
      .catch(function () {});
  }

  function fetchJobs() {
    state.loading = true;
    showLoading();
    hideError();

    var params = {
      page: state.page,
      limit: state.perPage,
    };
    if (state.activeTab !== 'all') params.status = state.activeTab;
    if (state.searchQuery) params.search = state.searchQuery;
    if (state.employmentType) params.employment_type = state.employmentType;
    if (state.experienceLevel) params.experience_level = state.experienceLevel;

    AngaziaAPI.admin.jobs(params)
      .then(function (res) {
        var data = res && res.data ? res.data : res;
        state.jobs = (data && data.jobs) || [];
        state.total = data ? data.total : 0;
        state.totalPages = data ? data.total_pages : 0;

        hideLoading();
        render();
      })
      .catch(function (err) {
        state.loading = false;
        hideLoading();
        showError(err.message || 'Failed to load jobs');
      });
  }

  function render() {
    var tbody = document.getElementById('aj-table-body');
    var totalEl = document.getElementById('aj-total-count');
    var emptyEl = document.getElementById('aj-empty');
    var table = document.getElementById('aj-table');

    if (!state.jobs.length) {
      if (emptyEl) emptyEl.style.display = 'flex';
      if (table) table.style.display = 'none';
      if (totalEl) totalEl.textContent = '0';
      hidePagination();
      updateBulkBar();
      return;
    }

    if (emptyEl) emptyEl.style.display = 'none';
    if (table) table.style.display = '';
    if (totalEl) totalEl.textContent = state.total.toString();

    var html = '';
    for (var i = 0; i < state.jobs.length; i++) {
      var j = state.jobs[i];
      var sClass = statusBadgeClass(j.status);
      var postedDate = formatDate(j.posted_at || j.postedAt);
      var companyLink = '/admin/companies/' + escapeHtml(j.company_id || j.employer_id || '');
      var jobLink = '/admin/jobs/' + escapeHtml(j.id);
      var employmentType = j.employment_type || j.employmentType || '';
      var hasPending = j.status === 'pending';

      html += '<tr data-job-id="' + escapeHtml(j.id) + '">';
      html += '  <td class="aj-th-check"><input type="checkbox" class="aj-row-check" data-job-id="' + escapeHtml(j.id) + '" /></td>';
      html += '  <td><a href="' + jobLink + '" class="aj-table-link"><span class="aj-table-title">' + escapeHtml(j.title) + '</span></a></td>';
      html += '  <td><a href="' + companyLink + '" class="aj-table-company">';
      if (j.company_logo) {
        html += '    <img src="' + escapeHtml(j.company_logo) + '" alt="" class="aj-table-clogo">';
      } else {
        html += '    <span class="aj-table-clogo-placeholder">' + getInitials(j.company_name) + '</span>';
      }
      html += '    <span>' + escapeHtml(j.company_name || 'Unknown') + '</span>';
      html += '  </a></td>';
      html += '  <td><span class="aj-status-badge ' + sClass + '">' + (j.status || 'unknown') + '</span></td>';
      html += '  <td><span class="aj-table-muted">' + escapeHtml(employmentType || '-') + '</span></td>';
      html += '  <td><span class="aj-table-number">' + (j.applications_count || j.applicationsCount || 0) + '</span></td>';
      html += '  <td><span class="aj-table-number">' + (j.views_count || j.viewsCount || 0) + '</span></td>';
      html += '  <td><span class="aj-table-muted">' + postedDate + '</span></td>';
      html += '  <td><div class="aj-row-actions">';
      if (hasPending) {
        html += '    <button class="aj-btn-sm aj-btn-sm-accept" data-action="approve" data-job-id="' + escapeHtml(j.id) + '">Approve</button>';
        html += '    <button class="aj-btn-sm aj-btn-sm-reject" data-action="reject" data-job-id="' + escapeHtml(j.id) + '">Reject</button>';
      }
      html += '    <a href="' + jobLink + '" class="aj-btn-sm aj-btn-sm-ghost">View</a>';
      html += '  </div></td>';
      html += '</tr>';
    }

    tbody.innerHTML = html;
    updatePagination();
    updateBulkBar();
  }

  function showLoading() {
    var loadingEl = document.getElementById('aj-loading');
    var table = document.getElementById('aj-table');
    var emptyEl = document.getElementById('aj-empty');
    if (loadingEl) loadingEl.style.display = 'flex';
    if (table) table.style.display = 'none';
    if (emptyEl) emptyEl.style.display = 'none';
    hidePagination();
  }

  function hideLoading() {
    var loadingEl = document.getElementById('aj-loading');
    if (loadingEl) loadingEl.style.display = 'none';
  }

  function showError(msg) {
    var errorEl = document.getElementById('aj-error');
    var errorText = document.getElementById('aj-error-text');
    var table = document.getElementById('aj-table');
    var emptyEl = document.getElementById('aj-empty');
    if (errorEl) errorEl.style.display = 'block';
    if (errorText && msg) errorText.textContent = msg;
    if (table) table.style.display = 'none';
    if (emptyEl) emptyEl.style.display = 'none';
    hidePagination();
  }

  function hideError() {
    var errorEl = document.getElementById('aj-error');
    if (errorEl) errorEl.style.display = 'none';
  }

  function hidePagination() {
    var pagiEl = document.getElementById('aj-pagination');
    if (pagiEl) pagiEl.style.display = 'none';
  }

  function updatePagination() {
    var pagiEl = document.getElementById('aj-pagination');
    var infoEl = document.getElementById('aj-pagi-info');
    var btnsEl = document.getElementById('aj-pagi-btns');
    if (!pagiEl || !infoEl || !btnsEl) return;

    if (state.totalPages <= 1) {
      pagiEl.style.display = 'none';
      return;
    }
    pagiEl.style.display = 'flex';

    var start = (state.page - 1) * state.perPage + 1;
    var end = Math.min(state.page * state.perPage, state.total);
    infoEl.textContent = start + '-' + end + ' of ' + state.total;

    btnsEl.innerHTML = '';

    var prevBtn = document.createElement('button');
    prevBtn.className = 'aj-pagi-btn';
    prevBtn.textContent = 'Prev';
    prevBtn.disabled = state.page <= 1;
    prevBtn.addEventListener('click', function () {
      if (state.page > 1) { state.page--; fetchJobs(); }
    });
    btnsEl.appendChild(prevBtn);

    var maxVisible = 5;
    var ps = Math.max(1, state.page - Math.floor(maxVisible / 2));
    var pe = Math.min(state.totalPages, ps + maxVisible - 1);
    if (pe - ps < maxVisible - 1) ps = Math.max(1, pe - maxVisible + 1);

    if (ps > 1) {
      var firstBtn = document.createElement('button');
      firstBtn.className = 'aj-pagi-btn';
      firstBtn.textContent = '1';
      firstBtn.addEventListener('click', function () { state.page = 1; fetchJobs(); });
      btnsEl.appendChild(firstBtn);
      if (ps > 2) {
        var dots = document.createElement('span');
        dots.className = 'aj-pagi-btn';
        dots.style.border = 'none';
        dots.style.cursor = 'default';
        dots.textContent = '...';
        btnsEl.appendChild(dots);
      }
    }

    for (var i = ps; i <= pe; i++) {
      (function (pageNum) {
        var btn = document.createElement('button');
        btn.className = 'aj-pagi-btn' + (pageNum === state.page ? ' active' : '');
        btn.textContent = pageNum;
        btn.addEventListener('click', function () { state.page = pageNum; fetchJobs(); });
        btnsEl.appendChild(btn);
      })(i);
    }

    if (pe < state.totalPages) {
      if (pe < state.totalPages - 1) {
        var dots2 = document.createElement('span');
        dots2.className = 'aj-pagi-btn';
        dots2.style.border = 'none';
        dots2.style.cursor = 'default';
        dots2.textContent = '...';
        btnsEl.appendChild(dots2);
      }
      var lastBtn = document.createElement('button');
      lastBtn.className = 'aj-pagi-btn';
      lastBtn.textContent = state.totalPages;
      lastBtn.addEventListener('click', function () { state.page = state.totalPages; fetchJobs(); });
      btnsEl.appendChild(lastBtn);
    }

    var nextBtn = document.createElement('button');
    nextBtn.className = 'aj-pagi-btn';
    nextBtn.textContent = 'Next';
    nextBtn.disabled = state.page >= state.totalPages;
    nextBtn.addEventListener('click', function () {
      if (state.page < state.totalPages) { state.page++; fetchJobs(); }
    });
    btnsEl.appendChild(nextBtn);
  }

  function updateBulkBar() {
    var bulkEl = document.getElementById('aj-bulk');
    var countEl = document.getElementById('aj-bulk-count');
    if (!bulkEl || !countEl) return;
    if (state.selectedIds.length === 0) {
      bulkEl.style.display = 'none';
      return;
    }
    bulkEl.style.display = 'flex';
    countEl.textContent = state.selectedIds.length + ' selected';
  }

  function toggleSelect(jobId) {
    var idx = state.selectedIds.indexOf(jobId);
    if (idx === -1) {
      state.selectedIds.push(jobId);
    } else {
      state.selectedIds.splice(idx, 1);
    }
    updateBulkBar();
  }

  function handleAction(action, jobId) {
    if (action === 'approve') {
      var btn = document.querySelector('[data-action="approve"][data-job-id="' + jobId + '"]');
      if (btn) btn.disabled = true;
      AngaziaAPI.admin.approveContent(jobId)
        .then(function () {
          showToast('Job listing approved!', 'success');
        })
        .catch(function (err) {
          showToast(err.message || 'Failed to approve job', 'error');
        })
        .then(function () { if (btn) btn.disabled = false; fetchJobs(); });
    } else if (action === 'reject') {
      state.pendingJobId = jobId;
      var reasonEl = document.getElementById('aj-reject-reason');
      if (reasonEl) reasonEl.value = '';
      document.getElementById('aj-reject-modal').style.display = 'flex';
    }
  }

  function submitReject() {
    if (!state.pendingJobId) return;
    var reason = document.getElementById('aj-reject-reason').value.trim();
    var confirmBtn = document.getElementById('aj-modal-confirm');
    confirmBtn.disabled = true;
    AngaziaAPI.admin.rejectContent(state.pendingJobId, { reason: reason })
      .then(function () {
        showToast('Job listing rejected', 'success');
        closeModal();
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to reject job', 'error');
      })
      .then(function () {
        confirmBtn.disabled = false;
        state.pendingJobId = null;
        fetchJobs();
      });
  }

  function closeModal() {
    document.getElementById('aj-reject-modal').style.display = 'none';
    state.pendingJobId = null;
  }

  function handleBulk(action) {
    var ids = state.selectedIds;
    if (!ids.length) return;
    var todo = 0;
    var done = 0;
    var errors = 0;
    ids.forEach(function (id) {
      todo++;
      var endpoint = '/admin/moderation/' + id + '/' + action;
      AngaziaAPI.post(endpoint)
        .then(function () {
          done++;
          if (done + errors === todo) {
            showToast('Bulk ' + action + ' completed for ' + done + ' job(s)', done > 0 ? 'success' : 'error');
          }
        })
        .catch(function () {
          errors++;
          if (done + errors === todo) {
            showToast('Completed with ' + errors + ' error(s)', errors === todo ? 'error' : 'warning');
          }
        });
    });
    state.selectedIds = [];
    updateBulkBar();
    setTimeout(fetchJobs, 500);
  }

  function switchTab(tab) {
    state.activeTab = tab;
    state.page = 1;
    var tabs = document.querySelectorAll('.aj-tab');
    tabs.forEach(function (t) {
      t.classList.toggle('active', t.getAttribute('data-tab') === tab);
    });
    fetchJobs();
  }

  function applyFilters() {
    state.page = 1;
    state.searchQuery = document.getElementById('aj-search').value.trim();
    state.statusFilter = document.getElementById('aj-status-filter').value;
    state.employmentType = document.getElementById('aj-type-filter').value;
    state.experienceLevel = document.getElementById('aj-exp-filter').value;
    fetchJobs();
  }

  document.addEventListener('DOMContentLoaded', function () {
    fetchStats();
    fetchJobs();

    var reloadBtn = document.querySelector('[data-action="reload"]');
    if (reloadBtn) {
      reloadBtn.addEventListener('click', function (e) {
        e.preventDefault();
        fetchJobs();
      });
    }

    var retryBtn = document.querySelector('[data-action="retry"]');
    if (retryBtn) {
      retryBtn.addEventListener('click', function () {
        fetchJobs();
      });
    }

    document.addEventListener('click', function (e) {
      var btn = e.target.closest('[data-action]');
      if (!btn) return;
      var action = btn.getAttribute('data-action');
      var jobId = btn.getAttribute('data-job-id');
      if ((action === 'approve' || action === 'reject') && jobId) {
        handleAction(action, jobId);
      }
    });

    document.getElementById('aj-modal-confirm').addEventListener('click', submitReject);
    document.getElementById('aj-modal-close').addEventListener('click', closeModal);
    document.getElementById('aj-modal-cancel').addEventListener('click', closeModal);
    document.getElementById('aj-reject-modal').addEventListener('click', function (e) {
      if (e.target === this) closeModal();
    });

    document.getElementById('aj-apply-filter').addEventListener('click', applyFilters);
    document.getElementById('aj-search').addEventListener('keypress', function (e) {
      if (e.key === 'Enter') applyFilters();
    });

    var tabs = document.querySelectorAll('.aj-tab');
    tabs.forEach(function (tab) {
      tab.addEventListener('click', function () {
        switchTab(this.getAttribute('data-tab'));
      });
    });

    document.getElementById('aj-select-all').addEventListener('change', function (e) {
      var checked = e.target.checked;
      var checkboxes = document.querySelectorAll('.aj-row-check');
      state.selectedIds = [];
      if (checked) {
        checkboxes.forEach(function (cb) {
          state.selectedIds.push(cb.getAttribute('data-job-id'));
        });
      }
      updateBulkBar();
    });

    document.addEventListener('change', function (e) {
      if (e.target.classList.contains('aj-row-check')) {
        var jobId = e.target.getAttribute('data-job-id');
        if (jobId) toggleSelect(jobId);
      }
    });

    document.getElementById('aj-bulk-approve').addEventListener('click', function () { handleBulk('approve'); });
    document.getElementById('aj-bulk-reject').addEventListener('click', function () { handleBulk('reject'); });
    document.getElementById('aj-bulk-clear').addEventListener('click', function () {
      state.selectedIds = [];
      var selectAll = document.getElementById('aj-select-all');
      if (selectAll) selectAll.checked = false;
      updateBulkBar();
    });
  });
})();
