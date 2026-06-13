(function () {
  var appId = window.location.pathname.split('/').pop();
  if (!appId) return;

  var els = {};
  var pendingAction = null;
  var candidateName = '';

  var actionIcons = {
    shortlist: { icon: '\u2B50', cls: 'icon-info', heading: 'Shortlist Candidate' },
    reject: { icon: '\u2715', cls: 'icon-danger', heading: 'Reject Candidate' },
    hire: { icon: '\u2714', cls: 'icon-success', heading: 'Hire Candidate' },
  };

  var actionConfigs = {
    shortlist: {
      title: 'Shortlist Candidate',
      message: 'They will be moved to shortlisted status and considered for the next stage.',
      endpoint: '/employer/applications/' + appId + '/shortlist',
      btnClass: 'emp-btn-success',
      btnLabel: 'Yes, Shortlist'
    },
    reject: {
      title: 'Reject Candidate',
      message: 'This will move them to rejected status. You can change their status later if needed.',
      endpoint: '/employer/applications/' + appId + '/reject',
      btnClass: 'emp-btn-danger',
      btnLabel: 'Yes, Reject'
    },
    hire: {
      title: 'Hire Candidate',
      message: 'This will finalize their status as hired and close out their application.',
      endpoint: '/employer/applications/' + appId + '/hire',
      btnClass: 'emp-btn-success',
      btnLabel: 'Yes, Hire'
    }
  };

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
      avatarImg: document.getElementById('ad-avatar-img'),
      profileLink: document.getElementById('ad-view-profile'),
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
      notesSave: document.getElementById('ad-notes-save'),
      notesStatus: document.getElementById('ad-notes-status'),
      shortlistBtn: document.getElementById('ad-shortlist-btn'),
      interviewBtn: document.getElementById('ad-interview-btn'),
      rejectBtn: document.getElementById('ad-reject-btn'),
      hireBtn: document.getElementById('ad-hire-btn'),
      confirmModal: document.getElementById('ad-confirm-modal'),
      confirmTitle: document.getElementById('ad-confirm-title'),
      confirmIcon: document.getElementById('ad-confirm-icon'),
      confirmHeading: document.getElementById('ad-confirm-heading'),
      confirmDesc: document.getElementById('ad-confirm-desc'),
      confirmYes: document.getElementById('ad-confirm-yes'),
      confirmYesLabel: document.getElementById('ad-confirm-yes-label'),
      confirmNo: document.getElementById('ad-confirm-no'),
      confirmClose: document.getElementById('ad-confirm-close'),
      interviewModal: document.getElementById('ad-interview-modal'),
      interviewClose: document.getElementById('ad-interview-modal-close'),
      interviewCancel: document.getElementById('ad-interview-cancel'),
      interviewConfirm: document.getElementById('ad-interview-confirm'),
      interviewDatetime: document.getElementById('ad-interview-datetime'),
      interviewType: document.getElementById('ad-interview-type'),
      interviewNotes: document.getElementById('ad-interview-notes'),
    };
    if (els.notesSave) {
      els.notesSave.addEventListener('click', saveNotes);
    }
    loadApplication();
    bindActions();
  }

  function showError(msg) {
    if (els.loading) els.loading.style.display = 'none';
    if (els.content) els.content.style.display = 'none';
    if (els.error) {
      els.error.style.display = 'block';
      if (els.errorMsg) els.errorMsg.textContent = msg || 'Unknown error';
    }
  }

  function loadApplication() {
    if (els.loading) els.loading.style.display = 'block';
    AngaziaAPI.get('/applications/' + appId).then(function (app) {
      if (els.loading) els.loading.style.display = 'none';
      if (!app) { showError('Application not found'); return; }
      try {
        renderApplication(app);
      } catch (e) {
        console.error('renderApplication error:', e);
        showError('Failed to render application details');
      }
    }).catch(function (err) {
      showError(err && err.message ? err.message : 'Failed to load application');
    });
  }

  function renderApplication(app) {
    var emp = app.employee || app.employee_profile || {};
    var job = app.job || {};
    candidateName = emp.full_name || app.candidate_name || app.name || 'Candidate';
    var candidateEmail = (emp.user && emp.user.email) || app.candidate_email || '';
    var avatarUrl = (emp.user && emp.user.avatar_url) || '';
    var candidateUserId = (emp.user && emp.user.id) || '';
    var initials = getInitials(candidateName);

    els.name.textContent = candidateName;
    els.email.textContent = candidateEmail;
    els.sub.textContent = 'Applied for ' + (job.title || 'Unknown Position');
    els.initials.textContent = initials;

    if (avatarUrl && els.avatarImg) {
      els.avatarImg.src = avatarUrl;
      els.avatarImg.style.display = '';
      if (els.initials) els.initials.style.display = 'none';
    } else {
      if (els.avatarImg) els.avatarImg.style.display = 'none';
      if (els.initials) els.initials.style.display = '';
    }

    if (candidateUserId && els.profileLink) {
      els.profileLink.href = '/employer/candidates/' + candidateUserId;
      els.profileLink.style.display = 'inline-block';
    } else if (els.profileLink) {
      els.profileLink.style.display = 'none';
    }
    els.status.textContent = (app.status || 'pending').toUpperCase();
    els.status.className = 'emp-status-badge ' + (app.status || 'pending');

    els.matchScore.textContent = (app.match_score || 0) + '%';
    els.skillMatch.textContent = (app.skill_match || 0) + '%';
    els.expMatch.textContent = (app.experience_match || 0) + '%';
    els.appliedDate.textContent = app.applied_at ? new Date(app.applied_at).toLocaleDateString() : '--';

    els.jobTitle.textContent = job.title || '--';
    els.jobLocation.textContent = job.location || job.Location || '--';
    els.jobType.textContent = job.employment_type || job.type || job.Type || '--';
    els.jobPosted.textContent = job.posted_at ? new Date(job.posted_at).toLocaleDateString() : job.created_at ? new Date(job.created_at).toLocaleDateString() : '--';

    if (app.cover_letter) {
      els.coverLetterCard.style.display = 'block';
      els.coverLetter.textContent = app.cover_letter;
    } else if (els.coverLetterCard) {
      els.coverLetterCard.style.display = 'none';
    }

    if (app.ai_insights) {
      els.aiInsightsCard.style.display = 'block';
      els.aiInsights.textContent = app.ai_insights;
    } else if (els.aiInsightsCard) {
      els.aiInsightsCard.style.display = 'none';
    }

    if (app.employer_notes && els.notes) {
      els.notes.value = app.employer_notes;
    }

    updateActionButtons(app.status);
    if (els.content) els.content.style.display = 'block';
  }

  function getInitials(name) {
    if (!name) return '??';
    return name.split(' ').map(function (n) { return n[0]; }).join('').toUpperCase().slice(0, 2);
  }

  function saveNotes() {
    var val = els.notes ? els.notes.value : '';
    if (els.notesSave) els.notesSave.disabled = true;
    if (els.notesStatus) els.notesStatus.textContent = 'Saving...';
    AngaziaAPI.post('/employer/applications/' + appId + '/notes', { notes: val }).then(function () {
      if (els.notesStatus) els.notesStatus.textContent = 'Saved';
      if (els.notesSave) els.notesSave.disabled = false;
      setTimeout(function () {
        if (els.notesStatus) els.notesStatus.textContent = '';
      }, 2000);
    }).catch(function () {
      if (els.notesStatus) els.notesStatus.textContent = 'Save failed';
      if (els.notesSave) els.notesSave.disabled = false;
    });
  }

  function updateActionButtons(status) {
    var btns = [els.shortlistBtn, els.interviewBtn, els.rejectBtn, els.hireBtn];
    btns.forEach(function (b) { if (b) { b.disabled = false; b.style.opacity = '1'; } });
    if (status === 'hired' || status === 'rejected' || status === 'withdrawn') {
      btns.forEach(function (b) { if (b) { b.disabled = true; b.style.opacity = '0.5'; } });
    }
  }

  /* ── Confirmation Modal ── */
  function setConfirmLoading(loading) {
    if (!els.confirmYes) return;
    els.confirmYes.disabled = loading;
    els.confirmYes.classList.toggle('emp-btn-loading', loading);
  }

  function showConfirmModal(action) {
    var cfg = actionConfigs[action];
    if (!cfg) return;
    var ico = actionIcons[action] || { icon: '\u26A0', cls: 'icon-warning', heading: cfg.title };
    els.confirmTitle.textContent = cfg.title;
    if (els.confirmIcon) {
      els.confirmIcon.textContent = ico.icon;
      els.confirmIcon.className = 'emp-modal-icon ' + ico.cls;
    }
    if (els.confirmHeading) {
      els.confirmHeading.textContent = candidateName ? ico.heading + ': ' + candidateName : ico.heading;
    }
    if (els.confirmDesc) els.confirmDesc.textContent = cfg.message;
    if (els.confirmYesLabel) els.confirmYesLabel.textContent = cfg.btnLabel;
    els.confirmYes.className = 'emp-btn ' + cfg.btnClass;
    setConfirmLoading(false);
    pendingAction = action;
    els.confirmModal.style.display = 'flex';
  }

  function hideConfirmModal() {
    els.confirmModal.style.display = 'none';
    setConfirmLoading(false);
    pendingAction = null;
  }

  function executePendingAction() {
    if (!pendingAction) return;
    var cfg = actionConfigs[pendingAction];
    if (!cfg) return;
    setConfirmLoading(true);
    AngaziaAPI.post(cfg.endpoint, {}).then(function () {
      hideConfirmModal();
      loadApplication();
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Application updated successfully', 'success');
      }
    }).catch(function (err) {
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast(err && err.message ? err.message : 'Action failed', 'error');
      }
      setConfirmLoading(false);
    });
  }

  /* ── Interview Modal ── */
  function setAdInterviewLoading(loading) {
    if (!els.interviewConfirm) return;
    els.interviewConfirm.disabled = loading;
    els.interviewConfirm.classList.toggle('emp-btn-loading', loading);
  }

  function showInterviewModal() {
    var tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    tomorrow.setHours(10, 0, 0, 0);
    if (els.interviewDatetime) els.interviewDatetime.value = tomorrow.toISOString().slice(0, 16);
    if (els.interviewType) els.interviewType.value = 'technical';
    if (els.interviewNotes) els.interviewNotes.value = '';
    setAdInterviewLoading(false);
    if (els.interviewModal) els.interviewModal.style.display = 'flex';
  }

  function hideInterviewModal() {
    if (els.interviewModal) els.interviewModal.style.display = 'none';
    setAdInterviewLoading(false);
  }

  function scheduleInterview() {
    var dt = els.interviewDatetime ? els.interviewDatetime.value : '';
    var type = els.interviewType ? els.interviewType.value : '';
    var notes = els.interviewNotes ? els.interviewNotes.value : '';
    if (!dt) {
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Please select an interview date and time', 'warning');
      }
      return;
    }
    setAdInterviewLoading(true);
    AngaziaAPI.applications.interview(appId, {
      interview_date: new Date(dt).toISOString(),
      interview_type: type,
      notes: notes
    }).then(function () {
      hideInterviewModal();
      loadApplication();
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Interview scheduled successfully!', 'success');
      }
    }).catch(function (err) {
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast(err && err.message ? err.message : 'Failed to schedule interview', 'error');
      }
      setAdInterviewLoading(false);
    });
  }

  function performAction(action) {
    if (action === 'interview') {
      showInterviewModal();
      return;
    }
    showConfirmModal(action);
  }

  function bindActions() {
    if (els.shortlistBtn) els.shortlistBtn.addEventListener('click', function () { performAction('shortlist'); });
    if (els.interviewBtn) els.interviewBtn.addEventListener('click', function () { performAction('interview'); });
    if (els.rejectBtn) els.rejectBtn.addEventListener('click', function () { performAction('reject'); });
    if (els.hireBtn) els.hireBtn.addEventListener('click', function () { performAction('hire'); });

    /* Interview modal */
    if (els.interviewClose) els.interviewClose.addEventListener('click', hideInterviewModal);
    if (els.interviewCancel) els.interviewCancel.addEventListener('click', hideInterviewModal);
    if (els.interviewConfirm) els.interviewConfirm.addEventListener('click', scheduleInterview);
    if (els.interviewModal) {
      els.interviewModal.addEventListener('click', function (e) {
        if (e.target === els.interviewModal) hideInterviewModal();
      });
    }

    /* Confirm modal */
    if (els.confirmYes) els.confirmYes.addEventListener('click', executePendingAction);
    if (els.confirmNo) els.confirmNo.addEventListener('click', hideConfirmModal);
    if (els.confirmClose) els.confirmClose.addEventListener('click', hideConfirmModal);
    if (els.confirmModal) {
      els.confirmModal.addEventListener('click', function (e) {
        if (e.target === els.confirmModal) hideConfirmModal();
      });
    }

    /* Escape key for all modals */
    document.addEventListener('keydown', function (e) {
      if (e.key !== 'Escape') return;
      if (els.interviewModal && els.interviewModal.style.display === 'flex') hideInterviewModal();
      if (els.confirmModal && els.confirmModal.style.display === 'flex') hideConfirmModal();
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
