(function () {
  var appId = window.location.pathname.split('/').pop();
  if (!appId) return;

  var els = {};
  function init() {
    els = {
      loading: document.getElementById('ad-loading'),
      error: document.getElementById('ad-error'),
      content: document.getElementById('ad-content'),
      errorMsg: document.getElementById('ad-error-msg'),
      name: document.getElementById('ad-candidate-name'),
      email: document.getElementById('ad-candidate-email'),
      sub: document.getElementById('ad-sub'),
      initials: document.getElementById('ad-initials'),
      status: document.getElementById('ad-status'),
      matchScore: document.getElementById('ad-match-score'),
      skillMatch: document.getElementById('ad-skill-match'),
      expMatch: document.getElementById('ad-exp-match'),
      appliedDate: document.getElementById('ad-applied-date'),
      jobTitle: document.getElementById('ad-job-title'),
      jobLocation: document.getElementById('ad-job-location'),
      jobType: document.getElementById('ad-job-type'),
      jobPosted: document.getElementById('ad-job-posted'),
      coverLetter: document.getElementById('ad-cover-letter'),
      coverLetterCard: document.getElementById('ad-cover-letter-card'),
      aiInsights: document.getElementById('ad-ai-insights'),
      aiInsightsCard: document.getElementById('ad-ai-insights-card'),
      notes: document.getElementById('ad-notes'),
      shortlistBtn: document.getElementById('ad-shortlist-btn'),
      interviewBtn: document.getElementById('ad-interview-btn'),
      rejectBtn: document.getElementById('ad-reject-btn'),
      hireBtn: document.getElementById('ad-hire-btn'),
    };
    loadApplication();
    bindActions();
  }

  function showError(msg) {
    els.loading.style.display = 'none';
    els.error.style.display = 'block';
    els.errorMsg.textContent = msg || 'Unknown error';
  }

  function loadApplication() {
    els.loading.style.display = 'block';
    AngaziaAPI.get('/applications/' + appId).then(function (app) {
      els.loading.style.display = 'none';
      if (!app) { showError('Application not found'); return; }
      renderApplication(app);
    }).catch(function (err) {
      showError(err && err.message ? err.message : 'Failed to load application');
    });
  }

  function renderApplication(app) {
    els.content.style.display = 'block';
    var emp = app.employee || {};
    var job = app.job || {};
    var name = emp.full_name || emp.name || emp.FullName || 'Unknown';
    var email = emp.email || emp.Email || '';
    var initial = name ? name.charAt(0).toUpperCase() : '?';

    els.name.textContent = name;
    els.email.textContent = email;
    els.sub.textContent = 'Applied for ' + (job.title || 'Unknown Position');
    els.initials.textContent = initial;
    els.status.textContent = (app.status || 'pending').toUpperCase();
    els.status.className = 'emp-status-badge ' + (app.status || 'pending');

    els.matchScore.textContent = (app.match_score || 0) + '%';
    els.skillMatch.textContent = (app.skill_match || 0) + '%';
    els.expMatch.textContent = (app.experience_match || 0) + '%';
    els.appliedDate.textContent = app.applied_at ? new Date(app.applied_at).toLocaleDateString() : '--';

    els.jobTitle.textContent = job.title || '--';
    els.jobLocation.textContent = job.location || job.Location || '--';
    els.jobType.textContent = job.type || job.Type || '--';
    els.jobPosted.textContent = job.created_at ? new Date(job.created_at).toLocaleDateString() : '--';

    if (app.cover_letter) {
      els.coverLetterCard.style.display = 'block';
      els.coverLetter.textContent = app.cover_letter;
    }

    if (app.ai_insights) {
      els.aiInsightsCard.style.display = 'block';
      els.aiInsights.textContent = app.ai_insights;
    }

    if (app.employer_notes) {
      els.notes.value = app.employer_notes;
    }

    updateActionButtons(app.status);
  }

  function updateActionButtons(status) {
    var btns = [els.shortlistBtn, els.interviewBtn, els.rejectBtn, els.hireBtn];
    btns.forEach(function (b) { b.disabled = false; b.style.opacity = '1'; });
    if (status === 'hired' || status === 'rejected' || status === 'withdrawn') {
      btns.forEach(function (b) { b.disabled = true; b.style.opacity = '0.5'; });
    }
  }

  function performAction(action) {
    var endpoint, data;
    if (action === 'interview') {
      document.getElementById('ad-interview-modal').style.display = 'flex';
      return;
    }
    switch (action) {
      case 'shortlist': endpoint = '/employer/applications/' + appId + '/shortlist'; break;
      case 'reject': endpoint = '/employer/applications/' + appId + '/reject'; break;
      case 'hire': endpoint = '/employer/applications/' + appId + '/hire'; break;
    }
    AngaziaAPI.post(endpoint, data || {}).then(function () {
      loadApplication();
    }).catch(function () {});
  }

  function scheduleInterview() {
    var dt = document.getElementById('ad-interview-datetime').value;
    var type = document.getElementById('ad-interview-type').value;
    var notes = document.getElementById('ad-interview-notes').value;
    if (!dt) return;
    AngaziaAPI.post('/employer/applications/' + appId + '/interview', {
      interview_date: new Date(dt).toISOString(),
      interview_type: type,
      interview_notes: notes,
    }).then(function () {
      document.getElementById('ad-interview-modal').style.display = 'none';
      loadApplication();
    }).catch(function () {});
  }

  function bindActions() {
    els.shortlistBtn.addEventListener('click', function () { performAction('shortlist'); });
    els.interviewBtn.addEventListener('click', function () { performAction('interview'); });
    els.rejectBtn.addEventListener('click', function () { performAction('reject'); });
    els.hireBtn.addEventListener('click', function () { performAction('hire'); });

    document.getElementById('ad-interview-modal-close').addEventListener('click', function () {
      document.getElementById('ad-interview-modal').style.display = 'none';
    });
    document.getElementById('ad-interview-cancel').addEventListener('click', function () {
      document.getElementById('ad-interview-modal').style.display = 'none';
    });
    document.getElementById('ad-interview-confirm').addEventListener('click', scheduleInterview);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
