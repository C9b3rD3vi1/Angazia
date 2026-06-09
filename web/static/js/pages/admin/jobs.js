(function () {
  'use strict';

  var state = {
    allJobs: [],
    filteredJobs: [],
    displayedJobs: [],
    selectedIds: [],
    page: 1,
    perPage: 20,
    pendingJobId: null,
    bulkAction: null,
  };

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  function collectJobsFromDOM() {
    var rows = document.querySelectorAll('#aj-table-body tr[data-job-id]');
    var jobs = [];
    rows.forEach(function (tr) {
      var jobId = tr.getAttribute('data-job-id');
      var titleEl = tr.querySelector('.aj-table-title');
      var companyEl = tr.querySelector('.aj-table-company span:last-child');
      var statusEl = tr.querySelector('.aj-status-badge');
      var appsEl = tr.querySelector('.aj-table-number');
      var viewsEl = tr.querySelectorAll('.aj-table-number');
      var dateEl = tr.querySelector('.aj-table-muted');
      var approveBtn = tr.querySelector('[data-action="approve"]');
      var rejectBtn = tr.querySelector('[data-action="reject"]');
      jobs.push({
        id: jobId,
        title: titleEl ? titleEl.textContent.trim().toLowerCase() : '',
        company: companyEl ? companyEl.textContent.trim().toLowerCase() : '',
        status: statusEl ? statusEl.textContent.trim().toLowerCase() : '',
        applications: appsEl ? parseInt(appsEl.textContent, 10) : 0,
        views: viewsEl.length > 1 ? parseInt(viewsEl[1].textContent, 10) : 0,
        date: dateEl ? dateEl.textContent.trim() : '',
        row: tr,
        hasApprove: !!approveBtn,
        hasReject: !!rejectBtn,
      });
    });
    return jobs;
  }

  function renderTable() {
    var tbody = document.getElementById('aj-table-body');
    var totalEl = document.getElementById('aj-total-count');
    var emptyEl = document.getElementById('aj-empty');
    var table = document.getElementById('aj-table');
    var noDataRow = tbody ? tbody.querySelector('.aj-no-data-row') : null;

    if (!state.filteredJobs.length) {
      if (emptyEl) emptyEl.style.display = 'flex';
      if (table) table.style.display = 'none';
      if (totalEl) totalEl.textContent = '0';
      return;
    }
    if (emptyEl) emptyEl.style.display = 'none';
    if (table) table.style.display = '';

    var start = (state.page - 1) * state.perPage;
    var end = Math.min(start + state.perPage, state.filteredJobs.length);
    state.displayedJobs = state.filteredJobs.slice(start, end);

    var ids = {};
    state.displayedJobs.forEach(function (j) { ids[j.id] = true; });

    state.allJobs.forEach(function (j) {
      var show = !!ids[j.id];
      j.row.style.display = show ? '' : 'none';
    });

    if (totalEl) totalEl.textContent = state.filteredJobs.length;
    if (noDataRow) noDataRow.style.display = 'none';
    updatePagination();
    updateBulkBar();
  }

  function updatePagination() {
    var pagiEl = document.getElementById('aj-pagination');
    var infoEl = document.getElementById('aj-pagi-info');
    var btnsEl = document.getElementById('aj-pagi-btns');
    if (!pagiEl || !infoEl || !btnsEl) return;

    var total = state.filteredJobs.length;
    var totalPages = Math.ceil(total / state.perPage);
    if (totalPages <= 1) {
      pagiEl.style.display = 'none';
      return;
    }
    pagiEl.style.display = 'flex';

    var start = (state.page - 1) * state.perPage + 1;
    var end = Math.min(state.page * state.perPage, total);
    infoEl.textContent = start + '-' + end + ' of ' + total;

    btnsEl.innerHTML = '';
    var prevBtn = document.createElement('button');
    prevBtn.className = 'aj-pagi-btn';
    prevBtn.textContent = 'Prev';
    prevBtn.disabled = state.page <= 1;
    prevBtn.addEventListener('click', function () {
      if (state.page > 1) { state.page--; renderTable(); }
    });
    btnsEl.appendChild(prevBtn);

    var maxVisible = 5;
    var ps = Math.max(1, state.page - Math.floor(maxVisible / 2));
    var pe = Math.min(totalPages, ps + maxVisible - 1);
    if (pe - ps < maxVisible - 1) ps = Math.max(1, pe - maxVisible + 1);

    if (ps > 1) {
      var firstBtn = document.createElement('button');
      firstBtn.className = 'aj-pagi-btn';
      firstBtn.textContent = '1';
      firstBtn.addEventListener('click', function () { state.page = 1; renderTable(); });
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
        btn.addEventListener('click', function () { state.page = pageNum; renderTable(); });
        btnsEl.appendChild(btn);
      })(i);
    }

    if (pe < totalPages) {
      if (pe < totalPages - 1) {
        var dots2 = document.createElement('span');
        dots2.className = 'aj-pagi-btn';
        dots2.style.border = 'none';
        dots2.style.cursor = 'default';
        dots2.textContent = '...';
        btnsEl.appendChild(dots2);
      }
      var lastBtn = document.createElement('button');
      lastBtn.className = 'aj-pagi-btn';
      lastBtn.textContent = totalPages;
      lastBtn.addEventListener('click', function () { state.page = totalPages; renderTable(); });
      btnsEl.appendChild(lastBtn);
    }

    var nextBtn = document.createElement('button');
    nextBtn.className = 'aj-pagi-btn';
    nextBtn.textContent = 'Next';
    nextBtn.disabled = state.page >= totalPages;
    nextBtn.addEventListener('click', function () {
      if (state.page < totalPages) { state.page++; renderTable(); }
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

    var allChecked = true;
    state.displayedJobs.forEach(function (j) {
      if (state.selectedIds.indexOf(j.id) === -1) allChecked = false;
    });
    var selectAll = document.getElementById('aj-select-all');
    if (selectAll) selectAll.checked = allChecked && state.displayedJobs.length > 0;
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
      AngaziaAPI.post('/admin/moderation/' + jobId + '/approve')
        .then(function () {
          showToast('Job listing approved!', 'success');
          removeJobRow(jobId);
        })
        .catch(function (err) {
          showToast(err.message || 'Failed to approve job', 'error');
        })
        .then(function () { if (btn) btn.disabled = false; });
    } else if (action === 'reject') {
      state.pendingJobId = jobId;
      document.getElementById('aj-reject-reason').value = '';
      document.getElementById('aj-reject-modal').style.display = 'flex';
    }
  }

  function removeJobRow(jobId) {
    var idx = state.allJobs.findIndex(function (j) { return j.id === jobId; });
    if (idx !== -1) {
      var job = state.allJobs[idx];
      if (job.row && job.row.remove) job.row.remove();
      state.allJobs.splice(idx, 1);
    }
    state.filteredJobs = state.filteredJobs.filter(function (j) { return j.id !== jobId; });
    var sidx = state.selectedIds.indexOf(jobId);
    if (sidx !== -1) state.selectedIds.splice(sidx, 1);
    var totalEl = document.getElementById('aj-total-count');
    if (totalEl) totalEl.textContent = state.filteredJobs.length;
    renderTable();
  }

  function submitReject() {
    if (!state.pendingJobId) return;
    var reason = document.getElementById('aj-reject-reason').value.trim();
    var confirmBtn = document.getElementById('aj-modal-confirm');
    confirmBtn.disabled = true;
    AngaziaAPI.post('/admin/moderation/' + state.pendingJobId + '/reject', { reason: reason })
      .then(function () {
        showToast('Job listing rejected', 'success');
        removeJobRow(state.pendingJobId);
        closeModal();
      })
      .catch(function (err) {
        showToast(err.message || 'Failed to reject job', 'error');
      })
      .then(function () {
        confirmBtn.disabled = false;
        state.pendingJobId = null;
      });
  }

  function closeModal() {
    document.getElementById('aj-reject-modal').style.display = 'none';
    state.pendingJobId = null;
  }

  function handleBulk(action) {
    var ids = state.selectedIds;
    if (!ids.length) return;
    state.bulkAction = action;
    var count = ids.length;
    var todo = 0;
    var done = 0;
    var errors = 0;
    ids.forEach(function (id) {
      todo++;
      var endpoint = action === 'approve'
        ? '/admin/moderation/' + id + '/approve'
        : '/admin/moderation/' + id + '/reject';
      AngaziaAPI.post(endpoint)
        .then(function () {
          done++;
          removeJobRow(id);
          if (done + errors === todo) {
            showToast('Bulk ' + action + ' completed for ' + done + ' job(s)', done > 0 ? 'success' : 'error');
            state.bulkAction = null;
          }
        })
        .catch(function () {
          errors++;
          if (done + errors === todo) {
            showToast('Completed with ' + errors + ' error(s)', errors === todo ? 'error' : 'warning');
            state.bulkAction = null;
          }
        });
    });
    state.selectedIds = [];
    updateBulkBar();
  }

  function applyFilters() {
    var query = document.getElementById('aj-search').value.trim().toLowerCase();
    var status = document.getElementById('aj-status-filter').value.toLowerCase();
    state.page = 1;
    state.filteredJobs = state.allJobs.filter(function (j) {
      var matchSearch = !query || j.title.indexOf(query) !== -1 || j.company.indexOf(query) !== -1;
      var matchStatus = !status || j.status === status;
      return matchSearch && matchStatus;
    });
    renderTable();
  }

  document.addEventListener('DOMContentLoaded', function () {
    state.allJobs = collectJobsFromDOM();
    state.filteredJobs = state.allJobs.slice();

    if (state.allJobs.length > 0) {
      renderTable();
    }

    var reloadBtn = document.querySelector('[data-action="reload"]');
    if (reloadBtn) {
      reloadBtn.addEventListener('click', function (e) {
        e.preventDefault();
        location.reload();
      });
    }

    document.addEventListener('click', function (e) {
      var btn = e.target.closest('[data-action]');
      if (!btn) return;
      var action = btn.getAttribute('data-action');
      var jobId = btn.getAttribute('data-job-id');
      if (action === 'approve' || action === 'reject') {
        if (jobId) handleAction(action, jobId);
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

    document.getElementById('aj-status-filter').addEventListener('change', applyFilters);

    document.getElementById('aj-select-all').addEventListener('change', function (e) {
      var checked = e.target.checked;
      state.displayedJobs.forEach(function (j) {
        var idx = state.selectedIds.indexOf(j.id);
        if (checked && idx === -1) state.selectedIds.push(j.id);
        if (!checked && idx !== -1) state.selectedIds.splice(idx, 1);
      });
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
