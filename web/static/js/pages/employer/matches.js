(function () {
  'use strict';

  var els = {
    loading: document.getElementById('mm-loading'),
    error: document.getElementById('mm-error'),
    content: document.getElementById('mm-content'),
    jobSelector: document.getElementById('job-selector'),
    matchesContainer: document.getElementById('matches-container'),
    matchCount: document.getElementById('match-count'),
    matchesEmpty: document.getElementById('matches-empty'),
    skillSection: document.getElementById('skill-gap-section'),
    skillChart: document.getElementById('skill-gap-chart'),
    skillEmpty: document.getElementById('skill-empty'),
    genQuestionsBtn: document.getElementById('gen-questions-btn'),
    modalOverlay: document.getElementById('mm-modal-overlay'),
    modal: document.getElementById('mm-modal'),
    modalTitle: document.getElementById('mm-modal-title'),
    modalBody: document.getElementById('mm-modal-body'),
    modalLoading: document.getElementById('mm-modal-loading'),
    modalClose: document.getElementById('mm-modal-close'),

    // Pool picker
    poolModal: document.getElementById('mm-pool-modal'),
    poolTitle: document.getElementById('mm-pool-title'),
    poolHeading: document.getElementById('mm-pool-heading'),
    poolDesc: document.getElementById('mm-pool-desc'),
    poolSelect: document.getElementById('mm-pool-select'),
    poolNewSection: document.getElementById('mm-new-pool-section'),
    poolNewName: document.getElementById('mm-new-pool-name'),
    poolCreateToggle: document.getElementById('mm-create-toggle'),
    poolSave: document.getElementById('mm-pool-save'),
    poolCancel: document.getElementById('mm-pool-cancel'),
    poolClose: document.getElementById('mm-pool-close'),
    poolError: document.getElementById('mm-pool-error'),
    poolErrorMsg: document.getElementById('mm-pool-error-msg'),

    // Interview
    interviewModal: document.getElementById('mm-interview-modal'),
    interviewHeading: document.getElementById('mm-interview-heading'),
    interviewDesc: document.getElementById('mm-interview-desc'),
    interviewDate: document.getElementById('mm-interview-date'),
    interviewTime: document.getElementById('mm-interview-time'),
    interviewType: document.getElementById('mm-interview-type'),
    interviewNotes: document.getElementById('mm-interview-notes'),
    interviewConfirm: document.getElementById('mm-interview-confirm'),
    interviewCancel: document.getElementById('mm-interview-cancel'),
    interviewClose: document.getElementById('mm-interview-close'),

    // Feedback
    feedbackModal: document.getElementById('mm-feedback-modal'),
    feedbackHeading: document.getElementById('mm-feedback-heading'),
    feedbackStars: document.getElementById('mm-feedback-stars'),
    feedbackStarLabel: document.getElementById('mm-feedback-star-label'),
    feedbackNotes: document.getElementById('mm-feedback-notes'),
    feedbackSubmit: document.getElementById('mm-feedback-submit'),
    feedbackCancel: document.getElementById('mm-feedback-cancel'),
    feedbackClose: document.getElementById('mm-feedback-close'),
    feedbackError: document.getElementById('mm-feedback-error'),
    feedbackErrorMsg: document.getElementById('mm-feedback-error-msg'),
  };

  var currentJobId = '';
  var currentMatches = [];

  // ── Pool Picker Modal (add to talent pool) ──

  var cachedPools = [];
  var pendingPoolCandidateId = null;
  var pendingPoolCandidateName = '';

  function setPoolLoading(loading) {
    if (!els.poolSave) return;
    els.poolSave.disabled = loading;
    els.poolSave.classList.toggle('emp-btn-loading', loading);
  }

  function showPoolError(msg) {
    if (els.poolError) els.poolError.style.display = msg ? 'block' : 'none';
    if (els.poolErrorMsg) els.poolErrorMsg.textContent = msg || '';
  }

  function loadPools() {
    return AngaziaAPI.talentPools.list({ limit: 100 })
      .then(function (resp) {
        cachedPools = resp && resp.pools ? resp.pools : (Array.isArray(resp) ? resp : []);
        if (els.poolSelect) {
          var html = '<option value="">— Select a pool —</option>';
          cachedPools.forEach(function (p) {
            html += '<option value="' + (p.id || p.ID) + '">' + (p.name || 'Unnamed') + '</option>';
          });
          els.poolSelect.innerHTML = html;
          els.poolSelect.value = '';
        }
      })
      .catch(function () { cachedPools = []; });
  }

  function openPoolPicker(candidateId, candidateName) {
    pendingPoolCandidateId = candidateId;
    pendingPoolCandidateName = candidateName || '';

    if (els.poolHeading) els.poolHeading.textContent = 'Add to Pool' + (pendingPoolCandidateName ? ': ' + pendingPoolCandidateName : '');
    if (els.poolDesc) els.poolDesc.textContent = 'Choose a pool to save this matched candidate for future reference.';

    if (els.poolNewSection) els.poolNewSection.style.display = 'none';
    if (els.poolNewName) els.poolNewName.value = '';
    showPoolError('');
    setPoolLoading(false);
    loadPools().then(function () {
      if (els.poolModal) els.poolModal.style.display = 'flex';
    });
  }

  function closePoolPicker() {
    if (els.poolModal) els.poolModal.style.display = 'none';
    setPoolLoading(false);
    showPoolError('');
    pendingPoolCandidateId = null;
    pendingPoolCandidateName = '';
  }

  function handlePoolSave() {
    var candidateId = pendingPoolCandidateId;
    var candidateName = pendingPoolCandidateName;
    if (!candidateId) return;

    var selectedPoolId = els.poolSelect ? els.poolSelect.value : '';
    var newPoolName = els.poolNewName ? els.poolNewName.value.trim() : '';

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
        .then(function (newPool) { return newPool.id || newPool.ID || (newPool.data && newPool.data.id); });
    }

    poolPromise.then(function (poolId) {
      return AngaziaAPI.talentPools.addCandidate(poolId, {
        employee_id: candidateId,
        job_id: currentJobId,
        notes: 'Added from AI Matches',
        match_score: 0
      }).then(function () {
        closePoolPicker();
        var btn = document.querySelector('[data-candidate-id="' + candidateId + '"]');
        if (btn) {
          var parent = btn.closest('.emp-card-actions, .mm-actions');
          if (parent) {
            var poolBtn = parent.querySelector('[data-action="addPool"]');
            if (poolBtn) { poolBtn.textContent = 'Added'; poolBtn.disabled = true; }
          }
        }
      });
    }).catch(function (err) {
      console.error('Failed to add candidate to pool:', err);
      setPoolLoading(false);
      showPoolError(err && err.message ? err.message : 'Failed to add candidate. Please try again.');
    });
  }

  // ── Interview Modal ──

  var pendingInterviewCandidateId = null;
  var pendingInterviewBtn = null;

  function setInterviewLoading(loading) {
    if (!els.interviewConfirm) return;
    els.interviewConfirm.disabled = loading;
    els.interviewConfirm.classList.toggle('emp-btn-loading', loading);
  }

  function openInterviewModal(candidateId, btn) {
    pendingInterviewCandidateId = candidateId;
    pendingInterviewBtn = btn;
    var name = getCandidateNameFromBtn(btn);
    if (els.interviewHeading) els.interviewHeading.textContent = 'Schedule Interview' + (name ? ': ' + name : '');
    if (els.interviewDesc) els.interviewDesc.textContent = 'Set the date, time, and type of interview for this candidate.';
    if (els.interviewDate) els.interviewDate.value = new Date().toISOString().slice(0, 10);
    if (els.interviewTime) els.interviewTime.value = '';
    if (els.interviewType) els.interviewType.value = 'technical';
    if (els.interviewNotes) els.interviewNotes.value = '';
    setInterviewLoading(false);
    if (els.interviewModal) els.interviewModal.style.display = 'flex';
  }

  function hideInterviewModal() {
    if (els.interviewModal) els.interviewModal.style.display = 'none';
    setInterviewLoading(false);
    pendingInterviewCandidateId = null;
    pendingInterviewBtn = null;
  }

  function executeInterview() {
    var employeeId = pendingInterviewCandidateId;
    if (!employeeId) return;
    var date = els.interviewDate ? els.interviewDate.value : '';
    if (!date) { if (els.interviewDate) els.interviewDate.focus(); return; }
    var time = els.interviewTime ? els.interviewTime.value : '';
    var type = els.interviewType ? els.interviewType.value : 'technical';
    var notes = els.interviewNotes ? els.interviewNotes.value : '';

    var scheduledAt = date + (time ? 'T' + time + ':00' : 'T09:00:00');

    setInterviewLoading(true);

    AngaziaAPI.applications.jobApplications(currentJobId).then(function (apps) {
      var appList = Array.isArray(apps) ? apps : (apps && apps.applications ? apps.applications : []);
      var application = null;
      for (var i = 0; i < appList.length; i++) {
        var a = appList[i];
        if (a.employee_id === employeeId || (a.employee && a.employee.id === employeeId) || a.candidate_id === employeeId) {
          application = a;
          break;
        }
      }
      if (!application) throw new Error('No application found for this candidate');
      return AngaziaAPI.applications.interview(application.id, {
        interview_date: new Date(scheduledAt).toISOString(),
        interview_type: type,
        notes: notes
      });
    }).then(function () {
      hideInterviewModal();
      var btn = pendingInterviewBtn;
      if (btn) { btn.textContent = 'Scheduled'; btn.disabled = true; }
    }).catch(function (err) {
      console.error('Failed to schedule interview:', err);
      setInterviewLoading(false);
    });
  }

  // ── Feedback Modal ──

  var pendingFeedbackMatchId = null;
  var pendingFeedbackRating = 0;
  var pendingFeedbackCandidateName = '';

  function setFeedbackLoading(loading) {
    if (!els.feedbackSubmit) return;
    els.feedbackSubmit.disabled = loading;
    els.feedbackSubmit.classList.toggle('emp-btn-loading', loading);
  }

  function showFeedbackError(msg) {
    if (els.feedbackError) els.feedbackError.style.display = msg ? 'block' : 'none';
    if (els.feedbackErrorMsg) els.feedbackErrorMsg.textContent = msg || '';
  }

  function openFeedbackModal(matchId, candidateName) {
    pendingFeedbackMatchId = matchId;
    pendingFeedbackRating = 0;
    pendingFeedbackCandidateName = candidateName || '';

    if (els.feedbackHeading) els.feedbackHeading.textContent = 'Rate ' + (pendingFeedbackCandidateName || 'this Candidate');
    if (els.feedbackNotes) els.feedbackNotes.value = '';
    if (els.feedbackStarLabel) els.feedbackStarLabel.textContent = 'Select rating';
    showFeedbackError('');
    setFeedbackLoading(false);

    // Reset stars
    if (els.feedbackStars) {
      els.feedbackStars.querySelectorAll('.emp-star').forEach(function (s) { s.classList.remove('active'); s.textContent = '\u2606'; });
    }

    if (els.feedbackModal) els.feedbackModal.style.display = 'flex';
  }

  function hideFeedbackModal() {
    if (els.feedbackModal) els.feedbackModal.style.display = 'none';
    setFeedbackLoading(false);
    showFeedbackError('');
    pendingFeedbackMatchId = null;
    pendingFeedbackRating = 0;
    pendingFeedbackCandidateName = '';
  }

  function handleFeedbackSubmit() {
    var matchId = pendingFeedbackMatchId;
    var rating = pendingFeedbackRating;
    if (!matchId) return;
    if (rating < 1 || rating > 5) {
      showFeedbackError('Please select a rating (1-5 stars).');
      return;
    }

    var notes = els.feedbackNotes ? els.feedbackNotes.value : '';
    showFeedbackError('');
    setFeedbackLoading(true);

    AngaziaAPI.matches.submitFeedback({ match_id: matchId, rating: rating, feedback: notes })
      .then(function () {
        hideFeedbackModal();
      })
      .catch(function (err) {
        console.error('Failed to submit feedback:', err);
        setFeedbackLoading(false);
        showFeedbackError(err && err.message ? err.message : 'Failed to submit feedback.');
      });
  }

  if (!els.content || !els.jobSelector) return;

  function showError(msg) {
    if (els.loading) els.loading.style.display = 'none';
    if (els.content) els.content.style.display = 'none';
    if (els.error) {
      els.error.textContent = msg;
      els.error.style.display = '';
    }
  }

  function hideError() {
    if (els.error) { els.error.style.display = 'none'; els.error.textContent = ''; }
  }

  function escapeHtml(t) {
    if (t == null) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(String(t)));
    return d.innerHTML;
  }

  function showToast(msg, type) {
    if (typeof AngaziaApp !== 'undefined' && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
      return;
    }
    var c = document.getElementById('toast-container');
    if (!c) {
      c = document.createElement('div');
      c.id = 'toast-container';
      c.style.cssText = 'position:fixed;bottom:16px;right:16px;z-index:9999;display:flex;flex-direction:column;gap:8px;';
      document.body.appendChild(c);
    }
    var t = document.createElement('div');
    var bg = type === 'success' ? '#00e5a0' : type === 'error' ? '#ef4444' : '#3b82f6';
    t.style.cssText = 'background:' + bg + ';color:#fff;padding:12px 20px;border-radius:10px;font-size:13px;font-family:var(--fm,sans-serif);box-shadow:0 4px 16px rgba(0,0,0,0.15);';
    t.textContent = msg;
    c.appendChild(t);
    setTimeout(function () { t.style.opacity = '0'; setTimeout(function () { t.remove(); }, 200); }, 3500);
  }

  function openModal(title, bodyHtml) {
    if (els.modalTitle) els.modalTitle.textContent = title || 'Details';
    if (els.modalBody) { els.modalBody.innerHTML = bodyHtml || ''; els.modalBody.style.display = ''; }
    if (els.modalLoading) els.modalLoading.style.display = 'none';
    if (els.modalOverlay) els.modalOverlay.style.display = 'flex';
  }

  function closeModal() {
    if (els.modalOverlay) els.modalOverlay.style.display = 'none';
  }

  function setModalLoading(loading) {
    if (els.modalBody) els.modalBody.style.display = loading ? 'none' : '';
    if (els.modalLoading) els.modalLoading.style.display = loading ? 'flex' : 'none';
  }

  function getScoreClass(score) {
    score = parseInt(score, 10);
    if (score >= 80) return 'score-high';
    if (score >= 60) return 'score-good';
    if (score >= 40) return 'score-medium';
    return 'score-low';
  }

  function getScoreColor(score) {
    score = parseInt(score, 10);
    if (score >= 80) return '#00e5a0';
    if (score >= 60) return '#3b82f6';
    if (score >= 40) return '#f5a623';
    return '#ef4444';
  }

  // ── Load Jobs ──

  function showIdleState() {
    hideError();
    if (els.loading) els.loading.style.display = 'none';
    if (els.content) els.content.style.display = '';
    if (els.matchesContainer) {
      els.matchesContainer.innerHTML =
        '<div class="emp-empty"><div class="emp-empty-icon">&#x1F50D;</div><h3 class="emp-empty-title">Select a job to see matches</h3><p class="emp-empty-desc">Choose a job from the dropdown above to view AI-recommended candidates.</p></div>';
    }
    if (els.matchCount) els.matchCount.textContent = '';
    if (els.genQuestionsBtn) els.genQuestionsBtn.style.display = 'none';
    if (els.skillSection) els.skillSection.style.display = 'none';
  }

  function loadJobs() {
    if (typeof AngaziaAPI === 'undefined') { showIdleState(); return; }
    AngaziaAPI.jobs.myJobs({ limit: 100 })
      .then(function (data) {
        var jobs = data && data.jobs ? data.jobs : (Array.isArray(data) ? data : []);
        els.jobSelector.innerHTML = '<option value="">— Choose a job —</option>';
        if (jobs.length) {
          jobs.forEach(function (j) {
            var opt = document.createElement('option');
            opt.value = j.id || '';
            opt.textContent = j.title || 'Untitled';
            els.jobSelector.appendChild(opt);
          });
          els.jobSelector.value = '';
        }
        showIdleState();
      })
      .catch(function () {
        showIdleState();
      });
  }

  // ── Load Matches ──

  function loadMatches(jobId) {
    if (!jobId || typeof AngaziaAPI === 'undefined') return;
    currentJobId = jobId;

    hideError();
    if (els.loading) els.loading.style.display = 'flex';
    if (els.content) els.content.style.display = 'none';

    if (els.genQuestionsBtn) { els.genQuestionsBtn.style.display = 'none'; els.genQuestionsBtn.dataset.jobId = jobId; }

    AngaziaAPI.matches.candidateMatches(jobId)
      .then(function (data) {
        if (els.loading) els.loading.style.display = 'none';
        if (els.content) els.content.style.display = '';

        var matches = Array.isArray(data) ? data : (data && data.matches ? data.matches : []);
        currentMatches = matches;

        renderMatches(matches);
        renderSkillGapAll(matches);
      })
      .catch(function (err) {
        showError(err.message || 'Failed to load matches. Please try again.');
      });
  }

  function renderMatches(matches) {
    if (!els.matchesContainer) return;

    if (!matches || !matches.length) {
      els.matchesContainer.innerHTML =
        '<div class="emp-empty"><div class="emp-empty-icon">&#x1F50D;</div><h3 class="emp-empty-title">No matches found</h3><p class="emp-empty-desc">No candidates matched this job. Try adjusting the job description or posting a new job.</p></div>';
      if (els.matchCount) els.matchCount.textContent = '0 candidates';
      return;
    }

    if (els.matchCount) els.matchCount.textContent = matches.length + ' candidate' + (matches.length > 1 ? 's' : '') + ' found';
    if (els.genQuestionsBtn) els.genQuestionsBtn.style.display = 'inline-flex';

    var list = document.createElement('div');
    list.className = 'emp-candidate-list';

    matches.forEach(function (m) {
      var cid = m.employee_id || m.EmployeeID || '';
      var name = m.candidate_name || m.CandidateName || m.full_name || m.FullName || 'Unknown';
      var headline = m.candidate_headline || m.CandidateHeadline || m.headline || m.Headline || '';
      var score = m.overall_score || m.OverallScore || 0;
      var location = m.candidate_location || m.CandidateLocation || m.location || m.Location || '';
      var experience = m.experience_years || m.ExperienceYears || 0;
      var skills = m.skills || m.Skills || [];
      var matchingSkills = m.matching_skills || m.MatchingSkills || [];
      var missingSkills = m.missing_skills || m.MissingSkills || [];
      var initials = m.candidate_initials || m.CandidateInitials || '';
      var summary = m.summary || m.Summary || '';
      var recommendation = m.recommendation || m.Recommendation || '';
      var matchId = m.match_id || m.MatchID || '';
      var avatarUrl = m.candidate_avatar || m.CandidateAvatar || '';

      if (!initials && name) {
        initials = name.split(' ').map(function (w) { return w[0]; }).join('').toUpperCase().slice(0, 2);
      }

      var card = document.createElement('div');
      card.className = 'emp-candidate-card';
      card.dataset.employeeId = cid;
      card.dataset.matchId = matchId;

      var scoreCls = getScoreClass(score);
      var avatarHtml = avatarUrl ? '<img src="' + avatarUrl + '" alt="' + escapeHtml(name) + '" style="width:100%;height:100%;object-fit:cover;border-radius:50%">' : '<span class="emp-candidate-initials">' + escapeHtml(initials) + '</span>';

      card.innerHTML =
        '<div class="emp-candidate-avatar">' + avatarHtml + '</div>' +
        '<div class="emp-candidate-info">' +
          '<div class="emp-candidate-top">' +
            '<h3 class="emp-candidate-name">' + escapeHtml(name) + '</h3>' +
            '<span class="emp-candidate-score ' + scoreCls + '">' + score + '%</span>' +
          '</div>' +
          (headline ? '<span class="emp-candidate-headline">' + escapeHtml(headline) + '</span>' : '') +
          (summary ? '<div class="emp-match-summary">' + escapeHtml(summary) + '</div>' : '') +
          (recommendation ? '<div class="emp-match-recommendation">&#x1F4A1; ' + escapeHtml(recommendation) + '</div>' : '') +
          '<div class="emp-candidate-tags">' +
            (Array.isArray(matchingSkills) && matchingSkills.length ? matchingSkills.slice(0, 4).map(function (s) {
              return '<span class="emp-tag emp-tag-matched">' + escapeHtml(typeof s === 'string' ? s : s.name || '') + '</span>';
            }).join('') : '') +
            (Array.isArray(skills) ? skills.slice(0, 3).map(function (s) {
              return '<span class="emp-tag">' + escapeHtml(typeof s === 'string' ? s : s.name || '') + '</span>';
            }).join('') : '') +
            (Array.isArray(missingSkills) && missingSkills.length ? missingSkills.slice(0, 2).map(function (s) {
              return '<span class="emp-tag emp-tag-missing">' + escapeHtml(typeof s === 'string' ? s : '') + '</span>';
            }).join('') : '') +
          '</div>' +
          '<div class="emp-candidate-meta">' +
            (location ? '<span>&#x1F4CD; ' + escapeHtml(location) + '</span>' : '') +
            (experience ? '<span>&#x1F4C5; ' + experience + ' yrs exp</span>' : '') +
          '</div>' +
          (matchId ? '<div class="emp-match-feedback">' +
            '<span style="font-size:10px;color:var(--muted2)">Was this match accurate?</span>' +
            '<button class="emp-btn emp-btn-xs emp-btn-ghost action-feedback" data-match-id="' + escapeHtml(matchId) + '" data-candidate-name="' + escapeHtml(name) + '">&#x1F4AC; Give Feedback</button>' +
          '</div>' : '') +
        '</div>' +
        '<div class="emp-candidate-actions emp-match-actions">' +
          '<button class="emp-btn emp-btn-sm emp-btn-outline action-view-profile" data-candidate-id="' + escapeHtml(cid) + '">View Profile</button>' +
          '<button class="emp-btn emp-btn-sm emp-btn-ghost action-add-pool" data-candidate-id="' + escapeHtml(cid) + '">+ Pool</button>' +
          '<button class="emp-btn emp-btn-sm emp-btn-ghost action-interview" data-candidate-id="' + escapeHtml(cid) + '">&#x1F4C5; Interview</button>' +
          '<button class="emp-btn emp-btn-sm emp-btn-ghost action-view-analysis" data-candidate-id="' + escapeHtml(cid) + '">&#x1F4CA; Analysis</button>' +
          '<button class="emp-btn emp-btn-sm emp-btn-ghost action-view-skillsgap" data-candidate-id="' + escapeHtml(cid) + '">&#x1F9E0; Skills Gap</button>' +
        '</div>';

      list.appendChild(card);
    });

    els.matchesContainer.innerHTML = '';
    els.matchesContainer.appendChild(list);
    bindCardActions();
    bindFeedback();
  }

  // ── Card Action Handlers ──

  function bindFeedback() {
    els.matchesContainer.querySelectorAll('.action-feedback').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var matchId = this.dataset.matchId;
        var name = this.dataset.candidateName || '';
        if (matchId) openFeedbackModal(matchId, name);
      });
    });
  }

  function bindCardActions() {
    els.matchesContainer.querySelectorAll('.action-view-profile').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var cid = this.dataset.candidateId;
        if (cid) window.location.href = '/employer/candidates/' + cid;
      });
    });

    els.matchesContainer.querySelectorAll('.action-add-pool').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var cid = this.dataset.candidateId;
        var name = getCandidateNameFromBtn(this);
        if (cid) openPoolPicker(cid, name);
      });
    });

    els.matchesContainer.querySelectorAll('.action-interview').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var cid = this.dataset.candidateId;
        if (cid) openInterviewModal(cid, this);
      });
    });

    els.matchesContainer.querySelectorAll('.action-view-analysis').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var cid = this.dataset.candidateId;
        if (cid) viewAnalysis(cid, currentJobId);
      });
    });

    els.matchesContainer.querySelectorAll('.action-view-skillsgap').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var cid = this.dataset.candidateId;
        if (cid) viewSkillsGap(cid, currentJobId);
      });
    });

    els.matchesContainer.querySelectorAll('.action-feedback').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var matchId = this.dataset.matchId;
        var name = this.dataset.candidateName || '';
        if (matchId) openFeedbackModal(matchId, name);
      });
    });
  }

  // ── Skills Gap Aggregation (aggregated chart) ──

  function renderSkillGapAll(matches) {
    if (!els.skillSection || !els.skillChart) return;
    if (!matches || !matches.length) {
      els.skillSection.style.display = 'none';
      return;
    }

    els.skillSection.style.display = '';
    if (els.skillEmpty) els.skillEmpty.style.display = 'none';

    var skillMap = {};
    var total = matches.length;

    matches.forEach(function (m) {
      var matching = m.matching_skills || m.MatchingSkills || [];
      var missing = m.missing_skills || m.MissingSkills || [];
      matching.forEach(function (s) {
        var name = typeof s === 'string' ? s : (s.name || s.Name || '');
        if (!name) return;
        if (!skillMap[name]) skillMap[name] = { matching: 0, missing: 0 };
        skillMap[name].matching++;
      });
      missing.forEach(function (s) {
        var name = typeof s === 'string' ? s : (s.name || s.Name || '');
        if (!name) return;
        if (!skillMap[name]) skillMap[name] = { matching: 0, missing: 0 };
        skillMap[name].missing++;
      });
    });

    var skills = Object.keys(skillMap).map(function (name) {
      return { name: name, score: Math.round((skillMap[name].matching / total) * 100) };
    });
    skills.sort(function (a, b) { return b.score - a.score; });
    skills = skills.slice(0, 10);

    if (!skills.length) {
      els.skillSection.style.display = 'none';
      return;
    }

    var html = '';
    skills.forEach(function (s) {
      html += '<div class="emp-skill-bar-row">' +
        '<span class="emp-skill-bar-label">' + escapeHtml(s.name) + '</span>' +
        '<div class="emp-skill-bar-track"><div class="emp-skill-bar-fill" style="width:' + s.score + '%"></div></div>' +
        '<span class="emp-skill-bar-pct">' + s.score + '%</span>' +
        '</div>';
    });
    els.skillChart.innerHTML = html;
  }

  // ── View Analysis (detailed modal per candidate) ──

  function viewAnalysis(candidateId, jobId) {
    if (!candidateId || !jobId) return;
    setModalLoading(true);
    openModal('Match Analysis', '');

    AngaziaAPI.matches.employerAnalysis(jobId, candidateId)
      .then(function (analysis) {
        setModalLoading(false);
        if (!analysis) {
          els.modalBody.innerHTML = '<p style="text-align:center;color:var(--muted)">No analysis available.</p>';
          return;
        }

        var scoreItems = [
          { label: 'Overall', value: analysis.overall_score || 0 },
          { label: 'Skills', value: analysis.skills_score || 0 },
          { label: 'Experience', value: analysis.experience_score || 0 },
          { label: 'Culture', value: analysis.culture_score || 0 },
          { label: 'Location', value: analysis.location_score || 0 },
        ];
        var scoreHtml = '';
        scoreItems.forEach(function (s) {
          scoreHtml += '<div class="mm-score-item"><div class="mm-score-value" style="color:' + getScoreColor(s.value) + '">' + s.value + '</div><div class="mm-score-label">' + s.label + '</div></div>';
        });

        var bodyHtml = '<div class="mm-analysis-scores">' + scoreHtml + '</div>';

        if (analysis.strong_points && analysis.strong_points.length) {
          bodyHtml += '<div class="mm-analysis-section"><div class="mm-analysis-label">Strong Points</div><ul style="margin:0;padding-left:20px">';
          analysis.strong_points.forEach(function (p) { bodyHtml += '<li style="padding:4px 0;color:var(--accent)">' + escapeHtml(p) + '</li>'; });
          bodyHtml += '</ul></div>';
        }

        if (analysis.weak_points && analysis.weak_points.length) {
          bodyHtml += '<div class="mm-analysis-section"><div class="mm-analysis-label">Areas for Improvement</div><ul style="margin:0;padding-left:20px">';
          analysis.weak_points.forEach(function (p) { bodyHtml += '<li style="padding:4px 0;color:var(--warn)">' + escapeHtml(p) + '</li>'; });
          bodyHtml += '</ul></div>';
        }

        if (analysis.matching_skills && analysis.matching_skills.length) {
          bodyHtml += '<div class="mm-analysis-section"><div class="mm-analysis-label">Matching Skills</div><div class="mm-tag-list">' +
            analysis.matching_skills.map(function (s) { return '<span class="emp-tag emp-tag-matched">' + escapeHtml(s) + '</span>'; }).join('') +
            '</div></div>';
        }

        if (analysis.missing_skills && analysis.missing_skills.length) {
          bodyHtml += '<div class="mm-analysis-section"><div class="mm-analysis-label">Missing Skills</div><div class="mm-tag-list">' +
            analysis.missing_skills.map(function (s) { return '<span class="emp-tag emp-tag-missing">' + escapeHtml(s) + '</span>'; }).join('') +
            '</div></div>';
        }

        if (analysis.summary) {
          bodyHtml += '<div class="mm-analysis-section"><div class="mm-analysis-label">Analysis</div><p style="margin:0;line-height:1.7">' + escapeHtml(analysis.summary) + '</p></div>';
        }

        if (analysis.recommendation) {
          bodyHtml += '<div class="mm-analysis-section"><div class="mm-analysis-label">Recommendation</div><p style="margin:0;line-height:1.7">' + escapeHtml(analysis.recommendation) + '</p></div>';
        }

        if (analysis.interview_tips && analysis.interview_tips.length) {
          bodyHtml += '<div class="mm-analysis-section"><div class="mm-analysis-label">Interview Tips</div><ol class="mm-question-list">';
          analysis.interview_tips.forEach(function (t) { bodyHtml += '<li>' + escapeHtml(t) + '</li>'; });
          bodyHtml += '</ol></div>';
        }

        els.modalBody.innerHTML = bodyHtml;
      })
      .catch(function (err) {
        setModalLoading(false);
        els.modalBody.innerHTML = '<p style="text-align:center;color:var(--danger)">Failed to load analysis: ' + escapeHtml(err.message || 'Unknown error') + '</p>';
      });
  }

  // ── View Skills Gap (detailed modal per candidate) ──

  function viewSkillsGap(candidateId, jobId) {
    if (!candidateId || !jobId) return;
    setModalLoading(true);
    openModal('Skills Gap Analysis', '');

    AngaziaAPI.matches.employerSkillsGap(jobId, candidateId)
      .then(function (analysis) {
        setModalLoading(false);
        if (!analysis) {
          els.modalBody.innerHTML = '<p style="text-align:center;color:var(--muted)">No skills gap data available.</p>';
          return;
        }

        var html = '';

        if (analysis.priority_level) {
          var pc = analysis.priority_level === 'high' ? 'critical' : analysis.priority_level === 'medium' ? 'important' : 'nice-to-have';
          html += '<div style="margin-bottom:16px"><span class="mm-gap-importance ' + pc + '">' + escapeHtml(analysis.priority_level.toUpperCase()) + ' PRIORITY</span></div>';
        }

        if (analysis.estimated_time_to_fill) {
          html += '<p style="font-size:13px;color:var(--muted2);margin:0 0 16px">Est. time to fill: <strong>' + escapeHtml(analysis.estimated_time_to_fill) + '</strong></p>';
        }

        if (analysis.transferable_skills && analysis.transferable_skills.length) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Transferable Skills</div><div class="mm-tag-list">' +
            analysis.transferable_skills.map(function (s) { return '<span class="emp-tag">' + escapeHtml(s) + '</span>'; }).join('') +
            '</div></div>';
        }

        if (analysis.missing_skills && analysis.missing_skills.length) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Missing Skills</div>';
          analysis.missing_skills.forEach(function (gap) {
            var ic = gap.importance === 'critical' ? 'critical' : gap.importance === 'important' ? 'important' : 'nice-to-have';
            html += '<div class="mm-gap-item">' +
              '<div class="mm-gap-skill">' + escapeHtml(gap.skill_name || gap.SkillName || '') +
              ' <span class="mm-gap-importance ' + ic + '">' + escapeHtml(gap.importance || '') + '</span></div>' +
              (gap.description ? '<div class="mm-gap-desc">' + escapeHtml(gap.description) + '</div>' : '') +
              (gap.learning_resources && gap.learning_resources.length ? '<div class="mm-gap-resources">Resources: ' + gap.learning_resources.map(function (r) { return escapeHtml(r); }).join(' &#183; ') + '</div>' : '') +
              '</div>';
          });
          html += '</div>';
        }

        if (analysis.improvement_plan) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Improvement Plan</div><p style="margin:0;line-height:1.7;white-space:pre-wrap">' + escapeHtml(analysis.improvement_plan) + '</p></div>';
        }

        if (analysis.recommended_courses && analysis.recommended_courses.length) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Recommended Courses</div>';
          analysis.recommended_courses.forEach(function (c) {
            var cName = c.name || c.Name || '';
            var cPlatform = c.platform || c.Platform || '';
            var cDuration = c.duration || c.Duration || '';
            var cDifficulty = c.difficulty || c.Difficulty || '';
            var cUrl = c.url || c.URL || '';
            html += '<div class="mm-gap-item" style="display:flex;justify-content:space-between;align-items:center">' +
              '<div><strong>' + escapeHtml(cName) + '</strong><br><span style="font-size:11px;color:var(--muted2)">' + escapeHtml(cPlatform) + ' &middot; ' + escapeHtml(cDuration) + ' &middot; ' + escapeHtml(cDifficulty) + '</span></div>' +
              (cUrl ? '<a href="' + escapeHtml(cUrl) + '" target="_blank" rel="noopener" class="emp-btn emp-btn-xs emp-btn-outline" style="flex-shrink:0">View</a>' : '') +
              '</div>';
          });
          html += '</div>';
        }

        if (!html) html = '<p style="text-align:center;color:var(--muted)">No skills gap data available.</p>';
        els.modalBody.innerHTML = html;
      })
      .catch(function (err) {
        setModalLoading(false);
        els.modalBody.innerHTML = '<p style="text-align:center;color:var(--danger)">Failed to load skills gap: ' + escapeHtml(err.message || 'Unknown error') + '</p>';
      });
  }

  // ── Generate Interview Questions ──

  function generateQuestions(jobId) {
    if (!jobId) return;
    setModalLoading(true);
    openModal('Interview Questions', '');

    AngaziaAPI.matches.interviewQuestions(jobId)
      .then(function (resp) {
        setModalLoading(false);
        var questions = resp && resp.questions ? resp.questions : (Array.isArray(resp) ? resp : []);
        if (!questions.length) {
          els.modalBody.innerHTML = '<p style="text-align:center;color:var(--muted)">No questions generated.</p>';
          return;
        }
        var html = '<p style="font-size:12px;color:var(--muted2);margin:0 0 12px">Role-specific interview questions to assess candidates:</p><ol class="mm-question-list">';
        questions.forEach(function (q) { html += '<li>' + escapeHtml(typeof q === 'string' ? q : (q.text || q.question || '')) + '</li>'; });
        html += '</ol>';
        els.modalBody.innerHTML = html;
      })
      .catch(function (err) {
        setModalLoading(false);
        els.modalBody.innerHTML = '<p style="text-align:center;color:var(--danger)">Failed to generate questions: ' + escapeHtml(err.message || 'Unknown error') + '</p>';
      });
  }

  // ── Init ──

  function getCandidateNameFromBtn(btn) {
    if (!btn) return '';
    var card = btn.closest('.emp-candidate-card') || btn.closest('.mm-candidate');
    if (!card) return '';
    var nameEl = card.querySelector('.emp-candidate-name') || card.querySelector('.mm-candidate-name');
    return nameEl ? nameEl.textContent.trim() : '';
  }

  function init() {
    loadJobs();

    els.jobSelector.addEventListener('change', function () {
      var jid = this.value;
      if (jid) {
        loadMatches(jid);
      } else {
        showIdleState();
      }
    });

    if (els.genQuestionsBtn) {
      els.genQuestionsBtn.addEventListener('click', function () {
        var jid = this.dataset.jobId;
        if (jid) generateQuestions(jid);
      });
    }

    if (els.modalClose) {
      els.modalClose.addEventListener('click', closeModal);
    }
    if (els.modalOverlay) {
      els.modalOverlay.addEventListener('click', function (e) {
        if (e.target === els.modalOverlay) closeModal();
      });
    }
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') {
        if (els.poolModal && els.poolModal.style.display === 'flex') closePoolPicker();
        else if (els.interviewModal && els.interviewModal.style.display === 'flex') hideInterviewModal();
        else if (els.feedbackModal && els.feedbackModal.style.display === 'flex') hideFeedbackModal();
        else closeModal();
      }
    });

    // Pool picker modal events
    if (els.poolSave) els.poolSave.addEventListener('click', handlePoolSave);
    if (els.poolCancel) els.poolCancel.addEventListener('click', closePoolPicker);
    if (els.poolClose) els.poolClose.addEventListener('click', closePoolPicker);
    if (els.poolModal) {
      els.poolModal.addEventListener('click', function (e) {
        if (e.target === els.poolModal) closePoolPicker();
      });
    }
    if (els.poolCreateToggle) {
      els.poolCreateToggle.addEventListener('click', function () {
        if (els.poolNewSection) els.poolNewSection.style.display = 'block';
        if (els.poolNewName) els.poolNewName.focus();
        if (els.poolSelect) els.poolSelect.value = '';
        this.style.display = 'none';
      });
    }
    if (els.poolSelect) {
      els.poolSelect.addEventListener('change', function () {
        if (this.value) {
          if (els.poolNewSection) els.poolNewSection.style.display = 'none';
          if (els.poolNewName) els.poolNewName.value = '';
          if (els.poolCreateToggle) els.poolCreateToggle.style.display = 'flex';
        }
        if (els.poolError) els.poolError.style.display = 'none';
      });
    }
    if (els.poolNewName) {
      els.poolNewName.addEventListener('input', function () {
        if (els.poolError) els.poolError.style.display = 'none';
      });
    }

    // Interview modal events
    if (els.interviewConfirm) els.interviewConfirm.addEventListener('click', executeInterview);
    if (els.interviewCancel) els.interviewCancel.addEventListener('click', hideInterviewModal);
    if (els.interviewClose) els.interviewClose.addEventListener('click', hideInterviewModal);
    if (els.interviewModal) {
      els.interviewModal.addEventListener('click', function (e) {
        if (e.target === els.interviewModal) hideInterviewModal();
      });
    }

    // Feedback modal events
    if (els.feedbackSubmit) els.feedbackSubmit.addEventListener('click', handleFeedbackSubmit);
    if (els.feedbackCancel) els.feedbackCancel.addEventListener('click', hideFeedbackModal);
    if (els.feedbackClose) els.feedbackClose.addEventListener('click', hideFeedbackModal);
    if (els.feedbackModal) {
      els.feedbackModal.addEventListener('click', function (e) {
        if (e.target === els.feedbackModal) hideFeedbackModal();
      });
    }
    if (els.feedbackStars) {
      els.feedbackStars.querySelectorAll('.emp-star').forEach(function (s) {
        s.addEventListener('click', function () {
          var val = parseInt(this.dataset.value, 10);
          pendingFeedbackRating = val;
          els.feedbackStars.querySelectorAll('.emp-star').forEach(function (star, i) {
            if (i < val) { star.classList.add('active'); star.textContent = '\u2605'; }
            else { star.classList.remove('active'); star.textContent = '\u2606'; }
          });
          var labels = ['', 'Poor', 'Below Average', 'Average', 'Good', 'Excellent'];
          if (els.feedbackStarLabel) els.feedbackStarLabel.textContent = labels[val] || val + ' stars';
          if (els.feedbackError) els.feedbackError.style.display = 'none';
        });
      });
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
