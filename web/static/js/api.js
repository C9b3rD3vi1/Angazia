var AngaziaAPI = (function () {
  const BASE_URL = '/api/v1';

  let accessToken = null;
  let refreshToken = null;
  let onUnauthorized = null;
  let onTokenRefreshed = null;

  function getAccessToken() {
    if (accessToken) return accessToken;
    accessToken = localStorage.getItem('angazia_access_token');
    return accessToken;
  }

  function getRefreshToken() {
    if (refreshToken) return refreshToken;
    refreshToken = localStorage.getItem('angazia_refresh_token');
    return refreshToken;
  }

  function setTokens(access, refresh) {
    accessToken = access;
    refreshToken = refresh;
    if (access) localStorage.setItem('angazia_access_token', access);
    else localStorage.removeItem('angazia_access_token');
    if (refresh) localStorage.setItem('angazia_refresh_token', refresh);
    else localStorage.removeItem('angazia_refresh_token');
  }

  function clearTokens() {
    accessToken = null;
    refreshToken = null;
    localStorage.removeItem('angazia_access_token');
    localStorage.removeItem('angazia_refresh_token');
  }

  function buildHeaders(extra) {
    var h = { 'Content-Type': 'application/json' };
    var token = getAccessToken();
    if (token) h['Authorization'] = 'Bearer ' + token;
    var ct = document.querySelector('meta[name="csrf-token"]');
    if (ct) h['X-CSRF-Token'] = ct.getAttribute('content');
    return Object.assign(h, extra || {});
  }

  function handleResponse(resp) {
    if (resp.status === 204) return null;
    return resp.json().then(function (body) {
      if (!resp.ok) {
        var err = new Error(body.error || body.message || 'Request failed');
        err.status = resp.status;
        err.body = body;
        if (resp.status === 401 && onUnauthorized) onUnauthorized(err);
        throw err;
      }
      return body.data !== undefined ? body.data : body;
    });
  }

  function refreshAccessToken() {
    var rt = getRefreshToken();
    if (!rt) return Promise.reject(new Error('No refresh token'));
    return fetch(BASE_URL + '/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: rt }),
    }).then(function (resp) {
      if (!resp.ok) {
        clearTokens();
        if (onUnauthorized) onUnauthorized(new Error('Refresh failed'));
        throw new Error('Session expired');
      }
      return resp.json().then(function (body) {
        var d = body.data || body;
        setTokens(d.access_token, d.refresh_token);
        if (onTokenRefreshed) onTokenRefreshed(d.access_token);
        return d.access_token;
      });
    });
  }

  function request(method, path, body, extraHeaders) {
    var url = BASE_URL + path;
    var opts = { method: method, headers: buildHeaders(extraHeaders) };
    if (body && method !== 'GET') opts.body = JSON.stringify(body);

    return fetch(url, opts)
      .then(function (resp) {
        if (resp.status === 401 && getRefreshToken()) {
          return refreshAccessToken().then(function () {
            opts.headers = buildHeaders(extraHeaders);
            return fetch(url, opts).then(handleResponse);
          });
        }
        return handleResponse(resp);
      });
  }

  function upload(path, formData, onProgress) {
    var url = BASE_URL + path;
    var token = getAccessToken();
    return new Promise(function (resolve, reject) {
      var xhr = new XMLHttpRequest();
      xhr.open('POST', url);
      if (token) xhr.setRequestHeader('Authorization', 'Bearer ' + token);
      if (onProgress && xhr.upload) {
        xhr.upload.addEventListener('progress', function (e) {
          if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
        });
      }
      xhr.onload = function () {
        try {
          var body = JSON.parse(xhr.responseText);
          if (xhr.status >= 200 && xhr.status < 300) resolve(body.data || body);
          else reject({ status: xhr.status, body: body });
        } catch (e) { reject(e); }
      };
      xhr.onerror = function () { reject(new Error('Network error')); };
      xhr.send(formData);
    });
  }

  function apiGet(path, params) {
    if (params) {
      var qs = Object.keys(params)
        .filter(function (k) { return params[k] !== undefined && params[k] !== null && params[k] !== ''; })
        .map(function (k) { return encodeURIComponent(k) + '=' + encodeURIComponent(params[k]); })
        .join('&');
      if (qs) path += '?' + qs;
    }
    return request('GET', path);
  }

  function apiPost(path, body) { return request('POST', path, body); }
  function apiPut(path, body) { return request('PUT', path, body); }
  function apiDelete(path, body) { return request('DELETE', path, body); }

  return {
    setTokens: setTokens,
    getAccessToken: getAccessToken,
    getRefreshToken: getRefreshToken,
    clearTokens: clearTokens,
    setOnUnauthorized: function (fn) { onUnauthorized = fn; },
    setOnTokenRefreshed: function (fn) { onTokenRefreshed = fn; },

    get: apiGet,
    post: apiPost,
    put: apiPut,
    del: apiDelete,
    upload: upload,

    auth: {
      register: function (data) { return apiPost('/auth/register', data); },
      login: function (data) { return apiPost('/auth/login', data); },
      refresh: function (token) { return apiPost('/auth/refresh', { refresh_token: token }); },
      logout: function () { return apiPost('/auth/logout'); },
      forgotPassword: function (data) { return apiPost('/auth/forgot-password', data); },
      resetPassword: function (data) { return apiPost('/auth/reset-password', data); },
      changePassword: function (data) { return apiPost('/auth/change-password', data); },
      verifyEmail: function (token) { return apiGet('/auth/verify-email/' + token); },
      resendVerification: function () { return apiPost('/auth/resend-verification'); },
      twoFA: {
        setup: function () { return apiPost('/auth/2fa/setup'); },
        verify: function (data) { return apiPost('/auth/2fa/verify', data); },
        disable: function (data) { return apiPost('/auth/2fa/disable', data); },
        status: function () { return apiGet('/auth/2fa/status'); },
        generateBackupCodes: function () { return apiPost('/auth/2fa/backup-codes/generate'); },
        getBackupCodes: function () { return apiGet('/auth/2fa/backup-codes'); },
        recovery: function (data) { return apiPost('/auth/2fa/recovery', data); },
        recover: function (data) { return apiGet('/auth/2fa/recover', data); },
      },
    },

    jobs: {
      list: function (params) { return apiGet('/jobs', params); },
      featured: function () { return apiGet('/jobs/featured'); },
      recent: function () { return apiGet('/jobs/recent'); },
      search: function (params) { return apiGet('/jobs/search', params); },
      get: function (id) { return apiGet('/jobs/' + id); },
      similar: function (id) { return apiGet('/jobs/' + id + '/similar'); },
      save: function (id) { return apiPost('/jobs/' + id + '/save'); },
      unsave: function (id) { return apiDelete('/jobs/' + id + '/save'); },
      saved: function () { return apiGet('/employee/saved-jobs'); },
      create: function (data) { return apiPost('/employer/jobs', data); },
      myJobs: function (params) { return apiGet('/employer/jobs', params); },
      update: function (id, data) { return apiPut('/employer/jobs/' + id, data); },
      delete: function (id) { return apiDelete('/employer/jobs/' + id); },
      close: function (id) { return apiPost('/employer/jobs/' + id + '/close'); },
    },

    applications: {
      apply: function (data) { return apiPost('/employee/applications', data); },
      myApplications: function (params) { return apiGet('/employee/applications', params); },
      stats: function () { return apiGet('/employee/applications/stats'); },
      withdraw: function (id) { return apiPost('/employee/applications/' + id + '/withdraw'); },
      get: function (id) { return apiGet('/applications/' + id); },
      companyApplications: function (params) { return apiGet('/employer/applications', params); },
      jobApplications: function (jobId, params) { return apiGet('/employer/jobs/' + jobId + '/applications', params); },
      shortlist: function (id) { return apiPost('/employer/applications/' + id + '/shortlist'); },
      reject: function (id) { return apiPost('/employer/applications/' + id + '/reject'); },
      interview: function (id, data) { return apiPost('/employer/applications/' + id + '/interview', data); },
      bulkShortlist: function (data) { return apiPost('/employer/applications/bulk-shortlist', data); },
    },

    companies: {
      get: function (id) { return apiGet('/companies/' + id); },
      badges: function (id) { return apiGet('/companies/' + id + '/badges'); },
      reviews: function (id, params) { return apiGet('/companies/' + id + '/reviews', params); },
      reviewStats: function (id) { return apiGet('/companies/' + id + '/reviews/stats'); },
      submitReview: function (id, data) { return apiPost('/companies/' + id + '/reviews', data); },
      markHelpful: function (id) { return apiPost('/reviews/' + id + '/helpful'); },
      myReviews: function () { return apiGet('/user/reviews'); },
      myCompany: function () { return apiGet('/employer/company'); },
      updateCompany: function (data) { return apiPut('/employer/company', data); },
      uploadLogo: function (fd, cb) { return upload('/employer/company/logo', fd, cb); },
      verify: function () { return apiPost('/employer/company/verify'); },
      verificationStatus: function () { return apiGet('/employer/company/verification'); },
      getBadges: function () { return apiGet('/employer/company/badges'); },
      team: function () { return apiGet('/employer/team'); },
      inviteTeam: function (data) { return apiPost('/employer/team/invite', data); },
      removeTeamMember: function (id) { return apiDelete('/employer/team/' + id); },
      analytics: function () { return apiGet('/employer/analytics'); },
    },

    notifications: {
      list: function (params) { return apiGet('/notifications', params); },
      unread: function () { return apiGet('/notifications/unread'); },
      counts: function () { return apiGet('/notifications/counts'); },
      get: function (id) { return apiGet('/notifications/' + id); },
      markRead: function (id) { return apiPost('/notifications/' + id + '/read'); },
      markAllRead: function () { return apiPost('/notifications/read-all'); },
      markMultipleRead: function (ids) { return apiPost('/notifications/read-multiple', { ids: ids }); },
      archive: function (id) { return apiPost('/notifications/' + id + '/archive'); },
      delete: function (id) { return apiDelete('/notifications/' + id); },
      deleteAll: function () { return apiDelete('/notifications'); },
      getPreferences: function () { return apiGet('/notifications/preferences'); },
      updatePreferences: function (data) { return apiPut('/notifications/preferences', data); },
    },

    search: {
      jobs: function (params) { return apiGet('/search/jobs', params); },
      jobFacets: function (params) { return apiGet('/search/jobs/facets', params); },
      companies: function (params) { return apiGet('/search/companies', params); },
      candidates: function (params) { return apiGet('/search/candidates', params); },
      autoComplete: function (params) { return apiGet('/search/auto-complete', params); },
      popular: function () { return apiGet('/search/popular'); },
      history: function () { return apiGet('/search/history'); },
      saved: function () { return apiGet('/search/saved'); },
      saveSearch: function (data) { return apiPost('/search/saved', data); },
      deleteSaved: function (id) { return apiDelete('/search/saved/' + id); },
      runSaved: function (id) { return apiGet('/search/saved/' + id + '/run'); },
    },

    matches: {
      jobMatches: function (params) { return apiGet('/employee/matches/jobs', params); },
      coverLetter: function (data) { return apiPost('/employee/matches/cover-letter', data); },
      skillsGap: function (jobId) { return apiGet('/employee/matches/skills-gap/' + jobId); },
      analysis: function (jobId, empId) { return apiGet('/employee/matches/analysis/' + jobId + '/' + empId); },
      candidateMatches: function (jobId) { return apiGet('/employer/matches/candidates/' + jobId); },
      interviewQuestions: function (jobId) { return apiGet('/employer/matches/interview-questions/' + jobId); },
    },

    alerts: {
      list: function (params) { return apiGet('/alerts', params); },
      get: function (id) { return apiGet('/alerts/' + id); },
      create: function (data) { return apiPost('/alerts/search', data); },
      update: function (id, data) { return apiPut('/alerts/' + id, data); },
      delete: function (id) { return apiDelete('/alerts/' + id); },
      test: function (id) { return apiPost('/alerts/' + id + '/test'); },
      settings: function () { return apiGet('/alerts/settings'); },
      updateSettings: function (data) { return apiPut('/alerts/settings', data); },
      history: function (params) { return apiGet('/alerts/history', params); },
    },

    talentPools: {
      list: function (params) { return apiGet('/employer/talent-pools', params); },
      stats: function () { return apiGet('/employer/talent-pools/stats'); },
      get: function (id) { return apiGet('/employer/talent-pools/' + id); },
      create: function (data) { return apiPost('/employer/talent-pools', data); },
      update: function (id, data) { return apiPut('/employer/talent-pools/' + id, data); },
      delete: function (id) { return apiDelete('/employer/talent-pools/' + id); },
      poolStats: function (id) { return apiGet('/employer/talent-pools/' + id + '/stats'); },
      searchPool: function (id, params) { return apiGet('/employer/talent-pools/' + id + '/search', params); },
      candidates: function (id, params) { return apiGet('/employer/talent-pools/' + id + '/candidates', params); },
      addCandidate: function (id, data) { return apiPost('/employer/talent-pools/' + id + '/candidates', data); },
      updateCandidate: function (poolId, candId, data) { return apiPut('/employer/talent-pools/' + poolId + '/candidates/' + candId, data); },
      removeCandidate: function (poolId, candId) { return apiDelete('/employer/talent-pools/' + poolId + '/candidates/' + candId); },
      markContacted: function (poolId, candId) { return apiPost('/employer/talent-pools/' + poolId + '/candidates/' + candId + '/contact'); },
      markHired: function (poolId, candId) { return apiPost('/employer/talent-pools/' + poolId + '/candidates/' + candId + '/hire'); },
    },

    analytics: {
      employerDashboard: function () { return apiGet('/employer/analytics/dashboard'); },
      employerTrends: function () { return apiGet('/employer/analytics/trends'); },
      employerFunnel: function () { return apiGet('/employer/analytics/funnel'); },
      employerJobs: function () { return apiGet('/employer/analytics/jobs'); },
      employerJobDetail: function (id) { return apiGet('/employer/analytics/jobs/' + id); },
      timeToHire: function () { return apiGet('/employer/analytics/time-to-hire'); },
      quality: function () { return apiGet('/employer/analytics/quality'); },
      sources: function () { return apiGet('/employer/analytics/sources'); },
      export: function () { return apiGet('/employer/analytics/export'); },
      candidateDashboard: function () { return apiGet('/employee/analytics/dashboard'); },
      profileStrength: function () { return apiGet('/employee/analytics/profile-strength'); },
      appStats: function () { return apiGet('/employee/analytics/applications/stats'); },
      monthlyStats: function () { return apiGet('/employee/analytics/applications/monthly'); },
      successRates: function () { return apiGet('/employee/analytics/success-rates'); },
      skillGap: function () { return apiGet('/employee/analytics/skill-gap'); },
      marketPositioning: function () { return apiGet('/employee/analytics/market-positioning'); },
      recommendations: function () { return apiGet('/employee/analytics/recommendations'); },
      recentActivity: function () { return apiGet('/employee/analytics/recent-activity'); },
    },

    admin: {
      platformStats: function () { return apiGet('/admin/stats/platform'); },
      userStats: function () { return apiGet('/admin/stats/users'); },
      jobStats: function () { return apiGet('/admin/stats/jobs'); },
      engagementStats: function () { return apiGet('/admin/stats/engagement'); },
      users: function (params) { return apiGet('/admin/users', params); },
      userDetail: function (id) { return apiGet('/admin/users/' + id); },
      suspendUser: function (id) { return apiPost('/admin/users/' + id + '/suspend'); },
      activateUser: function (id) { return apiPost('/admin/users/' + id + '/activate'); },
      deleteUser: function (id) { return apiDelete('/admin/users/' + id); },
      verifyUser: function (id) { return apiPost('/admin/users/' + id + '/verify'); },
      moderation: function (params) { return apiGet('/admin/moderation', params); },
      approveContent: function (id) { return apiPost('/admin/moderation/' + id + '/approve'); },
      rejectContent: function (id) { return apiPost('/admin/moderation/' + id + '/reject'); },
      settings: function () { return apiGet('/admin/settings'); },
      updateSetting: function (key, data) { return apiPut('/admin/settings/' + key, data); },
      reportReasons: function () { return apiGet('/admin/report-reasons'); },
      auditLogs: function (params) { return apiGet('/admin/audit-logs', params); },
      report: function (data) { return apiPost('/report', data); },
    },

    plans: {
      list: function () { return apiGet('/plans'); },
      get: function (id) { return apiGet('/plans/' + id); },
      adminList: function () { return apiGet('/admin/plans'); },
      adminGet: function (id) { return apiGet('/admin/plans/' + id); },
      adminCreate: function (data) { return apiPost('/admin/plans', data); },
      adminUpdate: function (id, data) { return apiPut('/admin/plans/' + id, data); },
      adminDelete: function (id) { return apiDelete('/admin/plans/' + id); },
      adminToggle: function (id) { return apiPost('/admin/plans/' + id + '/toggle'); },
    },

    subscriptions: {
      plans: function () { return apiGet('/subscriptions/plans'); },
      current: function () { return apiGet('/subscriptions/current'); },
      subscribe: function (data) { return apiPost('/subscriptions/subscribe', data); },
      cancel: function () { return apiPost('/subscriptions/cancel'); },
      reactivate: function () { return apiPost('/subscriptions/reactivate'); },
      upgrade: function (data) { return apiPost('/subscriptions/upgrade', data); },
      downgrade: function (data) { return apiPost('/subscriptions/downgrade', data); },
      proration: function (data) { return apiPost('/subscriptions/proration', data); },
      retryPayment: function () { return apiPost('/subscriptions/retry-payment'); },
      invoices: function () { return apiGet('/subscriptions/invoices'); },
      paymentMethods: function () { return apiGet('/subscriptions/payment-methods'); },
      addPaymentMethod: function (data) { return apiPost('/subscriptions/payment-methods', data); },
      removePaymentMethod: function (id) { return apiDelete('/subscriptions/payment-methods/' + id); },
      setDefaultPaymentMethod: function (id) { return apiPut('/subscriptions/payment-methods/' + id + '/default'); },
    },

    profile: {
      get: function () { return apiGet('/profile'); },
      update: function (data) { return apiPut('/profile', data); },
      completion: function () { return apiGet('/employee/profile/completion'); },
      suggestedSkills: function () { return apiGet('/employee/skills/suggested'); },
      wizard: function () { return apiGet('/employee/profile/wizard'); },
      uploadResume: function (fd, cb) { return upload('/employee/resume/upload', fd, cb); },
    },

    github: {
      auth: function () { window.location.href = BASE_URL + '/github/auth'; },
    },

    preferences: {
      get: function () { return apiGet('/preferences'); },
      update: function (data) { return apiPut('/preferences', data); },
    },
  };
})();
