(function () {
  'use strict';

  var jobId = null;

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
    fieldSkills: qs('field-skills'),
    fieldExperience: qs('field-experience'),
    fieldDeadline: qs('field-deadline'),
    submitBtn: null
  };

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

  function toast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
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

    var salaryMin = fd.get('salary_min');
    if (salaryMin) data.salary_min = parseInt(salaryMin, 10);
    var salaryMax = fd.get('salary_max');
    if (salaryMax) data.salary_max = parseInt(salaryMax, 10);

    var workType = fd.get('work_type');
    data.is_remote = workType === 'remote';
    data.is_hybrid = workType === 'hybrid';

    var deadline = fd.get('deadline');
    if (deadline) data.expires_at = deadline;

    var btn = els.formEl.querySelector('button[type="submit"]');
    els.submitBtn = btn;
    btn.disabled = true;
    btn.textContent = 'Saving...';

    AngaziaAPI.jobs.update(jobId, data).then(function () {
      toast('Job updated successfully!', 'success');
      setTimeout(function () { window.location.href = '/employer/jobs/' + jobId; }, 1200);
    }).catch(function (err) {
      toast(err.message || 'Failed to update job', 'error');
      btn.disabled = false;
      btn.textContent = 'Save Changes';
    });
  }

  function initRequirementsPreview() {
    var reqField = els.fieldRequirements;
    var reqPreview = document.getElementById('requirements-preview');
    if (!reqField || !reqPreview) return;
    reqField.addEventListener('input', function () {
      var lines = reqField.value.split('\n').filter(function(l) { return l.trim(); });
      if (lines.length === 0) {
        reqPreview.classList.remove('has-items');
        reqPreview.innerHTML = '';
        return;
      }
      reqPreview.classList.add('has-items');
      var html = '<ul>';
      for (var i = 0; i < lines.length; i++) {
        var text = lines[i].trim();
        if (text) html += '<li>' + escapeHtml(text) + '</li>';
      }
      html += '</ul>';
      reqPreview.innerHTML = html;
    });
    function escapeHtml(s) {
      var d = document.createElement('div');
      d.appendChild(document.createTextNode(s));
      return d.innerHTML;
    }
  }

  function init() {
    jobId = getJobId();
    if (!jobId || jobId === 'job-edit') {
      window.location.href = '/employer/jobs';
      return;
    }
    if (els.formEl) els.formEl.addEventListener('submit', handleSubmit);
    initRequirementsPreview();
    loadJob();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
