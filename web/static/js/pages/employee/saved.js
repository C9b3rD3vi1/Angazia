(function () {
  'use strict';

  var elements = {};

  function init() {
    elements = {
      loading: document.getElementById('saved-loading'),
      error: document.getElementById('saved-error'),
      errorMsg: document.getElementById('saved-error-msg'),
      content: document.getElementById('saved-content'),
      grid: document.getElementById('saved-grid'),
      subtitle: document.getElementById('saved-subtitle'),
    };
    loadSavedJobs();
  }

  async function loadSavedJobs() {
    showLoading(true);
    showError(false);
    try {
      var response = await AngaziaAPI.jobs.saved();
      var jobs = [];
      if (response && response.data) jobs = response.data;
      else if (Array.isArray(response)) jobs = response;
      else if (response && response.jobs) jobs = response.jobs;
      renderJobs(jobs);
      showLoading(false);
      showContent();
    } catch (err) {
      showLoading(false);
      showError(true, err.message || 'Failed to load saved jobs');
    }
  }

  function renderJobs(jobs) {
    if (!elements.grid) return;
    if (!jobs || jobs.length === 0) {
      elements.grid.innerHTML =
        '<div class="empty-state"><div class="emp-empty-icon">\uD83D\uDD16</div><p class="emp-empty-text">No saved jobs yet. Save jobs you\'re interested in to review later.</p><a href="/employee/jobs" class="emp-btn emp-btn-primary">Browse Jobs \u2192</a></div>';
      if (elements.subtitle) elements.subtitle.textContent = '';
      return;
    }
    if (elements.subtitle) elements.subtitle.textContent = jobs.length + ' position' + (jobs.length !== 1 ? 's' : '') + ' you\'ve bookmarked';
    elements.grid.innerHTML = jobs.map(function (job) { return createCard(job); }).join('');
    document.querySelectorAll('.emp-saved-rm').forEach(function (btn) {
      btn.addEventListener('click', function (e) {
        e.preventDefault();
        e.stopPropagation();
        var jobId = btn.getAttribute('data-job-id');
        if (jobId) unsaveJob(jobId);
      });
    });
  }

  function createCard(job) {
    var companyInitials = getInitials(job.employer && job.employer.company_name ? job.employer.company_name : (job.company_name || 'Unknown'));
    var location = job.location || 'Remote';
    var type = job.employment_type || 'Full-time';
    var match = job.match_score || 0;
    var salary = formatSalary(job.salary_min, job.salary_max);
    var logo = job.employer && job.employer.company_logo ? job.employer.company_logo : (job.company_logo || null);
    var companyName = job.employer && job.employer.company_name ? job.employer.company_name : (job.company_name || 'Unknown');
    return (
      '<a href="/employee/jobs/' + job.id + '" class="emp-saved-card">' +
      '<div class="emp-saved-top">' +
      '<div class="emp-saved-brand">' +
      (logo ? '<img src="' + logo + '" alt="' + escapeHtml(companyName) + '" class="emp-saved-logo">' : '<span class="emp-saved-logo-init">' + escapeHtml(companyInitials) + '</span>') +
      '</div>' +
      '<button class="emp-saved-rm" data-job-id="' + job.id + '" title="Remove">&times;</button>' +
      '</div>' +
      '<h3 class="emp-saved-title">' + escapeHtml(job.title) + '</h3>' +
      '<p class="emp-saved-company">' + escapeHtml(companyName) + '</p>' +
      '<div class="emp-saved-tags"><span class="emp-tag">' + escapeHtml(location) + '</span><span class="emp-tag">' + escapeHtml(type) + '</span></div>' +
      '<div class="emp-saved-foot">' +
      '<span class="emp-saved-match">\uD83C\uDFAF ' + match + '% Match</span>' +
      (salary ? '<span class="emp-saved-salary">' + salary + '</span>' : '') +
      '</div>' +
      '<span class="emp-saved-apply">Apply Now \u2192</span>' +
      '</a>'
    );
  }

  async function unsaveJob(jobId) {
    try {
      await AngaziaAPI.jobs.unsave(jobId);
      showToast('Job removed from saved', 'success');
      await loadSavedJobs();
    } catch (err) {
      showToast(err.message || 'Failed to remove job', 'error');
    }
  }

  function formatSalary(min, max) {
    if (!min && !max) return '';
    var fmt = function (n) {
      if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
      if (n >= 1000) return (n / 1000).toFixed(0) + 'K';
      return n;
    };
    if (min && max) return fmt(min) + ' - ' + fmt(max);
    if (min) return fmt(min) + '+';
    return 'Up to ' + fmt(max);
  }

  function getInitials(name) { if (!name) return '?'; return name.split(' ').map(function (n) { return n[0]; }).join('').toUpperCase().slice(0, 2); }

  function escapeHtml(text) { if (!text) return ''; var d = document.createElement('div'); d.appendChild(document.createTextNode(text)); return d.innerHTML; }

  function showLoading(show) { if (elements.loading) elements.loading.style.display = show ? 'flex' : 'none'; if (elements.content && !show) elements.content.style.display = 'block'; if (elements.content && show) elements.content.style.display = 'none'; }

  function showContent() { if (elements.content) elements.content.style.display = 'block'; }

  function showError(show, message) { if (elements.error) { elements.error.style.display = show ? 'flex' : 'none'; if (elements.errorMsg && message) elements.errorMsg.textContent = message; } if (elements.content && show) elements.content.style.display = 'none'; }

  function showToast(message, type) { if (window.AngaziaApp && window.AngaziaApp.showToast) { window.AngaziaApp.showToast(message, type); } else { alert(message); } }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();
})();
