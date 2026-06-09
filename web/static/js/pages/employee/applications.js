/**
 * Employee Applications Page
 * Handles loading, filtering, and managing job applications
 */

(function() {
  'use strict';

  // State
  let allApplications = [];
  let currentStatus = 'all';
  let currentPage = 1;
  let itemsPerPage = 10;
  let totalPages = 1;

  // DOM Elements
  let elements = {};

  // Initialize page
  async function init() {
    cacheElements();
    attachEventListeners();
    await loadApplications();
  }

  // Cache DOM elements
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
      tabsContainer: document.getElementById('application-tabs')
    };
  }

  // Attach event listeners
  function attachEventListeners() {
    // Tab clicks
    elements.tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        const status = tab.getAttribute('data-status');
        if (status && status !== currentStatus) {
          switchTab(status);
        }
      });
    });

    // Pagination
    if (elements.prevBtn) {
      elements.prevBtn.addEventListener('click', () => {
        if (currentPage > 1) {
          currentPage--;
          renderApplications();
          updatePagination();
        }
      });
    }

    if (elements.nextBtn) {
      elements.nextBtn.addEventListener('click', () => {
        if (currentPage < totalPages) {
          currentPage++;
          renderApplications();
          updatePagination();
        }
      });
    }
  }

  // Load applications from API
  async function loadApplications() {
    showLoading(true);
    showError(false);

    try {
      const response = await AngaziaAPI.applications.myApplications({ limit: 100 });
      
      // Handle response format
      let applications = [];
      if (response && response.data) {
        applications = response.data;
      } else if (Array.isArray(response)) {
        applications = response;
      } else if (response && response.applications) {
        applications = response.applications;
      }
      
      allApplications = applications.map(formatApplication);
      
      // Update tab counts
      updateTabCounts();
      
      // Reset pagination
      currentPage = 1;
      currentStatus = 'all';
      
      // Update active tab UI
      updateActiveTab('all');
      
      // Render applications
      renderApplications();
      updatePagination();
      
      showLoading(false);
      showContent();
      
    } catch (error) {
      console.error('Failed to load applications:', error);
      showLoading(false);
      showError(true, error.message || 'Failed to load your applications');
    }
  }

  // Format application data
  function formatApplication(app) {
    return {
      id: app.id,
      job_id: app.job_id,
      title: app.job?.title || app.job_title || 'Unknown Position',
      company: app.job?.employer?.company_name || app.company_name || 'Unknown Company',
      company_logo: app.job?.employer?.company_logo || null,
      location: app.job?.location || 'Remote',
      type: app.job?.employment_type || 'Full-time',
      status: app.status || 'pending',
      match_score: app.match_score || 0,
      applied_at: app.applied_at,
      viewed_at: app.viewed_at,
      interview_date: app.interview_date,
      interview_type: app.interview_type,
      employer_notes: app.employer_notes,
      next_step: getNextStep(app)
    };
  }

  // Determine next step based on status
  function getNextStep(app) {
    switch (app.status) {
      case 'pending':
        return 'Awaiting employer review';
      case 'viewed':
        return 'Employer has viewed your application';
      case 'shortlisted':
        return 'You have been shortlisted! Expect interview invite';
      case 'interview':
        return `Interview scheduled${app.interview_date ? ` for ${formatDate(app.interview_date)}` : ''}`;
      case 'hired':
        return 'Congratulations! You got the job!';
      case 'rejected':
        return 'Application not selected. Keep applying!';
      case 'withdrawn':
        return 'Application withdrawn';
      default:
        return 'Application under review';
    }
  }

  // Update tab counts
  function updateTabCounts() {
    const counts = {
      all: allApplications.length,
      pending: allApplications.filter(a => a.status === 'pending').length,
      viewed: allApplications.filter(a => a.status === 'viewed').length,
      shortlisted: allApplications.filter(a => a.status === 'shortlisted').length,
      interview: allApplications.filter(a => a.status === 'interview').length,
      hired: allApplications.filter(a => a.status === 'hired').length,
      rejected: allApplications.filter(a => a.status === 'rejected').length
    };
    
    updateElementText('tab-all-count', counts.all);
    updateElementText('tab-pending-count', counts.pending);
    updateElementText('tab-viewed-count', counts.viewed);
    updateElementText('tab-shortlisted-count', counts.shortlisted);
    updateElementText('tab-interview-count', counts.interview);
    updateElementText('tab-hired-count', counts.hired);
    updateElementText('tab-rejected-count', counts.rejected);
  }

  // Switch tab
  function switchTab(status) {
    currentStatus = status;
    currentPage = 1;
    updateActiveTab(status);
    renderApplications();
    updatePagination();
  }

  // Update active tab UI
  function updateActiveTab(activeStatus) {
    elements.tabs.forEach(tab => {
      const tabStatus = tab.getAttribute('data-status');
      if (tabStatus === activeStatus) {
        tab.classList.add('active');
      } else {
        tab.classList.remove('active');
      }
    });
  }

  // Get filtered applications based on current status
  function getFilteredApplications() {
    if (currentStatus === 'all') {
      return allApplications;
    }
    return allApplications.filter(app => app.status === currentStatus);
  }

  // Render applications list
  function renderApplications() {
    if (!elements.list) return;
    
    const filtered = getFilteredApplications();
    totalPages = Math.ceil(filtered.length / itemsPerPage);
    const start = (currentPage - 1) * itemsPerPage;
    const paginatedApps = filtered.slice(start, start + itemsPerPage);
    
    if (paginatedApps.length === 0) {
      elements.list.innerHTML = `
        <div class="emp-empty">
          <div class="emp-empty-icon">📋</div>
          <p class="emp-empty-text">No applications found in this category.</p>
          <a href="/employee/jobs" class="emp-btn emp-btn-primary">Browse Jobs →</a>
        </div>
      `;
      return;
    }
    
    elements.list.innerHTML = paginatedApps.map(app => createApplicationCard(app)).join('');
    
    // Attach withdraw handlers
    document.querySelectorAll('.emp-withdraw-btn').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        e.stopPropagation();
        const appId = btn.getAttribute('data-app-id');
        if (appId) await withdrawApplication(appId);
      });
    });
    
    // Attach card click handlers
    document.querySelectorAll('.emp-app-card').forEach(card => {
      card.addEventListener('click', (e) => {
        if (e.target.closest('.emp-withdraw-btn')) return;
        const appId = card.getAttribute('data-app-id');
        if (appId) window.location.href = `/applications/${appId}`;
      });
    });
  }

  // Create application card HTML
  function createApplicationCard(app) {
    const companyInitials = getInitials(app.company);
    const appliedDate = formatRelativeDate(app.applied_at);
    const statusClass = getStatusClass(app.status);
    const statusLabel = getStatusLabel(app.status);
    const showWithdraw = app.status === 'pending' || app.status === 'viewed';
    
    // Get match score color
    const matchScoreColor = app.match_score >= 80 ? '#10b981' : app.match_score >= 60 ? '#f59e0b' : '#ef4444';
    const matchMessage = app.match_score >= 80 ? 'Excellent match!' : app.match_score >= 60 ? 'Good match' : 'Consider improving your profile';
    
    return `
      <div class="emp-app-card" data-app-id="${app.id}" data-job-id="${app.job_id}">
        <div class="emp-app-head">
          <div class="emp-app-brand">
            ${app.company_logo ? 
              `<img src="${app.company_logo}" alt="${escapeHtml(app.company)}" class="emp-app-logo">` : 
              `<span class="emp-app-logo-init">${escapeHtml(companyInitials)}</span>`
            }
            <div>
              <h3 class="emp-app-title">${escapeHtml(app.title)}</h3>
              <p class="emp-app-company">${escapeHtml(app.company)}</p>
            </div>
          </div>
          <span class="emp-badge emp-badge-${statusClass}">${statusLabel}</span>
        </div>
        
        <div class="emp-app-meta">
          <span>📅 Applied ${appliedDate}</span>
          <span>📍 ${escapeHtml(app.location)}</span>
          <span>💼 ${escapeHtml(app.type)}</span>
        </div>
        
        ${app.match_score > 0 ? `
          <div class="emp-app-match">
            <span>🎯 Match Score:</span>
            <span class="emp-app-match-score" style="color: ${matchScoreColor}">${app.match_score}%</span>
            <span class="emp-app-match-message">${matchMessage}</span>
          </div>
        ` : ''}
        
        ${app.next_step ? `
          <div class="emp-app-next">
            <span class="emp-app-next-icon">📌</span>
            <span>Next step: <strong>${escapeHtml(app.next_step)}</strong></span>
          </div>
        ` : ''}
        
        ${app.employer_notes ? `
          <div class="emp-app-notes">
            <span class="emp-app-next-icon">📝</span>
            <span><strong>Employer note:</strong> ${escapeHtml(app.employer_notes)}</span>
          </div>
        ` : ''}
        
        ${app.interview_date ? `
          <div class="emp-app-notes">
            <span class="emp-app-next-icon">📅</span>
            <span><strong>Interview:</strong> ${formatDate(app.interview_date)} (${app.interview_type || 'scheduled'})</span>
          </div>
        ` : ''}
        
        ${showWithdraw ? `
          <div class="emp-app-withdraw">
            <button class="emp-withdraw-btn" data-app-id="${app.id}">Withdraw Application</button>
          </div>
        ` : ''}
      </div>
    `;
  }

  // Withdraw application
  async function withdrawApplication(appId) {
    const confirmed = await confirmDialog('Are you sure you want to withdraw this application? This action cannot be undone.');
    if (!confirmed) return;
    
    try {
      await AngaziaAPI.applications.withdraw(appId);
      showToast('Application withdrawn successfully', 'success');
      await loadApplications(); // Reload
    } catch (error) {
      console.error('Withdraw failed:', error);
      showToast(error.message || 'Failed to withdraw application', 'error');
    }
  }

  // Update pagination UI
  function updatePagination() {
    const filtered = getFilteredApplications();
    totalPages = Math.ceil(filtered.length / itemsPerPage);
    
    if (elements.pageInfo) {
      elements.pageInfo.textContent = `Page ${currentPage} of ${totalPages || 1}`;
    }
    
    if (elements.prevBtn) {
      elements.prevBtn.disabled = currentPage <= 1;
    }
    
    if (elements.nextBtn) {
      elements.nextBtn.disabled = currentPage >= totalPages;
    }
    
    if (elements.pagination) {
      elements.pagination.style.display = totalPages > 1 ? 'flex' : 'none';
    }
  }

  // UI State Management
  function showLoading(show) {
    if (elements.loading) elements.loading.style.display = show ? 'block' : 'none';
    if (elements.content && !show) elements.content.style.display = 'block';
    if (elements.content && show) elements.content.style.display = 'none';
  }

  function showContent() {
    if (elements.content) elements.content.style.display = 'block';
  }

  function showError(show, message = '') {
    if (elements.error) {
      elements.error.style.display = show ? 'block' : 'none';
      if (elements.errorMsg && message) elements.errorMsg.textContent = message;
    }
    if (elements.content && show) elements.content.style.display = 'none';
  }

  // Helper Functions
  function getStatusClass(status) {
    const statusMap = {
      'pending': 'pending',
      'viewed': 'viewed',
      'shortlisted': 'shortlisted',
      'interview': 'interview',
      'hired': 'hired',
      'rejected': 'rejected'
    };
    return statusMap[status] || 'pending';
  }

  function getStatusLabel(status) {
    const labelMap = {
      'pending': 'Pending Review',
      'viewed': 'Application Viewed',
      'shortlisted': 'Shortlisted',
      'interview': 'Interview',
      'hired': 'Hired',
      'rejected': 'Not Selected'
    };
    return labelMap[status] || status;
  }

  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
  }

  function formatDate(dateStr) {
    if (!dateStr) return 'N/A';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-KE', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function formatRelativeDate(dateStr) {
    if (!dateStr) return 'N/A';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);
    
    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins} min ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    return formatDate(dateStr);
  }

  function updateElementText(id, value) {
    const el = document.getElementById(id);
    if (el) el.textContent = value;
  }

  function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(text));
    return div.innerHTML;
  }

  function confirmDialog(message) {
    if (window.AngaziaApp && window.AngaziaApp.confirmDialog) {
      return window.AngaziaApp.confirmDialog(message);
    }
    return Promise.resolve(confirm(message));
  }

  function showToast(message, type) {
    if (window.AngaziaApp && window.AngaziaApp.showToast) {
      window.AngaziaApp.showToast(message, type);
    } else {
      alert(message);
    }
  }

  // Start
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();