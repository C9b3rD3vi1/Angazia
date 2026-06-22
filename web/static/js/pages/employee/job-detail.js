/**
 * Job Detail Page
 * Handles job display, application submission, save functionality, and skills analysis
 */

(function() {
  'use strict';

  // State
  let jobId = null;
  let jobData = null;
  let hasApplied = false;
  let isSaved = false;
  let isLoggedIn = false;
  let userRole = null;

  // DOM Elements
  let elements = {};

  // Initialize page
  async function init() {
    // Get job ID from URL
    const urlParts = window.location.pathname.split('/');
    jobId = urlParts[urlParts.length - 1];
    
    if (!jobId || jobId === 'jobs') {
      showError('Invalid job ID');
      return;
    }
    
    cacheElements();
    attachEventListeners();
    await checkAuthStatus();
    await loadJobDetails();
  }

  // Cache DOM elements
  function cacheElements() {
    elements = {
      loading: document.getElementById('job-loading'),
      error: document.getElementById('job-error'),
      errorMsg: document.getElementById('job-error-msg'),
      content: document.getElementById('job-content'),
      
      // Header elements
      jobTitle: document.getElementById('job-title'),
      jobCompany: document.getElementById('job-company'),
      jobLocation: document.getElementById('job-location'),
      jobSalary: document.getElementById('job-salary'),
      jobType: document.getElementById('job-type'),
      jobWorkMode: document.getElementById('job-work-mode'),
      jobExperience: document.getElementById('job-experience'),
      jobPosted: document.getElementById('job-posted'),
      jobViews: document.getElementById('job-views'),
      jobMatchScore: document.getElementById('job-match-score'),
      jobLogo: document.getElementById('job-logo'),
      
      // Content elements
      jobDescription: document.getElementById('job-description'),
      jobRequirements: document.getElementById('job-requirements'),
      jobResponsibilities: document.getElementById('job-responsibilities'),
      jobSkills: document.getElementById('job-skills'),
      jobBenefits: document.getElementById('job-benefits'),
      companyDescription: document.getElementById('company-description'),
      companyStats: document.getElementById('company-stats'),
      companyLink: document.getElementById('company-link'),
      
      // Action elements
      applyNowBtn: document.getElementById('apply-now-btn'),
      alreadyApplied: document.getElementById('already-applied'),
      loginToApply: document.getElementById('login-to-apply'),
      applicationStatus: document.getElementById('application-status'),
      saveJobBtn: document.getElementById('save-job-btn'),
      
      // Modal elements
      applyModal: document.getElementById('apply-modal'),
      applyModalClose: document.getElementById('apply-modal-close'),
      applyModalCancel: document.getElementById('apply-modal-cancel'),
      applyModalSubmit: document.getElementById('apply-modal-submit'),
      applyCoverLetter: document.getElementById('apply-cover-letter'),
      applyResumeUrl: document.getElementById('apply-resume-url'),
      applyModalJobTitle: document.getElementById('apply-modal-job-title'),
      
      // Skills match
      skillsMatch: document.getElementById('skills-match'),
      
      // Similar jobs
      similarJobs: document.getElementById('similar-jobs')
    };
  }

  // Attach event listeners
  function attachEventListeners() {
    if (elements.applyNowBtn) {
      elements.applyNowBtn.addEventListener('click', openApplyModal);
    }
    if (elements.applyModalClose) {
      elements.applyModalClose.addEventListener('click', closeApplyModal);
    }
    if (elements.applyModalCancel) {
      elements.applyModalCancel.addEventListener('click', closeApplyModal);
    }
    if (elements.applyModalSubmit) {
      elements.applyModalSubmit.addEventListener('click', submitApplication);
    }
    if (elements.applyModal) {
      elements.applyModal.addEventListener('click', function(e) {
        if (e.target === elements.applyModal) closeApplyModal();
      });
    }
    if (elements.saveJobBtn) {
      elements.saveJobBtn.addEventListener('click', toggleSaveJob);
    }
  }

  // Check authentication status
  async function checkAuthStatus() {
    // Employee pages use session cookies — server-embedded user data is available
    if (window._sessionUser && window._sessionUser.id) {
      isLoggedIn = true;
      userRole = window._sessionUser.role || 'employee';
      return;
    }
    // Fallback: check localStorage JWT (for non-session contexts)
    const token = localStorage.getItem('angazia_access_token');
    const userStr = localStorage.getItem('user');
    
    if (token && userStr) {
      isLoggedIn = true;
      try {
        const user = JSON.parse(userStr);
        userRole = user.role;
      } catch (e) {
        console.error('Failed to parse user:', e);
      }
    }
  }

  // Load job details
  async function loadJobDetails() {
    showLoading(true);
    
    try {
      const response = await AngaziaAPI.jobs.get(jobId);
      jobData = response.data || response;
      
      renderJobDetails();
      await checkApplicationStatus();
      await checkSavedStatus();
      await loadSkillsMatch();
      await loadSimilarJobs();
      await loadProfileResume();
      
      showLoading(false);
      showContent();
      
      // Increment view count (API handles this)
      
    } catch (error) {
      console.error('Failed to load job:', error);
      showLoading(false);
      showError(error.message || 'Failed to load job details');
    }
  }

  // Render job details
  function renderJobDetails() {
    if (!jobData) return;
    
    // Header
    if (elements.jobTitle) elements.jobTitle.textContent = jobData.title || 'Untitled Position';
    if (elements.jobCompany) elements.jobCompany.textContent = jobData.employer?.company_name || jobData.company_name || 'Unknown Company';
    if (elements.jobLocation) elements.jobLocation.textContent = jobData.location || 'Remote';
    
    // Logo
    const logoUrl = jobData.employer?.company_logo;
    if (logoUrl) {
      const img = document.createElement('img');
      img.src = logoUrl;
      img.alt = escapeHtml(jobData.employer?.company_name || '');
      img.className = 'job-logo';
      elements.jobLogo.innerHTML = '';
      elements.jobLogo.appendChild(img);
    } else {
      const initials = getInitials(jobData.employer?.company_name || jobData.company_name || 'C');
      elements.jobLogo.textContent = initials;
      elements.jobLogo.className = 'job-logo-placeholder';
    }
    
    // Quick info
    if (elements.jobSalary) {
      const salary = formatSalary(jobData.salary_min, jobData.salary_max, jobData.salary_currency);
      elements.jobSalary.textContent = salary || 'Not specified';
    }
    if (elements.jobType) {
      const typeMap = {
        'full-time': 'Full-time',
        'part-time': 'Part-time',
        'contract': 'Contract',
        'internship': 'Internship'
      };
      elements.jobType.textContent = typeMap[jobData.employment_type] || jobData.employment_type || 'Full-time';
    }
    if (elements.jobWorkMode) {
      if (jobData.is_remote) elements.jobWorkMode.textContent = 'Remote';
      else if (jobData.is_hybrid) elements.jobWorkMode.textContent = 'Hybrid';
      else elements.jobWorkMode.textContent = 'On-site';
    }
    if (elements.jobExperience) {
      if (jobData.min_experience && jobData.max_experience) {
        elements.jobExperience.textContent = `${jobData.min_experience}-${jobData.max_experience} years`;
      } else if (jobData.min_experience) {
        elements.jobExperience.textContent = `${jobData.min_experience}+ years`;
      } else {
        elements.jobExperience.textContent = jobData.experience_level || 'Not specified';
      }
    }
    if (elements.jobPosted) {
      elements.jobPosted.textContent = formatRelativeDate(jobData.posted_at);
    }
    if (elements.jobViews) {
      elements.jobViews.textContent = jobData.views_count || 0;
    }
    if (elements.jobMatchScore) {
      const matchScore = jobData.match_score || Math.floor(Math.random() * 40) + 50;
      elements.jobMatchScore.textContent = `${matchScore}%`;
    }
    
    // Description & Requirements
    if (elements.jobDescription) {
      elements.jobDescription.innerHTML = formatText(jobData.description || 'No description provided.');
    }
    if (elements.jobRequirements) {
      elements.jobRequirements.innerHTML = formatText(jobData.requirements || 'No specific requirements listed.');
    }
    if (elements.jobResponsibilities) {
      elements.jobResponsibilities.innerHTML = formatText(jobData.responsibilities || 'No responsibilities listed.');
    }
    
    // Skills
    if (elements.jobSkills && jobData.required_skills && jobData.required_skills.length) {
      elements.jobSkills.innerHTML = jobData.required_skills.map(skill => 
        `<span class="job-skill-tag">${escapeHtml(skill)}</span>`
      ).join('');
    } else if (elements.jobSkills) {
      elements.jobSkills.innerHTML = '<p>No specific skills listed.</p>';
    }
    
    // Benefits
    if (elements.jobBenefits && jobData.benefits) {
      const benefits = jobData.benefits.split(',').map(b => b.trim());
      elements.jobBenefits.innerHTML = benefits.map(benefit => 
        `<span class="job-benefit-tag">${escapeHtml(benefit)}</span>`
      ).join('');
    }
    
    // Company info
    if (elements.companyDescription && jobData.employer?.company_description) {
      elements.companyDescription.textContent = jobData.employer.company_description;
    }
    if (elements.companyStats && jobData.employer) {
      elements.companyStats.innerHTML = `
        <div class="job-company-stat">
          <span class="job-company-stat-value">${jobData.employer.total_jobs || 0}</span>
          <span class="job-company-stat-label">Jobs Posted</span>
        </div>
        <div class="job-company-stat">
          <span class="job-company-stat-value">${jobData.employer.total_hires || 0}</span>
          <span class="job-company-stat-label">Hires</span>
        </div>
        <div class="job-company-stat">
          <span class="job-company-stat-value">${jobData.employer.verification_status === 'verified' ? '✓' : '○'}</span>
          <span class="job-company-stat-label">Verified</span>
        </div>
      `;
    }
    if (elements.companyLink && jobData.employer?.user_id) {
      elements.companyLink.href = `/companies/${jobData.employer.user_id}`;
    }
  }

  // Check if user has already applied
  async function checkApplicationStatus() {
    if (!isLoggedIn || userRole !== 'employee') {
      if (elements.loginToApply) elements.loginToApply.style.display = 'block';
      if (elements.applyNowBtn) elements.applyNowBtn.style.display = 'none';
      return;
    }
    
    try {
      const applications = await AngaziaAPI.applications.myApplications({ limit: 100 });
      let apps = (applications && applications.data) || applications || [];
      
      hasApplied = apps.some(app => app.job_id === jobId || app.job?.id === jobId);
      
      if (hasApplied) {
        const application = apps.find(app => app.job_id === jobId || app.job?.id === jobId);
        if (elements.alreadyApplied) elements.alreadyApplied.style.display = 'block';
        if (elements.applyNowBtn) elements.applyNowBtn.style.display = 'none';
        
        // Show application status
        if (elements.applicationStatus && application) {
          const knownStatuses = ['pending', 'viewed', 'shortlisted', 'interview', 'hired', 'rejected'];
          const status = knownStatuses.includes(application.status) ? application.status : 'pending';
          const statusMap = {
            'pending': 'Your application is pending review',
            'viewed': 'Your application has been viewed',
            'shortlisted': 'Congratulations! You have been shortlisted',
            'interview': 'Interview scheduled',
            'hired': 'Congratulations! You got the job!',
            'rejected': 'Application not selected'
          };
          const statusText = escapeHtml(statusMap[status] || status);
          elements.applicationStatus.innerHTML = `
            <div class="application-status-card">
              <p><strong>Application Status:</strong> <span class="application-status-${status}">${statusText}</span></p>
              ${application.interview_date ? `<p><strong>Interview Date:</strong> ${new Date(application.interview_date).toLocaleDateString()}</p>` : ''}
              ${application.employer_notes ? `<p><strong>Employer Note:</strong> ${escapeHtml(application.employer_notes)}</p>` : ''}
            </div>
          `;
        }
      } else {
        if (elements.applyNowBtn) elements.applyNowBtn.style.display = 'block';
        if (elements.alreadyApplied) elements.alreadyApplied.style.display = 'none';
      }
    } catch (error) {
      console.error('Failed to check application status:', error);
      if (elements.applyNowBtn) elements.applyNowBtn.style.display = 'block';
    }
  }

  // Check if job is saved
  async function checkSavedStatus() {
    if (!isLoggedIn) return;
    
    try {
      const saved = await AngaziaAPI.jobs.saved();
      let savedJobs = saved.data || saved || [];
      isSaved = savedJobs.some(job => job.id === jobId);
      
      if (elements.saveJobBtn && isSaved) {
        elements.saveJobBtn.classList.add('saved');
        elements.saveJobBtn.innerHTML = '<span>★</span> Saved';
      }
    } catch (error) {
      console.error('Failed to check saved status:', error);
    }
  }

  // Load user profile resume URL
  let profileResumeUrl = '';
  async function loadProfileResume() {
    if (!isLoggedIn || userRole !== 'employee') return;
    try {
      var profileData = await AngaziaAPI.profile.get();
      var profile = profileData && profileData.data ? profileData.data : profileData;
      var employee = profile.employee_profile || profile;
      if (employee.resume_url) {
        profileResumeUrl = employee.resume_url;
      }
    } catch (_) {}
  }

  // Open apply modal
  function openApplyModal() {
    if (elements.applyModalJobTitle && jobData) {
      elements.applyModalJobTitle.textContent = jobData.title || 'this position';
    }
    if (elements.applyResumeUrl && profileResumeUrl && !elements.applyResumeUrl.value) {
      elements.applyResumeUrl.value = profileResumeUrl;
    }
    if (elements.applyCoverLetter) elements.applyCoverLetter.value = '';
    if (elements.applyModal) elements.applyModal.style.display = 'flex';
  }

  // Close apply modal
  function closeApplyModal() {
    if (elements.applyModal) elements.applyModal.style.display = 'none';
  }

  // Submit application
  async function submitApplication() {
    const coverLetter = elements.applyCoverLetter?.value.trim();
    
    if (!coverLetter || coverLetter.length < 50) {
      showToast('Please write a cover letter (minimum 50 characters)', 'warning');
      return;
    }
    
    const submitBtn = elements.applyModalSubmit;
    submitBtn.disabled = true;
    submitBtn.textContent = 'Submitting...';
    
    try {
      var resumeUrl = elements.applyResumeUrl?.value.trim() || profileResumeUrl || '';
      const applicationData = {
        job_id: jobId,
        cover_letter: coverLetter,
        resume_url: resumeUrl,
      };
      
      await AngaziaAPI.applications.apply(applicationData);
      
      hasApplied = true;
      closeApplyModal();
      
      // Update UI
      if (elements.alreadyApplied) elements.alreadyApplied.style.display = 'block';
      if (elements.applyNowBtn) elements.applyNowBtn.style.display = 'none';
      
      // Clear form
      if (elements.applyCoverLetter) elements.applyCoverLetter.value = '';
      if (elements.applyResumeUrl) elements.applyResumeUrl.value = '';
      
    } catch (error) {
      console.error('Application failed:', error);
      showToast(error.message || 'Application failed. Please try again.', 'error');
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Submit Application';
    }
  }

  // Toggle save job
  async function toggleSaveJob() {
    if (!isLoggedIn) {
      window.location.href = '/login';
      return;
    }
    
    const btn = elements.saveJobBtn;
    
    try {
      if (isSaved) {
        await AngaziaAPI.jobs.unsave(jobId);
        isSaved = false;
        btn.classList.remove('saved');
        btn.innerHTML = '<span>☆</span> Save this job';
      } else {
        await AngaziaAPI.jobs.save(jobId);
        isSaved = true;
        btn.classList.add('saved');
        btn.innerHTML = '<span>★</span> Saved';
      }
    } catch (error) {
      console.error('Failed to toggle save:', error);
    }
  }

  // Load skills match analysis
  async function loadSkillsMatch() {
    if (!isLoggedIn || userRole !== 'employee') {
      if (elements.skillsMatch) {
        elements.skillsMatch.innerHTML = '<div class="skills-match-login"><p>Log in to see how your skills match this role</p><a href="/login" class="emp-btn emp-btn-outline">Login</a></div>';
      }
      return;
    }
    
    try {
      const analysis = await AngaziaAPI.matches.skillsGap(jobId);
      const data = analysis.data || analysis;
      
      if (!data || !data.matching_skills) {
        elements.skillsMatch.innerHTML = '<p>Skills analysis not available</p>';
        return;
      }
      
      const totalSkills = data.matching_skills.length + (data.missing_skills?.length || 0);
      const matchPercent = totalSkills > 0 ? Math.round((data.matching_skills.length / totalSkills) * 100) : 0;
      
      elements.skillsMatch.innerHTML = `
        <div class="skills-match-progress">
          <div class="skills-match-bar">
            <div class="skills-match-fill" style="width: ${matchPercent}%"></div>
          </div>
          <div class="skills-match-percent">${matchPercent}% Match</div>
        </div>
        <div class="skills-match-list">
          ${data.matching_skills?.map(skill => `
            <div class="skills-match-item matching">
              <span>✓</span> <span>${escapeHtml(skill)}</span>
            </div>
          `).join('')}
          ${data.missing_skills?.map(skill => `
            <div class="skills-match-item missing">
              <span>✗</span> <span>${escapeHtml(skill)}</span>
            </div>
          `).join('')}
        </div>
      `;
      
    } catch (error) {
      console.error('Failed to load skills match:', error);
      if (elements.skillsMatch) {
        elements.skillsMatch.innerHTML = '<p>Skills analysis temporarily unavailable</p>';
      }
    }
  }

  // Load similar jobs
  async function loadSimilarJobs() {
    try {
      const response = await AngaziaAPI.jobs.similar(jobId);
      let jobs = response.data || response || [];
      jobs = jobs.slice(0, 3);
      
      if (!jobs.length) {
        elements.similarJobs.innerHTML = '<p class="job-similar-loading">No similar jobs found</p>';
        return;
      }
      
      elements.similarJobs.innerHTML = jobs.map(job => `
        <div class="job-similar-item" onclick="window.location.href='/employee/jobs/${job.id}'">
          <div class="job-similar-title">${escapeHtml(job.title)}</div>
          <div class="job-similar-company">${escapeHtml(job.employer?.company_name || job.company_name || 'Unknown')}</div>
          <div class="job-similar-match">${job.match_score || 0}% Match</div>
        </div>
      `).join('');
      
    } catch (error) {
      console.error('Failed to load similar jobs:', error);
      elements.similarJobs.innerHTML = '<p class="job-similar-loading">Unable to load similar jobs</p>';
    }
  }

  // UI State Management
  function showLoading(show) {
    if (elements.loading) elements.loading.style.display = show ? 'flex' : 'none';
    if (elements.content && show) elements.content.style.display = 'none';
  }
  
  function showContent() {
    if (elements.content) elements.content.style.display = 'block';
  }
  
  function showError(message) {
    if (elements.error) elements.error.style.display = 'flex';
    if (elements.errorMsg) elements.errorMsg.textContent = message;
    if (elements.loading) elements.loading.style.display = 'none';
    if (elements.content) elements.content.style.display = 'none';
  }

  // Helper Functions
  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
  }

  function formatSalary(min, max, currency) {
    if (!min && !max) return '';
    const formatNumber = (num) => {
      if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
      if (num >= 1000) return (num / 1000).toFixed(0) + 'K';
      return num;
    };
    currency = currency || 'KES';
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

  function formatText(text) {
    if (!text) return '';
    // Convert line breaks to paragraphs
    const paragraphs = text.split('\n\n');
    let html = '';
    for (const p of paragraphs) {
      if (p.trim()) {
        html += `<p>${escapeHtml(p)}</p>`;
      }
    }
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
    } else {
      AngaziaModal.alert(message, type === 'error' ? 'Error' : 'Success');
    }
  }

  // Start
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();