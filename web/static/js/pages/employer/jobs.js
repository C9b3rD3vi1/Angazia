/**
 * Employer Jobs Management
 * Handles loading, displaying, filtering, and managing job listings
 */

(function() {
  'use strict';

  // DOM Elements
  let elements = {};

  // Initialize the page
  function init() {
    cacheElements();
    attachEventListeners();
    loadJobs();
  }

  // Cache DOM elements
  function cacheElements() {
    elements = {
      loading: document.getElementById('jobs-loading'),
      content: document.getElementById('jobs-content'),
      empty: document.getElementById('jobs-empty'),
      error: document.getElementById('jobs-error'),
      errorMsg: document.getElementById('jobs-error-msg'),
      tbody: document.getElementById('jobs-tbody'),
      filterSelect: document.getElementById('job-status-filter')
    };
  }

  // Attach event listeners
  function attachEventListeners() {
    if (elements.filterSelect) {
      elements.filterSelect.addEventListener('change', function() {
        filterJobs(this.value);
      });
    }
  }

  // Load jobs from API
  async function loadJobs() {
    showLoading();
    
    try {
      const response = await AngaziaAPI.jobs.myJobs();
      
      // Handle different response formats
      let jobs = [];
      
      // Check if response is an array
      if (Array.isArray(response)) {
        jobs = response;
      }
      // Check if response has data property that is an array
      else if (response && response.data && Array.isArray(response.data)) {
        jobs = response.data;
      }
      // Check if response has jobs property that is an array
      else if (response && response.jobs && Array.isArray(response.jobs)) {
        jobs = response.jobs;
      }
      // Check if response has items property (pagination wrapper)
      else if (response && response.items && Array.isArray(response.items)) {
        jobs = response.items;
      }
      else {
        console.warn('Unexpected response format:', response);
        jobs = [];
      }
      
      hideLoading();
      
      if (!jobs || jobs.length === 0) {
        showEmpty();
        return;
      }
      
      renderJobsTable(jobs);
      showContent();
      
    } catch (error) {
      console.error('Failed to load jobs:', error);
      hideLoading();
      showError(error.message || 'Failed to load your job listings');
    }
  }

  // Render jobs table
  function renderJobsTable(jobs) {
    if (!elements.tbody) return;
    
    if (!Array.isArray(jobs) || jobs.length === 0) {
      elements.tbody.innerHTML = '';
      showEmpty();
      return;
    }
    
    elements.tbody.innerHTML = jobs.map(job => createJobRow(job)).join('');
    
    // Show/hide filter based on jobs count
    if (elements.filterSelect && jobs.length > 0) {
      elements.filterSelect.style.display = 'inline-block';
    }
  }

  // Create a single job row HTML
  function createJobRow(job) {
    const status = job.is_active ? 'active' : 'closed';
    const statusLabel = job.is_active ? 'Active' : 'Closed';
    const postedDate = formatDate(job.posted_at || job.created_at);
    const escapedTitle = escapeHtml(job.title || 'Untitled Job');
    const jobId = job.id || job.ID;
    const employmentType = job.employment_type || job.employmentType || 'Full-time';
    const applicationsCount = job.applications_count || job.applicationsCount || 0;
    const viewsCount = job.views_count || job.viewsCount || 0;
    
    return `
      <tr data-status="${status}" data-job-id="${jobId}">
        <td class="emp-job-cell">
          <a href="/employer/jobs/${jobId}" class="emp-job-title">${escapedTitle}</a>
          <span class="emp-job-type">${escapeHtml(employmentType)}</span>
        </td>
        <td class="emp-status-cell">
          <span class="emp-status ${status}">${statusLabel}</span>
        </td>
        <td class="emp-stat-cell">${applicationsCount}</td>
        <td class="emp-stat-cell">${viewsCount}</td>
        <td class="emp-date-cell">${postedDate}</td>
        <td class="emp-actions-cell">
          <a href="/employer/job-applications/${jobId}" class="emp-action-link">View Apps</a>
          ${job.is_active ? 
            `<button class="emp-action-btn emp-action-close" data-job-id="${jobId}" data-job-title="${escapedTitle.replace(/'/g, "\\'")}">Close</button>` : 
            ''
          }
          <button class="emp-action-btn emp-action-delete" data-job-id="${jobId}" data-job-title="${escapedTitle.replace(/'/g, "\\'")}">Delete</button>
        </td>
      </tr>
    `;
  }

  // Filter jobs by status
  function filterJobs(status) {
    const rows = document.querySelectorAll('#jobs-tbody tr');
    
    rows.forEach(row => {
      const rowStatus = row.getAttribute('data-status');
      if (status === 'all' || rowStatus === status) {
        row.style.display = '';
      } else {
        row.style.display = 'none';
      }
    });
  }

  // Close a job
  async function closeJob(jobId, jobTitle) {
    if (!confirm(`Are you sure you want to close "${jobTitle}"? You can re-open it later.`)) {
      return;
    }
    
    showToast('Closing job...', 'info');
    
    try {
      await AngaziaAPI.jobs.close(jobId);
      loadJobs(); // Reload the list
    } catch (error) {
      console.error('Failed to close job:', error);
    }
  }

  // Delete a job
  async function deleteJob(jobId, jobTitle) {
    if (!confirm(`⚠️ Are you sure you want to permanently delete "${jobTitle}"?\n\nThis action cannot be undone.`)) {
      return;
    }
    
    showToast('Deleting job...', 'info');
    
    try {
      await AngaziaAPI.jobs.delete(jobId);
      loadJobs(); // Reload the list
    } catch (error) {
      console.error('Failed to delete job:', error);
    }
  }

  // Handle action button clicks (event delegation)
  function handleActionClick(event) {
    const target = event.target;
    
    if (target.classList.contains('emp-action-close')) {
      const jobId = target.getAttribute('data-job-id');
      const jobTitle = target.getAttribute('data-job-title');
      if (jobId) closeJob(jobId, jobTitle);
    }
    
    if (target.classList.contains('emp-action-delete')) {
      const jobId = target.getAttribute('data-job-id');
      const jobTitle = target.getAttribute('data-job-title');
      if (jobId) deleteJob(jobId, jobTitle);
    }
  }

  // UI State Management
  function showLoading() {
    if (elements.loading) elements.loading.style.display = 'block';
    if (elements.content) elements.content.style.display = 'none';
    if (elements.empty) elements.empty.style.display = 'none';
    if (elements.error) elements.error.style.display = 'none';
  }

  function hideLoading() {
    if (elements.loading) elements.loading.style.display = 'none';
  }

  function showContent() {
    if (elements.content) elements.content.style.display = 'block';
  }

  function showEmpty() {
    if (elements.empty) elements.empty.style.display = 'block';
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
      font-family: monospace;
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

  // Add animation style if not exists
  function addToastStyles() {
    if (!document.querySelector('#toast-style')) {
      const style = document.createElement('style');
      style.id = 'toast-style';
      style.textContent = `
        @keyframes slideIn {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
      `;
      document.head.appendChild(style);
    }
  }

  // Set up event delegation for action buttons
  function setupEventDelegation() {
    if (elements.tbody) {
      elements.tbody.addEventListener('click', handleActionClick);
    }
  }

  // Start the page
  function start() {
    addToastStyles();
    setupEventDelegation();
    init();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', start);
  } else {
    start();
  }
})();