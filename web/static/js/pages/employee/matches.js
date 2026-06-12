(function () {
  'use strict';

  document.addEventListener('DOMContentLoaded', function () {
    setupEventListeners();
  });

  function setupEventListeners() {
    document.addEventListener('click', function (e) {
      var btn = e.target.closest('[data-action]');
      if (!btn) return;

      var action = btn.getAttribute('data-action');
      var jobId = btn.getAttribute('data-id') || btn.getAttribute('data-job');
      var employeeId = btn.getAttribute('data-employee');
      var matchId = btn.getAttribute('data-match-id');

      switch (action) {
        case 'apply-job':
          handleApply(jobId);
          break;
        case 'view-job':
          window.location.href = '/employee/jobs/' + jobId;
          break;
        case 'analyze-match':
          handleAnalyze(jobId, employeeId);
          break;
        case 'skills-gap':
          handleSkillsGap(jobId);
          break;
        case 'cover-letter':
          handleCoverLetter(jobId);
          break;
      }
    });

    var refreshBtn = document.getElementById('refresh-matches');
    if (refreshBtn) {
      refreshBtn.addEventListener('click', function () {
        refreshBtn.disabled = true;
        refreshBtn.textContent = '⟳ Refreshing...';
        showLoading(true);
        AngaziaAPI.matches.jobMatches({ limit: 20 })
          .then(function (data) {
            renderMatches(data || []);
            showToast('Matches refreshed!', 'success');
          })
          .catch(function (err) {
            showToast('Failed to refresh: ' + (err.message || 'Unknown error'), 'error');
          })
          .finally(function () {
            refreshBtn.disabled = false;
            refreshBtn.textContent = '🔄 Refresh Matches';
            showLoading(false);
          });
      });
    }

    var modalClose = document.getElementById('mm-modal-close');
    if (modalClose) {
      modalClose.addEventListener('click', closeModal);
    }
    var overlay = document.getElementById('mm-modal-overlay');
    if (overlay) {
      overlay.addEventListener('click', function (e) {
        if (e.target === overlay) closeModal();
      });
    }
  }

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type || 'info');
    } else {
      console.log('[' + (type || 'info') + '] ' + msg);
    }
  }

  function showLoading(on) {
    var el = document.getElementById('mm-loading');
    var content = document.getElementById('mm-content');
    if (el) el.style.display = on ? '' : 'none';
    if (content) content.style.display = on ? 'none' : '';
  }

  function renderMatches(matches) {
    var container = document.getElementById('matches-container');
    if (!container) return;

    if (!matches || matches.length === 0) {
      container.innerHTML =
        '<div class="emp-empty-state">' +
        '<span class="emp-empty-icon">🔍</span>' +
        '<h3>No matches found</h3>' +
        '<p>Complete your profile and add skills to get AI-powered job recommendations.</p>' +
        '<div class="emp-empty-actions">' +
        '<a href="/employee/settings" class="emp-btn emp-btn-primary">Complete Profile</a>' +
        '<a href="/employee/skills" class="emp-btn emp-btn-outline">Add Skills</a>' +
        '</div></div>';
      return;
    }

    var html = '';
    matches.forEach(function (m) {
      var ci = (m.company_name || m.CompanyName || '').split(' ').map(function (w) { return w.charAt(0); }).join('').toUpperCase().slice(0, 2);
      var logo = m.company_logo || m.CompanyLogo || '';
      var matchingSkills = m.matching_skills || m.MatchingSkills || [];
      var missingSkills = m.missing_skills || m.MissingSkills || [];
      var recommendation = m.recommendation || m.Recommendation || 'consider';
      var recLabel = 'Needs Review';
      var recIcon = '📋';
      if (recommendation === 'hire') { recLabel = 'Excellent Match'; recIcon = '🏆'; }
      else if (recommendation === 'interview') { recLabel = 'Strong Match'; recIcon = '✅'; }
      else if (recommendation === 'consider') { recLabel = 'Potential Match'; recIcon = '💭'; }

      html += '<div class="emp-rec-item" data-job-id="' + (m.job_id || m.JobID) + '" data-match-id="' + (m.match_id || m.MatchID || '') +'">' +
        '<div class="emp-rec-header">' +
        (logo ? '<img class="emp-rec-logo" src="' + logo + '" alt="' + (m.company_name || m.CompanyName || '') + '">' : '<div class="emp-rec-logo-placeholder">' + ci + '</div>') +
        '<div class="emp-rec-info"><h4 class="emp-rec-title">' + (m.job_title || m.JobTitle || 'Position') + '</h4>' +
        '<p class="emp-rec-company">' + (m.company_name || m.CompanyName || '') + '</p></div>' +
        '<div class="emp-rec-match-score">' +
        '<svg class="emp-score-ring" viewBox="0 0 36 36"><path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#2a3a3a" stroke-width="3"/>' +
        '<path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#00e5a0" stroke-width="3" stroke-dasharray="' + (m.overall_score || m.OverallScore || 0) + ', 100"/></svg>' +
        '<span class="emp-score-text">' + (m.overall_score || m.OverallScore || 0) + '%</span></div></div>' +
        '<div class="emp-rec-scores">' +
        '<span class="emp-score-chip skills">🛠️ ' + (m.skills_score || m.SkillsScore || 0) + '%</span>' +
        '<span class="emp-score-chip experience">💼 ' + (m.experience_score || m.ExperienceScore || 0) + '%</span>' +
        '<span class="emp-score-chip culture">🤝 ' + (m.culture_score || m.CultureScore || 0) + '%</span>' +
        '<span class="emp-score-chip location">📍 ' + (m.location_score || m.LocationScore || 0) + '%</span>' +
        '</div>' +
        '<div class="emp-rec-skills">' +
        matchingSkills.map(function (s) { return '<span class="emp-skill-match">✓ ' + s + '</span>'; }).join('') +
        missingSkills.map(function (s) { return '<span class="emp-skill-missing">+ ' + s + '</span>'; }).join('') +
        '</div>' +
        (m.summary || m.Summary ? '<p class="emp-rec-summary">' + (m.summary || m.Summary) + '</p>' : '') +
        '<div class="emp-rec-tag-row">' +
        '<span class="emp-rec-badge ' + recommendation + '">' + recIcon + ' ' + recLabel + '</span>' +
        '</div>' +
        '<div class="emp-rec-actions">' +
        '<button class="emp-btn-sm emp-btn-primary" data-action="apply-job" data-id="' + (m.job_id || m.JobID) + '">Apply Now →</button>' +
        '<button class="emp-btn-sm emp-btn-outline" data-action="view-job" data-id="' + (m.job_id || m.JobID) + '">View Job</button>' +
        '<button class="emp-btn-sm emp-btn-ghost" data-action="analyze-match" data-job="' + (m.job_id || m.JobID) + '" data-employee="' + (m.employee_id || m.EmployeeID || '') + '">🔍 Analysis</button>' +
        '<button class="emp-btn-sm emp-btn-ghost" data-action="skills-gap" data-job="' + (m.job_id || m.JobID) + '">📚 Skills Gap</button>' +
        '<button class="emp-btn-sm emp-btn-ghost" data-action="cover-letter" data-job="' + (m.job_id || m.JobID) + '">✉️ Cover Letter</button>' +
        '</div></div>';
    });
    container.innerHTML = html;
  }

  function handleApply(jobId) {
    if (!jobId) return;
    AngaziaAPI.applications.apply({ job_id: jobId })
      .then(function () { showToast('Application submitted successfully!', 'success'); })
      .catch(function (err) { showToast(err.body && err.body.error ? err.body.error : 'Failed to apply', 'error'); });
  }

  function handleAnalyze(jobId, employeeId) {
    openModal('Match Analysis');
    showModalLoading(true);
    if (!employeeId) {
      // Fetch from profile
      AngaziaAPI.profile.get().then(function (profile) {
        employeeId = profile.user_id || profile.UserID || profile.id || profile.ID;
        return AngaziaAPI.matches.analysis(jobId, employeeId);
      }).then(renderAnalysis).catch(handleModalError);
    } else {
      AngaziaAPI.matches.analysis(jobId, employeeId)
        .then(renderAnalysis)
        .catch(handleModalError);
    }
  }

  function renderAnalysis(data) {
    showModalLoading(false);
    var body = document.getElementById('mm-modal-body');
    if (!body) return;

    var d = data || {};
    var html = '<div class="mm-analysis-section">' +
      '<div class="mm-analysis-scores">' +
      '<div class="mm-score-item"><div class="mm-score-value" style="color:' + scoreColor(d.overall_score || 0) + '">' + (d.overall_score || 0) + '%</div><div class="mm-score-label">Overall Match</div></div>' +
      '<div class="mm-score-item"><div class="mm-score-value">' + (d.skills_score || 0) + '%</div><div class="mm-score-label">Skills</div></div>' +
      '<div class="mm-score-item"><div class="mm-score-value">' + (d.experience_score || 0) + '%</div><div class="mm-score-label">Experience</div></div>' +
      '<div class="mm-score-item"><div class="mm-score-value">' + (d.culture_score || 0) + '%</div><div class="mm-score-label">Culture</div></div>' +
      '<div class="mm-score-item"><div class="mm-score-value">' + (d.location_score || 0) + '%</div><div class="mm-score-label">Location</div></div>' +
      '</div></div>';

    if (d.summary) html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Summary</div><p>' + d.summary + '</p></div>';
    if (d.recommendation) html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Recommendation</div><p><strong>' + d.recommendation.toUpperCase() + '</strong></p></div>';

    if (d.strong_points && d.strong_points.length) {
      html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Strengths</div><ul>' +
        d.strong_points.map(function (p) { return '<li>✅ ' + p + '</li>'; }).join('') + '</ul></div>';
    }
    if (d.weak_points && d.weak_points.length) {
      html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Areas to Improve</div><ul>' +
        d.weak_points.map(function (p) { return '<li>⚠️ ' + p + '</li>'; }).join('') + '</ul></div>';
    }
    if (d.matching_skills && d.matching_skills.length) {
      html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Matching Skills</div><div class="mm-tag-list">' +
        d.matching_skills.map(function (s) { return '<span class="emp-skill-match">✓ ' + s + '</span>'; }).join('') + '</div></div>';
    }
    if (d.missing_skills && d.missing_skills.length) {
      html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Missing Skills</div><div class="mm-tag-list">' +
        d.missing_skills.map(function (s) { return '<span class="emp-skill-missing">+ ' + s + '</span>'; }).join('') + '</div></div>';
    }
    if (d.interview_tips && d.interview_tips.length) {
      html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Interview Tips</div><ul>' +
        d.interview_tips.map(function (t) { return '<li>💡 ' + t + '</li>'; }).join('') + '</ul></div>';
    }

    body.innerHTML = html;
  }

  function handleSkillsGap(jobId) {
    if (!jobId) return;
    openModal('Skills Gap Analysis');
    showModalLoading(true);

    AngaziaAPI.matches.skillsGap(jobId)
      .then(function (data) {
        showModalLoading(false);
        var body = document.getElementById('mm-modal-body');
        if (!body) return;

        var d = data || {};
        var html = '';

        if (d.missing_skills && d.missing_skills.length) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Skills to Acquire</div>';
          d.missing_skills.forEach(function (gap) {
            var imp = gap.importance || 'important';
            html += '<div class="mm-gap-item">' +
              '<div class="mm-gap-skill">' + (gap.skill_name || gap.Name || '') + '</div>' +
              '<span class="mm-gap-importance ' + imp + '">' + imp.toUpperCase() + '</span>' +
              (gap.description ? '<div class="mm-gap-desc">' + gap.description + '</div>' : '') +
              (gap.learning_resources && gap.learning_resources.length
                ? '<div class="mm-gap-resources">Resources: ' + gap.learning_resources.map(function (r) { return '<a href="' + (typeof r === 'string' ? r : r.url || '#') + '" target="_blank">' + (typeof r === 'string' ? 'Learn' : r.name || 'Resource') + '</a>'; }).join(', ') + '</div>'
                : '') +
              '</div>';
          });
          html += '</div>';
        }

        if (d.transferable_skills && d.transferable_skills.length) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Transferable Skills</div><ul>' +
            d.transferable_skills.map(function (s) { return '<li>🔄 ' + s + '</li>'; }).join('') + '</ul></div>';
        }

        if (d.recommended_courses && d.recommended_courses.length) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Recommended Courses</div><ul>';
          d.recommended_courses.forEach(function (c) {
            html += '<li><a href="' + (c.url || '#') + '" target="_blank"><strong>' + (c.name || 'Course') + '</strong></a> — ' +
              (c.platform || '') + ' (' + (c.duration || '') + ', ' + (c.difficulty || '') + ')</li>';
          });
          html += '</ul></div>';
        }

        if (d.improvement_plan) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Improvement Plan</div><p>' + d.improvement_plan + '</p></div>';
        }

        if (d.estimated_time_to_fill) {
          html += '<div class="mm-analysis-section"><div class="mm-analysis-label">Estimated Time to Fill Gaps</div><p><strong>' + d.estimated_time_to_fill + '</strong></p></div>';
        }

        if (!html) html = '<p>No skills gap data available for this job.</p>';
        body.innerHTML = html;
      })
      .catch(handleModalError);
  }

  function handleCoverLetter(jobId) {
    if (!jobId) return;
    openModal('Generated Cover Letter');
    showModalLoading(true);

    var coverLetterText = '';

    function doFetch() {
      return AngaziaAPI.matches.coverLetter({ job_id: jobId });
    }

    doFetch()
      .then(function (data) {
        showModalLoading(false);
        var body = document.getElementById('mm-modal-body');
        if (!body) return;

        coverLetterText = (data && data.cover_letter) || (data && typeof data === 'string' ? data : '');
        var html = '<div class="mm-analysis-section">' +
          '<div style="display:flex;justify-content:flex-end;margin-bottom:12px">' +
          '<button class="emp-btn-sm emp-btn-outline" onclick="navigator.clipboard.writeText(\'' + coverLetterText.replace(/'/g, "\\'") + '\').then(function(){showToast(\'Copied!\',\'success\')}).catch(function(){})">📋 Copy</button>' +
          '</div>' +
          '<div style="white-space:pre-wrap;font-size:13px;line-height:1.7;color:var(--text);padding:16px;background:var(--s2);border:1px solid var(--border);border-radius:8px;">' +
          (coverLetterText || 'No cover letter generated.') +
          '</div></div>';
        body.innerHTML = html;
      })
      .catch(handleModalError);
  }

  function openModal(title) {
    var overlay = document.getElementById('mm-modal-overlay');
    var titleEl = document.getElementById('mm-modal-title');
    var body = document.getElementById('mm-modal-body');
    if (overlay) overlay.style.display = '';
    if (titleEl) titleEl.textContent = title || 'Analysis';
    if (body) body.innerHTML = '';
  }

  function closeModal() {
    var overlay = document.getElementById('mm-modal-overlay');
    var loading = document.getElementById('mm-modal-loading');
    if (overlay) overlay.style.display = 'none';
    if (loading) loading.style.display = 'none';
  }

  function showModalLoading(on) {
    var loading = document.getElementById('mm-modal-loading');
    var body = document.getElementById('mm-modal-body');
    if (loading) loading.style.display = on ? '' : 'none';
    if (body && on) body.innerHTML = '';
  }

  function handleModalError(err) {
    showModalLoading(false);
    var body = document.getElementById('mm-modal-body');
    if (body) body.innerHTML = '<div class="mm-alert mm-alert-error">⚠️ ' + (err && err.message ? err.message : 'Failed to load data') + '</div>';
  }

  function scoreColor(score) {
    if (score >= 80) return 'var(--accent)';
    if (score >= 60) return '#3b82f6';
    if (score >= 40) return 'var(--warn)';
    return '#ef4444';
  }
})();
