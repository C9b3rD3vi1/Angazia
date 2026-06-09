/**
 * Employer Applications Management
 * Handles viewing, filtering, and managing job applications
 */

(function() {
  'use strict';

  let currentPage = 1;
  let totalPages = 1;
  let isLoading = false;
  let selectedIds = [];
  let currentFilters = {};
  let pendingInterviewId = null;

  // DOM Elements
  const elements = {
    loading: document.getElementById('applications-loading'),
    error: document.getElementById('applications-error'),
    errorMsg: document.getElementById('applications-error-msg'),
    content: document.getElementById('applications-content'),
    tableBody: document.getElementById('app-table-body'),
    selectAll: document.getElementById('select-all-apps'),
    bulkBar: document.getElementById('bulk-bar'),
    selectedCount: document.getElementById('selected-count'),
    clearSelection: document.getElementById('clear-selection'),
    resultHint: document.getElementById('result-hint'),
    pagination: document.getElementById('pagination'),
    pageInfo: document.getElementById('page-info'),
    prevBtn: document.getElementById('prev-page'),
    nextBtn: document.getElementById('next-page'),
    // Modal elements
    interviewModal: document.getElementById('interview-modal'),
    interviewDatetime: document.getElementById('interview-datetime'),
    interviewType: document.getElementById('interview-type'),
    interviewNotes: document.getElementById('interview-notes'),
    interviewConfirm: document.getElementById('interview-modal-confirm'),
    interviewCancel: document.getElementById('interview-modal-cancel'),
    interviewClose: document.getElementById('interview-modal-close')
  };

  // Add styles
  function addStyles() {
    if (!document.querySelector('#emp-app-styles')) {
      const style = document.createElement('style');
      style.id = 'emp-app-styles';
      style.textContent = `
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
        @keyframes slideIn {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
      `;
      document.head.appendChild(style);
    }
  }

  // Show toast notification
  function showToast(message, type) {
    if (window.AngaziaApp && window.AngaziaApp.showToast) {
      window.AngaziaApp.showToast(message, type);
      return;
    }
    
    const toast = document.createElement('div');
    const bgColor = type === 'success' ? '#10b981' : type === 'error' ? '#ef4444' : type === 'warning' ? '#f59e0b' : '#3b82f6';
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

  // Escape HTML
  function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(text));
    return div.innerHTML;
  }

  // Get initials
  function getInitials(name) {
    if (!name) return '??';
    return name.split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  }

  // Time ago formatter
  function timeAgo(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffSec = Math.floor(diffMs / 1000);
    
    if (diffSec < 60) return 'just now';
    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}m ago`;
    const diffHour = Math.floor(diffMin / 60);
    if (diffHour < 24) return `${diffHour}h ago`;
    const diffDay = Math.floor(diffHour / 24);
    if (diffDay < 7) return `${diffDay}d ago`;
    return date.toLocaleDateString();
  }

  // Load applications
  async function loadApplications() {
    if (isLoading) return;
    isLoading = true;
    
    showLoading();
    
    try {
      const params = {
        page: currentPage,
        limit: 20,
        ...currentFilters
      };
      
      const response = await AngaziaAPI.applications.companyApplications(params);
      
      let applications = [];
      let total = 0;
      
      if (response && response.data) {
        applications = response.data.applications || response.data || [];
        total = response.data.total || applications.length;
        totalPages = response.data.total_pages || Math.ceil(total / 20);
      } else if (response && response.applications) {
        applications = response.applications;
        total = response.total || applications.length;
      } else if (Array.isArray(response)) {
        applications = response;
        total = applications.length;
      }
      
      hideLoading();
      
      if (!applications || applications.length === 0) {
        showEmpty();
        return;
      }
      
      renderTable(applications);
      updatePagination(total);
      showContent();
      
    } catch (error) {
      console.error('Failed to load applications:', error);
      hideLoading();
      showError(error.message || 'Failed to load applications');
    } finally {
      isLoading = false;
    }
  }

  // Render applications table
  function renderTable(applications) {
    if (!elements.tableBody) return;
    
    elements.tableBody.innerHTML = applications.map(app => {
      const candidateName = app.candidate_name || app.employee?.full_name || 'Unknown Candidate';
      const candidateEmail = app.candidate_email || app.employee?.user?.email || '';
      const jobTitle = app.job_title || app.job?.title || 'Unknown Job';
      const status = app.status || 'pending';
      const matchScore = app.match_score || 0;
      const appliedAt = app.applied_at || app.created_at;
      const initials = getInitials(candidateName);
      
      return `
        <tr class="emp-app-row" data-id="${app.id}" data-status="${status}">
          <td><input type="checkbox" class="emp-app-cb" value="${app.id}"></td>
          <td>
            <div class="emp-app-candidate">
              <span class="emp-app-avatar">${initials}</span>
              <div>
                <a href="/employer/applications/${app.id}" class="emp-app-name">${escapeHtml(candidateName)}</a>
                <span class="emp-app-email">${escapeHtml(candidateEmail)}</span>
              </div>
            </div>
          </td>
          <td><span class="emp-table-muted">${escapeHtml(jobTitle)}</span></td>
          <td><span class="emp-status-badge ${status}">${status}</span></td>
          <td><span class="emp-match-score">${matchScore}%</span></td>
          <td><span class="emp-table-muted">${timeAgo(appliedAt)}</span></td>
          <td>
            <div class="emp-app-actions">
              ${status !== 'shortlisted' && status !== 'interview' && status !== 'hired' ? 
                `<button class="emp-action-btn" data-action="shortlist" data-id="${app.id}" title="Shortlist">⭐</button>` : ''
              }
              ${status !== 'rejected' && status !== 'hired' ? 
                `<button class="emp-action-btn" data-action="reject" data-id="${app.id}" title="Reject">❌</button>` : ''
              }
              ${status === 'shortlisted' ? 
                `<button class="emp-action-btn" data-action="interview" data-id="${app.id}" title="Schedule Interview">📅</button>` : ''
              }
              ${status === 'interview' ? 
                `<button class="emp-action-btn" data-action="hire" data-id="${app.id}" title="Mark as Hired">🎉</button>` : ''
              }
              <a href="/employer/applications/${app.id}" class="emp-action-btn" title="View Details">👁️</a>
            </div>
          </td>
        </tr>
      `;
    }).join('');
    
    // Re-attach event listeners
    attachActionListeners();
    attachCheckboxListeners();
    
    // Update result hint
    if (elements.resultHint) {
      elements.resultHint.textContent = `Showing ${applications.length} applications`;
    }
  }

  // Attach action button listeners
  function attachActionListeners() {
    document.querySelectorAll('.emp-action-btn[data-action]').forEach(btn => {
      btn.removeEventListener('click', handleActionClick);
      btn.addEventListener('click', handleActionClick);
    });
  }

  // Handle action button clicks
  async function handleActionClick(e) {
    e.stopPropagation();
    const btn = e.currentTarget;
    const action = btn.getAttribute('data-action');
    const id = btn.getAttribute('data-id');
    
    if (!id) return;
    
    switch (action) {
      case 'shortlist':
        await shortlistApplication(id, btn);
        break;
      case 'reject':
        await rejectApplication(id, btn);
        break;
      case 'interview':
        openInterviewModal(id);
        break;
      case 'hire':
        await hireCandidate(id, btn);
        break;
    }
  }

  // Shortlist application
  async function shortlistApplication(id, btn) {
    try {
      if (btn) btn.disabled = true;
      
      await AngaziaAPI.applications.shortlist(id);
      showToast('Application shortlisted!', 'success');
      
      // Update UI
      const row = document.querySelector(`.emp-app-row[data-id="${id}"]`);
      if (row) {
        const statusBadge = row.querySelector('.emp-status-badge');
        if (statusBadge) {
          statusBadge.className = 'emp-status-badge shortlisted';
          statusBadge.textContent = 'shortlisted';
        }
        // Update actions
        const actionsDiv = row.querySelector('.emp-app-actions');
        if (actionsDiv) {
          actionsDiv.innerHTML = `
            <button class="emp-action-btn" data-action="interview" data-id="${id}" title="Schedule Interview">📅</button>
            <a href="/employer/applications/${id}" class="emp-action-btn" title="View Details">👁️</a>
          `;
        }
        attachActionListeners();
      }
      
    } catch (error) {
      console.error('Shortlist failed:', error);
      showToast(error.message || 'Failed to shortlist application', 'error');
    } finally {
      if (btn) btn.disabled = false;
    }
  }

  // Reject application
  async function rejectApplication(id, btn) {
    if (!confirm('Are you sure you want to reject this application?')) return;
    
    try {
      if (btn) btn.disabled = true;
      
      await AngaziaAPI.applications.reject(id);
      showToast('Application rejected', 'success');
      
      // Update UI
      const row = document.querySelector(`.emp-app-row[data-id="${id}"]`);
      if (row) {
        const statusBadge = row.querySelector('.emp-status-badge');
        if (statusBadge) {
          statusBadge.className = 'emp-status-badge rejected';
          statusBadge.textContent = 'rejected';
        }
      }
      
    } catch (error) {
      console.error('Reject failed:', error);
      showToast(error.message || 'Failed to reject application', 'error');
    } finally {
      if (btn) btn.disabled = false;
    }
  }

  // Open interview modal
  function openInterviewModal(id) {
    pendingInterviewId = id;
    if (elements.interviewModal) {
      elements.interviewModal.style.display = 'flex';
      // Set default datetime to tomorrow at 10 AM
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      tomorrow.setHours(10, 0, 0);
      if (elements.interviewDatetime) {
        elements.interviewDatetime.value = tomorrow.toISOString().slice(0, 16);
      }
      if (elements.interviewType) elements.interviewType.value = 'technical';
      if (elements.interviewNotes) elements.interviewNotes.value = '';
    }
  }

  // Schedule interview
  async function scheduleInterview() {
    if (!pendingInterviewId) return;
    
    const datetime = elements.interviewDatetime?.value;
    const interviewType = elements.interviewType?.value;
    const notes = elements.interviewNotes?.value;
    
    if (!datetime) {
      showToast('Please select an interview date and time', 'warning');
      return;
    }
    
    try {
      await AngaziaAPI.applications.interview(pendingInterviewId, {
        scheduled_at: datetime,
        interview_type: interviewType,
        notes: notes
      });
      
      showToast('Interview scheduled successfully!', 'success');
      closeInterviewModal();
      
      // Update UI
      const row = document.querySelector(`.emp-app-row[data-id="${pendingInterviewId}"]`);
      if (row) {
        const statusBadge = row.querySelector('.emp-status-badge');
        if (statusBadge) {
          statusBadge.className = 'emp-status-badge interview';
          statusBadge.textContent = 'interview';
        }
        // Update actions
        const actionsDiv = row.querySelector('.emp-app-actions');
        if (actionsDiv) {
          actionsDiv.innerHTML = `
            <button class="emp-action-btn" data-action="hire" data-id="${pendingInterviewId}" title="Mark as Hired">🎉</button>
            <a href="/employer/applications/${pendingInterviewId}" class="emp-action-btn" title="View Details">👁️</a>
          `;
        }
        attachActionListeners();
      }
      
      pendingInterviewId = null;
      
    } catch (error) {
      console.error('Schedule interview failed:', error);
      showToast(error.message || 'Failed to schedule interview', 'error');
    }
  }

  // Close interview modal
  function closeInterviewModal() {
    if (elements.interviewModal) {
      elements.interviewModal.style.display = 'none';
    }
    pendingInterviewId = null;
  }

  // Hire candidate
  async function hireCandidate(id, btn) {
    if (!confirm('Mark this candidate as hired?')) return;
    
    try {
      if (btn) btn.disabled = true;
      
      await AngaziaAPI.applications.hire(id);
      showToast('Candidate marked as hired!', 'success');
      
      const row = document.querySelector(`.emp-app-row[data-id="${id}"]`);
      if (row) {
        const statusBadge = row.querySelector('.emp-status-badge');
        if (statusBadge) {
          statusBadge.className = 'emp-status-badge hired';
          statusBadge.textContent = 'hired';
        }
        const actionsDiv = row.querySelector('.emp-app-actions');
        if (actionsDiv) {
          actionsDiv.innerHTML = `<a href="/employer/applications/${id}" class="emp-action-btn" title="View Details">👁️</a>`;
        }
      }
      
    } catch (error) {
      console.error('Hire failed:', error);
      showToast(error.message || 'Failed to mark as hired', 'error');
    } finally {
      if (btn) btn.disabled = false;
    }
  }

  // Bulk actions
  async function bulkShortlist() {
    if (selectedIds.length === 0) {
      showToast('Select at least one application', 'warning');
      return;
    }
    
    try {
      await AngaziaAPI.applications.bulkShortlist({ application_ids: selectedIds });
      showToast(`${selectedIds.length} application(s) shortlisted`, 'success');
      
      // Update UI for all selected rows
      selectedIds.forEach(id => {
        const row = document.querySelector(`.emp-app-row[data-id="${id}"]`);
        if (row) {
          const statusBadge = row.querySelector('.emp-status-badge');
          if (statusBadge) {
            statusBadge.className = 'emp-status-badge shortlisted';
            statusBadge.textContent = 'shortlisted';
          }
        }
      });
      
      clearSelection();
      
    } catch (error) {
      console.error('Bulk shortlist failed:', error);
      showToast(error.message || 'Bulk action failed', 'error');
    }
  }

  // Bulk reject
  async function bulkReject() {
    if (selectedIds.length === 0) {
      showToast('Select at least one application', 'warning');
      return;
    }
    
    if (!confirm(`Reject ${selectedIds.length} application(s)?`)) return;
    
    try {
      await AngaziaAPI.applications.bulkReject({ application_ids: selectedIds });
      
      showToast(`${selectedIds.length} application(s) rejected`, 'success');
      
      selectedIds.forEach(id => {
        const row = document.querySelector(`.emp-app-row[data-id="${id}"]`);
        if (row) {
          const statusBadge = row.querySelector('.emp-status-badge');
          if (statusBadge) {
            statusBadge.className = 'emp-status-badge rejected';
            statusBadge.textContent = 'rejected';
          }
        }
      });
      
      clearSelection();
      
    } catch (error) {
      console.error('Bulk reject failed:', error);
      showToast(error.message || 'Bulk action failed', 'error');
    }
  }

  // Checkbox handlers
  function attachCheckboxListeners() {
    document.querySelectorAll('.emp-app-cb').forEach(cb => {
      cb.removeEventListener('change', handleCheckboxChange);
      cb.addEventListener('change', handleCheckboxChange);
    });
  }

  function handleCheckboxChange(e) {
    const cb = e.currentTarget;
    const id = cb.value;
    
    if (cb.checked) {
      if (!selectedIds.includes(id)) selectedIds.push(id);
    } else {
      selectedIds = selectedIds.filter(s => s !== id);
    }
    
    updateBulkBar();
    updateSelectAll();
  }

  function updateBulkBar() {
    if (!elements.bulkBar) return;
    
    if (selectedIds.length > 0) {
      elements.bulkBar.style.display = 'flex';
      if (elements.selectedCount) {
        elements.selectedCount.textContent = `${selectedIds.length} selected`;
      }
    } else {
      elements.bulkBar.style.display = 'none';
    }
  }

  function updateSelectAll() {
    if (!elements.selectAll) return;
    
    const allCbs = document.querySelectorAll('.emp-app-cb');
    const allChecked = Array.from(allCbs).every(cb => cb.checked);
    const someChecked = Array.from(allCbs).some(cb => cb.checked);
    
    elements.selectAll.checked = allChecked;
    elements.selectAll.indeterminate = !allChecked && someChecked;
  }

  function selectAll() {
    const allCbs = document.querySelectorAll('.emp-app-cb');
    const isChecked = elements.selectAll?.checked || false;
    
    allCbs.forEach(cb => {
      cb.checked = isChecked;
      const id = cb.value;
      if (isChecked) {
        if (!selectedIds.includes(id)) selectedIds.push(id);
      } else {
        selectedIds = [];
      }
    });
    
    updateBulkBar();
  }

  function clearSelection() {
    selectedIds = [];
    document.querySelectorAll('.emp-app-cb').forEach(cb => {
      cb.checked = false;
    });
    if (elements.selectAll) elements.selectAll.checked = false;
    updateBulkBar();
  }

  // Pagination
  function updatePagination(total) {
    if (!elements.pagination) return;
    
    if (totalPages <= 1) {
      elements.pagination.style.display = 'none';
      return;
    }
    
    elements.pagination.style.display = 'flex';
    if (elements.pageInfo) {
      elements.pageInfo.textContent = `Page ${currentPage} of ${totalPages}`;
    }
    if (elements.prevBtn) elements.prevBtn.disabled = currentPage === 1;
    if (elements.nextBtn) elements.nextBtn.disabled = currentPage === totalPages;
  }

  function prevPage() {
    if (currentPage > 1) {
      currentPage--;
      loadApplications();
    }
  }

  function nextPage() {
    if (currentPage < totalPages) {
      currentPage++;
      loadApplications();
    }
  }

  // Filter initialization (simplified - you can expand with a proper filter component)
  function initFilters() {
    // For now, just load without filters
    loadApplications();
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

  function showEmpty() {
    if (elements.tableBody) {
      elements.tableBody.innerHTML = `
        <tr><td colspan="7" style="text-align:center;padding:60px;">
          <div class="emp-empty-icon">📋</div>
          <h3 style="margin: 16px 0 8px;">No applications yet</h3>
          <p style="color: var(--muted);">Applications will appear here when candidates apply to your jobs.</p>
        </td></tr>
      `;
    }
    if (elements.content) elements.content.style.display = 'block';
  }

  function showError(message) {
    if (elements.error) {
      if (elements.errorMsg) elements.errorMsg.textContent = message;
      elements.error.style.display = 'block';
    }
    if (elements.content) elements.content.style.display = 'none';
  }

  // Initialize event listeners
  function initEventListeners() {
    if (elements.selectAll) {
      elements.selectAll.addEventListener('change', selectAll);
    }
    if (elements.clearSelection) {
      elements.clearSelection.addEventListener('click', clearSelection);
    }
    if (elements.prevBtn) {
      elements.prevBtn.addEventListener('click', prevPage);
    }
    if (elements.nextBtn) {
      elements.nextBtn.addEventListener('click', nextPage);
    }
    if (elements.interviewConfirm) {
      elements.interviewConfirm.addEventListener('click', scheduleInterview);
    }
    if (elements.interviewCancel) {
      elements.interviewCancel.addEventListener('click', closeInterviewModal);
    }
    if (elements.interviewClose) {
      elements.interviewClose.addEventListener('click', closeInterviewModal);
    }
    
    // Bulk action buttons
    const bulkShortlistBtn = document.querySelector('[data-action="bulk-shortlist"]');
    if (bulkShortlistBtn) {
      bulkShortlistBtn.addEventListener('click', bulkShortlist);
    }
    const bulkRejectBtn = document.querySelector('[data-action="bulk-reject"]');
    if (bulkRejectBtn) {
      bulkRejectBtn.addEventListener('click', bulkReject);
    }
    
    // Close modal on overlay click
    if (elements.interviewModal) {
      elements.interviewModal.addEventListener('click', (e) => {
        if (e.target === elements.interviewModal) closeInterviewModal();
      });
    }
  }

  // Initialize the page
  function init() {
    addStyles();
    initEventListeners();
    initFilters();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();