(function () {
  var currentPage = 1;
  var pageSize = 20;
  var searchTerm = '';
  var deleteId = null;
  var searchTimer = null;

  function loadContacts() {
    var params = { page: currentPage, limit: pageSize };
    if (searchTerm) params.search = searchTerm;

    AngaziaAPI.get('/admin/contacts', params).then(function (data) {
      renderTable(data);
    }).catch(function () {
      document.getElementById('contacts-tbody').innerHTML =
        '<tr><td colspan="7" style="text-align:center;padding:40px;color:var(--muted)">Failed to load contacts</td></tr>';
    });
  }

  function renderTable(data) {
    var tbody = document.getElementById('contacts-tbody');
    var empty = document.getElementById('contacts-empty');
    var subs = data.submissions || [];

    if (subs.length === 0) {
      tbody.innerHTML = '';
      empty.style.display = 'block';
    } else {
      empty.style.display = 'none';
      tbody.innerHTML = subs.map(function (s) {
        var isUnread = !s.is_read;
        var date = new Date(s.created_at).toLocaleDateString('en-KE', {
          year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit'
        });
        var msgPreview = s.message.length > 100 ? s.message.substring(0, 100) + '...' : s.message;
        return '<tr class="' + (isUnread ? 'unread' : '') + '" data-id="' + s.id + '">' +
          '<td><span class="status-dot ' + (isUnread ? 'unread' : 'read') + '"></span></td>' +
          '<td><strong>' + escapeHtml(s.name) + '</strong></td>' +
          '<td><a href="mailto:' + escapeHtml(s.email) + '" style="color:var(--accent);text-decoration:none">' + escapeHtml(s.email) + '</a></td>' +
          '<td class="subject-cell">' + escapeHtml(s.subject || '—') + '</td>' +
          '<td><span class="msg-preview">' + escapeHtml(msgPreview) + '</span></td>' +
          '<td style="white-space:nowrap;font-size:12px;color:var(--muted)">' + date + '</td>' +
          '<td class="actions-cell">' +
            '<button class="action-btn" onclick="window._viewContact(\'' + s.id + '\')">View</button>' +
            (!isUnread ? '' : '<button class="action-btn" onclick="window._markRead(\'' + s.id + '\')">Read</button>') +
            '<button class="action-btn danger" onclick="window._deleteContact(\'' + s.id + '\')">Delete</button>' +
          '</td>' +
        '</tr>';
      }).join('');
    }

    // Stats
    document.getElementById('stat-total').textContent = data.total || 0;
    document.getElementById('stat-unread').textContent = data.unread_count || 0;
    document.getElementById('stat-read').textContent = ((data.total || 0) - (data.unread_count || 0));
    document.getElementById('stat-messages').textContent = data.total || 0;

    // Pagination
    renderPagination(data);
  }

  function renderPagination(data) {
    var container = document.getElementById('contacts-pagination');
    var totalPages = data.total_pages || 1;
    if (totalPages <= 1) { container.innerHTML = ''; return; }

    var html = '';
    if (currentPage > 1) {
      html += '<button class="action-btn" onclick="window._goPage(' + (currentPage - 1) + ')">Prev</button>';
    }
    for (var i = Math.max(1, currentPage - 2); i <= Math.min(totalPages, currentPage + 2); i++) {
      html += '<button class="action-btn" style="' + (i === currentPage ? 'border-color:var(--accent);color:var(--accent)' : '') + '" onclick="window._goPage(' + i + ')">' + i + '</button>';
    }
    if (currentPage < totalPages) {
      html += '<button class="action-btn" onclick="window._goPage(' + (currentPage + 1) + ')">Next</button>';
    }
    container.innerHTML = html;
  }

  window._goPage = function (page) {
    currentPage = page;
    loadContacts();
  };

  window._viewContact = function (id) {
    AngaziaAPI.get('/admin/contacts/' + id).then(function (sub) {
      document.getElementById('detail-name').textContent = sub.name;
      document.getElementById('detail-email').innerHTML = '<a href="mailto:' + escapeHtml(sub.email) + '">' + escapeHtml(sub.email) + '</a>';
      document.getElementById('detail-subject').textContent = sub.subject || '—';
      document.getElementById('detail-date').textContent = new Date(sub.created_at).toLocaleString('en-KE');
      document.getElementById('detail-message').textContent = sub.message;
      document.getElementById('detail-modal').classList.add('open');
      // Auto mark as read
      if (!sub.is_read) {
        AngaziaAPI.post('/admin/contacts/' + id + '/read').then(function () {
          loadContacts();
        }).catch(function () {});
      }
    }).catch(function (err) {
      if (AngaziaApp && AngaziaApp.showToast) AngaziaApp.showToast('Failed to load contact details', 'error');
    });
  };

  window._markRead = function (id) {
    AngaziaAPI.post('/admin/contacts/' + id + '/read').then(function () {
      loadContacts();
    }).catch(function () {});
  };

  window._deleteContact = function (id) {
    deleteId = id;
    document.getElementById('delete-modal').classList.add('open');
  };

  function closeDetailModal() {
    document.getElementById('detail-modal').classList.remove('open');
  }
  window.closeDetailModal = closeDetailModal;

  function closeDeleteModal() {
    document.getElementById('delete-modal').classList.remove('open');
    deleteId = null;
  }
  window.closeDeleteModal = closeDeleteModal;

  document.getElementById('confirm-delete-btn').addEventListener('click', function () {
    if (!deleteId) return;
    AngaziaAPI.del('/admin/contacts/' + deleteId).then(function () {
      closeDeleteModal();
      loadContacts();
    }).catch(function () {});
  });

  document.getElementById('search-input').addEventListener('input', function () {
    var val = this.value.trim();
    clearTimeout(searchTimer);
    searchTimer = setTimeout(function () {
      searchTerm = val;
      currentPage = 1;
      loadContacts();
    }, 300);
  });

  function escapeHtml(str) {
    if (!str) return '';
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  loadContacts();
})();
