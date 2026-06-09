// Employee Dashboard - Complete Functionality

(function() {
  'use strict';

  let currentPage = 1;
  let isLoading = false;

  // Initialize dashboard on load
  document.addEventListener('DOMContentLoaded', function() {
    initDashboard();
    initCharts();
    initAIAssistant();
    initResumeUpload();
    initInfiniteScroll();
  });

  function initDashboard() {
    // Load dashboard data
    loadDashboardData();
    
    // Setup event listeners
    setupEventListeners();
    
    // Start real-time updates
    startRealtimeUpdates();
  }

  async function loadDashboardData() {
    showLoading(true);
    
    try {
      // Load all dashboard data in parallel
      const [stats, recommendations, activity, interviews] = await Promise.all([
        AngaziaAPI.analytics.candidateDashboard(),
        AngaziaAPI.matches.jobMatches({ limit: 5 }),
        AngaziaAPI.analytics.recentActivity(),
        AngaziaAPI.applications.myApplications({ status: 'interview' })
      ]);
      
      updateDashboardUI({ stats, recommendations, activity, interviews });
      
    } catch (error) {
      console.error('Failed to load dashboard:', error);
      AngaziaApp.showToast('Failed to load dashboard data', 'error');
    } finally {
      showLoading(false);
    }
  }

  function updateDashboardUI(data) {
    // Update stats
    if (data.stats) {
      updateStatsCards(data.stats);
    }
    
    // Update recommendations - ensure it's an array
    if (data.recommendations) {
      const recommendations = Array.isArray(data.recommendations) ? data.recommendations : 
                             (data.recommendations.data || data.recommendations.jobs || []);
      updateRecommendations(recommendations);
    }
    
    // Update activity timeline - ensure it's an array
    if (data.activity) {
      const activities = Array.isArray(data.activity) ? data.activity : 
                        (data.activity.data || data.activity.activities || []);
      updateActivityTimeline(activities);
    }
    
    // Update interviews - ensure it's an array
    if (data.interviews) {
      const interviews = Array.isArray(data.interviews) ? data.interviews : 
                         (data.interviews.data || data.interviews.applications || []);
      updateInterviewsList(interviews);
    }
  }

  function updateStatsCards(stats) {
    const statsMap = {
      'profile-views': stats.profile_views || stats.profileViews || 0,
      'applications': stats.total_applications || stats.applications || 0,
      'saved-jobs': stats.saved_jobs || stats.savedJobs || 0,
      'interviews': stats.interview_count || stats.interviews || 0
    };
    
    Object.entries(statsMap).forEach(([key, value]) => {
      const el = document.querySelector(`[data-stat="${key}"]`);
      if (el) el.textContent = value;
    });
  }

  function updateRecommendations(recommendations) {
    const container = document.getElementById('recommendations-list');
    if (!container) return;
    
    if (!recommendations || recommendations.length === 0) {
      container.innerHTML = `
        <div class="emp-empty-state">
          <span class="emp-empty-icon">🔍</span>
          <p>No recommendations yet. Complete your profile to get matched!</p>
          <a href="/employee/profile" class="emp-link">Complete Profile →</a>
        </div>
      `;
      return;
    }
    
    container.innerHTML = recommendations.map(job => `
      <div class="emp-rec-item" data-job-id="${job.id || job.ID}">
        <div class="emp-rec-header">
          ${job.company_logo || job.CompanyLogo ? 
            `<img src="${job.company_logo || job.CompanyLogo}" class="emp-rec-logo">` : 
            `<div class="emp-rec-logo-placeholder">${getInitials(job.company || job.Company)}</div>`
          }
          <div class="emp-rec-info">
            <h4 class="emp-rec-title">${escapeHtml(job.title || job.Title)}</h4>
            <p class="emp-rec-company">${escapeHtml(job.company || job.Company)}</p>
          </div>
          <div class="emp-rec-match-score">
            <svg class="emp-score-ring" viewBox="0 0 36 36">
              <path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#2a3a3a" stroke-width="3"/>
              <path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#00e5a0" stroke-width="3" stroke-dasharray="${job.match_score || job.MatchScore || 0}, 100"/>
            </svg>
            <span class="emp-score-text">${job.match_score || job.MatchScore || 0}%</span>
          </div>
        </div>
        <div class="emp-rec-tags">
          <span class="emp-tag">📍 ${job.location || job.Location || 'Remote'}</span>
          <span class="emp-tag">💼 ${job.employment_type || job.EmploymentType || 'Full-time'}</span>
          ${job.salary || job.Salary ? `<span class="emp-tag emp-tag-salary">💰 ${formatSalary(job.salary || job.Salary)}</span>` : ''}
        </div>
        <div class="emp-rec-skills">
          ${(job.matching_skills || job.MatchingSkills || []).slice(0, 3).map(s => `<span class="emp-skill-match">✓ ${escapeHtml(s)}</span>`).join('')}
          ${(job.missing_skills || job.MissingSkills || []).slice(0, 2).map(s => `<span class="emp-skill-missing">+ ${escapeHtml(s)}</span>`).join('')}
        </div>
        <div class="emp-rec-footer">
          <span class="emp-rec-posted">📅 ${timeAgo(job.posted_at || job.PostedAt || new Date())}</span>
          <button class="emp-btn-sm emp-btn-primary" onclick="window.applyToJob && window.applyToJob('${job.id || job.ID}')">Apply Now →</button>
        </div>
      </div>
    `).join('');
  }

  function updateActivityTimeline(activities) {
    const container = document.querySelector('.emp-timeline');
    if (!container) return;
    
    if (!activities || activities.length === 0) {
      container.innerHTML = `
        <div class="emp-empty-state">
          <span class="emp-empty-icon">📭</span>
          <p>No recent activity</p>
        </div>
      `;
      return;
    }
    
    container.innerHTML = activities.map(activity => `
      <div class="emp-timeline-item">
        <div class="emp-timeline-dot ${activity.type || activity.Type || 'application'}"></div>
        <div class="emp-timeline-content">
          <p class="emp-timeline-text">${escapeHtml(activity.text || activity.Text || activity.message || 'Activity')}</p>
          <span class="emp-timeline-time">${timeAgo(activity.created_at || activity.CreatedAt || activity.timestamp)}</span>
        </div>
      </div>
    `).join('');
  }

  function updateInterviewsList(interviews) {
    const container = document.querySelector('.emp-interviews-list');
    if (!container) return;
    
    // Safety check - ensure interviews is an array
    if (!interviews || !Array.isArray(interviews) || interviews.length === 0) {
      container.innerHTML = `
        <div class="emp-empty-state">
          <span class="emp-empty-icon">🎯</span>
          <p>No upcoming interviews</p>
          <a href="/employee/jobs" class="emp-link">Browse jobs →</a>
        </div>
      `;
      return;
    }
    
    container.innerHTML = interviews.map(interview => {
      const interviewDate = interview.interview_date || interview.InterviewDate || interview.date;
      const date = interviewDate ? new Date(interviewDate) : new Date();
      return `
        <div class="emp-interview-item">
          <div class="emp-interview-date">
            <span class="emp-interview-day">${date.getDate()}</span>
            <span class="emp-interview-month">${date.toLocaleString('default', { month: 'short' })}</span>
          </div>
          <div class="emp-interview-details">
            <h4 class="emp-interview-role">${escapeHtml(interview.job_title || interview.JobTitle || interview.role || 'Interview')}</h4>
            <p class="emp-interview-company">${escapeHtml(interview.company || interview.Company || 'Company')}</p>
            <div class="emp-interview-meta">
              <span>🕐 ${formatTime(interviewDate || interview.date)}</span>
              <span>📞 ${interview.interview_type || interview.InterviewType || 'Virtual'}</span>
            </div>
          </div>
          <div class="emp-interview-actions">
            <button class="emp-btn-icon" onclick="window.confirmInterview && window.confirmInterview('${interview.id || interview.ID}')" title="Confirm">✅</button>
            <button class="emp-btn-icon" onclick="window.rescheduleInterview && window.rescheduleInterview('${interview.id || interview.ID}')" title="Reschedule">📅</button>
            <button class="emp-btn-icon" onclick="window.addToCalendar && window.addToCalendar('${interview.id || interview.ID}')" title="Add to Calendar">📆</button>
          </div>
        </div>
      `;
    }).join('');
  }

  function initCharts() {
    // Application status chart
    const ctx = document.getElementById('app-status-chart');
    if (ctx) {
      // Chart.js or custom rendering
    }
  }

  function initAIAssistant() {
    const modal = document.getElementById('ai-assistant-modal');
    const openBtn = document.getElementById('ai-assistant-btn');
    const closeBtn = modal?.querySelector('[data-close="ai-assistant-modal"]');
    const sendBtn = document.getElementById('chat-send');
    const input = document.getElementById('chat-input');
    const messagesContainer = document.getElementById('chat-messages');
    
    if (openBtn) {
      openBtn.addEventListener('click', () => {
        if (modal) modal.style.display = 'flex';
      });
    }
    
    if (closeBtn) {
      closeBtn.addEventListener('click', () => {
        if (modal) modal.style.display = 'none';
      });
    }
    
    if (sendBtn && input) {
      sendBtn.addEventListener('click', () => sendChatMessage());
      input.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') sendChatMessage();
      });
    }
    
    async function sendChatMessage() {
      const message = input.value.trim();
      if (!message) return;
      
      // Add user message to chat
      addChatMessage(message, 'user');
      input.value = '';
      
      // Show typing indicator
      addTypingIndicator();
      
      try {
        // Send to AI API
        const response = await AngaziaAPI.matches.aiChat({ message });
        removeTypingIndicator();
        addChatMessage(response.reply || response.message || 'Response received', 'bot');
      } catch (error) {
        console.error('AI Chat error:', error);
        removeTypingIndicator();
        addChatMessage('Sorry, I encountered an error. Please try again.', 'bot');
      }
    }
    
    function addChatMessage(message, sender) {
      if (!messagesContainer) return;
      const messageDiv = document.createElement('div');
      messageDiv.className = `emp-chat-message emp-chat-${sender}`;
      messageDiv.innerHTML = `
        <div class="emp-chat-avatar">${sender === 'user' ? '👤' : '🤖'}</div>
        <div class="emp-chat-bubble">${escapeHtml(message)}</div>
      `;
      messagesContainer.appendChild(messageDiv);
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
    
    function addTypingIndicator() {
      if (!messagesContainer) return;
      const indicator = document.createElement('div');
      indicator.id = 'typing-indicator';
      indicator.className = 'emp-chat-message emp-chat-bot';
      indicator.innerHTML = `
        <div class="emp-chat-avatar">🤖</div>
        <div class="emp-chat-bubble">Typing<span class="typing-dots">...</span></div>
      `;
      messagesContainer.appendChild(indicator);
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
    
    function removeTypingIndicator() {
      const indicator = document.getElementById('typing-indicator');
      if (indicator) indicator.remove();
    }
  }

  function initResumeUpload() {
    const uploadArea = document.getElementById('upload-area');
    const fileInput = document.getElementById('resume-file');
    const browseBtn = document.getElementById('browse-btn');
    const modal = document.getElementById('resume-modal');
    const closeBtn = modal?.querySelector('[data-close="resume-modal"]');
    
    if (uploadArea) {
      uploadArea.addEventListener('click', () => fileInput?.click());
      uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('drag-over');
      });
      uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('drag-over');
      });
      uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('drag-over');
        const file = e.dataTransfer.files[0];
        if (file) uploadResume(file);
      });
    }
    
    if (browseBtn) {
      browseBtn.addEventListener('click', () => fileInput?.click());
    }
    
    if (fileInput) {
      fileInput.addEventListener('change', (e) => {
        if (e.target.files?.[0]) uploadResume(e.target.files[0]);
      });
    }
    
    if (closeBtn) {
      closeBtn.addEventListener('click', () => {
        if (modal) modal.style.display = 'none';
      });
    }
    
    async function uploadResume(file) {
      const formData = new FormData();
      formData.append('resume', file);
      
      const progressDiv = document.getElementById('upload-progress');
      const uploadAreaDiv = document.getElementById('upload-area');
      const progressFill = document.getElementById('upload-progress-fill');
      const statusText = document.getElementById('upload-status');
      
      if (uploadAreaDiv) uploadAreaDiv.style.display = 'none';
      if (progressDiv) progressDiv.style.display = 'block';
      
      try {
        // Simulate progress
        let progress = 0;
        const interval = setInterval(() => {
          progress += 10;
          if (progressFill) progressFill.style.width = `${progress}%`;
          if (statusText) statusText.textContent = `Uploading... ${progress}%`;
          if (progress >= 100) clearInterval(interval);
        }, 200);
        
        await AngaziaAPI.profile.uploadResume(formData);
        clearInterval(interval);
        if (progressFill) progressFill.style.width = '100%';
        if (statusText) statusText.textContent = 'Processing resume...';
        
        // Simulate processing
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        AngaziaApp.showToast('Resume uploaded and parsed successfully!', 'success');
        if (modal) modal.style.display = 'none';
        
        // Reload dashboard to show new data
        setTimeout(() => location.reload(), 1500);
        
      } catch (error) {
        console.error('Upload error:', error);
        AngaziaApp.showToast('Failed to upload resume', 'error');
        if (uploadAreaDiv) uploadAreaDiv.style.display = 'block';
        if (progressDiv) progressDiv.style.display = 'none';
      }
    }
  }

  function initInfiniteScroll() {
    window.addEventListener('scroll', () => {
      if (isLoading) return;
      
      const scrollPosition = window.innerHeight + window.scrollY;
      const threshold = document.body.offsetHeight - 500;
      
      if (scrollPosition >= threshold) {
        loadMoreRecommendations();
      }
    });
  }

  async function loadMoreRecommendations() {
    isLoading = true;
    currentPage++;
    
    try {
      const data = await AngaziaAPI.matches.jobMatches({ page: currentPage, limit: 5 });
      const container = document.getElementById('recommendations-list');
      
      if (container && data && data.length) {
        const newHtml = data.map(job => renderJobCard(job)).join('');
        container.insertAdjacentHTML('beforeend', newHtml);
      }
    } catch (error) {
      console.error('Failed to load more recommendations:', error);
    } finally {
      isLoading = false;
    }
  }

  function renderJobCard(job) {
    return `
      <div class="emp-rec-item" data-job-id="${job.id || job.ID}">
        <div class="emp-rec-header">
          <div class="emp-rec-info">
            <h4 class="emp-rec-title">${escapeHtml(job.title || job.Title)}</h4>
            <p class="emp-rec-company">${escapeHtml(job.company || job.Company)}</p>
          </div>
          <div class="emp-rec-match-score">
            <span class="emp-score-text">${job.match_score || job.MatchScore || 0}%</span>
          </div>
        </div>
        <div class="emp-rec-footer">
          <button class="emp-btn-sm emp-btn-primary" onclick="window.applyToJob && window.applyToJob('${job.id || job.ID}')">Apply Now →</button>
        </div>
      </div>
    `;
  }

  function setupEventListeners() {
    // Apply to job buttons
    document.addEventListener('click', (e) => {
      const applyBtn = e.target.closest('[data-action="apply-job"]');
      if (applyBtn) {
        const jobId = applyBtn.getAttribute('data-id');
        if (jobId) window.applyToJob && window.applyToJob(jobId);
      }
    });
  }

  function startRealtimeUpdates() {
    // Refresh data every 30 seconds
    setInterval(() => {
      loadDashboardData();
    }, 30000);
  }

  function showLoading(show) {
    const loader = document.getElementById('dashboard-loader');
    if (loader) loader.style.display = show ? 'flex' : 'none';
  }

  // Helper functions
  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
  }

  function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(String(text)));
    return div.innerHTML;
  }

  function timeAgo(dateStr) {
    if (!dateStr) return 'recently';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);
    
    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins} min ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    return date.toLocaleDateString();
  }

  function formatSalary(salary) {
    const num = typeof salary === 'string' ? parseInt(salary) : salary;
    if (isNaN(num)) return 'N/A';
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(0)}K`;
    return num.toString();
  }

  function formatTime(dateStr) {
    if (!dateStr) return 'TBD';
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return 'TBD';
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  }

  // Expose global functions safely
  window.applyToJob = async function(jobId) {
    if (!jobId) return;
    const confirmed = AngaziaApp && AngaziaApp.confirmDialog ? 
      await AngaziaApp.confirmDialog('Would you like to apply for this position?') : true;
    if (!confirmed) return;
    
    try {
      await AngaziaAPI.applications.apply({ job_id: jobId });
      if (AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Application submitted successfully!', 'success');
      }
    } catch (error) {
      console.error('Apply error:', error);
      if (AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Failed to submit application', 'error');
      }
    }
  };
  
  window.confirmInterview = async function(interviewId) {
    if (!interviewId) return;
    const confirmed = AngaziaApp && AngaziaApp.confirmDialog ? 
      await AngaziaApp.confirmDialog('Confirm your interview attendance?') : true;
    if (!confirmed) return;
    
    try {
      await AngaziaAPI.applications.confirmInterview(interviewId);
      if (AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Interview confirmed!', 'success');
      }
    } catch (error) {
      console.error('Confirm error:', error);
      if (AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Failed to confirm interview', 'error');
      }
    }
  };
  
  window.rescheduleInterview = function(interviewId) {
    if (AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast('Rescheduling feature coming soon', 'info');
    }
  };
  
  window.addToCalendar = function(interviewId) {
    if (AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast('Calendar integration coming soon', 'info');
    }
  };
})();