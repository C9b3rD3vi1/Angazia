/**
 * Employer Company Profile Management
 * Handles viewing, editing, and verifying company profile
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

  // ========== EDIT PAGE FUNCTIONS ==========

  async function loadCompanyForEdit() {
    const loadingEl = document.getElementById('company-loading');
    const contentEl = document.getElementById('company-content');
    const errorEl = document.getElementById('company-error');
    
    if (loadingEl) loadingEl.style.display = 'block';
    if (contentEl) contentEl.style.display = 'none';
    if (errorEl) errorEl.style.display = 'none';
    
    try {
      const response = await AngaziaAPI.companies.myCompany();
      
      let profile = null;
      
      if (response && response.data) {
        profile = response.data.profile;
      } else if (response && response.profile) {
        profile = response.profile;
      } else {
        profile = response;
      }
      
      if (!profile) {
        throw new Error('Failed to load company data');
      }
      
      // Populate the form with existing data
      populateEditForm(profile);
      
      if (loadingEl) loadingEl.style.display = 'none';
      if (contentEl) contentEl.style.display = 'block';
      
    } catch (error) {
      console.error('Failed to load company for edit:', error);
      if (loadingEl) loadingEl.style.display = 'none';
      if (errorEl) {
        const errorMsg = document.getElementById('company-error-msg');
        if (errorMsg) errorMsg.textContent = error.message || 'Failed to load company data';
        errorEl.style.display = 'block';
      }
    }
  }

  function populateEditForm(profile) {
    // Company Name
    const companyNameInput = document.getElementById('company_name');
    if (companyNameInput && profile.company_name) {
      companyNameInput.value = profile.company_name;
    }
    
    // Industry
    const industrySelect = document.getElementById('industry');
    if (industrySelect && profile.industry) {
      industrySelect.value = profile.industry;
    }
    
    // Company Size
    const sizeSelect = document.getElementById('company_size');
    if (sizeSelect && profile.company_size) {
      sizeSelect.value = profile.company_size;
    }
    
    // Location
    const locationInput = document.getElementById('location');
    if (locationInput && profile.location) {
      locationInput.value = profile.location;
    }
    
    // Website
    const websiteInput = document.getElementById('website');
    if (websiteInput && profile.company_website) {
      websiteInput.value = profile.company_website;
    }
    
    // LinkedIn
    const linkedinInput = document.getElementById('linkedin');
    if (linkedinInput && profile.company_linkedin) {
      linkedinInput.value = profile.company_linkedin;
    }
    
    // Description
    const descriptionTextarea = document.getElementById('description');
    if (descriptionTextarea && profile.company_description) {
      descriptionTextarea.value = profile.company_description;
    }
    
    // Contact fields
    var phoneInput = document.getElementById('phone_number');
    var emailInput = document.getElementById('email_address');
    if (phoneInput && profile.phone_number) phoneInput.value = profile.phone_number;
    if (emailInput && profile.contact_email) emailInput.value = profile.contact_email;

    // Verification fields
    var businessRegInput = document.getElementById('business_registration_number');
    var taxIdInput = document.getElementById('tax_id');
    if (businessRegInput && companyData?.verification?.business_registration_number) {
      businessRegInput.value = companyData.verification.business_registration_number;
    }
    if (taxIdInput && companyData?.verification?.tax_id) {
      taxIdInput.value = companyData.verification.tax_id;
    }
    
    // Logo preview
    const previewImg = document.getElementById('logo-preview-img');
    const previewInitials = document.getElementById('logo-preview-initials');
    
    var logoUrl = profile.logo || profile.company_logo;
    if (previewImg && logoUrl) {
      previewImg.src = logoUrl;
      previewImg.style.display = 'block';
      if (previewInitials) previewInitials.style.display = 'none';
    } else if (previewInitials) {
      previewInitials.textContent = getInitials(profile.company_name);
      previewInitials.style.display = 'flex';
    }
  }

  function initLogoUpload() {
    const uploadInput = document.getElementById('logo-upload');
    const previewImg = document.getElementById('logo-preview-img');
    const previewInitials = document.getElementById('logo-preview-initials');
    const uploadBtn = document.getElementById('upload-logo-btn');
    
    if (!uploadInput) return;
    
    if (uploadBtn) {
      uploadBtn.addEventListener('click', function() {
        uploadInput.click();
      });
    }
    
    uploadInput.addEventListener('change', async function(e) {
      const file = e.target.files[0];
      if (!file) return;
      
      if (!file.type.match('image.*')) {
        showToast('Please select an image file (JPEG, PNG, WEBP)', 'error');
        return;
      }
      
      if (file.size > 2 * 1024 * 1024) {
        showToast('Image must be under 2MB', 'error');
        return;
      }
      
      const reader = new FileReader();
      reader.onload = function(e) {
        if (previewImg) {
          previewImg.src = e.target.result;
          previewImg.style.display = 'block';
          if (previewInitials) previewInitials.style.display = 'none';
        }
      };
      reader.readAsDataURL(file);
      
      const formData = new FormData();
      formData.append('logo', file);
      
      const progressBar = document.getElementById('upload-progress');
      if (progressBar) progressBar.style.display = 'block';
      
      try {
        await AngaziaAPI.companies.uploadLogo(formData, function(percent) {
          if (progressBar) {
            progressBar.style.width = percent + '%';
          }
        });
        
        showToast('Logo uploaded successfully!', 'success');
        
        setTimeout(() => {
          loadCompanyForEdit();
        }, 1000);
        
      } catch (error) {
        console.error('Upload failed:', error);
        showToast(error.message || 'Failed to upload logo', 'error');
      } finally {
        if (progressBar) {
          setTimeout(() => {
            progressBar.style.display = 'none';
            progressBar.style.width = '0%';
          }, 1000);
        }
      }
    });
  }

  function initFormSave() {
    const form = document.getElementById('form-company-edit');
    if (!form) return;
    
    form.addEventListener('submit', async function(e) {
      e.preventDefault();
      
      const formData = new FormData(form);
      const data = {
        company_name: formData.get('company_name'),
        company_website: formData.get('website'),
        company_linkedin: formData.get('linkedin'),
        company_description: formData.get('description'),
        industry: formData.get('industry'),
        company_size: formData.get('company_size'),
        location: formData.get('location'),
        phone_number: formData.get('phone_number'),
        email_address: formData.get('email_address'),
        business_registration_number: formData.get('business_registration_number'),
        tax_id: formData.get('tax_id'),
      };
      
      const submitBtn = form.querySelector('button[type="submit"]');
      const originalText = submitBtn.textContent;
      submitBtn.disabled = true;
      submitBtn.textContent = 'Saving...';
      
      try {
        await AngaziaAPI.companies.updateCompany(data);
        showToast('Company profile updated successfully!', 'success');
        
        setTimeout(() => {
          window.location.href = '/employer/company';
        }, 1500);
        
      } catch (error) {
        console.error('Save failed:', error);
        showToast(error.message || 'Failed to save company profile', 'error');
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
      }
    });
  }

  // ========== PAGE INITIALIZATION ==========

  function initViewPage() {
    addStyles();
    loadCompanyProfile();
    initVerificationRequest();
  }

  function initEditPage() {
    addStyles();
    loadCompanyForEdit();
    initLogoUpload();
    initFormSave();
  }

  // Determine which page we're on and initialize accordingly
  function init() {
    const path = window.location.pathname;
    
    if (path.includes('/employer/company-edit')) {
      initEditPage();
    } else if (path.includes('/employer/company')) {
      initViewPage();
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();