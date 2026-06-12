(function () {
  'use strict';

  document.addEventListener('DOMContentLoaded', function () {
    setupEventListeners();
    loadGitHubData();
  });

  function setupEventListeners() {
    document.addEventListener('click', function (e) {
      var btn = e.target.closest('[data-action]');
      if (!btn) return;
      var action = btn.getAttribute('data-action');
      switch (action) {
        case 'apply-job':
          handleApplyJob(btn.getAttribute('data-id'));
          break;
        case 'confirm-interview':
          handleConfirmInterview(btn.getAttribute('data-id'));
          break;
        case 'reschedule-interview':
          handleRescheduleInterview(btn.getAttribute('data-id'));
          break;
        case 'add-calendar':
          handleAddToCalendar(btn.getAttribute('data-id'));
          break;
        case 'complete-profile':
          window.location.href = '/employee/settings';
          break;
        case 'upload-resume':
          window.location.href = '/employee/skills';
          break;
        case 'github-connect':
          window.location.href = '/employee/skills';
          break;
        case 'skill-assessment':
          window.location.href = '/employee/skills';
          break;
        case 'refresh-recommendations':
          handleRefreshRecommendations(btn);
          break;
        case 'view-all-activity':
          window.location.href = '/employee/applications';
          break;
        case 'analyze-skills':
          window.location.href = '/employee/skills';
          break;
        case 'learn-skill':
          handleLearnSkill(btn.getAttribute('data-skill'));
          break;
        case 'github-sync':
          handleGitHubSync(btn);
          break;
      }
    });
  }

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type || 'info');
    } else {
      console.log('[' + (type || 'info') + '] ' + msg);
    }
  }

  /* ========== GitHub Widget ========== */

  function loadGitHubData() {
    var card = document.getElementById('gh-card');
    var notConnected = document.getElementById('gh-not-connected');
    var loading = document.getElementById('gh-loading');
    var connected = document.getElementById('gh-connected');

    if (!card) return;

    // Check if GitHub is connected via the page's User data
    var userEl = document.querySelector('[data-user-github]');
    var githubConnected = userEl ? userEl.getAttribute('data-user-github') === 'true' : null;

    if (githubConnected === false) {
      notConnected.style.display = '';
      return;
    }

    loading.style.display = '';

    AngaziaAPI.github.profile()
      .then(function (profile) {
        loading.style.display = 'none';
        connected.style.display = '';
        renderGitHubProfile(profile);

        // Also load contributions
        AngaziaAPI.github.contributions(365)
          .then(function (contribs) {
            renderGitHubContributions(contribs);
          })
          .catch(function () {});
      })
      .catch(function () {
        loading.style.display = 'none';
        notConnected.style.display = '';
      });
  }

  function renderGitHubProfile(profile) {
    if (!profile) return;

    var avatarEl = document.getElementById('gh-avatar');
    var usernameEl = document.getElementById('gh-username');
    var bioEl = document.getElementById('gh-bio');
    var reposEl = document.getElementById('gh-repos');
    var followersEl = document.getElementById('gh-followers');

    if (avatarEl) {
      avatarEl.src = profile.github_avatar || '';
      avatarEl.alt = profile.github_username || 'GitHub';
    }
    if (usernameEl) {
      usernameEl.textContent = profile.github_username || '';
      usernameEl.href = profile.github_url || 'https://github.com/' + (profile.github_username || '');
    }
    if (bioEl) {
      bioEl.textContent = profile.github_bio || '';
    }
    if (reposEl) reposEl.textContent = profile.public_repos || 0;
    if (followersEl) followersEl.textContent = profile.followers || 0;

    var syncedEl = document.getElementById('gh-synced');
    if (syncedEl && profile.last_synced_at) {
      var d = new Date(profile.last_synced_at);
      syncedEl.textContent = 'Last synced: ' + d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
  }

  function renderGitHubContributions(contribs) {
    if (!contribs) return;

    var streakEl = document.getElementById('gh-streak');
    var commitsEl = document.getElementById('gh-commits');
    var activityEl = document.getElementById('gh-activity');

    if (streakEl) streakEl.textContent = contribs.current_streak || 0;
    if (commitsEl) commitsEl.textContent = contribs.total_commits || 0;
    if (activityEl && contribs.activity_level) activityEl.textContent = contribs.activity_level;
  }

  function handleGitHubSync(btn) {
    if (!btn) return;
    btn.disabled = true;
    btn.textContent = 'Syncing...';

    AngaziaAPI.github.sync()
      .then(function () {
        // Reload GitHub data after a short delay
        setTimeout(loadGitHubData, 3000);
      })
      .catch(function (err) {
        console.error('Sync failed:', err);
      })
      .finally(function () {
        btn.disabled = false;
        btn.textContent = '⟳ Sync';
      });
  }

  /* ========== Job Applications ========== */

  function handleApplyJob(jobId) {
    if (!jobId) return;
    if (window.applyToJob) {
      window.applyToJob(jobId);
    } else {
      AngaziaAPI.applications.apply({ job_id: jobId })
        .then(function () { })
        .catch(function (err) {
          console.error('Apply error:', err);
        });
    }
  }

  function handleConfirmInterview(interviewId) {
    if (!interviewId) return;
    if (!confirm('Confirm your attendance for this interview?')) return;
    AngaziaAPI.applications.interview(interviewId, { status: 'confirmed' })
      .then(function () { })
      .catch(function (err) {
        console.error('Confirm error:', err);
      });
  }

  function handleRescheduleInterview(interviewId) {
    if (!interviewId) return;
    window.location.href = '/employee/applications?filter=interview';
  }

  function handleAddToCalendar(interviewId) {
    if (!interviewId) return;
    var card = document.querySelector('[data-id="' + interviewId + '"]');
    if (!card) { showToast('Could not find interview details', 'error'); return; }
    var item = card.closest('.emp-interview-item');
    if (!item) return;
    var roleEl = item.querySelector('.emp-interview-role');
    var companyEl = item.querySelector('.emp-interview-company');
    var title = (roleEl ? roleEl.textContent : 'Interview') + ' at ' + (companyEl ? companyEl.textContent : 'Company');
    var startTime = new Date().toISOString().replace(/[-:]/g, '').split('.')[0] + 'Z';
    var endTime = new Date(Date.now() + 3600000).toISOString().replace(/[-:]/g, '').split('.')[0] + 'Z';
    var icsContent = [
      'BEGIN:VCALENDAR',
      'VERSION:2.0',
      'PRODID:-//Angazia//EN',
      'BEGIN:VEVENT',
      'SUMMARY:' + title,
      'DTSTART:' + startTime,
      'DTEND:' + endTime,
      'DESCRIPTION:Interview scheduled via Angazia',
      'END:VEVENT',
      'END:VCALENDAR'
    ].join('\r\n');
    var blob = new Blob([icsContent], { type: 'text/calendar;charset=utf-8' });
    var link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = 'interview.ics';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(link.href);
    showToast('Calendar file downloaded!', 'success');
  }

  function handleRefreshRecommendations(btn) {
    btn.disabled = true;
    btn.textContent = '⟳ Refreshing...';
    AngaziaAPI.matches.jobMatches({ limit: 6 })
      .then(function (matches) {
        showToast('Recommendations refreshed!', 'success');
        updateRecommendations(matches || []);
      })
      .catch(function (err) {
        console.error('Refresh error:', err);
        showToast('Failed to refresh recommendations', 'error');
      })
      .finally(function () {
        btn.disabled = false;
        btn.textContent = '🔄 Refresh';
      });
  }

  function updateRecommendations(matches) {
    var container = document.getElementById('recommendations-list');
    if (!container) return;
    if (!matches || matches.length === 0) {
      container.innerHTML =
        '<div class="emp-empty-state"><span class="emp-empty-icon">🔍</span><p>No recommendations yet. Complete your profile to get AI-powered job matches.</p>' +
        '<a href="/employee/settings" class="emp-btn-sm emp-btn-outline">Complete Profile →</a></div>';
      return;
    }
    var html = '';
    matches.forEach(function (m) {
      var jobId = m.job_id || m.JobID || '';
      var title = m.job_title || m.JobTitle || 'Position';
      var company = m.company_name || m.CompanyName || '';
      var logo = m.company_logo || m.CompanyLogo || '';
      var score = m.overall_score || m.OverallScore || 0;
      var mSkills = m.matching_skills || m.MatchingSkills || [];
      var posted = m.analyzed_at || m.AnalyzedAt || '';
      if (posted && typeof posted === 'string') {
        posted = new Date(posted).toLocaleDateString('en-KE', { day: 'numeric', month: 'short' });
      }
      var ci = company.split(' ').map(function (w) { return w.charAt(0); }).join('').toUpperCase().slice(0, 2);
      html += '<div class="emp-rec-item" data-job-id="' + jobId + '">' +
        '<div class="emp-rec-header">' +
        (logo ? '<img class="emp-rec-logo" src="' + logo + '" alt="' + company + '">' : '<div class="emp-rec-logo-placeholder">' + ci + '</div>') +
        '<div class="emp-rec-info"><h4 class="emp-rec-title">' + title + '</h4>' +
        '<p class="emp-rec-company">' + company + '</p></div>' +
        '<div class="emp-rec-match-score">' +
        '<svg class="emp-score-ring" viewBox="0 0 36 36"><path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#2a3a3a" stroke-width="3"/>' +
        '<path d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="#00e5a0" stroke-width="3" stroke-dasharray="' + score + ', 100"/></svg>' +
        '<span class="emp-score-text">' + score + '%</span></div></div>' +
        '<div class="emp-rec-skills">' +
        mSkills.map(function (s) { return '<span class="emp-skill-match">✓ ' + s + '</span>'; }).join('') +
        '</div>' +
        (posted ? '<div class="emp-rec-date">📅 ' + posted + '</div>' : '') +
        '<div class="emp-rec-footer"><button class="emp-btn-sm emp-btn-primary" data-action="apply-job" data-id="' + jobId + '">Apply Now →</button></div></div>';
    });
    container.innerHTML = html;
  }

  function handleLearnSkill(skill) {
    if (!skill) return;
    window.open('https://www.coursera.org/search?query=' + encodeURIComponent(skill), '_blank');
  }

  window.applyToJob = function (jobId) {
    if (!jobId) return;
    AngaziaAPI.applications.apply({ job_id: jobId })
      .then(function () { })
      .catch(function (err) {
        console.error('Apply error:', err);
      });
  };
})();
