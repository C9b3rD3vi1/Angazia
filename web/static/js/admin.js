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
    }
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

    if (!target.closest('#nav-dropdown, #nav-avatar-wrap')) {
      aCloseDropdown();
    }
  });

  function updateModerationBadge() {
    if (!window.AngaziaAPI) return;
    AngaziaAPI.admin.moderationPendingCount()
      .then(function (data) {
        var count = data.count || 0;
        var badge = document.getElementById('sidebar-moderation-badge');
        if (badge) {
          if (count > 0) {
            badge.textContent = count;
            badge.style.display = 'inline-flex';
          } else {
            badge.style.display = 'none';
          }
        }
      })
      .catch(function () {});
  }

  document.addEventListener('DOMContentLoaded', function () {
    aInitNavbar();
    aInitFlashFromQuery();

    if (window.AngaziaNotifications) {
      AngaziaNotifications.fetchUnreadCount();
    }

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

    setInterval(updateModerationBadge, 30000);
    setTimeout(updateModerationBadge, 5000);
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
