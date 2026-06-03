(function() {
  const btn = document.getElementById('btn-confirm-upgrade');
  const errDiv = document.getElementById('upgrade-error');
  const successDiv = document.getElementById('upgrade-success');
  if (!btn) return;

  btn.addEventListener('click', async function() {
    const plan = this.dataset.plan;
    errDiv.style.display = 'none';
    successDiv.style.display = 'none';
    this.disabled = true;
    this.textContent = 'Processing...';

    try {
      const res = await fetch('/api/v1/billing/upgrade', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ plan: plan })
      });
      const data = await res.json();
      if (!res.ok) {
        errDiv.textContent = data.error || data.message || 'Upgrade failed';
        errDiv.style.display = 'block';
        this.disabled = false;
        this.textContent = 'Confirm Upgrade';
        return;
      }
      successDiv.textContent = 'Successfully upgraded to ' + plan.charAt(0).toUpperCase() + plan.slice(1) + '!';
      successDiv.style.display = 'block';
      this.textContent = 'Upgraded!';
      setTimeout(function() { window.location.href = '/employer/billing'; }, 2000);
    } catch (err) {
      errDiv.textContent = 'Network error';
      errDiv.style.display = 'block';
      this.disabled = false;
      this.textContent = 'Confirm Upgrade';
    }
  });
})();
