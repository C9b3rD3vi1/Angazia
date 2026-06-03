(function() {
  const tbody = document.getElementById('invoices-tbody');
  if (!tbody) return;

  async function loadInvoices() {
    try {
      const res = await fetch('/api/v1/billing/invoices');
      if (!res.ok) {
        if (res.status === 401) { window.location.href = '/login?redirect=' + encodeURIComponent(window.location.pathname); return; }
        tbody.innerHTML = '<tr><td colspan="6" class="emp-table-empty">Failed to load invoices</td></tr>';
        return;
      }
      const data = await res.json();
      if (!data.data || data.data.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="emp-table-empty">No invoices yet</td></tr>';
        return;
      }
      tbody.innerHTML = data.data.map(inv => {
        const date = new Date(inv.created_at || inv.date).toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
        const statusClass = inv.status === 'paid' ? 'emp-badge-active' : inv.status === 'pending' ? 'emp-badge-warning' : 'emp-badge-inactive';
        return '<tr>' +
          '<td><strong>' + (inv.invoice_number || inv.id) + '</strong></td>' +
          '<td>' + date + '</td>' +
          '<td>' + (inv.plan_name || inv.plan || '-') + '</td>' +
          '<td>' + (inv.amount ? 'KSh ' + Number(inv.amount).toLocaleString() : '-') + '</td>' +
          '<td><span class="emp-badge ' + statusClass + '">' + (inv.status || 'unknown') + '</span></td>' +
          '<td>' + (inv.invoice_url ? '<a href="' + inv.invoice_url + '" class="emp-btn emp-btn-outline emp-btn-sm" target="_blank">Download</a>' : '') + '</td>' +
          '</tr>';
      }).join('');
    } catch (err) {
      tbody.innerHTML = '<tr><td colspan="6" class="emp-table-empty">Error loading invoices</td></tr>';
    }
  }

  loadInvoices();
})();
