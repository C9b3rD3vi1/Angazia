(function () {
  'use strict';

  var state = {
    testimonials: [],
    page: 1,
    limit: 10,
    total: 0,
    totalPages: 0,
    editingId: null,
    deletingId: null,
  };

  var els = {};

  function init() {
    cacheEls();
    bindEvents();
    fetchTestimonials();
  }

  function cacheEls() {
    var ids = [
      'et-loading','et-error','et-error-msg','et-content','et-empty','et-empty-add',
      'et-list','et-pagination','et-pagi-info','et-pagi-btns','et-add-btn',
      'et-modal','et-modal-title','et-modal-close','et-modal-cancel','et-modal-submit',
      'et-form-title','et-form-company','et-form-content','et-form-count','et-star-rating',
      'et-del-modal','et-del-close','et-del-cancel','et-del-confirm',
    ];
    ids.forEach(function (id) { els[id] = document.getElementById(id); });
  }

  function bindEvents() {
    els['et-add-btn'].addEventListener('click', function () { openCreateModal(); });
    els['et-empty-add'].addEventListener('click', function () { openCreateModal(); });
    els['et-modal-close'].addEventListener('click', closeModal);
    els['et-modal-cancel'].addEventListener('click', closeModal);
    els['et-modal-submit'].addEventListener('click', submitTestimonial);
    els['et-del-close'].addEventListener('click', closeDelModal);
    els['et-del-cancel'].addEventListener('click', closeDelModal);
    els['et-del-confirm'].addEventListener('click', confirmDelete);

    els['et-form-content'].addEventListener('input', function () {
      els['et-form-count'].textContent = this.value.length;
    });

    els['et-star-rating'].addEventListener('click', function (e) {
      var star = e.target.closest('.et-star');
      if (!star) return;
      var val = parseInt(star.getAttribute('data-value'));
      setRating(val);
    });

    els['et-star-rating'].addEventListener('mouseover', function (e) {
      var star = e.target.closest('.et-star');
      if (!star) return;
      var val = parseInt(star.getAttribute('data-value'));
      highlightStars(val);
    });

    els['et-star-rating'].addEventListener('mouseleave', function () {
      highlightStars(state.rating || 0);
    });

    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') { closeModal(); closeDelModal(); }
    });

    els['et-modal'].addEventListener('click', function (e) {
      if (e.target === els['et-modal']) closeModal();
    });
    els['et-del-modal'].addEventListener('click', function (e) {
      if (e.target === els['et-del-modal']) closeDelModal();
    });
  }

  function setRating(val) {
    state.rating = val;
    highlightStars(val);
  }

  function highlightStars(val) {
    var stars = els['et-star-rating'].querySelectorAll('.et-star');
    stars.forEach(function (s) {
      var v = parseInt(s.getAttribute('data-value'));
      s.classList.toggle('active', v <= val);
      s.textContent = v <= val ? '\u2605' : '\u2606';
    });
  }

  function fetchTestimonials() {
    show(els['et-loading']);
    hide(els['et-content']);
    hide(els['et-error']);

    AngaziaAPI.testimonials.mine({ page: state.page, limit: state.limit }).then(function (data) {
      hide(els['et-loading']);
      state.testimonials = data.testimonials || [];
      state.total = data.total || 0;
      state.totalPages = data.total_pages || 0;
      render();
    }).catch(function (err) {
      hide(els['et-loading']);
      show(els['et-error']);
      els['et-error-msg'].textContent = err.message || 'Failed to load testimonials';
    });
  }

  function render() {
    if (state.testimonials.length === 0) {
      show(els['et-empty']);
      hide(els['et-list']);
      hide(els['et-pagination']);
      return;
    }
    hide(els['et-empty']);
    show(els['et-content']);
    show(els['et-list']);

    var html = '';
    state.testimonials.forEach(function (t) {
      var starsHtml = renderStars(t.rating);
      var statusClass = t.is_approved ? 'et-status-approved' : 'et-status-pending';
      var statusText = t.is_approved ? 'Approved' : 'Pending Review';
      var initials = (t.user_name || '?').charAt(0).toUpperCase();
      var date = new Date(t.created_at).toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });

      html += '<div class="et-card" data-id="' + t.id + '">';
      html += '<div class="et-card-head">';
      html += '<div class="et-card-user">';
      html += '<div class="et-card-avatar">' + initials + '</div>';
      html += '<div class="et-card-info">';
      html += '<div class="et-card-name">' + escHtml(t.user_name) + '</div>';
      html += '<div class="et-card-meta">' + escHtml(t.user_title || '') + (t.user_title && t.company_name ? ' at ' : '') + escHtml(t.company_name || '') + '</div>';
      html += '</div></div>';
      html += '<div class="et-card-actions">';
      if (!t.is_approved) {
        html += '<button class="et-btn-edit" data-action="edit" data-id="' + t.id + '">Edit</button>';
      }
      html += '<button class="et-btn-del" data-action="delete" data-id="' + t.id + '">Delete</button>';
      html += '</div></div>';
      if (starsHtml) html += '<div class="et-view-stars">' + starsHtml + '</div>';
      html += '<div class="et-card-content">' + escHtml(t.content) + '</div>';
      html += '<div class="et-card-footer">';
      html += '<span class="et-status-badge ' + statusClass + '">' + statusText + '</span>';
      html += '<span>' + date + '</span>';
      html += '</div></div>';
    });
    els['et-list'].innerHTML = html;

    els['et-list'].querySelectorAll('[data-action="edit"]').forEach(function (btn) {
      btn.addEventListener('click', function () { openEditModal(this.getAttribute('data-id')); });
    });
    els['et-list'].querySelectorAll('[data-action="delete"]').forEach(function (btn) {
      btn.addEventListener('click', function () { openDeleteModal(this.getAttribute('data-id')); });
    });

    renderPagination();
  }

  function renderStars(rating) {
    if (!rating) return '';
    var s = '';
    for (var i = 1; i <= 5; i++) {
      s += '<span class="' + (i <= rating ? 'et-star-filled' : 'et-star-empty') + '">' + (i <= rating ? '\u2605' : '\u2606') + '</span>';
    }
    return s;
  }

  function renderPagination() {
    if (state.totalPages <= 1) {
      hide(els['et-pagination']);
      return;
    }
    show(els['et-pagination']);
    els['et-pagi-info'].textContent = 'Page ' + state.page + ' of ' + state.totalPages + ' (' + state.total + ' total)';
    var btns = '';
    btns += '<button class="emp-btn emp-btn-sm emp-btn-ghost" data-page="' + (state.page - 1) + '"' + (state.page <= 1 ? ' disabled' : '') + '>Previous</button>';
    btns += '<button class="emp-btn emp-btn-sm emp-btn-ghost" data-page="' + (state.page + 1) + '"' + (state.page >= state.totalPages ? ' disabled' : '') + '>Next</button>';
    els['et-pagi-btns'].innerHTML = btns;
    els['et-pagi-btns'].querySelectorAll('button').forEach(function (btn) {
      btn.addEventListener('click', function () {
        if (this.disabled) return;
        state.page = parseInt(this.getAttribute('data-page'));
        fetchTestimonials();
      });
    });
  }

  function openCreateModal() {
    state.editingId = null;
    state.rating = 0;
    els['et-modal-title'].textContent = 'New Testimonial';
    els['et-modal-submit'].textContent = 'Submit for Review';
    els['et-form-title'].value = '';
    els['et-form-company'].value = '';
    els['et-form-content'].value = '';
    els['et-form-count'].textContent = '0';
    setRating(0);
    show(els['et-modal']);
    els['et-form-content'].focus();
  }

  function openEditModal(id) {
    var t = findTestimonial(id);
    if (!t) return;
    state.editingId = id;
    state.rating = t.rating || 0;
    els['et-modal-title'].textContent = 'Edit Testimonial';
    els['et-modal-submit'].textContent = 'Update';
    els['et-form-title'].value = t.user_title || '';
    els['et-form-company'].value = t.company_name || '';
    els['et-form-content'].value = t.content || '';
    els['et-form-count'].textContent = (t.content || '').length;
    setRating(t.rating || 0);
    show(els['et-modal']);
  }

  function closeModal() {
    hide(els['et-modal']);
  }

  function submitTestimonial() {
    var content = els['et-form-content'].value.trim();
    if (!content || content.length < 20) {
      showToast('Content must be at least 20 characters', 'error');
      return;
    }

    var data = {
      content: content,
      user_title: els['et-form-title'].value.trim(),
      company_name: els['et-form-company'].value.trim(),
      rating: state.rating || 0,
    };

    els['et-modal-submit'].disabled = true;
    els['et-modal-submit'].textContent = 'Saving...';

    var promise;
    if (state.editingId) {
      promise = AngaziaAPI.testimonials.update(state.editingId, data);
    } else {
      promise = AngaziaAPI.testimonials.create(data);
    }

    promise.then(function () {
      closeModal();
      fetchTestimonials();
    }).catch(function (err) {
      showToast(err.message || 'Failed to save testimonial', 'error');
    }).finally(function () {
      els['et-modal-submit'].disabled = false;
      els['et-modal-submit'].textContent = state.editingId ? 'Update' : 'Submit for Review';
    });
  }

  function openDeleteModal(id) {
    state.deletingId = id;
    show(els['et-del-modal']);
  }

  function closeDelModal() {
    state.deletingId = null;
    hide(els['et-del-modal']);
  }

  function confirmDelete() {
    if (!state.deletingId) return;
    els['et-del-confirm'].disabled = true;
    els['et-del-confirm'].textContent = 'Deleting...';
    AngaziaAPI.testimonials.delete(state.deletingId).then(function () {
      closeDelModal();
      fetchTestimonials();
    }).catch(function (err) {
      showToast(err.message || 'Failed to delete testimonial', 'error');
    }).finally(function () {
      els['et-del-confirm'].disabled = false;
      els['et-del-confirm'].textContent = 'Delete';
    });
  }

  function findTestimonial(id) {
    for (var i = 0; i < state.testimonials.length; i++) {
      if (state.testimonials[i].id === id) return state.testimonials[i];
    }
    return null;
  }

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    }
  }

  function show(el) { if (el) el.style.display = ''; }
  function hide(el) { if (el) el.style.display = 'none'; }
  function escHtml(s) { if (!s) return ''; return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }

  document.addEventListener('DOMContentLoaded', init);
})();
