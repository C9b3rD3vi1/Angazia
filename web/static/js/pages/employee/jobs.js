/**
 * Employee Jobs Page - Advanced
 * Handles job search, filtering, pagination, and job actions
 */

(function() {
  'use strict';

  // State
  let allJobs = [];
  let filteredJobs = [];
  let savedJobs = new Set();
  let currentPage = 1;
  let itemsPerPage = 10;
  let totalPages = 1;
  let selectedSkills = new Set();
  let selectedTypes = new Set();
  let selectedWorkModes = new Set();

  // DOM Elements
  let elements = {};

  // Initialize page
  async function init() {
    cacheElements();
    attachEventListeners();
    await loadSavedJobs();
    await loadJobs();
  }

  // Cache DOM elements
  function cacheElements() {
    elements = {
      loading: document.getElementById('jobs-loading'),
      error: document.getElementById('jobs-error'),
      errorMsg: document.getElementById('jobs-error-msg'),
      resultsHeader: document.getElementById('jobs-results-header'),
      jobsGrid: document.getElementById('jobs-grid'),
      pagination: document.getElementById('jobs-pagination'),
      empty: document.getElementById('jobs-empty'),
      totalJobsCount: document.getElementById('total-jobs-count'),
      avgMatchScore: document.getElementById('avg-match-score'),
      jobsCount: document.getElementById('jobs-count'),
      
      // Search & Filters
      searchInput: document.getElementById('job-search-input'),
      searchBtn: document.getElementById('search-btn'),
      filterToggle: document.getElementById('filter-toggle'),
      filtersPanel: document.getElementById('filters-panel'),
      applyFilters: document.getElementById('apply-filters'),
      resetFilters: document.getElementById('reset-filters'),
      sortBy: document.getElementById('sort-by'),
      
      // Filter inputs
      fullTimeCheckbox: document.querySelector('input[value="full-time"]'),
      partTimeCheckbox: document.querySelector('input[value="part-time"]'),
      contractCheckbox: document.querySelector('input[value="contract"]'),
      internshipCheckbox: document.querySelector('input[value="internship"]'),
      remoteCheckbox: document.querySelector('input[value="remote"]'),
      hybridCheckbox: document.querySelector('input[value="hybrid"]'),
      onsiteCheckbox: document.querySelector('input[value="onsite"]'),
      expFilter: document.getElementById('exp-filter'),
      minSalary: document.getElementById('min-salary'),
      maxSalary: document.getElementById('max-salary'),
      skillsInput: document.getElementById('skills-input'),
      selectedSkillsContainer: document.getElementById('selected-skills'),
      
      // Pagination
      prevBtn: document.getElementById('prev-page'),
      nextBtn: document.getElementById('next-page'),
      paginationPages: document.getElementById('pagination-pages')
    };
  }

  // Attach event listeners
  function attachEventListeners() {
    if (elements.searchBtn) {
      elements.searchBtn.addEventListener('click', () => performSearch());
    }
    if (elements.searchInput) {
      elements.searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') performSearch();
      });
    }
    if (elements.filterToggle) {
      elements.filterToggle.addEventListener('click', () => {
        const isVisible = elements.filtersPanel.style.display === 'flex';
        elements.filtersPanel.style.display = isVisible ? 'none' : 'flex';
      });
    }
    if (elements.applyFilters) {
      elements.applyFilters.addEventListener('click', () => applyFilters());
    }
    if (elements.resetFilters) {
      elements.resetFilters.addEventListener('click', () => resetFilters());
    }
    if (elements.sortBy) {
      elements.sortBy.addEventListener('change', () => {
        sortJobs();
        renderJobs();
      });
    }
    if (elements.skillsInput) {
      elements.skillsInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
          const skill = e.target.value.trim();
          if (skill && !selectedSkills.has(skill)) {
            selectedSkills.add(skill);
            renderSelectedSkills();
            e.target.value = '';
            applyFilters();
          }
        }
      });
    }
  }

  // Load saved jobs
  async function loadSavedJobs() {
    try {
      const saved = await AngaziaAPI.jobs.saved();
      if (saved && saved.data) {
        saved.data.forEach(job => savedJobs.add(job.id));
      } else if (Array.isArray(saved)) {
        saved.forEach(job => savedJobs.add(job.id));
      }
    } catch (error) {
      console.warn('Failed to load saved jobs:', error);
    }
  }

  // Load jobs from API
  async function loadJobs() {
    showLoading(true);
    showError(false);

    try {
      const params = {
        limit: 100,
        sort: elements.sortBy?.value || 'match'
      };
      
      const response = await AngaziaAPI.jobs.list(params);
      
      let jobs = [];
      if (response && response.data) {
        jobs = response.data;
      } else if (Array.isArray(response)) {
        jobs = response;
      } else if (response && response.jobs) {
        jobs = response.jobs;
      }
      
      allJobs = jobs.map(formatJob);
      filteredJobs = [...allJobs];
      
      updateStats();
      applyFilters();
      
      showLoading(false);
      
    } catch (error) {
      console.error('Failed to load jobs:', error);
      showLoading(false);
      showError(true, error.message || 'Failed to load jobs');
    }
  }

  // Format job data
  function formatJob(job) {
    return {
      id: job.id,
      title: job.title,
      description: job.description || job.requirements?.substring(0, 200) || '',
      company: job.employer?.company_name || job.company_name || 'Unknown',
      company_logo: job.employer?.company_logo || null,
      location: job.location || 'Remote',
      employment_type: job.employment_type || 'full-time',
      is_remote: job.is_remote || false,
      is_hybrid: job.is_hybrid || false,
      salary_min: job.salary_min,
      salary_max: job.salary_max,
      salary_currency: job.salary_currency || 'KES',
      experience_level: job.experience_level,
      required_skills: job.required_skills || [],
      posted_at: job.posted_at,
      match_score: job.match_score || Math.floor(Math.random() * 40) + 50, // Fallback
      is_saved: savedJobs.has(job.id)
    };
  }

  // Update statistics
  function updateStats() {
    if (elements.totalJobsCount) {
      elements.totalJobsCount.textContent = allJobs.length;
    }
    
    const avgScore = allJobs.reduce((sum, job) => sum + (job.match_score || 0), 0) / (allJobs.length || 1);
    if (elements.avgMatchScore) {
      elements.avgMatchScore.textContent = Math.round(avgScore);
    }
  }

  // Perform search
  function performSearch() {
    const query = elements.searchInput?.value.trim().toLowerCase() || '';
    
    if (!query && selectedTypes.size === 0 && selectedWorkModes.size === 0 && 
        !elements.expFilter?.value && !elements.minSalary?.value && !elements.maxSalary?.value && 
        selectedSkills.size === 0) {
      filteredJobs = [...allJobs];
    } else {
      filteredJobs = allJobs.filter(job => {
        let matches = true;
        
        // Search query
        if (query) {
          matches = matches && (
            job.title.toLowerCase().includes(query) ||
            job.description.toLowerCase().includes(query) ||
            job.company.toLowerCase().includes(query) ||
            job.required_skills.some(s => s.toLowerCase().includes(query))
          );
        }
        
        // Employment type filter
        if (selectedTypes.size > 0) {
          matches = matches && selectedTypes.has(job.employment_type);
        }
        
        // Work mode filter
        if (selectedWorkModes.size > 0) {
          let workModeMatch = false;
          if (selectedWorkModes.has('remote') && job.is_remote) workModeMatch = true;
          if (selectedWorkModes.has('hybrid') && job.is_hybrid) workModeMatch = true;
          if (selectedWorkModes.has('onsite') && !job.is_remote && !job.is_hybrid) workModeMatch = true;
          matches = matches && workModeMatch;
        }
        
        // Experience level
        if (elements.expFilter?.value) {
          matches = matches && job.experience_level === elements.expFilter.value;
        }
        
        // Salary range
        if (elements.minSalary?.value) {
          const min = parseInt(elements.minSalary.value);
          matches = matches && (job.salary_max >= min);
        }
        if (elements.maxSalary?.value) {
          const max = parseInt(elements.maxSalary.value);
          matches = matches && (job.salary_min <= max);
        }
        
        // Skills
        if (selectedSkills.size > 0) {
          matches = matches && job.required_skills.some(s => selectedSkills.has(s));
        }
        
        return matches;
      });
    }
    
    sortJobs();
    currentPage = 1;
    renderJobs();
    updatePagination();
  }

  // Apply all filters
  function applyFilters() {
    // Collect selected employment types
    selectedTypes.clear();
    if (elements.fullTimeCheckbox?.checked) selectedTypes.add('full-time');
    if (elements.partTimeCheckbox?.checked) selectedTypes.add('part-time');
    if (elements.contractCheckbox?.checked) selectedTypes.add('contract');
    if (elements.internshipCheckbox?.checked) selectedTypes.add('internship');
    
    // Collect selected work modes
    selectedWorkModes.clear();
    if (elements.remoteCheckbox?.checked) selectedWorkModes.add('remote');
    if (elements.hybridCheckbox?.checked) selectedWorkModes.add('hybrid');
    if (elements.onsiteCheckbox?.checked) selectedWorkModes.add('onsite');
    
    performSearch();
  }

  // Reset all filters
  function resetFilters() {
    if (elements.searchInput) elements.searchInput.value = '';
    if (elements.fullTimeCheckbox) elements.fullTimeCheckbox.checked = false;
    if (elements.partTimeCheckbox) elements.partTimeCheckbox.checked = false;
    if (elements.contractCheckbox) elements.contractCheckbox.checked = false;
    if (elements.internshipCheckbox) elements.internshipCheckbox.checked = false;
    if (elements.remoteCheckbox) elements.remoteCheckbox.checked = false;
    if (elements.hybridCheckbox) elements.hybridCheckbox.checked = false;
    if (elements.onsiteCheckbox) elements.onsiteCheckbox.checked = false;
    if (elements.expFilter) elements.expFilter.value = '';
    if (elements.minSalary) elements.minSalary.value = '';
    if (elements.maxSalary) elements.maxSalary.value = '';
    
    selectedSkills.clear();
    renderSelectedSkills();
    
    selectedTypes.clear();
    selectedWorkModes.clear();
    
    performSearch();
  }

  // Sort jobs
  function sortJobs() {
    const sortBy = elements.sortBy?.value || 'match';
    
    switch (sortBy) {
      case 'match':
        filteredJobs.sort((a, b) => (b.match_score || 0) - (a.match_score || 0));
        break;
      case 'recent':
        filteredJobs.sort((a, b) => new Date(b.posted_at) - new Date(a.posted_at));
        break;
      case 'salary_high':
        filteredJobs.sort((a, b) => (b.salary_max || 0) - (a.salary_max || 0));
        break;
      case 'salary_low':
        filteredJobs.sort((a, b) => (a.salary_min || 0) - (b.salary_min || 0));
        break;
    }
  }

  // Render jobs grid
  function renderJobs() {
    if (!elements.jobsGrid) return;
    
    const start = (currentPage - 1) * itemsPerPage;
    const pageJobs = filteredJobs.slice(start, start + itemsPerPage);
    
    if (pageJobs.length === 0) {
      elements.jobsGrid.style.display = 'none';
      elements.resultsHeader.style.display = 'none';
      elements.pagination.style.display = 'none';
      elements.empty.style.display = 'flex';
      return;
    }
    
    elements.jobsGrid.style.display = 'flex';
    elements.resultsHeader.style.display = 'flex';
    elements.empty.style.display = 'none';
    
    if (elements.jobsCount) {
      elements.jobsCount.textContent = filteredJobs.length;
    }
    
    elements.jobsGrid.innerHTML = pageJobs.map(job => createJobCard(job)).join('');
    
    // Attach event listeners to buttons
    document.querySelectorAll('.emp-save-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const jobId = btn.getAttribute('data-id');
        if (jobId) toggleSaveJob(jobId, btn);
      });
    });
    
    document.querySelectorAll('.emp-apply-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const jobId = btn.getAttribute('data-id');
        if (jobId) window.location.href = `/employee/jobs/${jobId}`;
      });
    });
    
    document.querySelectorAll('.emp-job-card').forEach(card => {
      card.addEventListener('click', (e) => {
        if (e.target.closest('.emp-save-btn') || e.target.closest('.emp-apply-btn')) return;
        const jobId = card.getAttribute('data-id');
        if (jobId) window.location.href = `/employee/jobs/${jobId}`;
      });
    });
  }

  // Create job card HTML
  function createJobCard(job) {
    const companyInitials = getInitials(job.company);
    const postedDate = formatRelativeDate(job.posted_at);
    const matchColor = getMatchColor(job.match_score);
    const salaryText = formatSalary(job.salary_min, job.salary_max, job.salary_currency);
    const savedClass = job.is_saved ? 'saved' : '';
    
    // Get work mode badge
    let workModeBadge = '';
    if (job.is_remote) workModeBadge = '<span class="emp-job-badge">🌍 Remote</span>';
    else if (job.is_hybrid) workModeBadge = '<span class="emp-job-badge">🏢 Hybrid</span>';
    else workModeBadge = '<span class="emp-job-badge">📍 On-site</span>';
    
    // Get employment type badge
    const typeMap = {
      'full-time': '💼 Full-time',
      'part-time': '⏰ Part-time',
      'contract': '📄 Contract',
      'internship': '🎓 Internship'
    };
    const typeBadge = `<span class="emp-job-badge">${typeMap[job.employment_type] || job.employment_type}</span>`;
    
    // Get top 3 skills
    const skillsHtml = job.required_skills.slice(0, 3).map(skill => 
      `<span class="emp-job-badge">#${skill}</span>`
    ).join('');
    
    return `
      <div class="emp-job-card" data-id="${job.id}">
        <div class="emp-job-header">
          <div class="emp-job-logo-wrapper">
            ${job.company_logo ? 
              `<img src="${job.company_logo}" alt="${escapeHtml(job.company)}" class="emp-job-logo">` : 
              `<div class="emp-job-logo-placeholder">${escapeHtml(companyInitials)}</div>`
            }
          </div>
          <div class="emp-job-info">
            <h3 class="emp-job-title">${escapeHtml(job.title)}</h3>
            <div class="emp-job-company">${escapeHtml(job.company)}</div>
            <div class="emp-job-badges">
              ${workModeBadge}
              ${typeBadge}
              ${skillsHtml}
            </div>
          </div>
          <div class="emp-job-match">
            <div class="emp-match-score" style="color: ${matchColor}">${job.match_score}%</div>
            <div class="emp-match-label">Match</div>
          </div>
        </div>
        
        <div class="emp-job-description">
          ${escapeHtml(job.description.substring(0, 200))}${job.description.length > 200 ? '...' : ''}
        </div>
        
        <div class="emp-job-footer">
          <div class="emp-job-meta">
            <span class="emp-job-meta-item">📍 ${escapeHtml(job.location)}</span>
            <span class="emp-job-meta-item">📅 ${postedDate}</span>
            ${salaryText ? `<span class="emp-job-meta-item emp-job-salary">💰 ${salaryText}</span>` : ''}
          </div>
          <div class="emp-job-actions">
            <button class="emp-save-btn ${savedClass}" data-id="${job.id}">
              ${job.is_saved ? '★ Saved' : '☆ Save'}
            </button>
            <button class="emp-apply-btn" data-id="${job.id}">Apply Now →</button>
          </div>
        </div>
      </div>
    `;
  }

  // Toggle save job
  async function toggleSaveJob(jobId, btn) {
    const isSaved = btn.classList.contains('saved');
    
    try {
      if (isSaved) {
        await AngaziaAPI.jobs.unsave(jobId);
        savedJobs.delete(jobId);
        btn.classList.remove('saved');
        btn.textContent = '☆ Save';
      } else {
        await AngaziaAPI.jobs.save(jobId);
        savedJobs.add(jobId);
        btn.classList.add('saved');
        btn.textContent = '★ Saved';
      }
      
      // Update in allJobs
      allJobs.forEach(job => {
        if (job.id === jobId) job.is_saved = !isSaved;
      });
      
    } catch (error) {
      console.error('Failed to toggle save:', error);
    }
  }

  // Render selected skills
  function renderSelectedSkills() {
    if (!elements.selectedSkillsContainer) return;
    
    elements.selectedSkillsContainer.innerHTML = Array.from(selectedSkills).map(skill => `
      <span class="emp-skill-tag">
        ${escapeHtml(skill)}
        <span class="emp-skill-remove" data-skill="${escapeHtml(skill)}">×</span>
      </span>
    `).join('');
    
    document.querySelectorAll('.emp-skill-remove').forEach(btn => {
      btn.addEventListener('click', () => {
        const skill = btn.getAttribute('data-skill');
        selectedSkills.delete(skill);
        renderSelectedSkills();
        applyFilters();
      });
    });
  }

  // Pagination
  function updatePagination() {
    totalPages = Math.ceil(filteredJobs.length / itemsPerPage);
    
    if (elements.prevBtn) {
      elements.prevBtn.disabled = currentPage <= 1;
    }
    if (elements.nextBtn) {
      elements.nextBtn.disabled = currentPage >= totalPages;
    }
    
    if (elements.paginationPages && totalPages > 1) {
      elements.paginationPages.innerHTML = '';
      
      const startPage = Math.max(1, currentPage - 2);
      const endPage = Math.min(totalPages, currentPage + 2);
      
      for (let i = startPage; i <= endPage; i++) {
        const btn = document.createElement('button');
        btn.textContent = i;
        btn.className = `emp-page-btn ${i === currentPage ? 'active' : ''}`;
        btn.addEventListener('click', () => {
          currentPage = i;
          renderJobs();
          updatePagination();
          window.scrollTo({ top: 0, behavior: 'smooth' });
        });
        elements.paginationPages.appendChild(btn);
      }
    }
    
    if (elements.pagination) {
      elements.pagination.style.display = totalPages > 1 ? 'flex' : 'none';
    }
  }

  // UI State Management
  function showLoading(show) {
    if (elements.loading) elements.loading.style.display = show ? 'flex' : 'none';
    if (elements.resultsHeader && !show) elements.resultsHeader.style.display = 'flex';
    if (elements.jobsGrid && !show) elements.jobsGrid.style.display = 'flex';
  }

  function showError(show, message = '') {
    if (elements.error) elements.error.style.display = show ? 'flex' : 'none';
    if (elements.errorMsg && message) elements.errorMsg.textContent = message;
    if (elements.resultsHeader && show) elements.resultsHeader.style.display = 'none';
    if (elements.jobsGrid && show) elements.jobsGrid.style.display = 'none';
  }

  // Helper Functions
  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
  }

  function getMatchColor(score) {
    if (score >= 80) return '#10b981';
    if (score >= 60) return '#f59e0b';
    return '#ef4444';
  }

  function formatSalary(min, max, currency) {
    if (!min && !max) return '';
    const formatNumber = (num) => {
      if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
      if (num >= 1000) return (num / 1000).toFixed(0) + 'K';
      return num;
    };
    if (min && max) return `${currency} ${formatNumber(min)} - ${formatNumber(max)}`;
    if (min) return `${currency} ${formatNumber(min)}+`;
    if (max) return `Up to ${currency} ${formatNumber(max)}`;
    return '';
  }

  function formatRelativeDate(dateStr) {
    if (!dateStr) return 'Recently';
    const date = new Date(dateStr);
    const now = new Date();
    const diffDays = Math.floor((now - date) / 86400000);
    
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
    if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
    return `${Math.floor(diffDays / 365)} years ago`;
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
    } else {
      console.log(`[${type}] ${message}`);
    }
  }

  // Navigation
  if (elements.prevBtn) {
    elements.prevBtn.addEventListener('click', () => {
      if (currentPage > 1) {
        currentPage--;
        renderJobs();
        updatePagination();
        window.scrollTo({ top: 0, behavior: 'smooth' });
      }
    });
  }
  
  if (elements.nextBtn) {
    elements.nextBtn.addEventListener('click', () => {
      if (currentPage < totalPages) {
        currentPage++;
        renderJobs();
        updatePagination();
        window.scrollTo({ top: 0, behavior: 'smooth' });
      }
    });
  }

  // Start
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();