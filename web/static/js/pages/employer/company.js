/**
 * Employer Company Profile View
 * Handles loading and displaying the company profile (read-only)
 */

(function() {
  'use strict';

  let companyData = null;

  // Add styles
  function addStyles() {
    if (!document.querySelector('#emp-company-styles')) {
      const style = document.createElement('style');
      style.id = 'emp-company-styles';
      style.textContent = `
        @keyframes slideIn {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
        .emp-loading, .emp-error {
          text-align: center;
          padding: 60px 20px;
          background: var(--s1);
          border: 1px solid var(--border);
          border-radius: 12px;
        }
        .emp-spinner {
          width: 40px;
          height: 40px;
          border: 3px solid var(--border);
          border-top-color: var(--purple);
          border-radius: 50%;
          animation: spin 1s linear infinite;
          margin: 0 auto 16px;
        }
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `;
      document.head.appendChild(style);
    }
  }

  // Helper functions
  function getInitials(name) {
    if (!name) return '??';
    return name.split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  }

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

  // ========== VIEW PAGE FUNCTIONS ==========

  async function loadCompanyProfile() {
    const loadingEl = document.getElementById('company-loading');
    const contentEl = document.getElementById('company-content');
    const errorEl = document.getElementById('company-error');
    
    if (loadingEl) loadingEl.style.display = 'block';
    if (contentEl) contentEl.style.display = 'none';
    if (errorEl) errorEl.style.display = 'none';
    
    try {
      const response = await AngaziaAPI.companies.myCompany();
      
      let profile = null;
      let stats = null;
      let verification = null;
      
      if (response && response.data) {
        profile = response.data.profile;
        stats = response.data.stats;
        verification = response.data.verification;
      } else if (response && response.profile) {
        profile = response.profile;
        stats = response.stats;
        verification = response.verification;
      } else {
        profile = response;
      }
      
      if (!profile) {
        throw new Error('Failed to load company data');
      }
      
      companyData = { profile, stats, verification };
      renderCompanyProfile(profile, stats, verification);
      
      if (loadingEl) loadingEl.style.display = 'none';
      if (contentEl) contentEl.style.display = 'block';
      
    } catch (error) {
      console.error('Failed to load company:', error);
      if (loadingEl) loadingEl.style.display = 'none';
      if (errorEl) {
        const errorMsg = document.getElementById('company-error-msg');
        if (errorMsg) errorMsg.textContent = error.message || 'Failed to load company profile';
        errorEl.style.display = 'block';
      }
    }
  }

  function renderCompanyProfile(profile, stats, verification) {
    // Company logo
    const logoImg = document.getElementById('company-logo-img');
    const logoInitials = document.getElementById('company-logo-initials');
    
    var logoUrl = profile.logo || profile.company_logo;
    if (logoImg && logoUrl) {
      logoImg.src = logoUrl;
      logoImg.style.display = 'block';
      if (logoInitials) logoInitials.style.display = 'none';
    } else if (logoInitials) {
      logoInitials.textContent = getInitials(profile.company_name);
      logoInitials.style.display = 'flex';
      if (logoImg) logoImg.style.display = 'none';
    }
    
    // Company name
    const nameEl = document.getElementById('company-name');
    if (nameEl) nameEl.textContent = profile.company_name || 'Unnamed Company';
    
    // Verification badge
    const badgeEl = document.getElementById('verification-badge');
    if (badgeEl) {
      const isVerified = profile.verification_status === 'verified' || 
                        (verification && verification.status === 'approved');
      badgeEl.className = `emp-badge ${isVerified ? 'emp-badge-verified' : 'emp-badge-pending'}`;
      badgeEl.innerHTML = isVerified ? '✅ Verified Employer' : '⏳ Verification Pending';
    }
    
    // Stats
    const jobsCount = document.getElementById('stat-jobs');
    const hiresCount = document.getElementById('stat-hires');
    const activeCount = document.getElementById('stat-active');
    
    if (jobsCount) jobsCount.textContent = stats?.total_jobs || profile.total_jobs_posted || 0;
    if (hiresCount) hiresCount.textContent = stats?.total_hires || profile.total_hires || 0;
    if (activeCount) activeCount.textContent = stats?.active_jobs || 0;
    
    // Details
    const industry = document.getElementById('detail-industry');
    const size = document.getElementById('detail-size');
    const location = document.getElementById('detail-location');
    const website = document.getElementById('detail-website');
    const linkedin = document.getElementById('detail-linkedin');
    const description = document.getElementById('detail-description');
    
    if (industry) industry.textContent = profile.industry || 'Not specified';
    if (size) size.textContent = profile.company_size || 'Not specified';
    if (location) location.textContent = profile.location || 'Not specified';
    if (website) {
      if (profile.company_website) {
        website.innerHTML = `<a href="${profile.company_website}" class="emp-link" target="_blank" rel="noopener">${profile.company_website}</a>`;
      } else {
        website.innerHTML = '<span class="emp-na">Not set</span>';
      }
    }
    if (linkedin) {
      if (profile.company_linkedin) {
        linkedin.innerHTML = `<a href="${profile.company_linkedin}" class="emp-link" target="_blank" rel="noopener">${profile.company_linkedin}</a>`;
      } else {
        linkedin.innerHTML = '<span class="emp-na">Not set</span>';
      }
    }
    if (description) description.textContent = profile.company_description || 'No description provided.';
    
    // Verification details
    const businessReg = document.getElementById('detail-business-reg');
    const taxId = document.getElementById('detail-tax-id');
    if (businessReg) businessReg.textContent = verification?.business_registration_number || '-';
    if (taxId) taxId.textContent = verification?.tax_id || '-';

    var phoneEl = document.getElementById('detail-phone');
    var emailEl = document.getElementById('detail-email');
    if (phoneEl) phoneEl.textContent = profile.phone_number || '-';
    if (emailEl) emailEl.textContent = profile.contact_email || '-';
    
    // Show/hide verification request button
    const verifyBtn = document.getElementById('request-verify');
    if (verifyBtn) {
      const isVerified = profile.verification_status === 'verified' || 
                        (verification && verification.status === 'approved');
      if (isVerified) {
        verifyBtn.style.display = 'none';
        const verifyHint = document.querySelector('.emp-verify-hint');
        if (verifyHint) verifyHint.style.display = 'none';
      } else {
        verifyBtn.style.display = 'inline-flex';
      }
    }
  }

  function initVerificationRequest() {
    const verifyBtn = document.getElementById('request-verify');
    if (!verifyBtn) return;
    
    verifyBtn.addEventListener('click', async function() {
      const originalText = verifyBtn.textContent;
      verifyBtn.disabled = true;
      verifyBtn.textContent = 'Requesting...';
      
      try {
        await AngaziaAPI.companies.submitVerification();
        showToast('Verification request submitted successfully!', 'success');
        verifyBtn.textContent = 'Verification Requested';
        verifyBtn.disabled = true;
        
        setTimeout(() => {
          loadCompanyProfile();
        }, 2000);
        
      } catch (error) {
        console.error('Verification request failed:', error);
        showToast(error.message || 'Failed to submit verification request', 'error');
        verifyBtn.disabled = false;
        verifyBtn.textContent = originalText;
      }
    });
  }

  // ========== PAGE INITIALIZATION ==========

  function initViewPage() {
    addStyles();
    loadCompanyProfile();
    initVerificationRequest();
  }

  // Initialize the company view page
  function init() {
    initViewPage();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();