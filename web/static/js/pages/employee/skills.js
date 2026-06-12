(function () {
  'use strict';

  var state = {
    skills: [],
    experience: [],
    certifications: [],
    suggestions: [],
    gap: null,
    saving: false,
  };

  var els = {};

  var CATEGORY_MAP = {
    go: 'backend', python: 'backend', java: 'backend', rust: 'backend', c: 'backend',
    'c++': 'backend', 'c#': 'backend', ruby: 'backend', php: 'backend', scala: 'backend',
    kotlin: 'backend', swift: 'mobile', typescript: 'frontend', javascript: 'frontend',
    react: 'frontend', vue: 'frontend', angular: 'frontend', svelte: 'frontend',
    html: 'frontend', css: 'frontend', 'next.js': 'frontend', 'nuxt.js': 'frontend',
    docker: 'devops', kubernetes: 'devops', terraform: 'devops', jenkins: 'devops',
    ci: 'devops', cd: 'devops', ansible: 'devops', 'ci/cd': 'devops', aws: 'devops',
    gcp: 'devops', azure: 'devops', linux: 'devops', nginx: 'devops',
    postgresql: 'data', mysql: 'data', mongodb: 'data', redis: 'data', sql: 'data',
    elasticsearch: 'data', bigquery: 'data', spark: 'data', hadoop: 'data',
    tensorflow: 'data', pytorch: 'data', 'machine learning': 'data', 'deep learning': 'data',
    figma: 'design', sketch: 'design', 'ui/ux': 'design', photoshop: 'design',
    illustrator: 'design', 'user research': 'design', wireframing: 'design',
    git: 'tool', graphql: 'tool', rest: 'tool', grpc: 'tool', webpack: 'tool',
    vite: 'tool', jest: 'tool', cypress: 'tool', docker: 'devops',
    communication: 'soft', leadership: 'soft', teamwork: 'soft', 'problem solving': 'soft',
    'critical thinking': 'soft', 'time management': 'soft',
  };

  var CATEGORY_LABELS = {
    frontend: 'Frontend', backend: 'Backend', devops: 'DevOps',
    data: 'Data', mobile: 'Mobile', design: 'Design',
    language: 'Language', tool: 'Tool', soft: 'Soft Skill',
  };

  var CATEGORY_COLORS = {
    frontend: '#3b82f6', backend: '#10b981', devops: '#f59e0b',
    data: '#8b5cf6', mobile: '#ec4899', design: '#06b6d4',
    language: '#f97316', tool: '#6b7280', soft: '#84cc16',
  };

  function init() {
    els = {
      loading: document.getElementById('sk-loading'),
      error: document.getElementById('sk-error'),
      errorMsg: document.getElementById('sk-error-msg'),
      content: document.getElementById('sk-content'),
      cloud: document.getElementById('sk-cloud'),
      expList: document.getElementById('sk-exp-list'),
      certList: document.getElementById('sk-cert-list'),
      strength: document.getElementById('sk-strength'),
      gapContent: document.getElementById('sk-gap-content'),
      gapSection: document.getElementById('sk-gap-section'),
      matchesContent: document.getElementById('sk-matches-content'),
      matchesSection: document.getElementById('sk-matches-section'),
      countBadge: document.getElementById('sk-count-badge'),
      expCount: document.getElementById('sk-exp-count'),
      certCount: document.getElementById('sk-cert-count'),
      addBtn: document.getElementById('sk-add-btn'),
      addRow: document.getElementById('sk-add-row'),
      addInput: document.getElementById('sk-add-input'),
      addCategory: document.getElementById('sk-add-category'),
      addConfirm: document.getElementById('sk-add-confirm'),
      addCancel: document.getElementById('sk-add-cancel'),
      addExpBtn: document.getElementById('sk-add-exp-btn'),
      addCertBtn: document.getElementById('sk-add-cert-btn'),
      suggestSection: document.getElementById('sk-suggestions'),
      suggestToggle: document.getElementById('sk-suggest-toggle'),
      suggestBody: document.getElementById('sk-suggest-body'),
      modal: document.getElementById('sk-modal'),
      modalClose: document.getElementById('sk-modal-close'),
      modalCancel: document.getElementById('sk-modal-cancel'),
      modalSave: document.getElementById('sk-modal-save'),
      modalTitle: document.getElementById('sk-modal-title'),
      modalMode: document.getElementById('sk-modal-mode'),
      modalIdx: document.getElementById('sk-modal-idx'),
      modalExpFields: document.getElementById('sk-modal-exp-fields'),
      modalCertFields: document.getElementById('sk-modal-cert-fields'),
      expTitle: document.getElementById('sk-exp-title'),
      expCompany: document.getElementById('sk-exp-company'),
      expStart: document.getElementById('sk-exp-start'),
      expEnd: document.getElementById('sk-exp-end'),
      expDesc: document.getElementById('sk-exp-desc'),
      expCurrent: document.getElementById('sk-exp-current'),
      certName: document.getElementById('sk-cert-name'),
      certIssuer: document.getElementById('sk-cert-issuer'),
      certYear: document.getElementById('sk-cert-year'),
      toast: document.getElementById('sk-toast'),
      gapBtn: document.getElementById('skill-gap-btn'),
    };

    els.addBtn.addEventListener('click', showAddRow);
    els.addConfirm.addEventListener('click', confirmAddSkill);
    els.addCancel.addEventListener('click', hideAddRow);
    els.addInput.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') confirmAddSkill();
      if (e.key === 'Escape') hideAddRow();
    });
    els.addExpBtn.addEventListener('click', function () { openModal('experience'); });
    els.addCertBtn.addEventListener('click', function () { openModal('certification'); });
    els.expCurrent.addEventListener('change', function () {
      els.expEnd.disabled = els.expCurrent.checked;
      if (els.expCurrent.checked) els.expEnd.value = '';
    });
    els.modalClose.addEventListener('click', closeModal);
    els.modalCancel.addEventListener('click', closeModal);
    els.modalSave.addEventListener('click', saveModal);
    els.modal.addEventListener('click', function (e) { if (e.target === els.modal) closeModal(); });
    els.suggestToggle.addEventListener('click', toggleSuggestions);
    els.gapBtn.addEventListener('click', function () {
      var sec = els.gapSection;
      sec.style.display = sec.style.display === 'none' ? '' : 'none';
      if (sec.style.display !== 'none' && state.gap) renderGap();
    });

    initResumeSection();
    initGitHubSection();

    loadAll();
  }

  /* ========== Resume Upload ========== */

  function initResumeSection() {
    els.resumeInfo = document.getElementById('sk-resume-info');
    els.resumeEmpty = document.getElementById('sk-resume-empty');
    els.resumeName = document.getElementById('sk-resume-name');
    els.resumeSize = document.getElementById('sk-resume-size');
    els.resumeFile = document.getElementById('sk-resume-file');
    els.resumeUploadBtn = document.getElementById('sk-resume-upload-btn');
    els.resumeReplaceBtn = document.getElementById('sk-resume-replace');
    els.resumeRemoveBtn = document.getElementById('sk-resume-remove');
    els.resumeProgress = document.getElementById('sk-resume-progress');
    els.resumeProgressBar = document.getElementById('sk-resume-progress-bar');
    els.resumeProgressText = document.getElementById('sk-resume-progress-text');

    if (!els.resumeUploadBtn) return;

    els.resumeUploadBtn.addEventListener('click', function () { els.resumeFile.click(); });
    if (els.resumeReplaceBtn) els.resumeReplaceBtn.addEventListener('click', function () { els.resumeFile.click(); });
    if (els.resumeRemoveBtn) {
      els.resumeRemoveBtn.addEventListener('click', function () {
        AngaziaAPI.profile.update({ resume_url: '' }).then(function () {
          state.profile.resume_url = '';
          showResumeState(false);
        }).catch(function (err) {
        });
      });
    }
    els.resumeFile.addEventListener('change', function () {
      var file = els.resumeFile.files[0];
      if (!file) return;
      if (file.size > 10 * 1024 * 1024) {
        showToast('File must be under 10MB', true);
        return;
      }
      var fd = new FormData();
      fd.append('resume', file);
      els.resumeProgress.style.display = '';
      els.resumeProgressBar.style.width = '0%';
      els.resumeProgressText.textContent = 'Uploading...';
      AngaziaAPI.profile.uploadResume(fd, function (pct) {
        els.resumeProgressBar.style.width = pct + '%';
        els.resumeProgressText.textContent = pct + '%';
      }).then(function (data) {
        els.resumeProgress.style.display = 'none';
        var url = data && data.url ? data.url : (data && data.resume_url ? data.resume_url : '');
        if (url) {
          state.profile.resume_url = url;
          showResumeState(true, file.name, file.size);
        }
      }).catch(function (err) {
        els.resumeProgress.style.display = 'none';
      });
    });
  }

  function showResumeState(hasResume, name, size) {
    if (!els.resumeInfo || !els.resumeEmpty) return;
    if (hasResume) {
      els.resumeInfo.style.display = '';
      els.resumeEmpty.style.display = 'none';
      if (els.resumeName) els.resumeName.textContent = name || 'Resume';
      if (els.resumeSize) els.resumeSize.textContent = size ? formatFileSize(size) : '';
    } else {
      els.resumeInfo.style.display = 'none';
      els.resumeEmpty.style.display = '';
    }
  }

  function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1048576).toFixed(1) + ' MB';
  }

  /* ========== GitHub Integration ========== */

  function initGitHubSection() {
    els.ghNotConnected = document.getElementById('sk-github-not-connected');
    els.ghConnected = document.getElementById('sk-github-connected');
    els.ghLoading = document.getElementById('sk-github-loading');
    els.ghUsername = document.getElementById('sk-github-username');
    els.ghConnectBtn = document.getElementById('sk-github-connect-btn');
    els.ghDisconnectBtn = document.getElementById('sk-github-disconnect-btn');
    els.ghSyncBtn = document.getElementById('sk-github-sync-btn');

    if (!els.ghConnectBtn) return;

    els.ghConnectBtn.addEventListener('click', function () {
      AngaziaAPI.github.connect().then(function (data) {
        if (data && data.auth_url) {
          window.location.href = data.auth_url;
        }
      }).catch(function (err) {
        showToast(err && err.message ? err.message : 'Failed to connect GitHub', true);
      });
    });

    if (els.ghDisconnectBtn) {
      els.ghDisconnectBtn.addEventListener('click', function () {
        if (!confirm('Disconnect your GitHub account? This will remove your GitHub data from your profile.')) return;
        AngaziaAPI.github.disconnect().then(function () {
          state.profile.github_connected = false;
          state.profile.github_username = '';
          showGitHubState(false);
        }).catch(function (err) {
        });
      });
    }

    if (els.ghSyncBtn) {
      els.ghSyncBtn.addEventListener('click', function () {
        els.ghSyncBtn.disabled = true;
        els.ghSyncBtn.textContent = 'Syncing...';
        AngaziaAPI.github.sync().then(function () {
          els.ghSyncBtn.disabled = false;
          els.ghSyncBtn.textContent = '\u21BB Sync';
        }).catch(function (err) {
          els.ghSyncBtn.disabled = false;
          els.ghSyncBtn.textContent = '\u21BB Sync';
        });
      });
    }
  }

  function showGitHubState(connected) {
    if (!els.ghNotConnected || !els.ghConnected) return;
    els.ghNotConnected.style.display = connected ? 'none' : '';
    els.ghConnected.style.display = connected ? '' : 'none';
  }

  /* ========== Loading ========== */

  async function loadAll() {
    showLoading(true);
    showError(false);
    try {
      var [profileData, suggestedData] = await Promise.all([
        AngaziaAPI.profile.get(),
        AngaziaAPI.profile.suggestedSkills().catch(function () { return { skills: [] }; }),
      ]);
      var profile = profileData && profileData.data ? profileData.data : profileData;
      var employee = profile.employee_profile || {};
      state.skills = employee.skills || [];
      state.experience = employee.experience || [];
      state.certifications = employee.certifications || [];
      state.suggestions = (suggestedData && suggestedData.skills) || (suggestedData && suggestedData.data) || [];
      state.profile = employee;
      renderAll();
      showResumeState(!!employee.resume_url);
      showGitHubState(!!employee.github_connected);
      if (employee.github_connected && els.ghUsername) {
        els.ghUsername.textContent = employee.github_username || 'GitHub Connected';
      }
      showLoading(false);
      showContent();
      loadGap();
    } catch (err) {
      showLoading(false);
      showError(true, err.message || 'Failed to load profile');
    }
  }

  async function loadGap() {
    try {
      var data = await AngaziaAPI.analytics.skillGap();
      state.gap = data && data.data ? data.data : data;
      renderGap();
    } catch (_) { /* gap optional */ }
  }

  /* ========== Render All ========== */

  function renderAll() {
    renderSkills();
    renderExperience();
    renderCertifications();
    renderStrength();
    updateCounts();
    renderSuggestions();
  }

  function updateCounts() {
    if (els.countBadge) els.countBadge.textContent = state.skills.length;
    if (els.expCount) els.expCount.textContent = state.experience.length;
    if (els.certCount) els.certCount.textContent = state.certifications.length;
  }

  /* ========== Skills ========== */

  function getCategory(skill) {
    return CATEGORY_MAP[skill.toLowerCase()] || '';
  }

  function getCategoryColor(skill) {
    var cat = getCategory(skill);
    return CATEGORY_COLORS[cat] || 'var(--muted2)';
  }

  function renderSkills() {
    if (!els.cloud) return;
    if (!state.skills || state.skills.length === 0) {
      els.cloud.innerHTML = '<div class="sk-cloud-empty">No skills added yet. Click "+ Add Skill" above.</div>';
      return;
    }
    els.cloud.innerHTML = state.skills.map(function (s, idx) {
      var color = getCategoryColor(s);
      return '<span class="sk-skill" style="border-color:' + color + '40">' +
        '<span style="width:6px;height:6px;border-radius:50%;background:' + color + ';flex-shrink:0"></span>' +
        esc(s) +
        '<button class="sk-skill-rm" data-idx="' + idx + '" title="Remove">&times;</button>' +
        '</span>';
    }).join('');
    els.cloud.querySelectorAll('.sk-skill-rm').forEach(function (btn) {
      btn.addEventListener('click', function (e) { e.stopPropagation(); removeSkill(parseInt(this.getAttribute('data-idx'), 10)); });
    });
  }

  function showAddRow() {
    els.addBtn.style.display = 'none';
    els.addRow.style.display = 'flex';
    els.addInput.value = '';
    els.addCategory.value = '';
    els.addInput.focus();
  }

  function hideAddRow() {
    els.addRow.style.display = 'none';
    els.addBtn.style.display = '';
    els.addInput.value = '';
  }

  function confirmAddSkill() {
    var name = els.addInput.value.trim();
    if (!name) { els.addInput.focus(); return; }
    state.skills.push(name);
    renderSkills();
    hideAddRow();
    autoSave();
  }

  function removeSkill(idx) {
    state.skills.splice(idx, 1);
    renderSkills();
    autoSave();
  }

  /* ========== Suggestions ========== */

  function renderSuggestions() {
    if (!els.suggestSection || !state.suggestions || state.suggestions.length === 0) {
      if (els.suggestSection) els.suggestSection.style.display = 'none';
      return;
    }
    els.suggestSection.style.display = '';
    els.suggestBody.innerHTML = state.suggestions.map(function (s) {
      return '<span class="sk-suggest-chip" data-skill="' + esc(s) + '"><span class="sk-suggest-add">+</span> ' + esc(s) + '</span>';
    }).join('');
    els.suggestBody.querySelectorAll('.sk-suggest-chip').forEach(function (chip) {
      chip.addEventListener('click', function () {
        var name = this.getAttribute('data-skill');
        if (name && state.skills.indexOf(name) === -1) {
          state.skills.push(name);
          renderSkills();
          autoSave();
          this.remove();
          if (els.suggestBody.children.length === 0) els.suggestSection.style.display = 'none';
        }
      });
    });
  }

  function toggleSuggestions() {
    var show = els.suggestBody.style.display !== 'block';
    els.suggestBody.style.display = show ? 'block' : 'none';
    els.suggestToggle.textContent = show ? 'Hide' : 'Show';
  }

  /* ========== Experience ========== */

  function renderExperience() {
    if (!els.expList) return;
    if (!state.experience || state.experience.length === 0) {
      els.expList.innerHTML = '<div class="sk-tl-empty">No experience listed yet. Click "+ Add" above.</div>';
      return;
    }
    els.expList.innerHTML = state.experience.map(function (e, idx) {
      var start = e.start_date || '';
      var end = e.end_date || (e.current ? 'Present' : '');
      var dateStr = start && end ? start + ' \u2013 ' + end : start || end;
      return '<div class="sk-tl-item">' +
        '<div class="sk-tl-dot"></div>' +
        '<div class="sk-tl-head">' +
        '<div>' +
        '<div class="sk-tl-title">' + esc(e.title || '') + '</div>' +
        '<div class="sk-tl-org">' + esc(e.company || '') + (dateStr ? ' \u00B7 ' + esc(dateStr) : '') + '</div>' +
        (e.description ? '<div class="sk-tl-desc">' + esc(e.description) + '</div>' : '') +
        '</div>' +
        '<div class="sk-tl-actions">' +
        '<button class="sk-btn-icon sk-edit-exp" data-idx="' + idx + '" title="Edit">&#x270E;</button>' +
        '<button class="sk-btn-icon danger sk-del-exp" data-idx="' + idx + '" title="Delete">&times;</button>' +
        '</div>' +
        '</div>' +
        '</div>';
    }).join('');
    els.expList.querySelectorAll('.sk-edit-exp').forEach(function (btn) {
      btn.addEventListener('click', function () { editExperience(parseInt(this.getAttribute('data-idx'), 10)); });
    });
    els.expList.querySelectorAll('.sk-del-exp').forEach(function (btn) {
      btn.addEventListener('click', function () {
        state.experience.splice(parseInt(this.getAttribute('data-idx'), 10), 1);
        renderExperience();
        autoSave();
      });
    });
  }

  function editExperience(idx) {
    var e = state.experience[idx];
    els.expTitle.value = e.title || '';
    els.expCompany.value = e.company || '';
    els.expStart.value = e.start_date || '';
    els.expEnd.value = e.end_date || '';
    els.expDesc.value = e.description || '';
    els.expCurrent.checked = e.current || false;
    els.expEnd.disabled = els.expCurrent.checked;
    openModal('experience', idx);
  }

  /* ========== Certifications ========== */

  function renderCertifications() {
    if (!els.certList) return;
    if (!state.certifications || state.certifications.length === 0) {
      els.certList.innerHTML = '<div class="sk-cert-empty">No certifications added. Click "+ Add" above.</div>';
      return;
    }
    els.certList.innerHTML = state.certifications.map(function (c, idx) {
      return '<div class="sk-cert">' +
        '<div class="sk-cert-icon">&#x1F4DC;</div>' +
        '<div class="sk-cert-info">' +
        '<div class="sk-cert-name">' + esc(c.name || '') + '</div>' +
        '<div class="sk-cert-meta">' + esc(c.issuer || '') + (c.year ? ' \u00B7 ' + esc(c.year) : '') + '</div>' +
        '</div>' +
        '<div class="sk-cert-actions">' +
        '<button class="sk-btn-icon sk-edit-cert" data-idx="' + idx + '" title="Edit">&#x270E;</button>' +
        '<button class="sk-btn-icon danger sk-del-cert" data-idx="' + idx + '" title="Delete">&times;</button>' +
        '</div>' +
        '</div>';
    }).join('');
    els.certList.querySelectorAll('.sk-edit-cert').forEach(function (btn) {
      btn.addEventListener('click', function () { editCertification(parseInt(this.getAttribute('data-idx'), 10)); });
    });
    els.certList.querySelectorAll('.sk-del-cert').forEach(function (btn) {
      btn.addEventListener('click', function () {
        state.certifications.splice(parseInt(this.getAttribute('data-idx'), 10), 1);
        renderCertifications();
        autoSave();
      });
    });
  }

  function editCertification(idx) {
    var c = state.certifications[idx];
    els.certName.value = c.name || '';
    els.certIssuer.value = c.issuer || '';
    els.certYear.value = c.year || '';
    openModal('certification', idx);
  }

  /* ========== Profile Strength ========== */

  function renderStrength() {
    if (!els.strength) return;
    var p = state.profile || {};
    var score = p.profile_strength || 0;
    var skillsPct = p.skills_match_percent || calcSimplePct(p.skills && p.skills.length, 5);
    var expPct = p.experience_match_percent || (p.years_of_experience > 0 || (p.experience && p.experience.length > 0) ? 80 : 0);
    var locPct = p.location_match_percent || (p.location ? 100 : 0);
    var breakdown = [
      { label: 'Skills', pct: skillsPct, fill: '#00e5a0' },
      { label: 'Experience', pct: expPct, fill: '#3b82f6' },
      { label: 'Location', pct: locPct, fill: '#8b5cf6' },
    ];
    var r = 40, circ = 2 * Math.PI * r, offset = circ - (circ * score / 100);
    els.strength.innerHTML =
      '<div class="sk-str-ring">' +
      '<svg width="100" height="100" viewBox="0 0 100 100">' +
      '<circle cx="50" cy="50" r="' + r + '" fill="none" stroke="var(--border)" stroke-width="7"/>' +
      '<circle cx="50" cy="50" r="' + r + '" fill="none" stroke="#00e5a0" stroke-width="7" stroke-dasharray="' + circ + '" stroke-dashoffset="' + offset + '" stroke-linecap="round" transform="rotate(-90 50 50)" style="transition:stroke-dashoffset 0.6s"/>' +
      '<text x="50" y="46" text-anchor="middle" fill="var(--text)" font-family="var(--fh)" font-size="22" font-weight="700">' + score + '</text>' +
      '<text x="50" y="62" text-anchor="middle" fill="var(--muted2)" font-family="var(--fm)" font-size="8">Strength</text>' +
      '</svg>' +
      '</div>' +
      '<div class="sk-str-details">' +
      breakdown.map(function (b) {
        return '<div class="sk-str-row"><span class="sk-str-label">' + b.label + '</span><span class="sk-str-bar"><span style="width:' + b.pct + '%;background:' + b.fill + '"></span></span><span class="sk-str-val">' + b.pct + '%</span></div>';
      }).join('') +
      '</div>';
  }

  function calcSimplePct(val, max) {
    if (!val) return 0;
    var pct = Math.round((val / max) * 100);
    return pct > 100 ? 100 : pct;
  }

  /* ========== Skills Gap ========== */

  function renderGap() {
    if (!els.gapContent || !els.gapSection) return;
    var gap = state.gap;
    if (!gap) { els.gapSection.style.display = 'none'; return; }
    els.gapSection.style.display = '';

    var match = gap.matching_skills || [];
    var miss = gap.missing_skills || [];
    var rec = gap.recommended_skills || [];

    var html = '';

    if (match.length > 0) {
      html += '<div class="sk-gap-head">&#x2705; Matching Skills (' + match.length + ')</div>';
      html += '<div class="sk-gap-tags">' +
        match.slice(0, 8).map(function (s) { return '<span class="sk-gap-tag sk-gap-match">' + esc(s.name || s) + '</span>'; }).join('') +
        (match.length > 8 ? '<span class="sk-gap-tag">+' + (match.length - 8) + '</span>' : '') +
        '</div>';
    }

    if (miss.length > 0) {
      html += '<div class="sk-gap-head">&#x1F6A7; Missing Skills (' + miss.length + ')</div>';
      html += '<div class="sk-gap-tags">' +
        miss.slice(0, 8).map(function (s) { return '<span class="sk-gap-tag sk-gap-miss">' + esc(s.name || s) + '</span>'; }).join('') +
        (miss.length > 8 ? '<span class="sk-gap-tag">+' + (miss.length - 8) + '</span>' : '') +
        '</div>';
    }

    if (rec.length > 0) {
      html += '<div class="sk-gap-head">&#x1F4A1; Recommended to Learn</div>';
      html += '<div class="sk-gap-tags">' +
        rec.slice(0, 5).map(function (s) {
          return '<span class="sk-gap-tag sk-gap-rec" data-skill="' + esc(s.name || s) + '">+ ' + esc(s.name || s) + '</span>';
        }).join('') +
        '</div>';
    }

    if (!html) html = '<div class="sk-loading-sm">No gap data available yet.</div>';
    els.gapContent.innerHTML = html;

    els.gapContent.querySelectorAll('.sk-gap-rec').forEach(function (chip) {
      chip.addEventListener('click', function () {
        var name = this.getAttribute('data-skill');
        if (name && state.skills.indexOf(name) === -1) {
          state.skills.push(name);
          renderSkills();
          autoSave();
          this.remove();
        }
      });
    });

    renderJobMatches();
  }

  /* ========== Job Matches ========== */

  function renderJobMatches() {
    if (!els.matchesContent || !els.matchesSection) return;
    var gap = state.gap;
    var matches = (gap && gap.top_job_matches) || [];
    if (!matches || matches.length === 0) { els.matchesSection.style.display = 'none'; return; }
    els.matchesSection.style.display = '';

    els.matchesContent.innerHTML = matches.slice(0, 4).map(function (m) {
      var pct = m.match_score || 0;
      return '<div class="sk-match-card">' +
        '<div class="sk-match-title">' + esc(m.title || '') + '</div>' +
        '<div class="sk-match-company">' + esc(m.company || '') + '</div>' +
        '<div class="sk-match-score">' +
        '<span class="sk-str-bar"><span style="width:' + pct + '%;background:' + (pct >= 70 ? '#00e5a0' : pct >= 40 ? '#f59e0b' : '#ef4444') + '"></span></span>' +
        '<span class="sk-str-val">' + pct + '%</span>' +
        '</div>' +
        '</div>';
    }).join('');
  }

  /* ========== Modal ========== */

  function openModal(mode, editIdx) {
    els.modalMode.value = mode;
    els.modalIdx.value = editIdx != null ? editIdx : -1;
    if (mode === 'experience') {
      els.modalTitle.textContent = editIdx != null ? 'Edit Experience' : 'Add Experience';
      els.modalExpFields.style.display = '';
      els.modalCertFields.style.display = 'none';
    } else {
      els.modalTitle.textContent = editIdx != null ? 'Edit Certification' : 'Add Certification';
      els.modalExpFields.style.display = 'none';
      els.modalCertFields.style.display = '';
      if (editIdx == null) {
        els.certName.value = '';
        els.certIssuer.value = '';
        els.certYear.value = '';
      }
    }
    els.modal.style.display = 'flex';
  }

  function closeModal() {
    els.modal.style.display = 'none';
    els.modalIdx.value = -1;
  }

  function saveModal() {
    var mode = els.modalMode.value;
    var idx = parseInt(els.modalIdx.value, 10);
    if (mode === 'experience') {
      var title = els.expTitle.value.trim();
      var company = els.expCompany.value.trim();
      if (!title || !company) return;
      var item = {
        title: title,
        company: company,
        start_date: els.expStart.value.trim(),
        end_date: els.expCurrent.checked ? '' : els.expEnd.value.trim(),
        current: els.expCurrent.checked,
        description: els.expDesc.value.trim(),
      };
      if (idx >= 0 && idx < state.experience.length) {
        state.experience[idx] = item;
      } else {
        state.experience.push(item);
      }
    } else {
      var name = els.certName.value.trim();
      if (!name) return;
      var item = {
        name: name,
        issuer: els.certIssuer.value.trim(),
        year: els.certYear.value.trim(),
      };
      if (idx >= 0 && idx < state.certifications.length) {
        state.certifications[idx] = item;
      } else {
        state.certifications.push(item);
      }
    }
    closeModal();
    renderExperience();
    renderCertifications();
    renderStrength();
    updateCounts();
    autoSave();
  }

  /* ========== Auto Save ========== */

  function autoSave() {
    if (state.saving) return;
    state.saving = true;
    updateCounts();
    var updateData = {
      skills: state.skills,
      experience: state.experience,
      certifications: state.certifications,
    };
    if (state.profile && state.profile.resume_url) {
      updateData.resume_url = state.profile.resume_url;
    }
    AngaziaAPI.profile.update(updateData).then(function () {
      state.saving = false;
    }).catch(function (err) {
      state.saving = false;
    });
  }

  /* ========== Toast ========== */

  function showToast(msg, isError) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, isError ? 'error' : 'success');
    }
  }

  /* ========== Helpers ========== */

  function esc(text) {
    if (!text) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(text));
    return d.innerHTML;
  }

  function showLoading(show) {
    if (els.loading) els.loading.style.display = show ? 'flex' : 'none';
    if (els.content && show) els.content.style.display = 'none';
  }

  function showContent() { if (els.content) els.content.style.display = 'block'; }

  function showError(show, message) {
    if (els.error) { els.error.style.display = show ? 'flex' : 'none'; if (els.errorMsg && message) els.errorMsg.textContent = message; }
    if (els.content && show) els.content.style.display = 'none';
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();
})();
