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
    clearFiltersBtn: document.getElementById('clear-filters'),

    poolModal: document.getElementById('cand-pool-modal'),
    poolTitle: document.getElementById('cand-pool-title'),
    poolHeading: document.getElementById('cand-pool-heading'),
    poolDesc: document.getElementById('cand-pool-desc'),
    poolSelect: document.getElementById('cand-pool-select'),
    poolNewSection: document.getElementById('cand-new-pool-section'),
    poolNewName: document.getElementById('cand-new-pool-name'),
    poolCreateToggle: document.getElementById('cand-create-toggle'),
    poolSave: document.getElementById('cand-pool-save'),
    poolSaveLabel: document.getElementById('cand-pool-save').querySelector('.emp-btn-label'),
    poolCancel: document.getElementById('cand-pool-cancel'),
    poolClose: document.getElementById('cand-pool-close'),
    poolError: document.getElementById('cand-pool-error'),
    poolErrorMsg: document.getElementById('cand-pool-error-msg'),
  };

  let pendingSaveCandidateId = null;
  let pendingSaveCandidateName = '';
  let cachedPools = [];
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
        .ce-evidence {
          margin: 8px 0 6px;
          padding: 8px 10px;
          background: rgba(88, 166, 255, 0.04);
          border: 1px solid rgba(88, 166, 255, 0.12);
          border-radius: 6px;
          font-size: 11px;
        }
        .ce-evidence-head {
          display: flex;
          align-items: center;
          gap: 4px;
          flex-wrap: wrap;
          color: var(--muted);
        }
        .ce-evidence-icon { font-size: 13px; }
        .ce-evidence-user { font-weight: 600; color: #58a6ff; }
        .ce-evidence-dot { color: var(--border); font-size: 8px; }
        .ce-evidence-stat { color: var(--text); }
        .ce-evidence-streak { color: #f59e0b; }
        .ce-evidence-langs {
          display: flex;
          gap: 4px;
          flex-wrap: wrap;
          margin: 4px 0 2px;
        }
        .ce-evidence-lang {
          font-size: 9px;
          padding: 1px 7px;
          background: rgba(139, 92, 246, 0.08);
          color: var(--purple);
          border-radius: 3px;
          letter-spacing: 0.3px;
        }
        .ce-evidence-meta {
          display: flex;
          align-items: center;
          gap: 4px;
          flex-wrap: wrap;
          font-size: 10px;
          color: var(--muted2);
          margin-top: 2px;
        }
        .ce-evidence-score { font-weight: 500; }
        .ce-evidence-badge {
          padding: 1px 6px;
          border-radius: 3px;
          font-size: 9px;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          background: rgba(0, 229, 160, 0.1);
          color: var(--accent);
        }
        .ce-evidence-synced { font-size: 9px; color: var(--muted2); }
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

  function fmtNum(n) {
    if (!n) return '0';
    if (n >= 1000) return (n / 1000).toFixed(n >= 10000 ? 0 : 1).replace(/\.0$/, '') + 'K';
    return String(n);
  }

  function fmtDate(d) {
    if (!d) return '';
    var dt = new Date(d);
    var now = new Date();
    var diff = Math.floor((now - dt) / 86400000);
    if (diff === 0) return 'today';
    if (diff === 1) return 'yesterday';
    if (diff < 30) return diff + 'd ago';
    if (diff < 365) return Math.floor(diff / 30) + 'mo ago';
    return Math.floor(diff / 365) + 'yr ago';
  }

  function ghCategory(score) {
    if (score >= 80) return 'Exceptional';
    if (score >= 60) return 'Active Contributor';
    if (score >= 40) return 'Casual Developer';
    return 'Getting Started';
  }

  function renderEvidence(gh) {
    if (!gh) return '';
    var langs = gh.repo_languages || {};
    var topLangs = Object.keys(langs).sort(function (a, b) { return (langs[b] || 0) - (langs[a] || 0); }).slice(0, 4);
    var category = ghCategory(gh.overall_score);
    return '<div class="ce-evidence">' +
      '<div class="ce-evidence-head">' +
        '<span class="ce-evidence-icon">&#x1F43E;</span>' +
        '<span class="ce-evidence-user">' + escapeHtml(gh.github_username) + '</span>' +
        '<span class="ce-evidence-dot">&middot;</span>' +
        '<span class="ce-evidence-stat">' + (gh.public_repos || 0) + ' repos</span>' +
        '<span class="ce-evidence-dot">&middot;</span>' +
        '<span class="ce-evidence-stat">' + fmtNum(gh.total_commits) + ' commits</span>' +
        (gh.contribution_streak ? '<span class="ce-evidence-dot">&middot;</span><span class="ce-evidence-stat ce-evidence-streak">&#x1F525; ' + gh.contribution_streak + ' day streak</span>' : '') +
      '</div>' +
      (topLangs.length ? '<div class="ce-evidence-langs">' + topLangs.map(function (l) { return '<span class="ce-evidence-lang">' + escapeHtml(l) + '</span>'; }).join('') + '</div>' : '') +
      '<div class="ce-evidence-meta">' +
        '<span class="ce-evidence-score" style="color:' + (gh.activity_score >= 60 ? 'var(--accent)' : 'var(--warning)') + '">Activity: ' + (gh.activity_score || 0) + '/100</span>' +
        '<span class="ce-evidence-dot">&middot;</span>' +
        '<span class="ce-evidence-score">Quality: ' + (gh.quality_score || 0) + '/100</span>' +
        '<span class="ce-evidence-dot">&middot;</span>' +
        '<span class="ce-evidence-badge">' + category + '</span>' +
        (gh.last_synced_at ? '<span class="ce-evidence-dot">&middot;</span><span class="ce-evidence-synced">Synced ' + fmtDate(gh.last_synced_at) + '</span>' : '') +
      '</div>' +
    '</div>';
  }

  // Render candidates list
  function renderCandidates(candidates) {
    if (!elements.candidatesList) return;
    
    elements.candidatesList.innerHTML = candidates.map(function (candidate) {
      var candidateData = candidate.data || candidate;
      var candidateId = candidateData.user_id || candidateData.id;
      var fullName = candidateData.full_name || candidateData.name || 'Anonymous';
      var headline = candidateData.headline || candidate.title || 'Tech Professional';
      var skills = candidateData.skills || [];
      var location = candidateData.location || 'Location not specified';
      var yearsExp = candidateData.years_of_experience || 0;
      var matchScore = candidate.score || candidateData.match_score || 0;
      var githubConnected = candidateData.github_connected || false;
      var avatarUrl = (candidateData.user && candidateData.user.avatar_url) || '';
      var gh = candidateData.github_profile || null;
      
      return '<div class="emp-candidate-card" data-candidate-id="' + candidateId + '" onclick="window.location.href=\'/employer/candidates/' + candidateId + '\'">' +
        '<div class="emp-candidate-avatar">' +
          (avatarUrl ? '<img src="' + avatarUrl + '" alt="' + escapeHtml(fullName) + '" style="width:100%;height:100%;object-fit:cover;border-radius:50%">' : '<span class="emp-candidate-initials">' + getInitials(fullName) + '</span>') +
        '</div>' +
        '<div class="emp-candidate-info">' +
          '<div class="emp-candidate-top">' +
            '<h3 class="emp-candidate-name">' + escapeHtml(fullName) + '</h3>' +
            '<span class="emp-candidate-score">' + matchScore + '% match</span>' +
          '</div>' +
          '<span class="emp-candidate-headline">' + escapeHtml(headline) + '</span>' +
          '<div class="emp-candidate-tags">' +
            skills.slice(0, 5).map(function (skill) { return '<span class="emp-tag">' + escapeHtml(skill) + '</span>'; }).join('') +
            (skills.length > 5 ? '<span class="emp-tag">+' + (skills.length - 5) + '</span>' : '') +
          '</div>' +
          renderEvidence(gh) +
          '<div class="emp-candidate-meta">' +
            '<span>&#x1F4CD; ' + escapeHtml(location) + '</span>' +
            '<span>&#x1F4BC; ' + yearsExp + ' yr' + (yearsExp !== 1 ? 's' : '') + ' exp</span>' +
          '</div>' +
        '</div>' +
        '<div class="emp-candidate-actions">' +
          '<button class="emp-btn emp-btn-sm emp-btn-outline msg-candidate-btn" data-id="' + candidateId + '" onclick="event.stopPropagation()">&#x1F4AC; Message</button>' +
          '<button class="emp-btn emp-btn-sm emp-btn-outline save-candidate-btn" data-id="' + candidateId + '" onclick="event.stopPropagation()">&#x2764;&#xFE0F; Save</button>' +
        '</div>' +
      '</div>';
    }).join('');
    
    // Add event listeners to save buttons (re-attach because innerHTML was replaced)
    document.querySelectorAll('.save-candidate-btn').forEach(btn => {
      btn.removeEventListener('click', handleSaveClick);
      btn.addEventListener('click', handleSaveClick);
    });
    document.querySelectorAll('.msg-candidate-btn').forEach(btn => {
      btn.removeEventListener('click', handleMessageClick);
      btn.addEventListener('click', handleMessageClick);
    });
    
    // Update results title
    if (elements.resultsTitle) {
      const count = candidates.length;
      elements.resultsTitle.textContent = `${count} candidate${count !== 1 ? 's' : ''} found`;
    }
  }

  // ── Pool Picker Modal ──

  function setPoolLoading(loading) {
    if (!elements.poolSave) return;
    elements.poolSave.disabled = loading;
    elements.poolSave.classList.toggle('emp-btn-loading', loading);
  }

  function showPoolError(msg) {
    if (elements.poolError) elements.poolError.style.display = msg ? 'block' : 'none';
    if (elements.poolErrorMsg) elements.poolErrorMsg.textContent = msg || '';
  }

  function loadPools() {
    return AngaziaAPI.talentPools.list({ limit: 100 })
      .then(function (resp) {
        cachedPools = resp && resp.pools ? resp.pools : (Array.isArray(resp) ? resp : []);
        if (elements.poolSelect) {
          var html = '<option value="">— Select a pool —</option>';
          cachedPools.forEach(function (p) {
            html += '<option value="' + (p.id || p.ID) + '">' + (p.name || 'Unnamed') + '</option>';
          });
          elements.poolSelect.innerHTML = html;
          elements.poolSelect.value = '';
        }
      })
      .catch(function () {
        cachedPools = [];
      });
  }

  function openPoolPicker(candidateId, candidateName) {
    pendingSaveCandidateId = candidateId;
    pendingSaveCandidateName = candidateName || '';

    if (elements.poolHeading) elements.poolHeading.textContent = 'Save' + (pendingSaveCandidateName ? ': ' + pendingSaveCandidateName : ' to Talent Pool');

    // Reset UI
    if (elements.poolNewSection) elements.poolNewSection.style.display = 'none';
    if (elements.poolNewName) elements.poolNewName.value = '';
    if (elements.poolDesc) elements.poolDesc.textContent = 'Choose a pool or create a new one to save this candidate for future reference.';
    showPoolError('');
    setPoolLoading(false);
    loadPools().then(function () {
      if (elements.poolModal) elements.poolModal.style.display = 'flex';
    });
  }

  function closePoolPicker() {
    if (elements.poolModal) elements.poolModal.style.display = 'none';
    setPoolLoading(false);
    showPoolError('');
    pendingSaveCandidateId = null;
    pendingSaveCandidateName = '';
  }

  function handlePoolSave() {
    var candidateId = pendingSaveCandidateId;
    var candidateName = pendingSaveCandidateName;
    if (!candidateId) return;

    var selectedPoolId = elements.poolSelect ? elements.poolSelect.value : '';
    var newPoolName = elements.poolNewName ? elements.poolNewName.value.trim() : '';

    // Validate
    if (!selectedPoolId && !newPoolName) {
      showPoolError('Please select a pool or enter a name for a new one.');
      return;
    }
    if (newPoolName && newPoolName.length < 2) {
      showPoolError('Pool name must be at least 2 characters.');
      return;
    }

    showPoolError('');
    setPoolLoading(true);

    var poolPromise;
    if (selectedPoolId) {
      poolPromise = Promise.resolve(selectedPoolId);
    } else {
      poolPromise = AngaziaAPI.talentPools.create({ name: newPoolName })
        .then(function (newPool) {
          return newPool.id || newPool.ID || (newPool.data && newPool.data.id);
        });
    }

    poolPromise.then(function (poolId) {
      return AngaziaAPI.talentPools.addCandidate(poolId, {
        employee_id: candidateId,
        notes: 'Saved from candidate search',
        match_score: 0
      }).then(function () {
        closePoolPicker();
        var btn = document.querySelector('.save-candidate-btn[data-id="' + candidateId + '"]');
        if (btn) {
          btn.textContent = '✅ Saved';
          btn.disabled = true;
          btn.style.opacity = '0.6';
        }
        showToast((candidateName || 'Candidate') + ' saved to talent pool!', 'success');
      });
    }).catch(function (err) {
      console.error('Failed to save candidate:', err);
      setPoolLoading(false);
      if (err && err.message) {
        showPoolError(err.message);
      } else {
        showPoolError('Failed to save candidate. Please try again.');
      }
    });
  }

  // Handle message button click
  function handleMessageClick(e) {
    e.stopPropagation();
    var btn = e.currentTarget;
    var candidateId = btn.getAttribute('data-id');
    if (!candidateId) return;
    var card = btn.closest('.emp-candidate-card');
    var nameEl = card ? card.querySelector('.emp-candidate-name') : null;
    var candidateName = nameEl ? nameEl.textContent.trim() : 'Candidate';
    btn.disabled = true;
    btn.textContent = '...';
    AngaziaAPI.messages.create({ recipient_id: candidateId, subject: 'Regarding ' + candidateName, content: '' })
      .then(function (conv) {
        window.location.href = '/employer/messages';
      })
      .catch(function () {
        btn.disabled = false;
        btn.textContent = '\uD83D\uDCAC Message';
        showToast('Failed to open conversation. Please try again.', 'error');
      });
  }

  // Handle save button click
  function handleSaveClick(e) {
    e.stopPropagation();
    var btn = e.currentTarget;
    var candidateId = btn.getAttribute('data-id');
    if (!candidateId) return;
    var nameEl = btn.closest('.emp-candidate-card').querySelector('.emp-candidate-name');
    var candidateName = nameEl ? nameEl.textContent.trim() : '';
    openPoolPicker(candidateId, candidateName);
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

    // Pool picker modal events
    if (elements.poolSave) elements.poolSave.addEventListener('click', handlePoolSave);
    if (elements.poolCancel) elements.poolCancel.addEventListener('click', closePoolPicker);
    if (elements.poolClose) elements.poolClose.addEventListener('click', closePoolPicker);
    if (elements.poolModal) {
      elements.poolModal.addEventListener('click', function (e) {
        if (e.target === elements.poolModal) closePoolPicker();
      });
    }
    if (elements.poolCreateToggle) {
      elements.poolCreateToggle.addEventListener('click', function () {
        if (elements.poolNewSection) elements.poolNewSection.style.display = 'block';
        if (elements.poolNewName) elements.poolNewName.focus();
        if (elements.poolSelect) elements.poolSelect.value = '';
        this.style.display = 'none';
      });
    }
    if (elements.poolSelect) {
      elements.poolSelect.addEventListener('change', function () {
        if (this.value) {
          if (elements.poolNewSection) elements.poolNewSection.style.display = 'none';
          if (elements.poolNewName) elements.poolNewName.value = '';
          if (elements.poolCreateToggle) elements.poolCreateToggle.style.display = 'flex';
        }
        showPoolError('');
      });
    }
    if (elements.poolNewName) {
      elements.poolNewName.addEventListener('input', function () {
        showPoolError('');
      });
    }
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') {
        if (elements.poolModal && elements.poolModal.style.display === 'flex') {
          closePoolPicker();
        }
      }
    });
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