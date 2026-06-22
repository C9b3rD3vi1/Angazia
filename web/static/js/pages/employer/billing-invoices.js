(function () {
  'use strict';

  var tbody = document.getElementById('invoices-tbody');
  if (!tbody) return;

  function formatDate(dateStr) {
    if (!dateStr) return '-';
    var d = new Date(dateStr);
    return d.toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function statusBadge(status) {
    var cls = 'emp-badge-inactive';
    var label = status || 'unknown';
    if (status === 'paid') { cls = 'emp-badge-active'; label = 'Paid'; }
    else if (status === 'pending') { cls = 'emp-badge-warning'; label = 'Pending'; }
    else if (status === 'overdue') { cls = 'emp-badge-warning'; label = 'Overdue'; }
    return '<span class="emp-badge ' + cls + '">' + label + '</span>';
  }

  function loadInvoices() {
    AngaziaAPI.subscriptions.invoices()
      .then(function (data) {
        var invoices = data && data.invoices;
        if (!invoices || !invoices.length) {
          tbody.innerHTML = '<tr><td colspan="6" class="emp-table-empty">No invoices yet</td></tr>';
          return;
        }
        tbody.innerHTML = invoices.map(function (inv) {
          var date = formatDate(inv.created_at || inv.due_date);
          var badge = statusBadge(inv.status);
          return '<tr>' +
            '<td><strong>' + (inv.invoice_number || (inv && inv.id && inv.id.slice(0, 8))) + '</strong></td>' +
            '<td>' + date + '</td>' +
            '<td>' + (inv.plan_name || '-') + '</td>' +
            '<td>' + (inv.total || inv.amount ? 'KSh ' + Number(inv.total || inv.amount).toLocaleString() : '-') + '</td>' +
            '<td>' + badge + '</td>' +
            '<td>' + (inv.pdf_url ? '<a href="' + inv.pdf_url + '" class="emp-btn emp-btn-outline emp-btn-sm" target="_blank">Download</a>' : '') + '</td>' +
            '</tr>';
        }).join('');
      })
      .catch(function () {
        tbody.innerHTML = '<tr><td colspan="6" class="emp-table-empty">Failed to load invoices</td></tr>';
      });
  }

  loadInvoices();
})();
