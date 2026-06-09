(function () {
  var candidateId = window.location.pathname.split('/').pop();
  var candidateName = '';

  var loadingEl, errorEl, errorMsgEl, contentEl;
  var saveBtn, modalOverlay, modalClose, modalCancel, modalConfirm, modalNameEl;
  var poolSelect, poolNotes, poolLoading, poolForm, poolEmpty, poolError, poolErrorMsg;
  var toastEl, badgeEl, badgeWrap;

  function cacheEls() {
    loadingEl = document.getElementById('cd-loading');
    errorEl = document.getElementById('cd-error');
    errorMsgEl = document.getElementById('cd-error-msg');
    contentEl = document.getElementById('cd-content');

    saveBtn = document.getElementById('cd-save-btn');
    modalOverlay = document.getElementById('cd-pool-modal');
    modalClose = document.getElementById('cd-pool-modal-close');
    modalCancel = document.getElementById('cd-pool-cancel');
    modalConfirm = document.getElementById('cd-pool-confirm');
    modalNameEl = document.getElementById('cd-modal-candidate-name');
    poolSelect = document.getElementById('cd-pool-select');
    poolNotes = document.getElementById('cd-pool-notes');
    poolLoading = document.getElementById('cd-pool-loading');
    poolForm = document.getElementById('cd-pool-form');
    poolEmpty = document.getElementById('cd-pool-empty');
    poolError = document.getElementById('cd-pool-error');
    poolErrorMsg = document.getElementById('cd-pool-error-msg');
    toastEl = document.getElementById('cd-toast');
    badgeEl = document.getElementById('cd-pool-badge');
    badgeWrap = document.getElementById('cd-pool-badge-wrap');
  }

  function showLoading() {
    if (loadingEl) loadingEl.style.display = '';
    if (errorEl) errorEl.style.display = 'none';
    if (contentEl) contentEl.style.display = 'none';
  }

  function showError(msg) {
    if (loadingEl) loadingEl.style.display = 'none';
    if (errorEl) errorEl.style.display = '';
    if (errorMsgEl) errorMsgEl.innerText = msg || 'An unexpected error occurred.';
    if (contentEl) contentEl.style.display = 'none';
  }

  function showContent() {
    if (loadingEl) loadingEl.style.display = 'none';
    if (errorEl) errorEl.style.display = 'none';
    if (contentEl) contentEl.style.display = '';
  }

  function showToast(msg, type) {
    if (!toastEl) return;
    toastEl.innerText = msg;
    toastEl.className = 'emp-toast ' + (type || 'success');
    toastEl.style.display = '';
    setTimeout(function () {
      toastEl.style.display = 'none';
    }, 3500);
  }

  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(function (w) { return w[0]; }).join('').toUpperCase().slice(0, 2);
  }

  function updatePoolBadge() {
    if (!badgeWrap || !badgeEl) return;
    AngaziaAPI.candidates.pools(candidateId)
      .then(function (pools) {
        if (pools && pools.length > 0) {
          var names = pools.map(function (p) { return p.name; }).join(', ');
          badgeEl.innerText = '\u2713 In Talent Pool: ' + names;
          badgeEl.className = 'emp-status-badge active';
          if (saveBtn) saveBtn.style.display = 'none';
        } else {
          badgeEl.innerText = 'Not in Talent Pool';
          badgeEl.className = 'emp-status-badge';
          if (saveBtn) saveBtn.style.display = '';
        }
        badgeWrap.style.display = '';
      })
      .catch(function () {
        if (badgeWrap) badgeWrap.style.display = 'none';
      });
  }

  function renderProfile(data) {
    var profile = data.employee_profile || {};
    candidateName = profile.full_name || 'Candidate';

    document.title = candidateName + ' - Candidate Profile - Angazia';

    var nameEl = document.getElementById('cd-name');
    if (nameEl) nameEl.innerText = candidateName;

    var headlineEl = document.getElementById('cd-headline');
    if (headlineEl) headlineEl.innerText = profile.headline || '';

    var fullnameEl = document.getElementById('cd-fullname');
    if (fullnameEl) fullnameEl.innerText = candidateName;

    var headline2El = document.getElementById('cd-headline2');
    if (headline2El) headline2El.innerText = profile.headline || '';

    var initialsEl = document.getElementById('cd-initials');
    if (initialsEl) initialsEl.innerText = getInitials(candidateName);

    var availEl = document.getElementById('cd-availability');
    if (availEl) {
      if (profile.is_available) {
        availEl.innerText = 'Available';
        availEl.className = 'emp-status-badge active';
      } else {
        availEl.innerText = 'Not Available';
        availEl.className = 'emp-status-badge';
      }
    }

    var expYearsEl = document.getElementById('cd-exp-years');
    if (expYearsEl) expYearsEl.innerText = profile.years_of_experience || 0;

    var expLevelEl = document.getElementById('cd-exp-level');
    if (expLevelEl) expLevelEl.innerText = profile.experience_level || 'N/A';

    var locationEl = document.getElementById('cd-location');
    if (locationEl) locationEl.innerText = profile.location || 'N/A';

    var remoteEl = document.getElementById('cd-remote');
    if (remoteEl) remoteEl.innerText = profile.is_remote_only ? 'Yes' : 'No';

    var bioEl = document.getElementById('cd-bio');
    if (bioEl) bioEl.innerText = profile.bio || 'No bio provided.';

    var skillsEl = document.getElementById('cd-skills');
    if (skillsEl) {
      if (profile.skills && profile.skills.length > 0) {
        skillsEl.innerHTML = profile.skills.map(function (s) {
          return '<span class="emp-tag">' + s + '</span>';
        }).join('');
      } else {
        skillsEl.innerHTML = '<span class="emp-tag">No skills listed</span>';
      }
    }

    var linksEl = document.getElementById('cd-links');
    if (linksEl) {
      var linksHtml = '';
      if (profile.resume_url) linksHtml += '<a href="' + profile.resume_url + '" target="_blank">View Resume</a>';
      if (profile.portfolio_url) linksHtml += '<a href="' + profile.portfolio_url + '" target="_blank">Portfolio</a>';
      if (profile.linkedin_url) linksHtml += '<a href="' + profile.linkedin_url + '" target="_blank">LinkedIn</a>';
      linksEl.innerHTML = linksHtml || '<span class="emp-text-muted">No links provided.</span>';
    }

    var githubEl = document.getElementById('cd-github');
    if (githubEl) {
      var gh = profile.github_profile;
      if (gh && gh.github_username) {
        githubEl.innerHTML =
          '<a href="' + (gh.github_url || 'https://github.com/' + gh.github_username) + '" target="_blank">' +
          '@' + gh.github_username +
          (gh.public_repos ? ' (' + gh.public_repos + ' repos)' : '') +
          '</a>';
      } else {
        githubEl.innerHTML = '<span class="emp-text-muted">No GitHub profile linked.</span>';
      }
    }

    updatePoolBadge();
  }

  function loadCandidate() {
    showLoading();
    AngaziaAPI.candidates.detail(candidateId)
      .then(function (data) {
        showContent();
        renderProfile(data);
      })
      .catch(function (err) {
        showError(err.body && err.body.error ? err.body.error : err.message);
      });
  }

  function openPoolModal() {
    if (!modalOverlay) return;
    if (modalNameEl) modalNameEl.innerText = candidateName;
    modalOverlay.style.display = '';
    if (poolLoading) poolLoading.style.display = '';
    if (poolForm) poolForm.style.display = 'none';
    if (poolEmpty) poolEmpty.style.display = 'none';
    if (poolError) poolError.style.display = 'none';
    if (poolSelect) poolSelect.innerHTML = '<option value="">Select a pool...</option>';
    if (poolNotes) poolNotes.value = '';
    if (modalConfirm) {
      modalConfirm.disabled = true;
      modalConfirm.textContent = 'Confirm';
    }

    AngaziaAPI.talentPools.list()
      .then(function (resp) {
        if (poolLoading) poolLoading.style.display = 'none';
        var pools = resp && resp.pools ? resp.pools : (Array.isArray(resp) ? resp : []);
        if (!pools || pools.length === 0) {
          if (poolEmpty) poolEmpty.style.display = '';
          return;
        }
        if (poolForm) poolForm.style.display = '';
        if (poolSelect) {
          poolSelect.innerHTML = '<option value="">Select a pool...</option>';
          // Sort so "Saved Candidates" appears first
          pools.sort(function (a, b) {
            var na = (a.name || a.Name || '').toLowerCase();
            var nb = (b.name || b.Name || '').toLowerCase();
            if (na === 'saved candidates') return -1;
            if (nb === 'saved candidates') return 1;
            return 0;
          });
          pools.forEach(function (p) {
            var opt = document.createElement('option');
            opt.value = p.id || p.ID;
            opt.textContent = (p.name || p.Name) + (p.candidate_count !== undefined ? ' (' + p.candidate_count + ')' : '');
            poolSelect.appendChild(opt);
          });
        }
      })
      .catch(function (err) {
        if (poolLoading) poolLoading.style.display = 'none';
        if (poolError) poolError.style.display = '';
        if (poolErrorMsg) poolErrorMsg.innerText = err.body && err.body.error ? err.body.error : 'Failed to load pools.';
      });
  }

  function closePoolModal() {
    if (modalOverlay) modalOverlay.style.display = 'none';
  }

  function addToPool() {
    if (!poolSelect || !poolSelect.value) return;
    var poolId = poolSelect.value;
    var notes = poolNotes ? poolNotes.value : '';
    if (modalConfirm) {
      modalConfirm.disabled = true;
      modalConfirm.textContent = 'Adding...';
    }

    AngaziaAPI.candidates.pools(candidateId)
      .then(function (existingPools) {
        if (existingPools && existingPools.length > 0) {
          closePoolModal();
          showToast(candidateName + ' is already in a talent pool.', 'success');
          updatePoolBadge();
          return;
        }
        AngaziaAPI.talentPools.addCandidate(poolId, {
          employee_id: candidateId,
          notes: notes
        })
          .then(function () {
            closePoolModal();
            showToast(candidateName + ' added to talent pool.', 'success');
            updatePoolBadge();
          })
          .catch(function (err) {
            if (modalConfirm) {
              modalConfirm.disabled = false;
              modalConfirm.textContent = 'Confirm';
            }
            if (poolError) poolError.style.display = '';
            if (poolErrorMsg) poolErrorMsg.innerText = err.body && err.body.error ? err.body.error : 'Failed to add candidate to pool.';
          });
      })
      .catch(function () {
        AngaziaAPI.talentPools.addCandidate(poolId, {
          employee_id: candidateId,
          notes: notes
        })
          .then(function () {
            closePoolModal();
            showToast(candidateName + ' added to talent pool.', 'success');
            updatePoolBadge();
          })
          .catch(function (err) {
            if (modalConfirm) {
              modalConfirm.disabled = false;
              modalConfirm.textContent = 'Confirm';
            }
            if (poolError) poolError.style.display = '';
            if (poolErrorMsg) poolErrorMsg.innerText = err.body && err.body.error ? err.body.error : 'Failed to add candidate to pool.';
          });
      });
  }

  function initSaveHandlers() {
    if (saveBtn) {
      saveBtn.addEventListener('click', openPoolModal);
    }
    if (modalClose) {
      modalClose.addEventListener('click', closePoolModal);
    }
    if (modalCancel) {
      modalCancel.addEventListener('click', closePoolModal);
    }
    if (modalOverlay) {
      modalOverlay.addEventListener('click', function (e) {
        if (e.target === modalOverlay) closePoolModal();
      });
    }
    if (poolSelect) {
      poolSelect.addEventListener('change', function () {
        if (modalConfirm) modalConfirm.disabled = !this.value;
      });
    }
    if (modalConfirm) {
      modalConfirm.addEventListener('click', addToPool);
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    cacheEls();
    initSaveHandlers();
    if (candidateId) {
      loadCandidate();
    } else {
      showError('Invalid candidate ID.');
    }
  });
})();
