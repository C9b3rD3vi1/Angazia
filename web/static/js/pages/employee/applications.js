(function () {
  'use strict';

  let allApplications = [];
  let currentStatus = 'all';
  let currentPage = 1;
  let itemsPerPage = 10;
  let totalPages = 1;
  let elements = {};

  async function init() {
    cacheElements();
    attachEventListeners();
    await Promise.all([loadApplications(), loadStats()]);
  }

  function cacheElements() {
    elements = {
      loading: document.getElementById('applications-loading'),
      error: document.getElementById('applications-error'),
      errorMsg: document.getElementById('applications-error-msg'),
      content: document.getElementById('applications-content'),
      list: document.getElementById('applications-list'),
      pagination: document.getElementById('applications-pagination'),
      prevBtn: document.getElementById('prev-page'),
      nextBtn: document.getElementById('next-page'),
      pageInfo: document.getElementById('page-info'),
      tabs: document.querySelectorAll('.emp-tab'),
      tabsContainer: document.getElementById('application-tabs'),
      totalApps: document.getElementById('total-applications'),
      activeApps: document.getElementById('active-applications'),
      responseRate: document.getElementById('response-rate'),
    };
  }

  function attachEventListeners() {
    elements.tabs.forEach(function (tab) {
      tab.addEventListener('click', function () {
        var status = tab.getAttribute('data-status');
        if (status && status !== currentStatus) switchTab(status);
      });
    });
    if (elements.prevBtn) {
      elements.prevBtn.addEventListener('click', function () {
        if (currentPage > 1) { currentPage--; renderApplications(); updatePagination(); }
      });
    }
    if (elements.nextBtn) {
      elements.nextBtn.addEventListener('click', function () {
        if (currentPage < totalPages) { currentPage++; renderApplications(); updatePagination(); }
      });
    }
  }

  async function loadStats() {
    try {
      var resp = await AngaziaAPI.applications.stats();
      var stats = resp && resp.data ? resp.data : resp;
      if (elements.totalApps) elements.totalApps.textContent = stats.total_applications || stats.totalApplications || 0;
      if (elements.activeApps) {
        var active = (stats.pending_count || stats.pendingCount || 0) + (stats.shortlisted_count || stats.shortlistedCount || 0) + (stats.interview_count || stats.interviewCount || 0);
        elements.activeApps.textContent = active;
      }
      if (elements.responseRate) {
        var total = stats.total_applications || stats.totalApplications || 0;
        var responded = (stats.viewed_count || stats.viewedCount || 0) + (stats.shortlisted_count || stats.shortlistedCount || 0) + (stats.interview_count || stats.interviewCount || 0) + (stats.rejected_count || stats.rejectedCount || 0) + (stats.hired_count || stats.hiredCount || 0);
        elements.responseRate.textContent = total > 0 ? Math.round((responded / total) * 100) + '%' : '0%';
      }
    } catch (_) { /* stats are non-critical */ }
  }

  async function loadApplications() {
    showLoading(true);
    showError(false);
    try {
      var response = await AngaziaAPI.applications.myApplications({ limit: 100 });
      var applications = [];
      if (response && response.data) applications = response.data;
      else if (Array.isArray(response)) applications = response;
      else if (response && response.applications) applications = response.applications;
      allApplications = applications.map(formatApplication);
      updateTabCounts();
      currentPage = 1;
      currentStatus = 'all';
      updateActiveTab('all');
      renderApplications();
      updatePagination();
      showLoading(false);
      showContent();
    } catch (error) {
      showLoading(false);
      showError(true, error.message || 'Failed to load your applications');
    }
  }

  function formatApplication(app) {
    return {
      id: app.id,
      job_id: app.job_id,
      title: app.job && app.job.title ? app.job.title : (app.job_title || 'Unknown Position'),
      company: app.job && app.job.employer ? app.job.employer.company_name : (app.company_name || 'Unknown Company'),
      company_logo: app.job && app.job.employer ? app.job.employer.company_logo : null,
      location: app.job && app.job.location ? app.job.location : 'Remote',
      type: app.job && app.job.employment_type ? app.job.employment_type : 'Full-time',
      status: app.status || 'pending',
      match_score: app.match_score || 0,
      applied_at: app.applied_at || app.created_at,
      viewed_at: app.viewed_at,
      interview_date: app.interview_date,
      interview_type: app.interview_type,
      employer_notes: app.employer_notes,
      next_step: getNextStep(app),
    };
  }

  function getNextStep(app) {
    switch (app.status) {
      case 'pending': return 'Awaiting employer review';
      case 'viewed': return 'Employer has viewed your application';
      case 'shortlisted': return 'You have been shortlisted! Expect interview invite';
      case 'interview': return 'Interview scheduled' + (app.interview_date ? ' for ' + formatDate(app.interview_date) : '');
      case 'hired': return 'Congratulations! You got the job!';
      case 'rejected': return 'Application not selected. Keep applying!';
      case 'withdrawn': return 'Application withdrawn';
      default: return 'Application under review';
    }
  }

  function updateTabCounts() {
    var counts = {
      all: allApplications.length,
      pending: allApplications.filter(function (a) { return a.status === 'pending'; }).length,
      viewed: allApplications.filter(function (a) { return a.status === 'viewed'; }).length,
      shortlisted: allApplications.filter(function (a) { return a.status === 'shortlisted'; }).length,
      interview: allApplications.filter(function (a) { return a.status === 'interview'; }).length,
      hired: allApplications.filter(function (a) { return a.status === 'hired'; }).length,
      rejected: allApplications.filter(function (a) { return a.status === 'rejected'; }).length,
    };
    updateElementText('tab-all-count', counts.all);
    updateElementText('tab-pending-count', counts.pending);
    updateElementText('tab-viewed-count', counts.viewed);
    updateElementText('tab-shortlisted-count', counts.shortlisted);
    updateElementText('tab-interview-count', counts.interview);
    updateElementText('tab-hired-count', counts.hired);
    updateElementText('tab-rejected-count', counts.rejected);
  }

  function switchTab(status) {
    currentStatus = status;
    currentPage = 1;
    updateActiveTab(status);
    renderApplications();
    updatePagination();
  }

  function updateActiveTab(activeStatus) {
    elements.tabs.forEach(function (tab) {
      tab.classList.toggle('active', tab.getAttribute('data-status') === activeStatus);
    });
  }

  function getFilteredApplications() {
    return currentStatus === 'all' ? allApplications : allApplications.filter(function (app) { return app.status === currentStatus; });
  }

  function renderApplications() {
    if (!elements.list) return;
    var filtered = getFilteredApplications();
    totalPages = Math.ceil(filtered.length / itemsPerPage);
    var start = (currentPage - 1) * itemsPerPage;
    var paginatedApps = filtered.slice(start, start + itemsPerPage);
    if (paginatedApps.length === 0) {
      elements.list.innerHTML =
        '<div class="emp-empty"><div class="emp-empty-icon">\uD83D\uDCCB</div><p class="emp-empty-text">No applications found in this category.</p><a href="/employee/jobs" class="emp-btn emp-btn-primary">Browse Jobs \u2192</a></div>';
      return;
    }
    elements.list.innerHTML = paginatedApps.map(createApplicationCard).join('');
    document.querySelectorAll('.emp-withdraw-btn').forEach(function (btn) {
      btn.addEventListener('click', function (e) {
        e.stopPropagation();
        var appId = btn.getAttribute('data-app-id');
        if (appId) withdrawApplication(appId);
      });
    });
  }

  function createApplicationCard(app) {
    var companyInitials = getInitials(app.company);
    var appliedDate = formatRelativeDate(app.applied_at);
    var statusClass = getStatusClass(app.status);
    var statusLabel = getStatusLabel(app.status);
    var showWithdraw = app.status === 'pending' || app.status === 'viewed';
    var matchScoreColor = app.match_score >= 80 ? '#10b981' : app.match_score >= 60 ? '#f59e0b' : '#ef4444';
    var matchMessage = app.match_score >= 80 ? 'Excellent match!' : app.match_score >= 60 ? 'Good match' : 'Consider improving your profile';
    return (
      '<div class="emp-app-card" data-app-id="' + app.id + '" data-job-id="' + app.job_id + '">' +
      '<div class="emp-app-head">' +
      '<div class="emp-app-brand">' +
      (app.company_logo
        ? '<img src="' + app.company_logo + '" alt="' + escapeHtml(app.company) + '" class="emp-app-logo">'
        : '<span class="emp-app-logo-init">' + escapeHtml(companyInitials) + '</span>') +
      '<div><h3 class="emp-app-title">' + escapeHtml(app.title) + '</h3><p class="emp-app-company">' + escapeHtml(app.company) + '</p></div>' +
      '</div>' +
      '<span class="emp-badge emp-badge-' + statusClass + '">' + statusLabel + '</span>' +
      '</div>' +
      '<div class="emp-app-meta">' +
      '<span>\uD83D\uDCC5 Applied ' + appliedDate + '</span>' +
      '<span>\uD83D\uDCCD ' + escapeHtml(app.location) + '</span>' +
      '<span>\uD83D\uDCBC ' + escapeHtml(app.type) + '</span>' +
      '</div>' +
      (app.match_score > 0
        ? '<div class="emp-app-match"><span>\uD83C\uDFAF Match Score:</span><span class="emp-app-match-score" style="color:' + matchScoreColor + '">' + app.match_score + '%</span><span class="emp-app-match-message">' + matchMessage + '</span></div>'
        : '') +
      (app.next_step
        ? '<div class="emp-app-next"><span class="emp-app-next-icon">\uD83D\uDCCC</span><span>Next step: <strong>' + escapeHtml(app.next_step) + '</strong></span></div>'
        : '') +
      (app.employer_notes
        ? '<div class="emp-app-notes"><span class="emp-app-next-icon">\uD83D\uDCDD</span><span><strong>Employer note:</strong> ' + escapeHtml(app.employer_notes) + '</span></div>'
        : '') +
      (app.interview_date
        ? '<div class="emp-app-notes"><span class="emp-app-next-icon">\uD83D\uDCC5</span><span><strong>Interview:</strong> ' + formatDate(app.interview_date) + (app.interview_type ? ' (' + app.interview_type + ')' : '') + '</span></div>'
        : '') +
      (showWithdraw
        ? '<div class="emp-app-withdraw"><button class="emp-withdraw-btn" data-app-id="' + app.id + '">Withdraw Application</button></div>'
        : '') +
      '</div>'
    );
  }

  async function withdrawApplication(appId) {
    var confirmed = await confirmDialog('Are you sure you want to withdraw this application? This action cannot be undone.');
    if (!confirmed) return;
    try {
      await AngaziaAPI.applications.withdraw(appId);
      showToast('Application withdrawn successfully', 'success');
      await Promise.all([loadApplications(), loadStats()]);
    } catch (error) {
      showToast(error.message || 'Failed to withdraw application', 'error');
    }
  }

  function updatePagination() {
    var filtered = getFilteredApplications();
    totalPages = Math.ceil(filtered.length / itemsPerPage);
    if (elements.pageInfo) elements.pageInfo.textContent = 'Page ' + currentPage + ' of ' + (totalPages || 1);
    if (elements.prevBtn) elements.prevBtn.disabled = currentPage <= 1;
    if (elements.nextBtn) elements.nextBtn.disabled = currentPage >= totalPages;
    if (elements.pagination) elements.pagination.style.display = totalPages > 1 ? 'flex' : 'none';
  }

  function showLoading(show) {
    if (elements.loading) elements.loading.style.display = show ? 'flex' : 'none';
    if (elements.content && !show) elements.content.style.display = 'block';
    if (elements.content && show) elements.content.style.display = 'none';
  }

  function showContent() { if (elements.content) elements.content.style.display = 'block'; }

  function showError(show, message) {
    if (elements.error) { elements.error.style.display = show ? 'flex' : 'none'; if (elements.errorMsg && message) elements.errorMsg.textContent = message; }
    if (elements.content && show) elements.content.style.display = 'none';
  }

  function getStatusClass(status) { return { pending: 'pending', viewed: 'viewed', shortlisted: 'shortlisted', interview: 'interview', hired: 'hired', rejected: 'rejected' }[status] || 'pending'; }

  function getStatusLabel(status) { return { pending: 'Pending Review', viewed: 'Application Viewed', shortlisted: 'Shortlisted', interview: 'Interview', hired: 'Hired', rejected: 'Not Selected' }[status] || status; }

  function getInitials(name) { if (!name) return '?'; return name.split(' ').map(function (n) { return n[0]; }).join('').toUpperCase().slice(0, 2); }

  function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    return new Date(dateStr).toLocaleDateString('en-KE', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function formatRelativeDate(dateStr) {
    if (!dateStr) return 'N/A';
    var date = new Date(dateStr);
    var now = new Date();
    var diffMs = now - date;
    var diffMins = Math.floor(diffMs / 60000);
    var diffHours = Math.floor(diffMs / 3600000);
    var diffDays = Math.floor(diffMs / 86400000);
    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return diffMins + ' min ago';
    if (diffHours < 24) return diffHours + ' hour' + (diffHours > 1 ? 's' : '') + ' ago';
    if (diffDays < 7) return diffDays + ' day' + (diffDays > 1 ? 's' : '') + ' ago';
    return formatDate(dateStr);
  }

  function updateElementText(id, value) { var el = document.getElementById(id); if (el) el.textContent = value; }

  function escapeHtml(text) { if (!text) return ''; var d = document.createElement('div'); d.appendChild(document.createTextNode(text)); return d.innerHTML; }

  function confirmDialog(message) { return window.AngaziaApp && window.AngaziaApp.confirmDialog ? window.AngaziaApp.confirmDialog(message) : Promise.resolve(confirm(message)); }

  function showToast(message, type) { if (window.AngaziaApp && window.AngaziaApp.showToast) { window.AngaziaApp.showToast(message, type); } else { alert(message); } }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();
})();
