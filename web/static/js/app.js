var AngaziaApp = (function () {
  var initialized = false;
  var currentUser = null;

  function init() {
    if (initialized) return;
    initialized = true;

    currentUser = AngaziaAuth.getUser();

    setupGlobalListeners();
    initTheme();
    initNavbar();
    initMobileMenu();
    initDropdowns();
    initTooltips();
    initNotificationCenter();

    if (AngaziaAuth.isLoggedIn()) {
      AngaziaNotifications.fetchUnreadCount();
    }

    initFlashFromQuery();

    emit('ready');
  }

  function setupGlobalListeners() {
    AngaziaAPI.setOnUnauthorized(function () {
      AngaziaAuth.clearAuth();
      window.location.href = '/logout';
    });

    AngaziaNotifications.on('received', function (notification) {
      updatePageCounts(notification);
    });
  }

  function initTheme() {
    var theme = localStorage.getItem('angazia_theme') || 'system';
    applyTheme(theme);
    var toggle = document.getElementById('theme-toggle');
    if (toggle) {
      toggle.addEventListener('click', function () {
        var current = localStorage.getItem('angazia_theme') || 'system';
        var next = current === 'light' ? 'dark' : current === 'dark' ? 'system' : 'light';
        localStorage.setItem('angazia_theme', next);
        applyTheme(next);
        updateThemeIcon(next);
      });
    }
  }

  function applyTheme(theme) {
    var isDark;
    if (theme === 'dark') isDark = true;
    else if (theme === 'light') isDark = false;
    else isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    document.documentElement.classList.toggle('dark', isDark);
    var meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.setAttribute('content', isDark ? '#1f2937' : '#ffffff');
  }

  function updateThemeIcon(theme) {
    var icons = document.querySelectorAll('.theme-icon');
    icons.forEach(function (el) {
      el.textContent = theme === 'dark' ? '\u2600' : theme === 'light' ? '\uD83C\uDF19' : '\uD83D\uDD04';
    });
  }

  function initNavbar() {
    if (!currentUser) return;
    updateNavbar(currentUser);
  }

  window.updateNavbar = function (user) {
    currentUser = user;
    var avatar = document.getElementById('user-avatar');
    var name = document.getElementById('user-name');
    var email = document.getElementById('user-email');
    var roleBadge = document.getElementById('user-role-badge');
    var loginBtn = document.getElementById('login-btn');
    var registerBtn = document.getElementById('register-btn');
    var userMenu = document.getElementById('user-menu');
    var dashboardLink = document.getElementById('dashboard-link');

    if (avatar && user.avatar_url) avatar.src = user.avatar_url;
    if (name) {
      var displayName = user.full_name || user.company_name || user.email;
      name.textContent = displayName;
    }
    if (email) email.textContent = user.email;
    if (roleBadge) roleBadge.innerHTML = AngaziaAuth.getRoleBadge(user.role);
    if (loginBtn) loginBtn.classList.add('hidden');
    if (registerBtn) registerBtn.classList.add('hidden');
    if (userMenu) userMenu.classList.remove('hidden');
    if (dashboardLink) dashboardLink.href = AngaziaAuth.getDashboardUrl();

    var sidebarLinks = document.querySelectorAll('[data-role-sidebar]');
    sidebarLinks.forEach(function (el) {
      el.classList.toggle('hidden', el.getAttribute('data-role') !== user.role);
    });
  };

  function initMobileMenu() {
    var toggle = document.getElementById('mobile-menu-toggle');
    var menu = document.getElementById('mobile-menu');
    if (toggle && menu) {
      toggle.addEventListener('click', function () {
        menu.classList.toggle('hidden');
        toggle.classList.toggle('active');
      });
    }
  }

  function initDropdowns() {
    document.addEventListener('click', function (e) {
      var dropdowns = document.querySelectorAll('.dropdown-menu');
      dropdowns.forEach(function (menu) {
        var trigger = menu.closest('.dropdown');
        if (trigger && !trigger.contains(e.target)) {
          menu.classList.add('hidden');
        }
      });
    });

    document.addEventListener('click', function (e) {
      var trigger = e.target.closest('[data-dropdown-toggle]');
      if (trigger) {
        var targetId = trigger.getAttribute('data-dropdown-toggle');
        var menu = document.getElementById(targetId);
        if (menu) {
          menu.classList.toggle('hidden');
          e.stopPropagation();
        }
      }
    });
  }

  function initTooltips() {
    document.querySelectorAll('[data-tooltip]').forEach(function (el) {
      el.addEventListener('mouseenter', function () {
        var text = el.getAttribute('data-tooltip');
        var tip = document.createElement('div');
        tip.className = 'tooltip absolute z-50 px-2 py-1 text-xs text-white bg-gray-900 dark:bg-gray-100 dark:text-gray-900 rounded shadow-lg whitespace-nowrap';
        tip.textContent = text;
        tip.style.pointerEvents = 'none';
        var rect = el.getBoundingClientRect();
        tip.style.top = (rect.top - 30) + 'px';
        tip.style.left = (rect.left + rect.width / 2 - tip.offsetWidth / 2) + 'px';
        document.body.appendChild(tip);
        el._tooltip = tip;
      });
      el.addEventListener('mouseleave', function () {
        if (el._tooltip) { el._tooltip.remove(); el._tooltip = null; }
      });
    });
  }

  function initNotificationCenter() {
    var bell = document.getElementById('notification-bell');
    var panel = document.getElementById('notification-panel');
    if (!bell || !panel) return;

    bell.addEventListener('click', function (e) {
      e.stopPropagation();
      var isOpen = !panel.classList.contains('hidden');
      panel.classList.toggle('hidden');
      if (!isOpen) {
        loadNotificationPanel();
      }
    });

    document.addEventListener('click', function () {
      if (panel && !panel.classList.contains('hidden')) {
        panel.classList.add('hidden');
      }
    });

    if (panel) {
      panel.addEventListener('click', function (e) { e.stopPropagation(); });
    }
  }

  function loadNotificationPanel() {
    var list = document.getElementById('notification-list');
    var empty = document.getElementById('notification-empty');
    var loading = document.getElementById('notification-loading');
    if (!list) return;

    if (loading) loading.classList.remove('hidden');
    if (empty) empty.classList.add('hidden');
    list.innerHTML = '';

    AngaziaNotifications.fetchUnread()
      .then(function (notifications) {
        if (loading) loading.classList.add('hidden');
        if (!notifications || notifications.length === 0) {
          if (empty) empty.classList.remove('hidden');
          return;
        }
        notifications.slice(0, 10).forEach(function (n) {
          list.appendChild(createNotificationItem(n));
        });
      })
      .catch(function () {
        if (loading) loading.classList.add('hidden');
        if (empty) empty.classList.remove('hidden');
      });
  }

  function createNotificationItem(n) {
    var el = document.createElement('div');
    el.className = 'flex items-start gap-3 p-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer border-b border-gray-100 dark:border-gray-700 last:border-0';
    if (!n.is_read) el.classList.add('bg-blue-50 dark:bg-blue-900/20');
    el.innerHTML = '<span class="text-lg flex-shrink-0 mt-0.5">' + getNotifIcon(n.type) + '</span>'
      + '<div class="flex-1 min-w-0">'
      + '<p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">' + escapeHtml(n.title) + '</p>'
      + '<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-2">' + escapeHtml(n.content || n.body) + '</p>'
      + '<p class="text-xs text-gray-400 mt-1">' + timeAgo(n.created_at) + '</p>'
      + '</div>'
      + '<div class="flex-shrink-0 flex flex-col gap-1">'
      + (!n.is_read ? '<button class="text-xs text-blue-600 hover:text-blue-800 mark-read-btn" data-id="' + n.id + '">Mark read</button>' : '')
      + '</div>';
    if (n.action_url) {
      el.addEventListener('click', function (e) {
        if (e.target.closest('.mark-read-btn')) return;
        if (!n.is_read) AngaziaNotifications.markAsRead(n.id);
        window.location.href = n.action_url;
      });
    }
    var markBtn = el.querySelector('.mark-read-btn');
    if (markBtn) {
      markBtn.addEventListener('click', function (e) {
        e.stopPropagation();
        AngaziaNotifications.markAsRead(n.id);
        el.classList.remove('bg-blue-50 dark:bg-blue-900/20');
        markBtn.remove();
      });
    }
    return el;
  }

  function getNotifIcon(type) {
    var icons = {
      application_received: '\uD83D\uDCE5',
      application_status: '\u2705',
      application_shortlisted: '\u2B50',
      application_rejected: '\u274C',
      interview_scheduled: '\uD83D\uDCC5',
      job_match: '\uD83D\uDC4D',
      job_alert: '\uD83D\uDD14',
      message: '\uD83D\uDCAC',
      system: '\u2699\uFE0F',
      security: '\uD83D\uDD12',
      verification: '\u270F\uFE0F',
      subscription: '\uD83D\uDCB3',
      payment: '\uD83D\uDCB0',
    };
    return icons[type] || '\uD83D\uDD14';
  }

  function escapeHtml(text) {
    if (!text) return '';
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(text));
    return div.innerHTML;
  }

  function timeAgo(dateStr) {
    if (!dateStr) return '';
    var date = new Date(dateStr);
    var now = new Date();
    var diffMs = now - date;
    var diffSec = Math.floor(diffMs / 1000);
    if (diffSec < 60) return 'just now';
    var diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return diffMin + 'm ago';
    var diffHour = Math.floor(diffMin / 60);
    if (diffHour < 24) return diffHour + 'h ago';
    var diffDay = Math.floor(diffHour / 24);
    if (diffDay < 7) return diffDay + 'd ago';
    var diffWeek = Math.floor(diffDay / 7);
    if (diffWeek < 4) return diffWeek + 'w ago';
    return date.toLocaleDateString();
  }

  function updatePageCounts(notification) {
    var countEl = document.getElementById('notif-count');
    if (countEl) {
      var current = parseInt(countEl.textContent) || 0;
      countEl.textContent = current + 1;
      countEl.classList.remove('hidden');
    }
  }

  function initFlashFromQuery() {
    var params = new URLSearchParams(window.location.search);
    var msg = params.get('flash');
    var type = params.get('type') || 'success';
    if (msg) {
      showToast(decodeURIComponent(msg), type);
      var url = new URL(window.location);
      url.searchParams.delete('flash');
      url.searchParams.delete('type');
      window.history.replaceState({}, '', url);
    }
  }

  function emit(evt) {
    document.dispatchEvent(new CustomEvent('angazia:' + evt));
  }

  function on(evt, fn) {
    document.addEventListener('angazia:' + evt, fn);
  }

  function showLoading(show) {
    var loader = document.getElementById('global-loader');
    if (loader) loader.classList.toggle('hidden', !show);
  }

  function showToast(message, type) {
    type = type || 'info';
    var container = document.getElementById('toast-container');
    if (!container) {
      container = document.createElement('div');
      container.id = 'toast-container';
      container.className = 'toast-container';
      document.body.appendChild(container);
    }
    var icons = {
      success: '\u2713',
      error: '\u2717',
      warning: '\u26A0',
      info: '\u2139',
    };
    var toast = document.createElement('div');
    toast.className = 'toast toast-' + type;
    toast.innerHTML = '<span class="toast-icon">' + (icons[type] || icons.info) + '</span><span class="toast-message">' + escapeHtml(message) + '</span><button class="toast-close" onclick="this.parentElement.remove()">&times;</button>';
    container.appendChild(toast);
    setTimeout(function () {
      toast.classList.add('toast-exit');
      setTimeout(function () { if (toast.parentNode) toast.remove(); }, 300);
    }, 4000);
  }

  function confirmDialog(message) {
    return new Promise(function (resolve) {
      var overlay = document.createElement('div');
      overlay.className = 'fixed inset-0 z-50 flex items-center justify-center bg-black/50';
      overlay.innerHTML = '<div class="bg-white dark:bg-gray-800 rounded-xl shadow-2xl p-6 max-w-md mx-4">'
        + '<p class="text-gray-700 dark:text-gray-300 mb-6">' + escapeHtml(message) + '</p>'
        + '<div class="flex justify-end gap-3">'
        + '<button class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 cancel-btn">Cancel</button>'
        + '<button class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 confirm-btn">Confirm</button>'
        + '</div></div>';
      document.body.appendChild(overlay);
      overlay.querySelector('.confirm-btn').addEventListener('click', function () { overlay.remove(); resolve(true); });
      overlay.querySelector('.cancel-btn').addEventListener('click', function () { overlay.remove(); resolve(false); });
      overlay.addEventListener('click', function (e) { if (e.target === overlay) { overlay.remove(); resolve(false); } });
    });
  }

  function handleFormErrors(form, errors) {
    form.querySelectorAll('.field-error').forEach(function (el) { el.remove(); });
    form.querySelectorAll('.is-invalid').forEach(function (el) { el.classList.remove('is-invalid'); });
    if (!errors) return;
    Object.keys(errors).forEach(function (field) {
      var input = form.querySelector('[name="' + field + '"]');
      if (input) {
        input.classList.add('is-invalid');
        var errorEl = document.createElement('p');
        errorEl.className = 'field-error text-red-500 text-xs mt-1';
        errorEl.textContent = Array.isArray(errors[field]) ? errors[field][0] : errors[field];
        input.parentNode.appendChild(errorEl);
      }
    });
  }

  function serializeForm(form) {
    var data = {};
    var fd = new FormData(form);
    fd.forEach(function (value, key) { data[key] = value; });
    return data;
  }

  function getCSRFToken() {
    var meta = document.querySelector('meta[name="csrf-token"]');
    return meta ? meta.getAttribute('content') : null;
  }

  function debounce(fn, delay) {
    var timer;
    return function () {
      var args = arguments;
      var ctx = this;
      clearTimeout(timer);
      timer = setTimeout(function () { fn.apply(ctx, args); }, delay);
    };
  }

  return {
    init: init,
    on: on,
    showToast: showToast,
    showLoading: showLoading,
    confirmDialog: confirmDialog,
    handleFormErrors: handleFormErrors,
    serializeForm: serializeForm,
    getCSRFToken: getCSRFToken,
    debounce: debounce,
    timeAgo: timeAgo,
    escapeHtml: escapeHtml,
  };
})();

(function () {
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () { AngaziaApp.init(); });
  } else {
    AngaziaApp.init();
  }
})();
