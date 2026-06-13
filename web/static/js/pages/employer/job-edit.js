(function () {
  'use strict';

  var jobId = null;
  var pendingData = null;

  function getJobId() {
    var parts = window.location.pathname.split('/');
    return parts[parts.length - 1];
  }

  function qs(id) { return document.getElementById(id); }

  var els = {
    loading: qs('edit-loading'),
    error: qs('edit-error'),
    errorMsg: qs('edit-error-msg'),
    form: qs('edit-form'),
    formEl: qs('form-job-edit'),
    titleSub: qs('job-title-sub'),
    fieldJobId: qs('field-job-id'),
    fieldTitle: qs('field-title'),
    fieldType: qs('field-type'),
    fieldLocation: qs('field-location'),
    fieldWorkType: qs('field-work-type'),
    fieldSalaryMin: qs('field-salary-min'),
    fieldSalaryMax: qs('field-salary-max'),
    fieldDescription: qs('field-description'),
    fieldRequirements: qs('field-requirements'),
    fieldResponsibilities: qs('field-responsibilities'),
    fieldNiceSkills: qs('field-nice-skills'),
    fieldSkills: qs('field-skills'),
    fieldExperience: qs('field-experience'),
    fieldDeadline: qs('field-deadline'),
    submitBtn: null,
    confirmModal: qs('je-confirm-modal'),
    confirmIcon: qs('je-confirm-icon'),
    confirmHeading: qs('je-confirm-heading'),
    confirmDesc: qs('je-confirm-desc'),
    confirmYes: qs('je-confirm-yes'),
    confirmYesLabel: qs('je-confirm-yes-label'),
    confirmNo: qs('je-confirm-no'),
    confirmClose: qs('je-confirm-close'),
  };

  function setJeLoading(loading) {
    if (!els.confirmYes) return;
    els.confirmYes.disabled = loading;
    els.confirmYes.classList.toggle('emp-btn-loading', loading);
  }

  function showJeModal(title, desc) {
    var data = pendingData;
    var jobTitle = data && data.title ? data.title : 'this job';
    if (els.confirmIcon) {
      els.confirmIcon.className = 'emp-modal-icon icon-info';
    }
    if (els.confirmHeading) els.confirmHeading.textContent = title + (jobTitle ? ': ' + jobTitle : '');
    if (els.confirmDesc) els.confirmDesc.textContent = desc;
    setJeLoading(false);
    if (els.confirmModal) els.confirmModal.style.display = 'flex';
  }

  function hideJeModal() {
    if (els.confirmModal) els.confirmModal.style.display = 'none';
    setJeLoading(false);
    pendingData = null;
  }

  function executeJeSave() {
    if (!pendingData) return;
    setJeLoading(true);
    AngaziaAPI.jobs.update(jobId, pendingData).then(function () {
      hideJeModal();
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Job updated successfully!', 'success');
      } else {
        alert('Job updated successfully!');
      }
      setTimeout(function () { window.location.href = '/employer/jobs/' + jobId; }, 1200);
    }).catch(function (err) {
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast(err.message || 'Failed to update job', 'error');
      } else {
        alert('Error: ' + (err.message || 'Failed to update job'));
      }
      setJeLoading(false);
    });
  }

  if (els.confirmYes) els.confirmYes.addEventListener('click', executeJeSave);
  if (els.confirmNo) els.confirmNo.addEventListener('click', hideJeModal);
  if (els.confirmClose) els.confirmClose.addEventListener('click', hideJeModal);
  if (els.confirmModal) {
    els.confirmModal.addEventListener('click', function (e) {
      if (e.target === els.confirmModal) hideJeModal();
    });
  }
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && els.confirmModal && els.confirmModal.style.display === 'flex') hideJeModal();
  });

  function showLoading() {
    if (els.loading) els.loading.style.display = 'block';
    if (els.error) els.error.style.display = 'none';
    if (els.form) els.form.style.display = 'none';
  }

  function showError(msg) {
    if (els.loading) els.loading.style.display = 'none';
    if (els.error) {
      if (els.errorMsg) els.errorMsg.textContent = msg;
      els.error.style.display = 'block';
    }
  }

  function showForm() {
    if (els.loading) els.loading.style.display = 'none';
    if (els.form) els.form.style.display = 'block';
  }

  function loadJob() {
    showLoading();

    AngaziaAPI.jobs.get(jobId).then(function (resp) {
      var job = resp && resp.data ? resp.data : resp;
      if (!job || !job.id) {
        showError('Job not found');
        return;
      }

      if (els.titleSub) els.titleSub.textContent = 'Editing: ' + (job.title || '');
      if (els.fieldJobId) els.fieldJobId.value = job.id;
      if (els.fieldTitle) els.fieldTitle.value = job.title || '';
      if (els.fieldType) els.fieldType.value = job.employment_type || 'full-time';
      if (els.fieldLocation) els.fieldLocation.value = job.location || '';
      if (els.fieldSalaryMin) els.fieldSalaryMin.value = job.salary_min || '';
      if (els.fieldSalaryMax) els.fieldSalaryMax.value = job.salary_max || '';
      if (els.fieldDescription) els.fieldDescription.value = job.description || '';
      if (els.fieldRequirements) {
        els.fieldRequirements.value = job.requirements || '';
        var evt = new Event('input');
        els.fieldRequirements.dispatchEvent(evt);
      }
      if (els.fieldExperience) els.fieldExperience.value = job.experience_level || 'any';

      if (els.fieldWorkType) {
        if (job.is_remote && job.is_hybrid) els.fieldWorkType.value = 'hybrid';
        else if (job.is_remote) els.fieldWorkType.value = 'remote';
        else els.fieldWorkType.value = 'onsite';
      }

      if (els.fieldSkills && job.required_skills) {
        els.fieldSkills.value = (Array.isArray(job.required_skills) ? job.required_skills : []).join(', ');
      }

      if (els.fieldNiceSkills && job.nice_to_have_skills) {
        els.fieldNiceSkills.value = (Array.isArray(job.nice_to_have_skills) ? job.nice_to_have_skills : []).join(', ');
      }

      if (els.fieldResponsibilities && job.responsibilities) {
        els.fieldResponsibilities.value = job.responsibilities;
      }

      if (els.fieldDeadline && job.expires_at) {
        els.fieldDeadline.value = job.expires_at.split('T')[0];
      }

      showForm();
    }).catch(function (err) {
      showError(err.message || 'Failed to load job');
    });
  }

  function handleSubmit(e) {
    e.preventDefault();
    if (!els.formEl) return;

    var fd = new FormData(els.formEl);
    var data = {
      title: fd.get('title'),
      description: fd.get('description'),
      requirements: fd.get('requirements'),
      location: fd.get('location'),
      employment_type: fd.get('type'),
      experience_level: fd.get('experience_level'),
      salary_currency: 'KES'
    };

    var skills = fd.get('skills');
    if (skills) data.required_skills = skills.split(',').map(function (s) { return s.trim(); }).filter(Boolean);

    var niceSkills = fd.get('nice_to_have_skills');
    if (niceSkills) data.nice_to_have_skills = niceSkills.split(',').map(function (s) { return s.trim(); }).filter(Boolean);

    var responsibilities = fd.get('responsibilities');
    if (responsibilities) data.responsibilities = responsibilities;

    var salaryMin = fd.get('salary_min');
    if (salaryMin) data.salary_min = parseInt(salaryMin, 10);
    var salaryMax = fd.get('salary_max');
    if (salaryMax) data.salary_max = parseInt(salaryMax, 10);

    var workType = fd.get('work_type');
    data.is_remote = workType === 'remote';
    data.is_hybrid = workType === 'hybrid';

    var deadline = fd.get('deadline');
    if (deadline) data.expires_at = deadline;

    pendingData = data;
    showJeModal('Save Changes', 'Changes will be visible immediately to candidates.');
  }

  function init() {
    jobId = getJobId();
    if (!jobId || jobId === 'job-edit') {
      window.location.href = '/employer/jobs';
      return;
    }
    if (els.formEl) els.formEl.addEventListener('submit', handleSubmit);
    loadJob();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
