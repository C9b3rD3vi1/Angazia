(function () {
  'use strict';

  var API = '/api/v1/employer';
  var state = {
    members: [],
    invitations: [],
    pendingAction: null
  };

  var els = {};

  function cache() {
    els.loading = document.getElementById('team-loading');
    els.error = document.getElementById('team-error');
    els.errorMsg = document.getElementById('team-error-msg');
    els.retryBtn = document.getElementById('team-retry-btn');
    els.content = document.getElementById('team-content');
    els.membersBody = document.getElementById('team-members-body');
    els.membersEmpty = document.getElementById('team-members-empty');
    els.memberCount = document.getElementById('team-member-count');
    els.invSection = document.getElementById('invitations-section');
    els.invBody = document.getElementById('invitations-body');
    els.invEmpty = document.getElementById('invitations-empty');
    els.invCount = document.getElementById('invitation-count');
    els.inviteBtn = document.getElementById('invite-member-btn');

    els.inviteModal = document.getElementById('invite-modal');
    els.inviteClose = document.getElementById('invite-modal-close');
    els.inviteCancel = document.getElementById('invite-modal-cancel');
    els.inviteSubmit = document.getElementById('invite-modal-submit');
    els.inviteEmail = document.getElementById('invite-email');
    els.inviteRole = document.getElementById('invite-role');
    els.inviteEmailHint = document.getElementById('invite-email-hint');
    els.inviteRoleHint = document.getElementById('invite-role-hint');

    els.confirmModal = document.getElementById('confirm-modal');
    els.confirmIcon = document.getElementById('confirm-modal-icon');
    els.confirmTitle = document.getElementById('confirm-modal-title');
    els.confirmDesc = document.getElementById('confirm-modal-desc');
    els.confirmCancel = document.getElementById('confirm-modal-cancel');
    els.confirmSubmit = document.getElementById('confirm-modal-submit');
    els.confirmSubmitLabel = document.getElementById('confirm-modal-submit-label');

    els.roleModal = document.getElementById('role-modal');
    els.roleClose = document.getElementById('role-modal-close');
    els.roleCancel = document.getElementById('role-modal-cancel');
    els.roleSubmit = document.getElementById('role-modal-submit');
    els.roleSelect = document.getElementById('role-select');
    els.roleMember = document.getElementById('role-modal-member');
    els.roleDesc = document.getElementById('role-modal-desc');
  }

  function show(el) { if (el) el.style.display = ''; }
  function hide(el) { if (el) el.style.display = 'none'; }

  function loading(isLoading) {
    if (isLoading) {
      show(els.loading);
      hide(els.error);
      hide(els.content);
    } else {
      hide(els.loading);
    }
  }

  function showError(msg) {
    hide(els.loading);
    hide(els.content);
    els.errorMsg.textContent = msg || 'An unexpected error occurred.';
    show(els.error);
  }

  function api(path, opts) {
    opts = opts || {};
    var headers = {};
    var token = localStorage.getItem('angazia_access_token');
    if (token) headers['Authorization'] = 'Bearer ' + token;
    if (opts.body) headers['Content-Type'] = 'application/json';
    return fetch(API + path, {
      method: opts.method || 'GET',
      headers: headers,
      body: opts.body ? JSON.stringify(opts.body) : undefined
    }).then(function (r) {
      return r.json().then(function (data) {
        if (!r.ok) throw new Error(data.message || data.error || 'Request failed');
        return data;
      });
    });
  }

  function flash(msg, type) {
    var f = document.getElementById('flash-message');
    if (f) {
      f.className = 'flash flash-' + (type || 'info');
      f.innerHTML = msg + '<button class="flash-close" onclick="this.parentElement.remove()">&times;</button>';
      f.style.display = '';
      setTimeout(function () {
        f.classList.add('flash-fade');
        setTimeout(function () { if (f.parentNode) f.parentNode.removeChild(f); }, 400);
      }, 5000);
    }
  }

  function formatDate(iso) {
    if (!iso) return '--';
    var d = new Date(iso);
    return d.toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function roleBadge(role) {
    var cls = role === 'admin' ? 'shortlisted' : role === 'recruiter' ? 'interview' : 'pending';
    return '<span class="emp-status-badge ' + cls + '">' + role + '</span>';
  }

  function renderMembers() {
    els.membersBody.innerHTML = '';
    if (!state.members.length) {
      hide(els.membersBody.closest('.emp-table-responsive'));
      show(els.membersEmpty);
      els.memberCount.textContent = '0 members';
      return;
    }
    show(els.membersBody.closest('.emp-table-responsive'));
    hide(els.membersEmpty);
    els.memberCount.textContent = state.members.length + ' member' + (state.members.length !== 1 ? 's' : '');

    state.members.forEach(function (m) {
      var tr = document.createElement('tr');
      var initials = m.full_name ? m.full_name.split(' ').map(function (s) { return s[0]; }).join('').toUpperCase().slice(0, 2) : '??';

      var actionsHtml = '';
      if (m.is_owner) {
        actionsHtml = '<span class="emp-table-muted">Owner</span>';
      } else {
        actionsHtml = '<button class="emp-btn emp-btn-xs emp-btn-ghost change-role-btn" data-id="' + m.id + '" data-name="' + m.full_name + '" data-role="' + m.role + '" title="Change role">&#x1F504;</button>' +
          '<button class="emp-btn emp-btn-xs emp-btn-ghost remove-member-btn" data-id="' + m.id + '" data-name="' + m.full_name + '" title="Remove from team" style="color:var(--danger)">&#x1F5D1;</button>';
      }

      var avatarHtml = m.avatar_url
        ? '<img src="' + m.avatar_url + '" alt="' + m.full_name + '" style="width:100%;height:100%;object-fit:cover;border-radius:50%">'
        : initials;

      tr.innerHTML = '<td><div style="display:flex;align-items:center;gap:10px">' +
        '<div class="emp-avatar">' + avatarHtml + '</div>' +
        '<span>' + m.full_name + '</span>' +
        '</div></td>' +
        '<td class="emp-table-muted">' + m.email + '</td>' +
        '<td>' + roleBadge(m.role) + '</td>' +
        '<td class="emp-table-muted">' + formatDate(m.joined_at) + '</td>' +
        '<td style="white-space:nowrap">' + actionsHtml + '</td>';
      els.membersBody.appendChild(tr);
    });
  }

  function renderInvitations() {
    els.invBody.innerHTML = '';
    if (!state.invitations.length) {
      hide(els.invSection);
      return;
    }
    show(els.invSection);
    hide(els.invEmpty);
    els.invCount.textContent = state.invitations.length + ' pending';

    state.invitations.forEach(function (inv) {
      var tr = document.createElement('tr');
      tr.innerHTML = '<td>' + inv.email + '</td>' +
        '<td>' + roleBadge(inv.role) + '</td>' +
        '<td class="emp-table-muted">' + formatDate(inv.created_at) + '</td>' +
        '<td class="emp-table-muted">' + formatDate(inv.expires_at) + '</td>' +
        '<td><button class="emp-btn emp-btn-xs emp-btn-ghost cancel-inv-btn" data-id="' + inv.id + '" data-email="' + inv.email + '" title="Cancel invitation" style="color:var(--danger)">Cancel</button></td>';
      els.invBody.appendChild(tr);
    });
  }

  function loadTeam() {
    loading(true);
    state.members = [];
    state.invitations = [];

    api('/team').then(function (res) {
      state.members = res.data || [];
      return api('/team/invitations');
    }).then(function (res) {
      state.invitations = res.data || [];
      loading(false);
      show(els.content);
      renderMembers();
      renderInvitations();
    }).catch(function (err) {
      showError(err.message);
    });
  }

  function openInviteModal() {
    els.inviteEmail.value = '';
    els.inviteRole.value = 'recruiter';
    els.inviteEmailHint.textContent = '';
    els.inviteRoleHint.textContent = '';
    els.inviteSubmit.querySelector('.emp-btn-label').textContent = 'Send Invitation';
    els.inviteSubmit.disabled = false;
    show(els.inviteModal);
    els.inviteEmail.focus();
  }

  function closeInviteModal() {
    hide(els.inviteModal);
  }

  function submitInvite() {
    var email = els.inviteEmail.value.trim();
    var role = els.inviteRole.value;

    els.inviteEmailHint.textContent = '';
    els.inviteRoleHint.textContent = '';

    if (!email) {
      els.inviteEmailHint.textContent = 'Email is required';
      els.inviteEmail.focus();
      return;
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      els.inviteEmailHint.textContent = 'Invalid email format';
      els.inviteEmail.focus();
      return;
    }

    els.inviteSubmit.disabled = true;
    els.inviteSubmit.classList.add('emp-btn-loading');

    api('/team/invite', { method: 'POST', body: { email: email, role: role } }).then(function () {
      closeInviteModal();
      flash('Invitation sent to ' + email, 'success');
      loadTeam();
    }).catch(function (err) {
      els.inviteSubmit.disabled = false;
      els.inviteSubmit.classList.remove('emp-btn-loading');
      els.inviteEmailHint.textContent = err.message;
    });
  }

  function openConfirmModal(title, desc, label, icon, cb) {
    state.pendingAction = cb;
    els.confirmTitle.textContent = title;
    els.confirmDesc.textContent = desc;
    els.confirmSubmitLabel.textContent = label || 'Confirm';
    els.confirmIcon.className = 'emp-modal-icon ' + (icon || 'icon-danger');
    els.confirmSubmit.disabled = false;
    els.confirmSubmit.classList.remove('emp-btn-loading');
    show(els.confirmModal);
  }

  function closeConfirmModal() {
    hide(els.confirmModal);
    state.pendingAction = null;
  }

  function executeConfirm() {
    if (typeof state.pendingAction !== 'function') return;
    els.confirmSubmit.disabled = true;
    els.confirmSubmit.classList.add('emp-btn-loading');
    state.pendingAction(function (err) {
      els.confirmSubmit.disabled = false;
      els.confirmSubmit.classList.remove('emp-btn-loading');
      if (err) {
        flash(err.message || 'Action failed', 'error');
      } else {
        closeConfirmModal();
        loadTeam();
      }
    });
  }

  function removeMember(id, name) {
    openConfirmModal(
      'Remove ' + name + '?',
      'This will remove them from your team. They will lose access to company resources.',
      'Remove',
      'icon-danger',
      function (done) {
        api('/team/' + id, { method: 'DELETE' }).then(function () {
          flash(name + ' removed from team', 'success');
          done();
        }).catch(function (err) { done(err); });
      }
    );
  }

  function cancelInvitation(id, email) {
    openConfirmModal(
      'Cancel invitation for ' + email + '?',
      'This will cancel the pending invitation. They will not be able to join using the invitation link.',
      'Cancel Invitation',
      'icon-warning',
      function (done) {
        api('/team/invitations/' + id, { method: 'DELETE' }).then(function () {
          flash('Invitation cancelled', 'success');
          done();
        }).catch(function (err) { done(err); });
      }
    );
  }

  function openRoleModal(id, name, role) {
    els.roleMember.textContent = name;
    els.roleSelect.value = role;
    els.roleSubmit.disabled = false;
    els.roleSubmit.classList.remove('emp-btn-loading');
    state.pendingMemberId = id;
    show(els.roleModal);
  }

  function closeRoleModal() {
    hide(els.roleModal);
    state.pendingMemberId = null;
  }

  function submitRoleChange() {
    var id = state.pendingMemberId;
    if (!id) return;

    var role = els.roleSelect.value;
    els.roleSubmit.disabled = true;
    els.roleSubmit.classList.add('emp-btn-loading');

    api('/team/' + id + '/role', { method: 'PUT', body: { role: role } }).then(function () {
      closeRoleModal();
      flash('Role updated successfully', 'success');
      loadTeam();
    }).catch(function (err) {
      els.roleSubmit.disabled = false;
      els.roleSubmit.classList.remove('emp-btn-loading');
      flash(err.message, 'error');
    });
  }

  function bindEvents() {
    els.retryBtn.addEventListener('click', loadTeam);
    els.inviteBtn.addEventListener('click', openInviteModal);
    els.inviteClose.addEventListener('click', closeInviteModal);
    els.inviteCancel.addEventListener('click', closeInviteModal);
    els.inviteSubmit.addEventListener('click', submitInvite);

    els.confirmCancel.addEventListener('click', closeConfirmModal);
    els.confirmSubmit.addEventListener('click', executeConfirm);

    els.roleClose.addEventListener('click', closeRoleModal);
    els.roleCancel.addEventListener('click', closeRoleModal);
    els.roleSubmit.addEventListener('click', submitRoleChange);

    els.inviteEmail.addEventListener('keydown', function (e) { if (e.key === 'Enter') submitInvite(); });

    els.membersBody.addEventListener('click', function (e) {
      var target = e.target.closest('.remove-member-btn');
      if (target) {
        removeMember(target.dataset.id, target.dataset.name);
        return;
      }
      target = e.target.closest('.change-role-btn');
      if (target) {
        openRoleModal(target.dataset.id, target.dataset.name, target.dataset.role);
      }
    });

    els.invBody.addEventListener('click', function (e) {
      var target = e.target.closest('.cancel-inv-btn');
      if (target) {
        cancelInvitation(target.dataset.id, target.dataset.email);
      }
    });

    els.inviteModal.addEventListener('click', function (e) {
      if (e.target === els.inviteModal) closeInviteModal();
    });
    els.confirmModal.addEventListener('click', function (e) {
      if (e.target === els.confirmModal) closeConfirmModal();
    });
    els.roleModal.addEventListener('click', function (e) {
      if (e.target === els.roleModal) closeRoleModal();
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    cache();
    bindEvents();
    loadTeam();
  });
})();
