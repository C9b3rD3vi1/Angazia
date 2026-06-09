(function () {
  var loadingEl, errorEl, errorMsgEl, contentEl, poolsContainer, emptyEl;
  var createBtn, createModal, createModalClose, createCancel, createConfirm, createName, createDesc, createError, createErrorMsg;
  var deleteModal, deleteModalClose, deleteCancel, deleteConfirm, deleteNameEl, deleteError, deleteErrorMsg;
  var toastEl;

  var deletePoolId = null;

  function cacheEls() {
    loadingEl = document.getElementById('tp-loading');
    errorEl = document.getElementById('tp-error');
    errorMsgEl = document.getElementById('tp-error-msg');
    contentEl = document.getElementById('tp-content');
    poolsContainer = document.getElementById('tp-pools');
    emptyEl = document.getElementById('tp-empty');

    createBtn = document.getElementById('tp-create-btn');
    createModal = document.getElementById('tp-create-modal');
    createModalClose = document.getElementById('tp-create-modal-close');
    createCancel = document.getElementById('tp-create-cancel');
    createConfirm = document.getElementById('tp-create-confirm');
    createName = document.getElementById('tp-create-name');
    createDesc = document.getElementById('tp-create-desc');
    createError = document.getElementById('tp-create-error');
    createErrorMsg = document.getElementById('tp-create-error-msg');

    deleteModal = document.getElementById('tp-delete-modal');
    deleteModalClose = document.getElementById('tp-delete-modal-close');
    deleteCancel = document.getElementById('tp-delete-cancel');
    deleteConfirm = document.getElementById('tp-delete-confirm');
    deleteNameEl = document.getElementById('tp-delete-name');
    deleteError = document.getElementById('tp-delete-error');
    deleteErrorMsg = document.getElementById('tp-delete-error-msg');

    toastEl = document.getElementById('tp-toast');
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
    setTimeout(function () { toastEl.style.display = 'none'; }, 3500);
  }

  function getInitials(name) {
    if (!name) return '?';
    return name.split(' ').map(function (w) { return w[0]; }).join('').toUpperCase().slice(0, 2);
  }

  function formatDate(dateStr) {
    if (!dateStr) return '';
    var d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function loadCandidatesForPool(poolId, containerEl, toggleEl) {
    AngaziaAPI.talentPools.candidates(poolId)
      .then(function (resp) {
        var candidates = resp && resp.candidates ? resp.candidates : [];
        if (!candidates || candidates.length === 0) {
          containerEl.innerHTML = '<div class="tp-candidate-empty">No candidates in this pool yet.</div>';
          if (toggleEl) toggleEl.textContent = 'show';
          return;
        }
        containerEl.innerHTML = candidates.map(function (c) {
          var name = c.employee_name || c.employee?.full_name || 'Unknown';
          var headline = c.employee_headline || c.employee?.headline || '';
          var initials = getInitials(name);
          var addedAt = formatDate(c.added_at || c.AddedAt);
          var candId = c.employee_id || c.EmployeeID || '';
          return '<div class="tp-candidate">' +
            '<div class="tp-candidate-main">' +
              '<div class="tp-candidate-avatar"><span class="tp-candidate-initials">' + initials + '</span></div>' +
              '<div class="tp-candidate-info">' +
                '<div class="tp-candidate-name">' + name + '</div>' +
                '<span class="tp-candidate-headline">' + headline + '</span>' +
              '</div>' +
              '<span class="tp-candidate-date">Added ' + addedAt + '</span>' +
            '</div>' +
            '<div class="tp-candidate-actions">' +
              '<a href="/employer/candidates/' + candId + '" class="emp-btn emp-btn-sm emp-btn-outline">View</a>' +
              '<button class="emp-btn emp-btn-sm emp-btn-ghost emp-btn-danger remove-candidate" data-pool-id="' + poolId + '" data-cand-id="' + candId + '">Remove</button>' +
            '</div>' +
          '</div>';
        }).join('');

        containerEl.querySelectorAll('.remove-candidate').forEach(function (btn) {
          btn.addEventListener('click', function () {
            var pid = this.getAttribute('data-pool-id');
            var cid = this.getAttribute('data-cand-id');
            AngaziaAPI.talentPools.removeCandidate(pid, cid)
              .then(function () {
                showToast('Candidate removed from pool.', 'success');
                loadAllPools();
              })
              .catch(function (err) {
                showToast(err.body && err.body.error ? err.body.error : 'Failed to remove candidate.', 'error');
              });
          });
        });

        if (toggleEl) toggleEl.textContent = 'hide';
      })
      .catch(function () {
        containerEl.innerHTML = '<div class="tp-candidate-empty">Failed to load candidates.</div>';
        if (toggleEl) toggleEl.textContent = 'show';
      });
  }

  function loadAllPools() {
    showLoading();
    AngaziaAPI.talentPools.list()
      .then(function (resp) {
        showContent();
        var pools = resp && resp.pools ? resp.pools : (Array.isArray(resp) ? resp : []);
        if (!pools || pools.length === 0) {
          poolsContainer.innerHTML = '';
          if (emptyEl) emptyEl.style.display = '';
          return;
        }
        if (emptyEl) emptyEl.style.display = 'none';

        poolsContainer.innerHTML = pools.map(function (p) {
          var pid = p.id || p.ID;
          var name = p.name || p.Name;
          var desc = p.description || p.Description;
          var count = p.candidate_count || p.CandidateCount || 0;
          return '<div class="tp-pool" data-pool-id="' + pid + '">' +
            '<div class="tp-pool-head">' +
              '<div class="tp-pool-head-left">' +
                '<span class="tp-pool-name">' + name + '</span>' +
                '<span class="tp-pool-count">' + count + '</span>' +
                (desc ? '<span class="emp-text-muted" style="font-size:11px;color:var(--muted2);">' + desc + '</span>' : '') +
              '</div>' +
              '<div class="tp-pool-head-actions">' +
                '<button class="emp-btn emp-btn-sm emp-btn-ghost emp-btn-danger delete-pool" data-pool-id="' + pid + '" data-pool-name="' + name + '">Delete</button>' +
              '</div>' +
            '</div>' +
            '<div class="tp-pool-body"><div class="tp-candidate-empty">Loading candidates...</div></div>' +
          '</div>';
        }).join('');

        poolsContainer.querySelectorAll('.delete-pool').forEach(function (btn) {
          btn.addEventListener('click', function (e) {
            e.stopPropagation();
            deletePoolId = this.getAttribute('data-pool-id');
            if (deleteNameEl) deleteNameEl.textContent = this.getAttribute('data-pool-name');
            if (deleteError) deleteError.style.display = 'none';
            if (deleteModal) deleteModal.style.display = '';
          });
        });

        poolsContainer.querySelectorAll('.tp-pool-head').forEach(function (head) {
          head.addEventListener('click', function () {
            var pool = this.closest('.tp-pool');
            var body = pool.querySelector('.tp-pool-body');
            var pid = pool.getAttribute('data-pool-id');
            if (body.style.display === 'none') {
              body.style.display = '';
              loadCandidatesForPool(pid, body);
            } else {
              body.style.display = body.style.display === 'none' ? '' : 'none';
            }
          });
        });

        poolsContainer.querySelectorAll('.tp-pool').forEach(function (pool) {
          var pid = pool.getAttribute('data-pool-id');
          var body = pool.querySelector('.tp-pool-body');
          loadCandidatesForPool(pid, body);
        });
      })
      .catch(function (err) {
        showError(err.body && err.body.error ? err.body.error : 'Failed to load talent pools.');
      });
  }

  function openCreateModal() {
    if (createName) createName.value = '';
    if (createDesc) createDesc.value = '';
    if (createError) createError.style.display = 'none';
    if (createConfirm) createConfirm.disabled = true;
    if (createModal) createModal.style.display = '';
  }

  function closeCreateModal() {
    if (createModal) createModal.style.display = 'none';
  }

  function closeDeleteModal() {
    if (deleteModal) deleteModal.style.display = 'none';
    deletePoolId = null;
  }

  function handleCreate() {
    var name = createName ? createName.value.trim() : '';
    if (!name || name.length < 3) {
      if (createError) { createError.style.display = ''; if (createErrorMsg) createErrorMsg.innerText = 'Pool name must be at least 3 characters.'; }
      return;
    }
    var desc = createDesc ? createDesc.value.trim() : '';
    if (createConfirm) { createConfirm.disabled = true; createConfirm.textContent = 'Creating...'; }

    AngaziaAPI.talentPools.create({ name: name, description: desc })
      .then(function () {
        closeCreateModal();
        showToast('Pool "' + name + '" created.', 'success');
        loadAllPools();
      })
      .catch(function (err) {
        if (createConfirm) { createConfirm.disabled = false; createConfirm.textContent = 'Create Pool'; }
        if (createError) {
          createError.style.display = '';
          if (createErrorMsg) createErrorMsg.innerText = err.body && err.body.error ? err.body.error : 'Failed to create pool.';
        }
      });
  }

  function handleDelete() {
    if (!deletePoolId) return;
    if (deleteConfirm) { deleteConfirm.disabled = true; deleteConfirm.textContent = 'Deleting...'; }

    AngaziaAPI.talentPools.delete(deletePoolId)
      .then(function () {
        closeDeleteModal();
        showToast('Pool deleted.', 'success');
        loadAllPools();
      })
      .catch(function (err) {
        if (deleteConfirm) { deleteConfirm.disabled = false; deleteConfirm.textContent = 'Delete'; }
        if (deleteError) {
          deleteError.style.display = '';
          if (deleteErrorMsg) deleteErrorMsg.innerText = err.body && err.body.error ? err.body.error : 'Failed to delete pool.';
        }
      });
  }

  function initHandlers() {
    if (createBtn) createBtn.addEventListener('click', openCreateModal);
    if (createModalClose) createModalClose.addEventListener('click', closeCreateModal);
    if (createCancel) createCancel.addEventListener('click', closeCreateModal);
    if (createModal) {
      createModal.addEventListener('click', function (e) { if (e.target === createModal) closeCreateModal(); });
    }
    if (createName) {
      createName.addEventListener('input', function () {
        if (createConfirm) createConfirm.disabled = this.value.trim().length < 3;
      });
    }
    if (createConfirm) createConfirm.addEventListener('click', handleCreate);

    if (deleteModalClose) deleteModalClose.addEventListener('click', closeDeleteModal);
    if (deleteCancel) deleteCancel.addEventListener('click', closeDeleteModal);
    if (deleteModal) {
      deleteModal.addEventListener('click', function (e) { if (e.target === deleteModal) closeDeleteModal(); });
    }
    if (deleteConfirm) deleteConfirm.addEventListener('click', handleDelete);
  }

  document.addEventListener('DOMContentLoaded', function () {
    cacheEls();
    initHandlers();
    loadAllPools();
  });
})();
