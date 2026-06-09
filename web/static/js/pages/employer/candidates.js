/**
 * Employer Candidates Page
 * Handles searching, filtering, and displaying candidates
 */

(function() {
  'use strict';

  let currentPage = 1;
  let totalPages = 1;
  let isLoading = false;

  // DOM Elements
  const elements = {
    loading: document.getElementById('candidates-loading'),
    error: document.getElementById('candidates-error'),
    errorMsg: document.getElementById('candidates-error-msg'),
    content: document.getElementById('candidates-content'),
    candidatesList: document.getElementById('candidates-list'),
    resultsTitle: document.getElementById('results-title'),
    resultHint: document.getElementById('result-hint'),
    pagination: document.getElementById('pagination'),
    pageInfo: document.getElementById('page-info'),
    prevBtn: document.getElementById('prev-page'),
    nextBtn: document.getElementById('next-page'),
    searchInput: document.getElementById('candidate-search'),
    searchBtn: document.getElementById('search-btn'),
    skillFilter: document.getElementById('skill-filter'),
    locationFilter: document.getElementById('location-filter'),
    expFilter: document.getElementById('exp-filter'),
    availabilityFilter: document.getElementById('availability-filter'),
    clearFiltersBtn: document.getElementById('clear-filters')
  };

  // Current filter state
  let currentFilters = {
    search: '',
    skill: '',
    location: '',
    experience: '',
    availability: '',
    page: 1,
    limit: 20
  };

  // Add styles
  function addStyles() {
    if (!document.querySelector('#emp-candidates-styles')) {
      const style = document.createElement('style');
      style.id = 'emp-candidates-styles';
      style.textContent = `
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
        @keyframes slideIn {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
        .emp-candidate-card {
          cursor: pointer;
        }
        .emp-candidate-card:hover {
          transform: translateY(-2px);
        }
        .emp-github-badge {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          background: rgba(88, 166, 255, 0.1);
          color: #58a6ff;
          padding: 2px 8px;
          border-radius: 4px;
          font-size: 10px;
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

  // Get initials from name
  function getInitials(name) {
    if (!name) return '??';
    return name.split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  }

  // Load candidates from API
  async function loadCandidates() {
    if (isLoading) return;
    isLoading = true;
    
    showLoading();
    
    try {
      // Build query params
      const params = {
        page: currentPage,
        limit: currentFilters.limit,
        q: currentFilters.search,
        skills: currentFilters.skill,
        location: currentFilters.location,
        experience_level: currentFilters.experience,
      };
      
      // Remove empty params
      Object.keys(params).forEach(key => {
        if (!params[key]) delete params[key];
      });
      
      const response = await AngaziaAPI.search.candidates(params);
      
      let candidates = [];
      let total = 0;
      
      // Handle the response format: { success: true, data: { results: [...], total, page, limit, total_pages } }
      if (response && response.data) {
        // Extract results from the nested data object
        if (response.data.results && Array.isArray(response.data.results)) {
          candidates = response.data.results;
          total = response.data.total || 0;
          totalPages = response.data.total_pages || 1;
          currentPage = response.data.page || 1;
        } else if (Array.isArray(response.data)) {
          candidates = response.data;
          total = candidates.length;
        }
      } else if (response && response.results && Array.isArray(response.results)) {
        candidates = response.results;
        total = response.total || 0;
      } else if (Array.isArray(response)) {
        candidates = response;
        total = response.length;
      }
      
      hideLoading();
      
      if (!candidates || candidates.length === 0) {
        showEmpty();
        return;
      }
      
      renderCandidates(candidates);
      updatePagination(total);
      showContent();
      
    } catch (error) {
      console.error('Failed to load candidates:', error);
      hideLoading();
      showError(error.message || 'Failed to load candidates');
    } finally {
      isLoading = false;
    }
  }

  // Render candidates list
  function renderCandidates(candidates) {
    if (!elements.candidatesList) return;
    
    elements.candidatesList.innerHTML = candidates.map(candidate => {
      // Extract candidate data from the result structure
      const candidateData = candidate.data || candidate;
      const candidateId = candidateData.user_id || candidateData.id;
      const fullName = candidateData.full_name || candidateData.name || 'Anonymous';
      const headline = candidateData.headline || candidate.title || 'Tech Professional';
      const skills = candidateData.skills || [];
      const location = candidateData.location || 'Location not specified';
      const yearsExp = candidateData.years_of_experience || 0;
      const matchScore = candidate.score || candidateData.match_score || 0;
      const githubConnected = candidateData.github_connected || false;
      
      return `
        <div class="emp-candidate-card" data-candidate-id="${candidateId}" onclick="window.location.href='/employer/candidates/${candidateId}'">
          <div class="emp-candidate-avatar">
            <span class="emp-candidate-initials">${getInitials(fullName)}</span>
          </div>
          <div class="emp-candidate-info">
            <div class="emp-candidate-top">
              <h3 class="emp-candidate-name">${escapeHtml(fullName)}</h3>
              <span class="emp-candidate-score">${matchScore}% match</span>
            </div>
            <span class="emp-candidate-headline">${escapeHtml(headline)}</span>
            <div class="emp-candidate-tags">
              ${skills.slice(0, 5).map(skill => `<span class="emp-tag">${escapeHtml(skill)}</span>`).join('')}
              ${skills.length > 5 ? `<span class="emp-tag">+${skills.length - 5}</span>` : ''}
              ${githubConnected ? `<span class="emp-github-badge">🐙 GitHub Connected</span>` : ''}
            </div>
            <div class="emp-candidate-meta">
              <span>📍 ${escapeHtml(location)}</span>
              <span>💼 ${yearsExp} yr${yearsExp !== 1 ? 's' : ''} exp</span>
            </div>
          </div>
          <div class="emp-candidate-actions">
            <button class="emp-btn emp-btn-sm emp-btn-outline save-candidate-btn" data-id="${candidateId}" onclick="event.stopPropagation()">❤️ Save</button>
          </div>
        </div>
      `;
    }).join('');
    
    // Add event listeners to save buttons (re-attach because innerHTML was replaced)
    document.querySelectorAll('.save-candidate-btn').forEach(btn => {
      btn.removeEventListener('click', handleSaveClick);
      btn.addEventListener('click', handleSaveClick);
    });
    
    // Update results title
    if (elements.resultsTitle) {
      const count = candidates.length;
      elements.resultsTitle.textContent = `${count} candidate${count !== 1 ? 's' : ''} found`;
    }
  }

  // Handle save button click
  async function handleSaveClick(e) {
    e.stopPropagation();
    const btn = e.currentTarget;
    const candidateId = btn.getAttribute('data-id');
    await saveCandidate(candidateId, btn);
  }

  // Save candidate to talent pool
  async function saveCandidate(candidateId, btn) {
    try {
      // Check if candidate is already in a pool
      var existingPools = await AngaziaAPI.candidates.pools(candidateId);
      if (existingPools && existingPools.length > 0) {
        if (btn) {
          btn.textContent = '✅ Saved';
          btn.disabled = true;
          btn.style.opacity = '0.6';
        }
        showToast('Candidate is already in your talent pool.', 'success');
        return;
      }

      // Get or create a default talent pool
      var resp = await AngaziaAPI.talentPools.list();
      var pools = resp && resp.pools ? resp.pools : (Array.isArray(resp) ? resp : []);
      
      // Find existing "Saved Candidates" pool by name to avoid creating duplicates
      var defaultPool = null;
      for (var i = 0; i < pools.length; i++) {
        var poolName = pools[i].name || pools[i].Name || '';
        if (poolName === 'Saved Candidates') {
          defaultPool = pools[i];
          break;
        }
      }
      // Fall back to first pool if no named match found
      if (!defaultPool && pools.length > 0) {
        defaultPool = pools[0];
      }
      
      var defaultPoolId = defaultPool ? (defaultPool.id || defaultPool.ID) : null;
      if (!defaultPoolId) {
        // Create a default pool if none exists
        var newPool = await AngaziaAPI.talentPools.create({
          name: 'Saved Candidates',
          description: 'Auto-generated pool for saved candidates'
        });
        defaultPoolId = newPool.id || newPool.ID;
      }
      
      if (defaultPoolId) {
        await AngaziaAPI.talentPools.addCandidate(defaultPoolId, {
          employee_id: candidateId,
          notes: 'Saved from candidate search',
          match_score: 0
        });
        
        showToast('Candidate saved to talent pool!', 'success');
        
        if (btn) {
          btn.textContent = '✅ Saved';
          btn.disabled = true;
          btn.style.opacity = '0.6';
        }
      }
    } catch (error) {
      console.error('Failed to save candidate:', error);
      showToast(error.body && error.body.error ? error.body.error : (error.message || 'Failed to save candidate'), 'error');
    }
  }

  // Update pagination controls
  function updatePagination(total) {
    if (!elements.pagination) return;
    
    const limit = currentFilters.limit;
    
    if (totalPages <= 1) {
      elements.pagination.style.display = 'none';
      return;
    }
    
    elements.pagination.style.display = 'flex';
    
    if (elements.pageInfo) {
      elements.pageInfo.textContent = `Page ${currentPage} of ${totalPages}`;
    }
    
    if (elements.prevBtn) {
      elements.prevBtn.disabled = currentPage === 1;
    }
    
    if (elements.nextBtn) {
      elements.nextBtn.disabled = currentPage === totalPages;
    }
  }

  // Previous page
  function prevPage() {
    if (currentPage > 1) {
      currentPage--;
      currentFilters.page = currentPage;
      loadCandidates();
    }
  }

  // Next page
  function nextPage() {
    if (currentPage < totalPages) {
      currentPage++;
      currentFilters.page = currentPage;
      loadCandidates();
    }
  }

  // Search function
  function searchCandidates() {
    currentPage = 1;
    currentFilters.search = elements.searchInput?.value.trim() || '';
    currentFilters.page = currentPage;
    loadCandidates();
  }

  // Apply filters
  function applyFilters() {
    currentPage = 1;
    currentFilters.skill = elements.skillFilter?.value || '';
    currentFilters.location = elements.locationFilter?.value || '';
    currentFilters.experience = elements.expFilter?.value || '';
    currentFilters.availability = elements.availabilityFilter?.value || '';
    currentFilters.page = currentPage;
    loadCandidates();
  }

  // Clear all filters
  function clearFilters() {
    if (elements.searchInput) elements.searchInput.value = '';
    if (elements.skillFilter) elements.skillFilter.value = '';
    if (elements.locationFilter) elements.locationFilter.value = '';
    if (elements.expFilter) elements.expFilter.value = '';
    if (elements.availabilityFilter) elements.availabilityFilter.value = '';
    
    currentFilters = {
      search: '',
      skill: '',
      location: '',
      experience: '',
      availability: '',
      page: 1,
      limit: 20
    };
    currentPage = 1;
    loadCandidates();
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
    if (elements.candidatesList) {
      elements.candidatesList.innerHTML = `
        <div class="emp-empty">
          <div class="emp-empty-icon">🔍</div>
          <h3 class="emp-empty-title">No candidates found</h3>
          <p class="emp-empty-desc">Try adjusting your filters or broadening your search criteria.</p>
        </div>
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

  // Escape HTML
  function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(text));
    return div.innerHTML;
  }

  // Debounce search input
  function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }

  // Initialize event listeners
  function initEventListeners() {
    if (elements.searchBtn) {
      elements.searchBtn.addEventListener('click', searchCandidates);
    }
    
    if (elements.searchInput) {
      elements.searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
          searchCandidates();
        }
      });
    }
    
    if (elements.skillFilter) {
      elements.skillFilter.addEventListener('change', applyFilters);
    }
    
    if (elements.locationFilter) {
      elements.locationFilter.addEventListener('change', applyFilters);
    }
    
    if (elements.expFilter) {
      elements.expFilter.addEventListener('change', applyFilters);
    }
    
    if (elements.availabilityFilter) {
      elements.availabilityFilter.addEventListener('change', applyFilters);
    }
    
    if (elements.clearFiltersBtn) {
      elements.clearFiltersBtn.addEventListener('click', clearFilters);
    }
    
    if (elements.prevBtn) {
      elements.prevBtn.addEventListener('click', prevPage);
    }
    
    if (elements.nextBtn) {
      elements.nextBtn.addEventListener('click', nextPage);
    }
  }

  // Initialize the page
  function init() {
    addStyles();
    initEventListeners();
    loadCandidates();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();