/**
 * Employer Job Detail Page
 * Handles viewing, editing, closing, and deleting job listings
 */

(function() {
  'use strict';

  let jobId = null;
  let jobData = null;
  let elements = {};
  let pendingJobAction = null;

  const jobActionIcons = {
    close: { icon: '\uD83D\uDD12', cls: 'icon-warning', heading: 'Close Job' },
    delete: { icon: '\uD83D\uDDD1', cls: 'icon-danger', heading: 'Delete Job' },
  };

  // Get job ID from URL
  function getJobIdFromUrl() {
    const pathParts = window.location.pathname.split('/');
    return pathParts[pathParts.length - 1];
  }

  // Cache DOM elements
  function cacheElements() {
    elements = {
      loading: document.getElementById('job-loading'),
      content: document.getElementById('job-content'),
      error: document.getElementById('job-error'),
      errorMsg: document.getElementById('job-error-msg'),
      jobTitle: document.getElementById('job-title'),
      jobStatus: document.getElementById('job-status'),
      statusBadge: document.getElementById('job-status-badge'),
      
      // Stats
      statApplications: document.getElementById('stat-applications'),
      statViews: document.getElementById('stat-views'),
      statShortlisted: document.getElementById('stat-shortlisted'),
      statHired: document.getElementById('stat-hired'),
      
      // Details
      detailTitle: document.getElementById('detail-title'),
      detailType: document.getElementById('detail-type'),
      detailLocation: document.getElementById('detail-location'),
      detailWorkType: document.getElementById('detail-work-type'),
      detailSalary: document.getElementById('detail-salary'),
      detailExperience: document.getElementById('detail-experience'),
      detailPosted: document.getElementById('detail-posted'),
      detailDeadline: document.getElementById('detail-deadline'),
      detailDescription: document.getElementById('detail-description'),
      detailResponsibilities: document.getElementById('detail-responsibilities'),
      detailRequirements: document.getElementById('detail-requirements'),
      detailSkills: document.getElementById('detail-skills'),
      detailNiceSkills: document.getElementById('detail-nice-skills'),
      
      // Applications
      recentApplications: document.getElementById('recent-applications'),
      
      // Buttons
      editBtn: document.getElementById('edit-job-btn'),
      closeBtn: document.getElementById('close-job-btn'),
      deleteBtn: document.getElementById('delete-job-btn'),

      confirmModal: document.getElementById('jd-confirm-modal'),
      confirmTitle: document.getElementById('jd-confirm-title'),
      confirmIcon: document.getElementById('jd-confirm-icon'),
      confirmHeading: document.getElementById('jd-confirm-heading'),
      confirmDesc: document.getElementById('jd-confirm-desc'),
      confirmYes: document.getElementById('jd-confirm-yes'),
      confirmYesLabel: document.getElementById('jd-confirm-yes-label'),
      confirmNo: document.getElementById('jd-confirm-no'),
      confirmClose: document.getElementById('jd-confirm-close')
    };
  }

  // Attach event listeners
  function attachEventListeners() {
    if (elements.editBtn) {
      elements.editBtn.addEventListener('click', () => {
        window.location.href = `/employer/job-edit/${jobId}`;
      });
    }
    
    if (elements.closeBtn) {
      elements.closeBtn.addEventListener('click', () => closeJob());
    }
    
    if (elements.deleteBtn) {
      elements.deleteBtn.addEventListener('click', () => deleteJob());
    }

    if (elements.confirmYes) elements.confirmYes.addEventListener('click', executeJobAction);
    if (elements.confirmNo) elements.confirmNo.addEventListener('click', hideJobConfirmModal);
    if (elements.confirmClose) elements.confirmClose.addEventListener('click', hideJobConfirmModal);
    if (elements.confirmModal) {
      elements.confirmModal.addEventListener('click', (e) => {
        if (e.target === elements.confirmModal) hideJobConfirmModal();
      });
    }
  }

  // Load job details
  async function loadJobDetails() {
    showLoading();
    
    try {
      const response = await AngaziaAPI.jobs.get(jobId);
      
      // Handle response format
      let job = response;
      if (response && response.data) {
        job = response.data;
      }
      
      if (!job || !job.id) {
        throw new Error('Job not found');
      }
      
      jobData = job;
      renderJobDetails(job);
      await loadRecentApplications();
      
      hideLoading();
      showContent();
      
    } catch (error) {
      console.error('Failed to load job:', error);
      hideLoading();
      showError(error.message || 'Failed to load job details');
    }
  }

  // Load recent applications for this job
  async function loadRecentApplications() {
    if (!elements.recentApplications) return;
    
    try {
      const response = await AngaziaAPI.applications.jobApplications(jobId, { limit: 5 });
      
      let applications = [];
      if (response && response.data && Array.isArray(response.data)) {
        applications = response.data;
      } else if (Array.isArray(response)) {
        applications = response;
      } else if (response && response.applications && Array.isArray(response.applications)) {
        applications = response.applications;
      }
      
      renderRecentApplications(applications);
      
    } catch (error) {
      console.error('Failed to load applications:', error);
      if (elements.recentApplications) {
        elements.recentApplications.innerHTML = '<p class="emp-muted">Failed to load applications</p>';
      }
    }
  }

  // Render job details
  function renderJobDetails(job) {
    // Page title and status
    if (elements.jobTitle) elements.jobTitle.textContent = job.title || 'Untitled Job';
    if (elements.jobStatus) {
      elements.jobStatus.textContent = job.is_active ? 'Active - Accepting Applications' : 'Closed - No Longer Accepting Applications';
    }
    
    // Status badge
    if (elements.statusBadge) {
      elements.statusBadge.className = `emp-status-badge ${job.is_active ? 'active' : 'closed'}`;
      elements.statusBadge.textContent = job.is_active ? 'Active' : 'Closed';
    }
    
    // Update button visibility based on status
    if (elements.closeBtn && !job.is_active) {
      elements.closeBtn.style.display = 'none';
    }
    
    // Statistics
    if (elements.statApplications) elements.statApplications.textContent = job.applications_count || 0;
    if (elements.statViews) elements.statViews.textContent = job.views_count || 0;
    if (elements.statShortlisted) elements.statShortlisted.textContent = job.shortlisted_count || 0;
    if (elements.statHired) elements.statHired.textContent = job.hired_count || 0;
    
    // Job details
    if (elements.detailTitle) elements.detailTitle.textContent = job.title || '-';
    if (elements.detailType) elements.detailType.textContent = formatEmploymentType(job.employment_type);
    if (elements.detailLocation) elements.detailLocation.textContent = job.location || 'Remote';
    if (elements.detailWorkType) elements.detailWorkType.textContent = formatWorkType(job);
    if (elements.detailSalary) elements.detailSalary.textContent = formatSalary(job.salary_min, job.salary_max, job.salary_currency);
    if (elements.detailExperience) elements.detailExperience.textContent = formatExperienceLevel(job.experience_level);
    if (elements.detailPosted) elements.detailPosted.textContent = formatDate(job.posted_at);
    if (elements.detailDeadline) elements.detailDeadline.textContent = formatDate(job.expires_at) || 'Not set';
    
    // Rich text content
    if (elements.detailDescription) {
      elements.detailDescription.innerHTML = formatText(job.description || 'No description provided.');
    }
    if (elements.detailResponsibilities) {
      elements.detailResponsibilities.innerHTML = formatText(job.responsibilities || 'No responsibilities listed.');
    }
    if (elements.detailRequirements) {
      elements.detailRequirements.innerHTML = formatRequirements(job.requirements);
    }
    
    // Skills
    if (elements.detailSkills) {
      const skills = job.required_skills || [];
      if (skills.length > 0) {
        elements.detailSkills.innerHTML = skills.map(skill => 
          `<span class="emp-skill-badge">${escapeHtml(skill)}</span>`
        ).join('');
      } else {
        elements.detailSkills.innerHTML = '<p class="emp-muted">No specific skills listed.</p>';
      }
    }
    if (elements.detailNiceSkills) {
      const niceSkills = job.nice_to_have_skills || [];
      if (niceSkills.length > 0) {
        elements.detailNiceSkills.innerHTML = niceSkills.map(skill => 
          `<span class="emp-skill-badge">${escapeHtml(skill)}</span>`
        ).join('');
      } else {
        elements.detailNiceSkills.innerHTML = '<p class="emp-muted">No nice-to-have skills listed.</p>';
      }
    }
  }

  // Render recent applications
  function renderRecentApplications(applications) {
    if (!elements.recentApplications) return;
    
    if (!applications || applications.length === 0) {
      elements.recentApplications.innerHTML = `
        <div class="emp-empty-small">
          <p>No applications yet.</p>
          <a href="/jobs/${jobId}" class="emp-link">Share this job</a>
        </div>
      `;
      return;
    }
    
    elements.recentApplications.innerHTML = applications.map(app => {
      const emp = app.employee || {};
      const name = emp.full_name || app.candidate_name || app.name || 'Candidate';
      const email = (emp.user && emp.user.email) || '';
      const avatarUrl = emp.user && emp.user.avatar_url ? emp.user.avatar_url : null;
      return `
      <div class="emp-application-item">
        <div class="emp-application-avatar">
          ${avatarUrl ? `<img src="${escapeHtml(avatarUrl)}" alt="" class="emp-avatar-img">` : `<div class="emp-avatar-placeholder">${name.charAt(0).toUpperCase()}</div>`}
        </div>
        <div class="emp-application-score">${app.match_score || 0}%</div>
        <div class="emp-application-info">
          <div class="emp-application-name">${escapeHtml(name)}</div>
          <div class="emp-application-details">
            <span>📅 Applied ${formatDate(app.applied_at)}</span>
            <span>${email ? `📧 ${escapeHtml(email)}` : ''}</span>
          </div>
        </div>
        <div class="emp-application-actions">
          <a href="/employer/applications/${app.id}" class="emp-link">Review →</a>
        </div>
      </div>`;
    }).join('');
  }

  function setJobConfirmLoading(loading) {
    if (!elements.confirmYes) return;
    elements.confirmYes.disabled = loading;
    elements.confirmYes.classList.toggle('emp-btn-loading', loading);
  }

  // Confirmation modal helpers
  function showJobConfirmModal(title, message, btnClass, btnLabel, action) {
    if (!elements.confirmModal) return;
    var ico = jobActionIcons[action] || { icon: '\u26A0', cls: 'icon-warning', heading: title };
    elements.confirmTitle.textContent = title;
    if (elements.confirmIcon) {
      elements.confirmIcon.textContent = ico.icon;
      elements.confirmIcon.className = 'emp-modal-icon ' + ico.cls;
    }
    if (elements.confirmHeading) {
      var jobTitle = jobData ? jobData.title : '';
      elements.confirmHeading.textContent = jobTitle ? ico.heading + ': ' + jobTitle : ico.heading;
    }
    if (elements.confirmDesc) elements.confirmDesc.textContent = message;
    if (elements.confirmYesLabel) elements.confirmYesLabel.textContent = btnLabel;
    elements.confirmYes.className = 'emp-btn ' + btnClass;
    setJobConfirmLoading(false);
    pendingJobAction = action;
    elements.confirmModal.style.display = 'flex';
  }

  function hideJobConfirmModal() {
    if (elements.confirmModal) elements.confirmModal.style.display = 'none';
    setJobConfirmLoading(false);
    pendingJobAction = null;
  }

  function executeJobAction() {
    if (!pendingJobAction) return;
    setJobConfirmLoading(true);
    if (pendingJobAction === 'close') {
      executeCloseJob().then(function () {
        hideJobConfirmModal();
      }).catch(function () {
        setJobConfirmLoading(false);
      });
    } else if (pendingJobAction === 'delete') {
      executeDeleteJob().then(function () {
        hideJobConfirmModal();
      }).catch(function () {
        setJobConfirmLoading(false);
      });
    } else {
      hideJobConfirmModal();
    }
  }

  // Close job
  function closeJob() {
    var title = jobData ? jobData.title : '';
    showJobConfirmModal(
      'Close Job',
      'Close "' + title + '"? It will no longer accept new applications. You can reopen it later.',
      'emp-btn-warning',
      'Yes, Close Job',
      'close'
    );
  }

  async function executeCloseJob() {
    try {
      await AngaziaAPI.jobs.close(jobId);
      jobData.is_active = false;
      renderJobDetails(jobData);
    } catch (error) {
      console.error('Failed to close job:', error);
      throw error;
    }
  }

  // Delete job
  function deleteJob() {
    var title = jobData ? jobData.title : '';
    showJobConfirmModal(
      'Delete Job',
      'Permanently delete "' + title + '"? This cannot be undone. All associated applications will also be removed.',
      'emp-btn-danger',
      'Yes, Delete Job',
      'delete'
    );
  }

  async function executeDeleteJob() {
    try {
      await AngaziaAPI.jobs.delete(jobId);
      window.location.href = '/employer/jobs';
    } catch (error) {
      console.error('Failed to delete job:', error);
      throw error;
    }
  }

  // UI State Management
  function showLoading() {
    if (elements.loading) elements.loading.style.display = 'block';
    if (elements.content) elements.content.style.display = 'none';
    if (elements.error) elements.error.style.display = 'none';
  }
  
  function hideLoading() {
    if (elements.loading) elements.loading.style.display = 'none';
  }
  
  function showContent() {
    if (elements.content) elements.content.style.display = 'block';
  }
  
  function showError(message) {
    if (elements.error) {
      if (elements.errorMsg) elements.errorMsg.textContent = message;
      elements.error.style.display = 'block';
    }
  }

  // Helper Functions
  function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-KE', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function formatEmploymentType(type) {
    const types = {
      'full-time': 'Full Time',
      'part-time': 'Part Time',
      'contract': 'Contract',
      'internship': 'Internship',
      'freelance': 'Freelance'
    };
    return types[type] || type || 'Full Time';
  }

  function formatWorkType(job) {
    if (job.is_remote && job.is_hybrid) return 'Remote/Hybrid';
    if (job.is_remote) return 'Remote';
    if (job.is_hybrid) return 'Hybrid';
    return 'On-site';
  }

  function formatSalary(min, max, currency) {
    if (!min && !max) return 'Not specified';
    const curr = currency || 'KES';
    const formatter = new Intl.NumberFormat('en-KE', { style: 'currency', currency: curr });
    if (min && max) return `${formatter.format(min)} - ${formatter.format(max)}`;
    if (min) return `${formatter.format(min)}+`;
    if (max) return `Up to ${formatter.format(max)}`;
    return 'Not specified';
  }

  function formatExperienceLevel(level) {
    const levels = {
      'entry': 'Entry Level (0-2 years)',
      'junior': 'Junior (1-3 years)',
      'mid': 'Mid Level (3-5 years)',
      'senior': 'Senior (5-8 years)',
      'lead': 'Lead (8+ years)'
    };
    return levels[level] || level || 'Not specified';
  }

  function formatText(text) {
    if (!text) return '';
    return escapeHtml(text).replace(/\n/g, '<br>');
  }

  function formatRequirements(text) {
    if (!text) return '<p class="emp-muted">No specific requirements listed.</p>';
    var lines = text.split('\n').filter(function(l) { return l.trim(); });
    if (lines.length === 0) return '<p class="emp-muted">No specific requirements listed.</p>';
    var html = '<ul class="emp-req-list">';
    for (var i = 0; i < lines.length; i++) {
      html += '<li class="emp-req-item">' + escapeHtml(lines[i].trim()) + '</li>';
    }
    html += '</ul>';
    return html;
  }

  function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(text));
    return div.innerHTML;
  }

  function showToast(message, type) {
    if (window.AngaziaApp && window.AngaziaApp.showToast) {
      window.AngaziaApp.showToast(message, type);
      return;
    }
    
    // Fallback toast
    const toast = document.createElement('div');
    const bgColor = type === 'success' ? '#10b981' : type === 'error' ? '#ef4444' : '#3b82f6';
    toast.style.cssText = `
      position: fixed;
      bottom: 20px;
      right: 20px;
      background: ${bgColor};
      color: white;
      padding: 12px 20px;
      border-radius: 8px;
      font-size: 13px;
      z-index: 9999;
      animation: slideIn 0.3s ease;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);
    
    setTimeout(() => {
      toast.style.opacity = '0';
      setTimeout(() => toast.remove(), 300);
    }, 3000);
  }

  // Initialize
  function init() {
    jobId = getJobIdFromUrl();
    if (!jobId || jobId === 'job-detail') {
      window.location.href = '/employer/jobs';
      return;
    }
    
    cacheElements();
    attachEventListeners();
    loadJobDetails();
  }

  // Add toast styles
  function addToastStyles() {
    if (!document.querySelector('#toast-style')) {
      const style = document.createElement('style');
      style.id = 'toast-style';
      style.textContent = `
        @keyframes slideIn {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
        .emp-muted { color: var(--muted); text-align: center; padding: 20px; }
        .emp-empty-small { text-align: center; padding: 30px; color: var(--muted); }
        .emp-req-list { list-style: none; padding: 0; margin: 0; }
        .emp-req-item { padding: 6px 0 6px 20px; font-size: 14px; line-height: 1.6; color: var(--text); position: relative; }
        .emp-req-item::before { content: "\\2713"; position: absolute; left: 0; color: var(--accent); font-weight: 700; }
        .emp-application-avatar { width: 36px; height: 36px; border-radius: 50%; overflow: hidden; flex-shrink: 0; margin-right: 12px; }
        .emp-avatar-img { width: 100%; height: 100%; object-fit: cover; border-radius: 50%; }
        .emp-avatar-placeholder { width: 100%; height: 100%; background: var(--purple); color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 14px; border-radius: 50%; }
      `;
      document.head.appendChild(style);
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
      addToastStyles();
      init();
    });
  } else {
    addToastStyles();
    init();
  }
})();