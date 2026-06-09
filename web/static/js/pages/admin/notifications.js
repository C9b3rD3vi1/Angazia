(function () {
  'use strict';

  var currentFilter = 'all';
  var currentPage = 1;
  var totalPages = 1;
  var totalCount = 0;
  var sortOrder = 'newest';
  var searchQuery = '';

  function render() {
    var loadingEl = document.getElementById('an-loading');
    var listEl = document.getElementById('an-list');
    var pagiEl = document.getElementById('an-pagination');
    var errorEl = document.getElementById('an-error');
    var markAllBtn = document.getElementById('an-mark-all-btn');

    if (loadingEl) loadingEl.style.display = '';
    if (listEl) listEl.style.display = 'none';
    if (pagiEl) pagiEl.style.display = 'none';
    if (errorEl) errorEl.style.display = 'none';

    var params = { page: currentPage, limit: 20 };
    if (currentFilter !== 'all') params.filter = currentFilter;
    if (searchQuery) params.search = searchQuery;
    if (sortOrder) params.sort = sortOrder;

    var endpoint = currentFilter === 'unread' ? '/notifications/unread' : '/notifications';

    AngaziaAPI.get(endpoint, params).then(function (data) {
      if (loadingEl) loadingEl.style.display = 'none';

      var items = data && data.notifications ? data.notifications : (data && data.data ? data.data : []);
      totalCount = data && data.total ? data.total : items.length;

      if (listEl) {
        if (items.length === 0) {
          listEl.innerHTML = '<div class="an-empty"><span class="an-empty-text">No notifications found.</span></div>';
          listEl.style.display = '';
        } else {
          var html = '';
          for (var i = 0; i < items.length; i++) {
            var n = items[i];
            var isRead = n.is_read || n.read_at;
            html += '<div class="an-item' + (isRead ? '' : ' an-item-unread') + '" data-id="' + (n.id || '') + '">' +
              '<div class="an-item-left">' +
              '<span class="an-item-icon">' + getIcon(n.type) + '</span>' +
              '</div>' +
              '<div class="an-item-body">' +
              '<div class="an-item-head">' +
              '<span class="an-item-title">' + escapeHtml(n.title || n.message || '') + '</span>' +
              (!isRead ? '<span class="an-item-dot"></span>' : '') +
              '</div>' +
              '<p class="an-item-msg">' + escapeHtml(n.message || n.body || '') + '</p>' +
              '<span class="an-item-time">' + formatTime(n.created_at) + '</span>' +
              '</div>' +
              '<div class="an-item-right">' +
              '<button class="an-item-action" data-action="mark-read" data-id="' + (n.id || '') + '" ' + (isRead ? 'style="display:none"' : '') + '>Mark Read</button>' +
              '</div>' +
              '</div>';
          }
          listEl.innerHTML = html;
          listEl.style.display = '';
        }
      }

      if (markAllBtn) {
        markAllBtn.style.display = items.length > 0 && currentFilter !== 'unread' ? '' : 'none';
      }

      totalPages = data && data.total_pages ? data.total_pages : Math.ceil(totalCount / 20);
      renderPagination();
    }).catch(function () {
      if (loadingEl) loadingEl.style.display = 'none';
      var errorEl2 = document.getElementById('an-error');
      if (errorEl2) errorEl2.style.display = '';
    });
  }

  function renderPagination() {
    var pagiEl = document.getElementById('an-pagination');
    if (!pagiEl) return;
    if (totalPages <= 1) { pagiEl.style.display = 'none'; return; }
    pagiEl.style.display = '';

    var infoEl = document.getElementById('an-pagi-info');
    if (infoEl) {
      infoEl.textContent = 'Page ' + currentPage + ' of ' + totalPages + ' (' + totalCount + ' total)';
    }

    var btnsEl = document.getElementById('an-pagi-btns');
    if (!btnsEl) return;
    var html = '';
    if (currentPage > 1) {
      html += '<button class="an-pagi-btn" data-page="' + (currentPage - 1) + '">Prev</button>';
    }
    for (var p = Math.max(1, currentPage - 2); p <= Math.min(totalPages, currentPage + 2); p++) {
      html += '<button class="an-pagi-btn' + (p === currentPage ? ' active' : '') + '" data-page="' + p + '">' + p + '</button>';
    }
    if (currentPage < totalPages) {
      html += '<button class="an-pagi-btn" data-page="' + (currentPage + 1) + '">Next</button>';
    }
    btnsEl.innerHTML = html;
  }

  function getIcon(type) {
    var t = (type || '').toLowerCase();
    if (t === 'application' || t === 'application_update') return 'A';
    if (t === 'interview' || t === 'interview_scheduled') return 'I';
    if (t === 'message' || t === 'new_message') return 'M';
    if (t === 'alert' || t === 'system_alert') return '!';
    if (t === 'job_match') return 'J';
    return 'N';
  }

  function formatTime(dateStr) {
    if (!dateStr) return '';
    var d = new Date(dateStr);
    var now = new Date();
    var diff = now - d;
    if (diff < 60000) return 'Just now';
    if (diff < 3600000) return Math.floor(diff / 60000) + 'm ago';
    if (diff < 86400000) return Math.floor(diff / 3600000) + 'h ago';
    if (diff < 604800000) return Math.floor(diff / 86400000) + 'd ago';
    return d.toLocaleDateString();
  }

  function escapeHtml(text) {
    if (!text) return '';
    var d = document.createElement('div');
    d.textContent = text;
    return d.innerHTML;
  }

  function markRead(id) {
    AngaziaAPI.post('/notifications/' + id + '/read').then(function () {
      render();
    }).catch(function () {});
  }

  function markAllRead() {
    AngaziaAPI.post('/notifications/read-all').then(function () {
      render();
    }).catch(function () {});
  }

  document.addEventListener('click', function (e) {
    var tab = e.target.closest('.an-tab');
    if (tab) {
      document.querySelectorAll('.an-tab').forEach(function (t) { t.classList.remove('active'); });
      tab.classList.add('active');
      currentFilter = tab.getAttribute('data-filter') || 'all';
      currentPage = 1;
      render();
      return;
    }

    var pageBtn = e.target.closest('.an-pagi-btn');
    if (pageBtn) {
      var p = parseInt(pageBtn.getAttribute('data-page'), 10);
      if (p > 0 && p <= totalPages) {
        currentPage = p;
        render();
      }
      return;
    }

    var markBtn = e.target.closest('[data-action="mark-read"]');
    if (markBtn) {
      var id = markBtn.getAttribute('data-id');
      if (id) markRead(id);
      return;
    }

    var markAllBtn = e.target.closest('#an-mark-all-btn');
    if (markAllBtn) {
      markAllRead();
      return;
    }
  });

  document.addEventListener('DOMContentLoaded', function () {
    var searchInput = document.getElementById('an-search-input');
    if (searchInput) {
      var debounceTimer;
      searchInput.addEventListener('input', function () {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(function () {
          searchQuery = searchInput.value.trim();
          currentPage = 1;
          render();
        }, 300);
      });
    }

    var sortSelect = document.getElementById('an-sort-select');
    if (sortSelect) {
      sortSelect.addEventListener('change', function () {
        sortOrder = sortSelect.value;
        currentPage = 1;
        render();
      });
    }

    render();
  });
})();
