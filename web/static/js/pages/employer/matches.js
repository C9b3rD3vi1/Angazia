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
  };

  var currentJobId = '';
  var currentMatches = [];

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
            '<button class="emp-feedback-btn" data-match-id="' + escapeHtml(matchId) + '" data-rating="1" title="Not accurate">&#x1F44E;</button>' +
            '<button class="emp-feedback-btn" data-match-id="' + escapeHtml(matchId) + '" data-rating="3" title="Somewhat accurate">&#x1F44D;</button>' +
            '<button class="emp-feedback-btn" data-match-id="' + escapeHtml(matchId) + '" data-rating="5" title="Very accurate">&#x1F31F;</button>' +
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
        if (cid) addToTalentPool(cid, currentJobId, this);
      });
    });

    els.matchesContainer.querySelectorAll('.action-interview').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var cid = this.dataset.candidateId;
        if (cid) scheduleInterview(cid, this);
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
  }

  function bindFeedback() {
    els.matchesContainer.querySelectorAll('.emp-feedback-btn').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var matchId = this.dataset.matchId;
        var rating = parseInt(this.dataset.rating, 10);
        if (!matchId) return;

        var allBtns = this.parentElement.querySelectorAll('.emp-feedback-btn');
        allBtns.forEach(function (b) { b.classList.remove('active'); b.style.opacity = '0.4'; });
        this.classList.add('active');
        this.style.opacity = '1';

        AngaziaAPI.matches.submitFeedback({ match_id: matchId, rating: rating })
          .then(function () { })
          .catch(function () { });
      });
    });
  }

  function addToTalentPool(candidateId, jobId, btn) {
    if (typeof AngaziaAPI === 'undefined') return;
    btn.disabled = true;
    btn.textContent = 'Adding...';

    AngaziaAPI.talentPools.list({ limit: 1 })
      .then(function (data) {
        var pools = data && data.pools ? data.pools : (Array.isArray(data) ? data : []);
        if (pools && pools.length) {
          return AngaziaAPI.talentPools.addCandidate(pools[0].id, { candidate_id: candidateId, job_id: jobId })
            .then(function () {
              btn.textContent = 'Added';
            });
        }
        return AngaziaAPI.talentPools.create({ name: 'AI Matches - ' + new Date().toLocaleDateString() })
          .then(function (p) {
            var poolId = p && (p.id || (p.data && p.data.id));
            if (!poolId) throw new Error('Could not create pool');
              return AngaziaAPI.talentPools.addCandidate(poolId, { candidate_id: candidateId, job_id: jobId })
                .then(function () {
                  btn.textContent = 'Added';
                });
          });
      })
      .catch(function (err) {
        console.error(err);
        btn.disabled = false;
        btn.textContent = '+ Pool';
      });
  }

  function scheduleInterview(candidateId, btn) {
    var date = prompt('Interview date (YYYY-MM-DD):');
    if (!date) return;
    var time = prompt('Interview time (optional, e.g. 10:00):');
    var scheduledAt = date + (time ? 'T' + time + ':00' : 'T09:00:00');

    btn.disabled = true;
    btn.textContent = 'Scheduling...';

    AngaziaAPI.applications.interview(candidateId, { scheduled_at: scheduledAt, job_id: currentJobId })
      .then(function () {
        btn.textContent = '&#x1F4C5; Interview';
        btn.disabled = false;
      })
      .catch(function (err) {
        console.error(err);
        btn.textContent = '&#x1F4C5; Interview';
        btn.disabled = false;
      });
  }

  function viewAnalysis(candidateId, jobId) {
    if (typeof AngaziaAPI === 'undefined') return;
    openModal('Match Analysis', '');
    setModalLoading(true);

    AngaziaAPI.matches.employerAnalysis(jobId, candidateId)
      .then(function (data) {
        var analysis = data && data.analysis ? data.analysis : data;
        var html = '';

        html += '<div class="mm-analysis-scores">';
        html += scoreBox('Overall', analysis.overall_score || 0, getScoreColor(analysis.overall_score));
        html += scoreBox('Skills', analysis.skills_score || 0, getScoreColor(analysis.skills_score));
        html += scoreBox('Experience', analysis.experience_score || 0, getScoreColor(analysis.experience_score));
        html += scoreBox('Culture', analysis.culture_score || 0, getScoreColor(analysis.culture_score));
        html += scoreBox('Location', analysis.location_score || 0, getScoreColor(analysis.location_score));
        html += '</div>';

        if (analysis.matching_skills && analysis.matching_skills.length) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x2705; Matching Skills</div>';
          html += '<div class="mm-tag-list">';
          analysis.matching_skills.forEach(function (s) {
            html += '<span class="emp-tag emp-tag-matched">' + escapeHtml(typeof s === 'string' ? s : '') + '</span>';
          });
          html += '</div></div>';
        }

        if (analysis.missing_skills && analysis.missing_skills.length) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x274C; Missing Skills</div>';
          html += '<div class="mm-tag-list">';
          analysis.missing_skills.forEach(function (s) {
            html += '<span class="emp-tag emp-tag-missing">' + escapeHtml(typeof s === 'string' ? s : '') + '</span>';
          });
          html += '</div></div>';
        }

        if (analysis.strong_points && analysis.strong_points.length) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x2B50; Strong Points</div>';
          html += '<ul style="padding-left:16px;margin:0">';
          analysis.strong_points.forEach(function (s) {
            html += '<li style="padding:3px 0;font-size:12px;color:var(--muted)">' + escapeHtml(typeof s === 'string' ? s : '') + '</li>';
          });
          html += '</ul></div>';
        }

        if (analysis.weak_points && analysis.weak_points.length) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x26A0; Areas to Improve</div>';
          html += '<ul style="padding-left:16px;margin:0">';
          analysis.weak_points.forEach(function (s) {
            html += '<li style="padding:3px 0;font-size:12px;color:var(--muted)">' + escapeHtml(typeof s === 'string' ? s : '') + '</li>';
          });
          html += '</ul></div>';
        }

        if (analysis.interview_tips && analysis.interview_tips.length) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x1F4AD; Interview Tips</div>';
          html += '<ul style="padding-left:16px;margin:0">';
          analysis.interview_tips.forEach(function (s) {
            html += '<li style="padding:3px 0;font-size:12px;color:var(--muted)">' + escapeHtml(typeof s === 'string' ? s : '') + '</li>';
          });
          html += '</ul></div>';
        }

        if (analysis.summary) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x1F4A1; Summary</div>';
          html += '<p style="margin:0;font-size:13px;color:var(--text)">' + escapeHtml(analysis.summary) + '</p></div>';
        }

        if (analysis.recommendation) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x1F3AF; Recommendation</div>';
          html += '<p style="margin:0;font-size:13px;color:var(--accent)">' + escapeHtml(analysis.recommendation) + '</p></div>';
        }

        if (analysis.analysis_metadata) {
          var meta = analysis.analysis_metadata;
          html += '<div style="margin-top:16px;padding-top:12px;border-top:1px solid var(--border);font-size:10px;color:var(--muted2)">';
          if (meta.provider) html += 'AI: ' + escapeHtml(meta.provider) + ' | ';
          if (meta.model) html += 'Model: ' + escapeHtml(meta.model) + ' | ';
          if (meta.processing_time_ms) html += 'Time: ' + meta.processing_time_ms + 'ms';
          html += '</div>';
        }

        setModalLoading(false);
        if (els.modalBody) { els.modalBody.innerHTML = html; els.modalBody.style.display = ''; }
      })
      .catch(function () {
        setModalLoading(false);
        if (els.modalBody) {
          els.modalBody.innerHTML = '<p style="color:var(--muted);text-align:center;padding:20px 0">Failed to load analysis. Please try again.</p>';
          els.modalBody.style.display = '';
        }
      });
  }

  function scoreBox(label, value, color) {
    return '<div class="mm-score-item">' +
      '<div class="mm-score-value" style="color:' + color + '">' + value + '%</div>' +
      '<div class="mm-score-label">' + label + '</div>' +
      '</div>';
  }

  function viewSkillsGap(candidateId, jobId) {
    if (typeof AngaziaAPI === 'undefined') return;
    openModal('Skills Gap Analysis', '');
    setModalLoading(true);

    AngaziaAPI.matches.employerSkillsGap(jobId, candidateId)
      .then(function (data) {
        var analysis = data && data.analysis ? data.analysis : data;
        var html = '';

        if (analysis.improvement_plan) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x1F4CB; Improvement Plan</div>';
          html += '<p style="margin:0 0 12px;font-size:13px;color:var(--text)">' + escapeHtml(analysis.improvement_plan) + '</p></div>';
        }

        if (analysis.estimated_time_to_fill) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x23F1; Estimated Time to Fill</div>';
          html += '<p style="margin:0 0 12px;font-size:13px;color:var(--text)">' + escapeHtml(analysis.estimated_time_to_fill) + '</p></div>';
        }

        if (analysis.priority_level) {
          var pCls = analysis.priority_level === 'high' ? 'critical' : (analysis.priority_level === 'medium' ? 'important' : '');
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">Priority Level</div>';
          html += '<span class="mm-gap-importance ' + pCls + '">' + escapeHtml(analysis.priority_level.toUpperCase()) + '</span>';
          html += '</div>';
        }

        if (analysis.missing_skills && analysis.missing_skills.length) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x274C; Missing Skills</div>';
          analysis.missing_skills.forEach(function (s) {
            var skillName = s.skill_name || s.SkillName || '';
            var importance = s.importance || s.Importance || '';
            var description = s.description || s.Description || '';
            var resources = s.learning_resources || s.LearningResources || [];
            var impCls = importance === 'critical' ? 'critical' : (importance === 'important' ? 'important' : 'nice-to-have');

            html += '<div class="mm-gap-item">';
            html += '<div class="mm-gap-skill">' + escapeHtml(skillName) + '</div>';
            if (importance) html += '<span class="mm-gap-importance ' + impCls + '">' + escapeHtml(importance.toUpperCase()) + '</span>';
            if (description) html += '<div class="mm-gap-desc">' + escapeHtml(description) + '</div>';
            if (resources && resources.length) {
              html += '<div class="mm-gap-resources">Resources: ';
              resources.forEach(function (r, i) {
                html += escapeHtml(typeof r === 'string' ? r : r.name || r.url || '');
                if (i < resources.length - 1) html += ', ';
              });
              html += '</div>';
            }
            html += '</div>';
          });
          html += '</div>';
        }

        if (analysis.transferable_skills && analysis.transferable_skills.length) {
          html += '<div class="mm-analysis-section">';
          html += '<div class="mm-analysis-label">&#x1F500; Transferable Skills</div>';
          html += '<div class="mm-tag-list">';
          analysis.transferable_skills.forEach(function (s) {
            html += '<span class="emp-tag">' + escapeHtml(typeof s === 'string' ? s : '') + '</span>';
          });
          html += '</div></div>';
        }

        setModalLoading(false);
        if (els.modalBody) { els.modalBody.innerHTML = html || '<p style="color:var(--muted);text-align:center;padding:20px 0">No skills gap data available.</p>'; els.modalBody.style.display = ''; }
      })
      .catch(function () {
        setModalLoading(false);
        if (els.modalBody) {
          els.modalBody.innerHTML = '<p style="color:var(--muted);text-align:center;padding:20px 0">Failed to load skills gap analysis.</p>';
          els.modalBody.style.display = '';
        }
      });
  }

  // ── Skills Gap Aggregate ──

  function renderSkillGapAll(matches) {
    if (!els.skillChart || !els.skillSection) return;

    var hasData = matches && matches.some(function (m) {
      return (m.matching_skills && m.matching_skills.length) || (m.missing_skills && m.missing_skills.length);
    });

    if (!hasData) {
      els.skillSection.style.display = 'none';
      return;
    }

    els.skillSection.style.display = 'block';

    var skillCounts = {};
    matches.forEach(function (m) {
      var skills = m.skills || m.Skills || [];
      var matching = m.matching_skills || m.MatchingSkills || [];
      var missing = m.missing_skills || m.MissingSkills || [];

      var all = [];
      if (Array.isArray(skills)) {
        skills.forEach(function (s) {
          var name = typeof s === 'string' ? s : (s.name || '');
          if (name) all.push(name);
        });
      }
      if (Array.isArray(matching)) {
        matching.forEach(function (s) {
          var name = typeof s === 'string' ? s : '';
          if (name && all.indexOf(name) === -1) all.push(name);
        });
      }
      if (Array.isArray(missing)) {
        missing.forEach(function (s) {
          var name = typeof s === 'string' ? s : '';
          if (name && all.indexOf(name) === -1) all.push(name);
        });
      }

      all.forEach(function (skill) {
        if (!skillCounts[skill]) skillCounts[skill] = { count: 0, total: 0 };
        skillCounts[skill].total++;
        if (matching.indexOf(skill) !== -1 || skills.indexOf(skill) !== -1) {
          skillCounts[skill].count++;
        }
      });
    });

    var sorted = Object.keys(skillCounts).sort(function (a, b) {
      return (skillCounts[b].count / skillCounts[b].total) - (skillCounts[a].count / skillCounts[a].total);
    }).slice(0, 10);

    if (!sorted.length) {
      els.skillSection.style.display = 'none';
      return;
    }

    els.skillChart.innerHTML = '';
    sorted.forEach(function (skill) {
      var info = skillCounts[skill];
      var pct = Math.round((info.count / info.total) * 100);
      var row = document.createElement('div');
      row.className = 'emp-skill-bar-row';

      var label = document.createElement('span');
      label.className = 'emp-skill-bar-label';
      label.textContent = skill;

      var track = document.createElement('div');
      track.className = 'emp-skill-bar-track';

      var fill = document.createElement('div');
      fill.className = 'emp-skill-bar-fill';
      fill.style.width = pct + '%';

      track.appendChild(fill);

      var pctEl = document.createElement('span');
      pctEl.className = 'emp-skill-bar-pct';
      pctEl.textContent = pct + '%';

      row.appendChild(label);
      row.appendChild(track);
      row.appendChild(pctEl);
      els.skillChart.appendChild(row);
    });
  }

  // ── Generate Interview Questions ──

  function generateQuestions(jobId) {
    if (!jobId || typeof AngaziaAPI === 'undefined') return;
    openModal('Interview Questions', '<p style="text-align:center;padding:20px 0;color:var(--muted2)">Generating questions...</p>');

    AngaziaAPI.matches.interviewQuestions(jobId)
      .then(function (data) {
        var questions = data && data.questions ? data.questions : (Array.isArray(data) ? data : []);
        var html = '';
        if (questions && questions.length) {
          html += '<ol class="mm-question-list">';
          questions.forEach(function (q) {
            var text = typeof q === 'string' ? q : (q.question || q.text || '');
            html += '<li>' + escapeHtml(text) + '</li>';
          });
          html += '</ol>';
        } else {
          html = '<p style="color:var(--muted2);text-align:center;padding:20px 0">No questions generated. Try with more job details.</p>';
        }
        if (els.modalBody) { els.modalBody.innerHTML = html; els.modalBody.style.display = ''; }
      })
      .catch(function () {
        if (els.modalBody) {
          els.modalBody.innerHTML = '<p style="color:var(--muted2);text-align:center;padding:20px 0">Failed to generate questions.</p>';
          els.modalBody.style.display = '';
        }
      });
  }

  // ── Init ──

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
      if (e.key === 'Escape') closeModal();
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
