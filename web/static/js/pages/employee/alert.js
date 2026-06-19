/**
 * Employee Job Alerts Page
 * Handles creating, editing, toggling, and deleting job alerts
 */

(function() {
  'use strict';

  let alerts = [];
  let editingAlertId = null;

  // DOM Elements
  let elements = {};

  // Initialize page
  async function init() {
    cacheElements();
    attachEventListeners();
    await loadAlerts();
  }

  // Cache DOM elements
  function cacheElements() {
    elements = {
      loading: document.getElementById('alerts-loading'),
      error: document.getElementById('alerts-error'),
      errorMsg: document.getElementById('alerts-error-msg'),
      content: document.getElementById('alerts-content'),
      list: document.getElementById('alerts-list'),
      createBtn: document.getElementById('create-alert-btn'),
      modal: document.getElementById('alert-modal'),
      modalTitle: document.getElementById('modal-title'),
      modalSave: document.getElementById('modal-save'),
      modalCancel: document.getElementById('modal-cancel'),
      modalClose: document.getElementById('modal-close'),
      alertForm: document.getElementById('alert-form')
    };
  }

  // Attach event listeners
  function attachEventListeners() {
    if (elements.createBtn) {
      elements.createBtn.addEventListener('click', () => openCreateModal());
    }
    
    if (elements.modalSave) {
      elements.modalSave.addEventListener('click', () => saveAlert());
    }
    
    if (elements.modalCancel || elements.modalClose) {
      const closeModal = () => closeModalWindow();
      if (elements.modalCancel) elements.modalCancel.addEventListener('click', closeModal);
      if (elements.modalClose) elements.modalClose.addEventListener('click', closeModal);
      
      if (elements.modal) {
        elements.modal.addEventListener('click', (e) => {
          if (e.target === elements.modal) closeModalWindow();
        });
      }
    }
  }

  // Load alerts from API
  async function loadAlerts() {
    showLoading(true);
    showError(false);

    try {
      const response = await AngaziaAPI.alerts.list();
      
      let alertsData = [];
      if (response && response.data && response.data.searches) {
        alertsData = response.data.searches;
      } else if (response && response.data && Array.isArray(response.data)) {
        alertsData = response.data;
      } else if (Array.isArray(response)) {
        alertsData = response;
      } else if (response && response.alerts) {
        alertsData = response.alerts;
      } else if (response && response.searches) {
        alertsData = response.searches;
      }
      
      alerts = alertsData.map(formatAlert);
      
      renderAlerts();
      
      showLoading(false);
      showContent();
      
    } catch (error) {
      console.error('Failed to load alerts:', error);
      showLoading(false);
      showError(true, error.message || 'Failed to load your alerts');
    }
  }

  // Format alert data
  function formatAlert(alert) {
    return {
      id: alert.id,
      name: alert.name || 'Untitled Alert',
      keywords: alert.filters?.keywords || alert.keywords || '',
      location: alert.filters?.location || alert.location || '',
      jobType: alert.filters?.job_type || alert.employment_type || '',
      salaryMin: alert.filters?.salary_min || alert.salary_min || null,
      salaryMax: alert.filters?.salary_max || alert.salary_max || null,
      experienceLevel: alert.filters?.experience_level || alert.experience_level || '',
      skills: alert.filters?.skills || (alert.required_skills || []),
      frequency: alert.frequency || 'daily',
      isActive: alert.is_active !== false,
      weeklyMatches: alert.weekly_matches || 0,
      lastSentAt: alert.last_sent_at
    };
  }

  // Render alerts list
  function renderAlerts() {
    if (!elements.list) return;
    
    if (!alerts || alerts.length === 0) {
      elements.list.innerHTML = `
        <div class="emp-empty">
          <div class="emp-empty-icon">🔔</div>
          <p class="emp-empty-text">No job alerts set up. Create alerts to get notified about jobs that match your preferences.</p>
          <button class="emp-btn emp-btn-primary" id="empty-create-btn">+ Create Your First Alert</button>
        </div>
      `;
      
      const emptyBtn = document.getElementById('empty-create-btn');
      if (emptyBtn) {
        emptyBtn.addEventListener('click', () => openCreateModal());
      }
      return;
    }
    
    elements.list.innerHTML = alerts.map(alert => createAlertCard(alert)).join('');
    
    alerts.forEach(alert => {
      const card = document.querySelector(`.emp-alert-card[data-alert-id="${alert.id}"]`);
      if (!card) return;
      
      const toggleBtn = card.querySelector('.emp-alert-toggle');
      const deleteBtn = card.querySelector('.emp-alert-delete');
      const editBtn = card.querySelector('.emp-alert-edit');
      
      if (toggleBtn) {
        toggleBtn.addEventListener('click', () => toggleAlert(alert.id));
      }
      if (deleteBtn) {
        deleteBtn.addEventListener('click', () => deleteAlert(alert.id));
      }
      if (editBtn) {
        editBtn.addEventListener('click', () => openEditModal(alert.id));
      }
    });
  }

  // Create alert card HTML
  function createAlertCard(alert) {
    const statusClass = alert.isActive ? 'emp-dot-green' : 'emp-dot-gray';
    const statusText = alert.isActive ? 'Active' : 'Paused';
    const toggleText = alert.isActive ? 'Pause' : 'Resume';
    
    let criteria = [];
    if (alert.location) criteria.push(`📍 ${escapeHtml(alert.location)}`);
    if (alert.jobType) criteria.push(`💼 ${escapeHtml(alert.jobType)}`);
    if (alert.keywords) criteria.push(`🔍 ${escapeHtml(alert.keywords)}`);
    if (alert.salaryMin || alert.salaryMax) {
      const salary = formatSalary(alert.salaryMin, alert.salaryMax);
      criteria.push(`💰 ${salary}`);
    }
    if (alert.experienceLevel) criteria.push(`📊 ${getExperienceLabel(alert.experienceLevel)}`);
    if (alert.skills && alert.skills.length) {
      const skillList = Array.isArray(alert.skills) ? alert.skills.slice(0, 3).join(', ') : alert.skills;
      criteria.push(`🛠️ ${escapeHtml(skillList)}`);
    }
    
    const frequencyLabel = alert.frequency === 'daily' ? 'Daily' : alert.frequency === 'weekly' ? 'Weekly' : 'Instant';
    
    return `
      <div class="emp-alert-card" data-alert-id="${alert.id}">
        <div class="emp-alert-head">
          <div class="emp-alert-info">
            <h3 class="emp-alert-name">${escapeHtml(alert.name)}</h3>
            <span class="emp-dot ${statusClass}"></span>
            <span class="emp-alert-state">${statusText}</span>
          </div>
          <div class="emp-alert-actions">
            <button class="emp-alert-edit" title="Edit Alert">✏️</button>
            <button class="emp-alert-toggle">${toggleText}</button>
            <button class="emp-alert-delete" title="Delete Alert">🗑️</button>
          </div>
        </div>
        
        <div class="emp-alert-criteria">
          ${criteria.map(c => `<span class="emp-alert-crit">${c}</span>`).join('')}
          ${criteria.length === 0 ? '<span class="emp-alert-crit">No filters set - all jobs</span>' : ''}
        </div>
        
        <div class="emp-alert-meta">
          <span>📅 Frequency: <strong>${frequencyLabel}</strong></span>
          <span>🎯 Matches this week: <strong>${alert.weeklyMatches}</strong></span>
          ${alert.lastSentAt ? `<span class="emp-alert-last-sent">Last sent: ${formatRelativeDate(alert.lastSentAt)}</span>` : ''}
        </div>
      </div>
    `;
  }

  // Open create modal
  function openCreateModal() {
    editingAlertId = null;
    elements.modalTitle.textContent = 'Create New Alert';
    resetForm();
    openModalWindow();
  }

  // Open edit modal
  function openEditModal(alertId) {
    const alert = alerts.find(a => a.id === alertId);
    if (!alert) return;
    
    editingAlertId = alertId;
    elements.modalTitle.textContent = 'Edit Alert';
    
    document.querySelector('[name="name"]').value = alert.name || '';
    document.querySelector('[name="keywords"]').value = Array.isArray(alert.keywords) ? alert.keywords.join(', ') : (alert.keywords || '');
    document.querySelector('[name="location"]').value = alert.location || '';
    document.querySelector('[name="job_type"]').value = alert.jobType || '';
    document.querySelector('[name="salary_min"]').value = alert.salaryMin || '';
    document.querySelector('[name="salary_max"]').value = alert.salaryMax || '';
    document.querySelector('[name="experience_level"]').value = alert.experienceLevel || '';
    document.querySelector('[name="skills"]').value = Array.isArray(alert.skills) ? alert.skills.join(', ') : (alert.skills || '');
    document.querySelector('[name="frequency"]').value = alert.frequency || 'daily';
    
    openModalWindow();
  }

  // Save alert
  async function saveAlert() {
    const form = elements.alertForm;
    const formData = new FormData(form);
    
    const filters = {
      keywords: formData.get('keywords') || '',
      location: formData.get('location') || '',
      job_type: formData.get('job_type') || '',
      salary_min: formData.get('salary_min') ? parseInt(formData.get('salary_min')) : null,
      salary_max: formData.get('salary_max') ? parseInt(formData.get('salary_max')) : null,
      experience_level: formData.get('experience_level') || '',
      skills: formData.get('skills') ? formData.get('skills').split(',').map(s => s.trim()).filter(Boolean) : []
    };
    
    const data = {
      name: formData.get('name'),
      filters: filters,
      frequency: formData.get('frequency') || 'daily'
    };
    
    if (!data.name) {
      showToast('Please enter an alert name', 'error');
      return;
    }
    
    const saveBtn = elements.modalSave;
    const originalText = saveBtn.textContent;
    saveBtn.disabled = true;
    saveBtn.textContent = 'Saving...';
    
    try {
      if (editingAlertId) {
        await AngaziaAPI.alerts.update(editingAlertId, data);
      } else {
        await AngaziaAPI.alerts.create(data);
      }
      
      closeModalWindow();
      await loadAlerts();
      
    } catch (error) {
      console.error('Save alert failed:', error);
      console.error(error);
    } finally {
      saveBtn.disabled = false;
      saveBtn.textContent = originalText;
    }
  }

  // Toggle alert
  async function toggleAlert(alertId) {
    const alert = alerts.find(a => a.id === alertId);
    if (!alert) return;
    
    try {
      await AngaziaAPI.alerts.update(alertId, { is_active: !alert.isActive });

      await loadAlerts();
    } catch (error) {
      console.error('Toggle alert failed:', error);
      console.error(error);
    }
  }

  // Delete alert
  async function deleteAlert(alertId) {
    const confirmed = await confirmDialog('Are you sure you want to delete this alert? You will no longer receive notifications for this search.');
    if (!confirmed) return;
    
    try {
      await AngaziaAPI.alerts.delete(alertId);

      await loadAlerts();
    } catch (error) {
      console.error('Delete alert failed:', error);
      console.error(error);
    }
  }

  // Reset form
  function resetForm() {
    if (!elements.alertForm) return;
    elements.alertForm.reset();
    const freqSelect = document.querySelector('[name="frequency"]');
    if (freqSelect) freqSelect.value = 'daily';
  }

  // Modal controls
  function openModalWindow() {
    if (elements.modal) elements.modal.style.display = 'flex';
    document.body.style.overflow = 'hidden';
  }

  function closeModalWindow() {
    if (elements.modal) elements.modal.style.display = 'none';
    document.body.style.overflow = '';
    resetForm();
    editingAlertId = null;
  }

  // UI State
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

  // Helpers
  function formatSalary(min, max) {
    const formatter = (n) => {
      if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
      if (n >= 1000) return (n / 1000).toFixed(0) + 'K';
      return n;
    };
    if (min && max) return `${formatter(min)} - ${formatter(max)}`;
    if (min) return `${formatter(min)}+`;
    if (max) return `Up to ${formatter(max)}`;
    return 'Not specified';
  }

  function getExperienceLabel(level) {
    const labels = { 'entry': 'Entry Level', 'junior': 'Junior', 'mid': 'Mid Level', 'senior': 'Senior', 'lead': 'Lead' };
    return labels[level] || level || 'Any Level';
  }

  function formatRelativeDate(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const now = new Date();
    const diffDays = Math.floor((now - date) / 86400000);
    if (diffDays === 0) return 'today';
    if (diffDays === 1) return 'yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    return date.toLocaleDateString('en-KE', { month: 'short', day: 'numeric' });
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
    return AngaziaModal.confirm(message);
  }

  function showToast(message, type) {
    if (window.AngaziaApp && window.AngaziaApp.showToast) {
      window.AngaziaApp.showToast(message, type);
    } else if (type) {
      AngaziaModal.alert(message, type === 'error' ? 'Error' : 'Success');
    } else {
      AngaziaModal.alert(message);
    }
  }

  // Start
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();