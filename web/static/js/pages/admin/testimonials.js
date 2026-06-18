(function () {
  'use strict';

  var state = {
    testimonials: [],
    page: 1,
    limit: 20,
    total: 0,
    totalPages: 0,
    search: '',
    status: '',
    role: '',
    loading: false,
  };

  var els = {};

  function init() {
    cacheEls();
    bindEvents();
    fetchTestimonials();
  }

  function cacheEls() {
    var ids = [
      'atm-search','atm-filter-status','atm-filter-role','atm-filter-btn',
      'atm-tbody','atm-loading','atm-error','atm-error-msg',
      'atm-empty','atm-empty-msg','atm-table-wrap','atm-pagination',
      'atm-pagi-info','atm-pagi-btns',
      'atm-stat-total','atm-stat-pending','atm-stat-approved','atm-stat-featured',
    ];
    ids.forEach(function (id) { els[id] = document.getElementById(id); });
  }

  function bindEvents() {
    els['atm-filter-btn'].addEventListener('click', function () {
      state.page = 1;
      state.search = els['atm-search'].value.trim();
      state.status = els['atm-filter-status'].value;
      state.role = els['atm-filter-role'].value;
      fetchTestimonials();
    });

    els['atm-search'].addEventListener('keydown', function (e) {
      if (e.key === 'Enter') {
        state.page = 1;
        state.search = els['atm-search'].value.trim();
        state.status = els['atm-filter-status'].value;
        state.role = els['atm-filter-role'].value;
        fetchTestimonials();
      }
    });

    document.querySelector('[data-action="reload"]').addEventListener('click', function () {
      fetchTestimonials();
    });
  }

  function fetchTestimonials() {
    state.loading = true;
    show(els['atm-loading']);
    hide(els['atm-error']);
    hide(els['atm-table-wrap']);
    hide(els['atm-empty']);

    var params = {
      page: state.page,
      limit: state.limit,
      search: state.search || undefined,
      status: state.status || undefined,
      role: state.role || undefined,
    };

    AngaziaAPI.testimonials.admin.list(params).then(function (data) {
      state.loading = false;
      hide(els['atm-loading']);
      state.testimonials = data.testimonials || [];
      state.total = data.total || 0;
      state.totalPages = data.total_pages || 0;
      render();
      updateStats();
    }).catch(function (err) {
      state.loading = false;
      hide(els['atm-loading']);
      show(els['atm-error']);
      els['atm-error-msg'].textContent = err.message || 'Failed to load testimonials';
    });
  }

  function render() {
    if (state.testimonials.length === 0) {
      hide(els['atm-table-wrap']);
      show(els['atm-empty']);
      els['atm-empty-msg'].textContent = state.search || state.status || state.role
        ? 'No testimonials match your filters.'
        : 'No testimonials have been submitted yet.';
      return;
    }

    hide(els['atm-empty']);
    show(els['atm-table-wrap']);

    var html = '';
    state.testimonials.forEach(function (t) {
      var statusClass = t.is_approved ? 'atm-status-approved' : 'atm-status-pending';
      var statusText = t.is_approved ? 'Approved' : 'Pending';
      var featuredIcon = t.is_featured ? '\u2B50' : '\u2606';
      var featuredClass = t.is_featured ? 'atm-featured-yes' : 'atm-featured-no';
      var stars = renderStars(t.rating);
      var initials = (t.user_name || '?').charAt(0).toUpperCase();
      var date = new Date(t.created_at).toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
      var contentPreview = (t.content || '').substring(0, 100) + ((t.content || '').length > 100 ? '...' : '');
      var userEmail = t.user ? t.user.email : '';

      html += '<tr>';
      html += '<td><div class="atm-user-cell"><div class="atm-user-avatar">' + initials + '</div><div><div class="atm-user-name">' + escHtml(t.user_name) + '</div><div class="atm-user-email">' + escHtml(userEmail) + '</div></div></div></td>';
      html += '<td><div class="atm-content-preview" title="' + escAttr(t.content) + '">' + escHtml(contentPreview) + '</div></td>';
      html += '<td>' + (stars || '-') + '</td>';
      html += '<td><span class="atm-status ' + statusClass + '">' + escHtml(t.role) + '</span></td>';
      html += '<td><span class="atm-status ' + statusClass + '">' + statusText + '</span></td>';
      html += '<td><span class="' + featuredClass + '">' + featuredIcon + '</span></td>';
      html += '<td style="font-size:13px;color:var(--text-muted,#6b7280);white-space:nowrap">' + date + '</td>';
      html += '<td><div class="atm-actions" data-id="' + t.id + '">';
      if (!t.is_approved) {
        html += '<button class="atm-approve" data-action="approve">Approve</button>';
        html += '<button class="atm-reject" data-action="reject">Reject</button>';
      }
      html += '<button class="atm-feature" data-action="feature">' + (t.is_featured ? 'Unfeature' : 'Feature') + '</button>';
      html += '<button class="atm-delete" data-action="delete">Delete</button>';
      html += '</div></td>';
      html += '</tr>';
    });
    els['atm-tbody'].innerHTML = html;

    els['atm-tbody'].querySelectorAll('[data-action="approve"]').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = this.closest('.atm-actions').getAttribute('data-id');
        adminAction('approve', id);
      });
    });
    els['atm-tbody'].querySelectorAll('[data-action="reject"]').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = this.closest('.atm-actions').getAttribute('data-id');
        adminAction('reject', id);
      });
    });
    els['atm-tbody'].querySelectorAll('[data-action="feature"]').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = this.closest('.atm-actions').getAttribute('data-id');
        adminAction('feature', id);
      });
    });
    els['atm-tbody'].querySelectorAll('[data-action="delete"]').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = this.closest('.atm-actions').getAttribute('data-id');
        if (confirm('Are you sure you want to delete this testimonial? This cannot be undone.')) {
          adminAction('delete', id);
        }
      });
    });

    renderPagination();
  }

  function adminAction(action, id) {
    var btn = document.querySelector('.atm-actions[data-id="' + id + '"] button');
    if (btn) { btn.disabled = true; btn.textContent = '...'; }

    var promise;
    switch (action) {
      case 'approve': promise = AngaziaAPI.testimonials.admin.approve(id); break;
      case 'reject': promise = AngaziaAPI.testimonials.admin.reject(id); break;
      case 'feature': promise = AngaziaAPI.testimonials.admin.feature(id); break;
      case 'delete': promise = AngaziaAPI.testimonials.admin.del(id); break;
    }

    promise.then(function () {
      fetchTestimonials();
    }).catch(function (err) {
      showToast(err.message || 'Action failed', 'error');
      if (btn) { btn.disabled = false; btn.textContent = action.charAt(0).toUpperCase() + action.slice(1); }
    });
  }

  function renderStars(rating) {
    if (!rating) return '';
    var s = '';
    for (var i = 1; i <= 5; i++) {
      s += '<span class="' + (i <= rating ? 'atm-star' : 'atm-star-empty') + '">' + (i <= rating ? '\u2605' : '\u2606') + '</span>';
    }
    return s;
  }

  function renderPagination() {
    if (state.totalPages <= 1) {
      hide(els['atm-pagination']);
      return;
    }
    show(els['atm-pagination']);
    els['atm-pagi-info'].textContent = 'Page ' + state.page + ' of ' + state.totalPages + ' (' + state.total + ' total)';
    var btns = '';
    btns += '<button class="au-btn au-btn-sm au-btn-ghost" data-page="' + (state.page - 1) + '"' + (state.page <= 1 ? ' disabled' : '') + '>Previous</button>';
    btns += '<button class="au-btn au-btn-sm au-btn-ghost" data-page="' + (state.page + 1) + '"' + (state.page >= state.totalPages ? ' disabled' : '') + '>Next</button>';
    els['atm-pagi-btns'].innerHTML = btns;
    els['atm-pagi-btns'].querySelectorAll('button').forEach(function (btn) {
      btn.addEventListener('click', function () {
        if (this.disabled) return;
        state.page = parseInt(this.getAttribute('data-page'));
        fetchTestimonials();
      });
    });
  }

  function updateStats() {
    var total = state.total;
    var pending = 0;
    var approved = 0;
    var featured = 0;
    state.testimonials.forEach(function (t) {
      if (t.is_approved) approved++;
      else pending++;
      if (t.is_featured) featured++;
    });
    els['atm-stat-total'].textContent = total;
    els['atm-stat-pending'].textContent = pending;
    els['atm-stat-approved'].textContent = approved;
    els['atm-stat-featured'].textContent = featured;
  }

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  function show(el) { if (el) el.style.display = ''; }
  function hide(el) { if (el) el.style.display = 'none'; }
  function escHtml(s) { if (!s) return ''; return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }
  function escAttr(s) { if (!s) return ''; return s.replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/'/g,'&#39;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

  document.addEventListener('DOMContentLoaded', init);
})();
