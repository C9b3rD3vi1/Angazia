/**
 * Employer Job Detail Page
 * Handles viewing, editing, closing, and deleting job listings
 */

(function() {
  'use strict';

  let jobId = null;
  let jobData = null;
  let elements = {};

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
      detailRequirements: document.getElementById('detail-requirements'),
      detailSkills: document.getElementById('detail-skills'),
      
      // Applications
      recentApplications: document.getElementById('recent-applications'),
      
      // Buttons
      editBtn: document.getElementById('edit-job-btn'),
      closeBtn: document.getElementById('close-job-btn'),
      deleteBtn: document.getElementById('delete-job-btn')
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
    
    elements.recentApplications.innerHTML = applications.map(app => `
      <div class="emp-application-item">
        <div class="emp-application-score">${app.match_score || 0}%</div>
        <div class="emp-application-info">
          <div class="emp-application-name">${escapeHtml(app.employee_name || 'Candidate')}</div>
          <div class="emp-application-details">
            <span>📅 Applied ${formatDate(app.applied_at)}</span>
            <span>📧 ${escapeHtml(app.employee_email || 'No email')}</span>
          </div>
        </div>
        <div class="emp-application-actions">
          <a href="/employer/application/${app.id}" class="emp-link">Review →</a>
        </div>
      </div>
    `).join('');
  }

  // Close job
  async function closeJob() {
    if (!confirm(`Are you sure you want to close "${jobData.title}"?`)) {
      return;
    }
    
    showToast('Closing job...', 'info');
    
    try {
      await AngaziaAPI.jobs.close(jobId);
      // Update UI
      jobData.is_active = false;
      renderJobDetails(jobData);
    } catch (error) {
      console.error('Failed to close job:', error);
    }
  }

  // Delete job
  async function deleteJob() {
    if (!confirm(`⚠️ Are you sure you want to permanently delete "${jobData.title}"?\n\nThis action cannot be undone.`)) {
      return;
    }
    
    showToast('Deleting job...', 'info');
    
    try {
      await AngaziaAPI.jobs.delete(jobId);
      // Redirect to jobs list
      setTimeout(() => {
        window.location.href = '/employer/jobs';
      }, 1500);
    } catch (error) {
      console.error('Failed to delete job:', error);
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