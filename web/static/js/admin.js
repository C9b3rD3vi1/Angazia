(function () {
  'use strict';

  function aInitNavbar() {
    var n = document.getElementById('navbar');
    if (n) {
      window.addEventListener('scroll', function () {
        if (window.scrollY > 50) n.classList.add('scrolled');
        else n.classList.remove('scrolled');
      });
    }
  }

  function aCloseDropdown(name) {
    if (name) {
      var dd = document.getElementById(name);
      if (dd) dd.style.display = 'none';
    } else {
      var dd = document.getElementById('nav-dropdown');
      if (dd) dd.style.display = 'none';
      var nd = document.getElementById('notif-dropdown');
      if (nd) nd.style.display = 'none';
    }
  }

  function aFetchNotifCount() {
    AngaziaAPI.get('/notifications/counts').then(function (d) {
      var badge = document.getElementById('nav-notif-badge');
      if (!badge) return;
      if (d && d.unread > 0) {
        badge.textContent = d.unread > 99 ? '99+' : d.unread;
        badge.style.display = '';
      } else {
        badge.style.display = 'none';
      }
    }).catch(function () {});
  }

  function aFetchNotifPreview() {
    AngaziaAPI.get('/notifications/unread').then(function (d) {
      var list = document.getElementById('notif-dropdown-list');
      if (!list) return;
      var items = d && d.notifications ? d.notifications : [];
      if (items.length === 0) {
        list.innerHTML = '<div class="notif-list-empty">No new notifications</div>';
        return;
      }
      var html = '';
      for (var i = 0; i < Math.min(items.length, 5); i++) {
        var n = items[i];
        html += '<a href="/admin/notifications" class="notif-dropdown-item">' +
          '<div class="notif-dd-content">' +
          '<span class="notif-dd-msg">' + (n.title || n.message || '') + '</span>' +
          '<span class="notif-dd-time">' + (n.created_at ? new Date(n.created_at).toLocaleDateString() : '') + '</span>' +
          '</div></a>';
      }
      list.innerHTML = html;
    }).catch(function () {});
  }

  document.addEventListener('click', function (e) {
    var target = e.target;

    var avatarBtn = target.closest('#avatar-btn, .nav-avatar');
    if (avatarBtn) {
      e.preventDefault();
      e.stopPropagation();
      var dd = document.getElementById('nav-dropdown');
      if (dd) {
        dd.style.display = dd.style.display === 'none' || !dd.style.display ? 'block' : 'none';
        var nd = document.getElementById('notif-dropdown');
        if (nd) nd.style.display = 'none';
      }
      return;
    }

    var logoutBtn = target.closest('[data-action="logout"]');
    if (logoutBtn) {
      e.preventDefault();
      if (window.AngaziaAPI) AngaziaAPI.clearTokens();
      window.location.href = '/admin/logout';
      return;
    }

    var notifBtn = target.closest('#nav-notif-btn, [data-action="notifications"]');
    if (notifBtn) {
      e.preventDefault();
      e.stopPropagation();
      var nd = document.getElementById('notif-dropdown');
      if (nd) {
        var isOpen = nd.style.display === 'block';
        nd.style.display = isOpen ? 'none' : 'block';
        if (!isOpen) {
          aFetchNotifPreview();
        }
        var dd = document.getElementById('nav-dropdown');
        if (dd) dd.style.display = 'none';
      }
      return;
    }

    if (!target.closest('#nav-dropdown, #nav-avatar-wrap, #notif-dropdown, #nav-notif-btn')) {
      aCloseDropdown();
    }
  });

  document.addEventListener('DOMContentLoaded', function () {
    aInitNavbar();
    aInitFlashFromQuery();
    aFetchNotifCount();
    setInterval(aFetchNotifCount, 30000);

    if (window.AngaziaAPI && AngaziaAPI.setOnUnauthorized) {
      AngaziaAPI.setOnUnauthorized(function () {
        localStorage.removeItem('angazia_access_token');
        localStorage.removeItem('angazia_refresh_token');
        if (window.location.pathname !== '/admin/login') {
          window.location.href = '/admin/login';
        }
      });
    }

    initFlashMessages();
  });

  function initFlashMessages() {
    var f = document.querySelector('#flash-message');
    if (f) {
      setTimeout(function () {
        f.classList.add('flash-fade');
        setTimeout(function () { if (f.parentNode) f.parentNode.removeChild(f); }, 400);
      }, 5000);
    }
  }

  function aInitFlashFromQuery() {
    var params = new URLSearchParams(window.location.search);
    var msg = params.get('flash');
    var type = params.get('type') || 'success';
    if (msg) {
      if (window.AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast(decodeURIComponent(msg), type);
      }
      var url = new URL(window.location);
      url.searchParams.delete('flash');
      url.searchParams.delete('type');
      window.history.replaceState({}, '', url);
    }
  }
})();
