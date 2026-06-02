var AngaziaAuth = (function () {
  var TOKEN_KEY = 'angazia_access_token';
  var REFRESH_KEY = 'angazia_refresh_token';
  var USER_KEY = 'angazia_user';

  function getToken() { return localStorage.getItem(TOKEN_KEY); }
  function getRefreshToken() { return localStorage.getItem(REFRESH_KEY); }
  function getUser() {
    try { return JSON.parse(localStorage.getItem(USER_KEY)); }
    catch (e) { return null; }
  }

  function isLoggedIn() { return !!getToken(); }

  function isEmployee() { var u = getUser(); return u && u.role === 'employee'; }
  function isEmployer() { var u = getUser(); return u && u.role === 'employer'; }
  function isAdmin() { var u = getUser(); return u && u.role === 'admin'; }

  function decodeToken(token) {
    try {
      var parts = token.split('.');
      if (parts.length !== 3) return null;
      return JSON.parse(atob(parts[1]));
    } catch (e) { return null; }
  }

  function saveAuthResult(data) {
    var access = data.access_token || data.token;
    var refresh = data.refresh_token;
    if (access) localStorage.setItem(TOKEN_KEY, access);
    if (refresh) localStorage.setItem(REFRESH_KEY, refresh);
    if (data.user) localStorage.setItem(USER_KEY, JSON.stringify(data.user));
    AngaziaAPI.setTokens(access, refresh);
    if (window.AngaziaWS) {
      AngaziaWS.connect(access);
    }
  }

  function clearAuth() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_KEY);
    localStorage.removeItem(USER_KEY);
    AngaziaAPI.clearTokens();
    if (window.AngaziaWS) AngaziaWS.disconnect();
  }

  function redirectAfterLogin(role) {
    var target = '/dashboard';
    switch (role) {
      case 'employee': target = '/employee/dashboard'; break;
      case 'employer': target = '/employer/dashboard'; break;
      case 'admin': target = '/admin/dashboard'; break;
    }
    window.location.href = target;
  }

  function login(email, password) {
    return AngaziaAPI.auth.login({ email: email, password: password })
      .then(function (data) {
        saveAuthResult(data);
        return data;
      });
  }

  function register(data) {
    return AngaziaAPI.auth.register(data)
      .then(function (result) {
        saveAuthResult(result);
        return result;
      });
  }

  function logout() {
    return AngaziaAPI.auth.logout()
      .catch(function () {})
      .then(function () {
        clearAuth();
        window.location.href = '/login';
      });
  }

  function forgotPassword(email) {
    return AngaziaAPI.auth.forgotPassword({ email: email });
  }

  function resetPassword(token, password) {
    return AngaziaAPI.auth.resetPassword({ token: token, password: password });
  }

  function changePassword(current, newPass) {
    return AngaziaAPI.auth.changePassword({ current_password: current, new_password: newPass });
  }

  function resendVerification() {
    return AngaziaAPI.auth.resendVerification();
  }

  function loadProfile() {
    return AngaziaAPI.profile.get().then(function (user) {
      if (user) {
        localStorage.setItem(USER_KEY, JSON.stringify(user));
        if (window.updateNavbar) window.updateNavbar(user);
      }
      return user;
    });
  }

  function getDashboardUrl() {
    var u = getUser();
    if (!u) return '/login';
    switch (u.role) {
      case 'employee': return '/employee/dashboard';
      case 'employer': return '/employer/dashboard';
      case 'admin': return '/admin/dashboard';
      default: return '/dashboard';
    }
  }

  function getRoleBadge(role) {
    var labels = { employee: 'Candidate', employer: 'Employer', admin: 'Admin' };
    var colors = { employee: 'blue', employer: 'green', admin: 'red' };
    return '<span class="badge badge-' + (colors[role] || 'gray') + ' text-xs">'
      + (labels[role] || role) + '</span>';
  }

  function initAuthWatcher() {
    var token = getToken();
    if (!token) return;
    var decoded = decodeToken(token);
    if (!decoded || (decoded.exp && decoded.exp * 1000 < Date.now())) {
      clearAuth();
      return;
    }
    AngaziaAPI.setTokens(token, getRefreshToken());
    if (window.AngaziaWS) AngaziaWS.connect(token);
  }

  return {
    getToken: getToken,
    getRefreshToken: getRefreshToken,
    getUser: getUser,
    isLoggedIn: isLoggedIn,
    isEmployee: isEmployee,
    isEmployer: isEmployer,
    isAdmin: isAdmin,
    decodeToken: decodeToken,
    login: login,
    register: register,
    logout: logout,
    forgotPassword: forgotPassword,
    resetPassword: resetPassword,
    changePassword: changePassword,
    resendVerification: resendVerification,
    loadProfile: loadProfile,
    saveAuthResult: saveAuthResult,
    clearAuth: clearAuth,
    redirectAfterLogin: redirectAfterLogin,
    getDashboardUrl: getDashboardUrl,
    getRoleBadge: getRoleBadge,
    initAuthWatcher: initAuthWatcher,
  };
})();

(function () {
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () { AngaziaAuth.initAuthWatcher(); });
  } else {
    AngaziaAuth.initAuthWatcher();
  }
})();
