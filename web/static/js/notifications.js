var AngaziaNotifications = (function () {
  var unreadCount = 0;
  var cachedNotifications = [];
  var pageSize = 20;
  var listeners = {};

  function on(evt, fn) {
    if (!listeners[evt]) listeners[evt] = [];
    listeners[evt].push(fn);
  }

  function emit(evt, data) {
    (listeners[evt] || []).forEach(function (fn) { try { fn(data); } catch (e) { console.error(e); } });
  }

  function fetchUnreadCount() {
    return AngaziaAPI.notifications.counts()
      .then(function (data) {
        unreadCount = data.unread || data.total_unread || 0;
        emit('countChanged', unreadCount);
        return unreadCount;
      })
      .catch(function () { return 0; });
  }

  function fetchNotifications(params) {
    return AngaziaAPI.notifications.list(params || { page: 1, limit: pageSize })
      .then(function (data) {
        var list = data.notifications || data.data || data || [];
        cachedNotifications = list;
        emit('loaded', list);
        return list;
      });
  }

  function fetchUnread() {
    return AngaziaAPI.notifications.unread()
      .then(function (data) {
        var list = data.notifications || data.data || data || [];
        emit('unreadLoaded', list);
        return list;
      });
  }

  function markAsRead(id) {
    return AngaziaAPI.notifications.markRead(id).then(function () {
      unreadCount = Math.max(0, unreadCount - 1);
      emit('countChanged', unreadCount);
      emit('read', id);
    });
  }

  function markAllAsRead() {
    return AngaziaAPI.notifications.markAllRead().then(function () {
      unreadCount = 0;
      emit('countChanged', 0);
      emit('allRead');
    });
  }

  function markMultipleAsRead(ids) {
    if (!ids || ids.length === 0) return Promise.resolve();
    return AngaziaAPI.notifications.markMultipleRead(ids).then(function () {
      unreadCount = Math.max(0, unreadCount - ids.length);
      emit('countChanged', unreadCount);
      emit('multipleRead', ids);
    });
  }

  function archiveNotification(id) {
    return AngaziaAPI.notifications.archive(id).then(function () {
      emit('archived', id);
    });
  }

  function deleteNotification(id) {
    return AngaziaAPI.notifications.delete(id).then(function () {
      emit('deleted', id);
    });
  }

  function deleteAllNotifications() {
    return AngaziaAPI.notifications.deleteAll().then(function () {
      unreadCount = 0;
      cachedNotifications = [];
      emit('countChanged', 0);
      emit('allDeleted');
    });
  }

  function fetchPreferences() {
    return AngaziaAPI.notifications.getPreferences();
  }

  function updatePreferences(prefs) {
    return AngaziaAPI.notifications.updatePreferences(prefs).then(function (data) {
      emit('preferencesUpdated', data);
      return data;
    });
  }

  function onNotificationReceived(notification) {
    unreadCount++;
    emit('countChanged', unreadCount);
    emit('received', notification);
    showToast(notification);
  }

  function showToast(notification) {
    var icon = getNotificationIcon(notification.type);
    var container = document.getElementById('notification-toast-container');
    if (!container) {
      container = document.createElement('div');
      container.id = 'notification-toast-container';
      container.className = 'fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm';
      document.body.appendChild(container);
    }
    var toast = document.createElement('div');
    toast.className = 'bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-4 transform transition-all duration-300 translate-x-0 opacity-100 cursor-pointer';
    toast.innerHTML = '<div class="flex items-start gap-3">'
      + '<span class="text-xl flex-shrink-0">' + icon + '</span>'
      + '<div class="flex-1 min-w-0">'
      + '<p class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">' + escapeHtml(notification.title || '') + '</p>'
      + '<p class="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">' + escapeHtml(notification.content || notification.body || '') + '</p>'
      + '</div>'
      + '<button onclick="this.parentElement.parentElement.remove()" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 flex-shrink-0">&times;</button>'
      + '</div>';
    if (notification.action_url) {
      toast.addEventListener('click', function () { window.location.href = notification.action_url; });
    }
    container.appendChild(toast);
    setTimeout(function () {
      toast.style.opacity = '0';
      toast.style.transform = 'translateX(100%)';
      setTimeout(function () { if (toast.parentNode) toast.remove(); }, 300);
    }, 5000);
  }

  function getNotificationIcon(type) {
    var icons = {
      application_received: '\uD83D\uDCE5',
      application_status: '\u2705',
      application_shortlisted: '\u2B50',
      application_rejected: '\u274C',
      application: '\uD83D\uDCE5',
      interview_scheduled: '\uD83D\uDCC5',
      interview: '\uD83D\uDCC5',
      job_match: '\uD83D\uDC4D',
      job_alert: '\uD83D\uDD14',
      job: '\uD83D\uDC4D',
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

  function updateBellBadge(count) {
    var badges = document.querySelectorAll('.notification-badge, .notif-badge');
    badges.forEach(function (el) {
      if (count > 0) {
        el.textContent = count > 99 ? '99+' : count;
        el.classList.remove('hidden');
      } else {
        el.classList.add('hidden');
      }
    });
  }

  function getUnreadCount() { return unreadCount; }

  on('countChanged', function (count) { updateBellBadge(count); });

  return {
    on: on,
    fetchUnreadCount: fetchUnreadCount,
    fetchNotifications: fetchNotifications,
    fetchUnread: fetchUnread,
    markAsRead: markAsRead,
    markAllAsRead: markAllAsRead,
    markMultipleAsRead: markMultipleAsRead,
    archiveNotification: archiveNotification,
    deleteNotification: deleteNotification,
    deleteAllNotifications: deleteAllNotifications,
    fetchPreferences: fetchPreferences,
    updatePreferences: updatePreferences,
    onNotificationReceived: onNotificationReceived,
    getUnreadCount: getUnreadCount,
  };
})();
