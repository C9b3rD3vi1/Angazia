(function () {
  'use strict';

  function showToast(msg, type) {
    if (window.AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
    } else {
      alert((type === 'error' ? 'Error: ' : '') + msg);
    }
  }

  var f = document.getElementById('al-form'), e = document.getElementById('al-error'), b = document.getElementById('al-submit');
  if (!f) return;
  f.addEventListener('submit', function (ev) {
    ev.preventDefault();
    if (e) e.style.display = 'none';
    b.disabled = true;
    b.textContent = 'Signing in\u2026';
    var email = document.getElementById('al-email').value.trim();
    var password = document.getElementById('al-password').value;
    if (!email || !password) {
      if (e) { e.textContent = 'Please enter email and password'; e.style.display = ''; }
      b.disabled = false; b.textContent = 'Sign In'; return;
    }
    AngaziaAPI.auth.adminLogin({ email: email, password: password }).then(function (data) {
      AngaziaAPI.setTokens(data.access_token, data.refresh_token);
      window.location.href = '/admin/dashboard';
    }).catch(function (err) {
      b.disabled = false; b.textContent = 'Sign In';
      var msg = (err.body && err.body.message) || err.message || 'Invalid credentials';
      if (e) { e.textContent = msg; e.style.display = ''; } else { showToast(msg, 'error'); }
    });
  });
})();
