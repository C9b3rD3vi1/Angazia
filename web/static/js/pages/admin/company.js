(function() {
    'use strict';

    var state = {
        allCompanies: [],
        page: 1,
        perPage: 20,
        pendingCompanyId: null,
    };

    function showToast(msg, type) {
        type = type || 'info';
        if (window.AngaziaApp && AngaziaApp.showToast) {
            AngaziaApp.showToast(msg, type);
        } else {
            console.log('[' + type.toUpperCase() + '] ' + msg);
            if (type === 'error') {
                alert('Error: ' + msg);
            }
        }
    }

    function getInitials(name) {
        if (!name || name === '') return '?';
        return name.split(' ').map(function(n) { return n[0]; }).join('').toUpperCase().slice(0, 2);
    }

    function formatDate(dateStr) {
        if (!dateStr) return 'N/A';
        try {
            var date = new Date(dateStr);
            return date.toLocaleDateString('en-KE', { year: 'numeric', month: 'short', day: 'numeric' });
        } catch (e) {
            return dateStr;
        }
    }

    function escapeHtml(text) {
        if (!text) return '';
        var div = document.createElement('div');
        div.appendChild(document.createTextNode(text));
        return div.innerHTML;
    }

    function getVerificationClass(status) {
        switch (status) {
            case 'verified': return 'verified';
            case 'pending': return 'pending';
            case 'rejected': return 'rejected';
            default: return 'unverified';
        }
    }

    function getVerificationText(status) {
        switch (status) {
            case 'verified': return 'Verified';
            case 'pending': return 'Pending';
            case 'rejected': return 'Rejected';
            default: return 'Unverified';
        }
    }

    async function fetchCompanies() {
        var response = await AngaziaAPI.admin.companies();
        var companies = [];
        if (response && response.data && response.data.companies) {
            companies = response.data.companies;
        } else if (response && response.companies) {
            companies = response.companies;
        } else if (response && response.data && Array.isArray(response.data)) {
            companies = response.data;
        } else if (Array.isArray(response)) {
            companies = response;
        }
        companies = companies.map(function(company) {
            return {
                id: company.id,
                name: company.name && company.name.trim() !== '' ? company.name : (company.company_name || 'Unnamed Company'),
                email: company.email || '',
                logo: company.logo || null,
                verification_status: company.verification_status || 'unverified',
                jobs_count: company.jobs_count || 0,
                created_at: company.created_at || '',
            };
        });
        return companies;
    }

    async function approveCompany(companyId) {
        return await AngaziaAPI.admin.verifyCompany(companyId);
    }

    async function rejectCompany(companyId, reason) {
        return await AngaziaAPI.admin.rejectCompany(companyId, { reason: reason });
    }

    function renderPendingVerifications(companies) {
        var container = document.getElementById('ac-pending-list');
        var countEl = document.getElementById('ac-pending-count');
        if (!container) return;

        var pendingCompanies = companies.filter(function(c) {
            return c.verification_status === 'unverified' || c.verification_status === 'pending';
        });

        if (countEl) countEl.textContent = pendingCompanies.length;

        if (!pendingCompanies || pendingCompanies.length === 0) {
            container.innerHTML = '<div class="ac-empty"><p class="ac-empty-text">No pending verification requests.</p></div>';
            return;
        }

        container.innerHTML = pendingCompanies.map(function(company) {
            var logoHtml = company.logo
                ? '<img src="' + escapeHtml(company.logo) + '" alt="" class="ac-pending-logo">'
                : '<span class="ac-pending-logo-placeholder">' + getInitials(company.name) + '</span>';
            return '<div class="ac-pending-item" data-company-id="' + company.id + '">'
                + '<div class="ac-pending-header">'
                + logoHtml
                + '<div class="ac-pending-info">'
                + '<a href="/admin/companies/' + company.id + '" class="ac-pending-name">' + escapeHtml(company.name) + '</a>'
                + '<span class="ac-pending-email">' + escapeHtml(company.email) + '</span>'
                + '</div></div>'
                + '<div class="ac-pending-meta">'
                + '<span>Joined ' + formatDate(company.created_at) + '</span>'
                + '<span>' + (company.jobs_count || 0) + ' jobs</span>'
                + '<span class="ac-status-badge ' + getVerificationClass(company.verification_status) + '">'
                + getVerificationText(company.verification_status)
                + '</span></div>'
                + '<div class="ac-pending-actions">'
                + '<button class="ac-btn-accept" data-action="approve" data-company-id="' + company.id + '">Approve</button>'
                + '<button class="ac-btn-reject" data-action="reject" data-company-id="' + company.id + '">Reject</button>'
                + '</div></div>';
        }).join('');
    }

    function renderCompaniesTable(companies) {
        var tbody = document.getElementById('ac-table-body');
        var countEl = document.getElementById('ac-total-count');
        if (!tbody) return;

        var paginated = companies.slice(0, state.perPage);
        if (countEl) countEl.textContent = companies.length;

        if (!companies || companies.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6"><div class="ac-empty"><p class="ac-empty-text">No companies found.</p></div></td></tr>';
            renderPagination(companies.length);
            return;
        }

        tbody.innerHTML = paginated.map(function(company) {
            var logoHtml = company.logo
                ? '<img src="' + escapeHtml(company.logo) + '" alt="" class="ac-table-logo">'
                : '<span class="ac-table-logo-placeholder">' + getInitials(company.name) + '</span>';
            var actionsHtml = '';
            if (company.verification_status !== 'verified') {
                actionsHtml += '<button class="ac-btn-sm ac-btn-sm-accept" data-action="approve" data-company-id="' + company.id + '">Approve</button>';
                actionsHtml += '<button class="ac-btn-sm ac-btn-sm-reject" data-action="reject" data-company-id="' + company.id + '">Reject</button>';
            }
            actionsHtml += '<a href="/admin/companies/' + company.id + '" class="ac-btn-sm ac-btn-sm-ghost">View</a>';
            return '<tr data-company-id="' + company.id + '">'
                + '<td><a href="/admin/companies/' + company.id + '" class="ac-table-link">'
                + logoHtml
                + '<span class="ac-table-name">' + escapeHtml(company.name) + '</span></a></td>'
                + '<td><span class="ac-table-muted">' + escapeHtml(company.email) + '</span></td>'
                + '<td><span class="ac-status-badge ' + getVerificationClass(company.verification_status) + '">'
                + getVerificationText(company.verification_status) + '</span></td>'
                + '<td><span class="ac-table-number">' + (company.jobs_count || 0) + '</span></td>'
                + '<td><span class="ac-table-muted">' + formatDate(company.created_at) + '</span></td>'
                + '<td><div class="ac-row-actions">' + actionsHtml + '</div></td>'
                + '</tr>';
        }).join('');

        renderPagination(companies.length);
    }

    function renderPagination(total) {
        var pagiEl = document.getElementById('ac-pagination');
        var infoEl = document.getElementById('ac-pagi-info');
        var btnsEl = document.getElementById('ac-pagi-btns');
        if (!pagiEl || !infoEl || !btnsEl) return;

        var totalPages = Math.ceil(total / state.perPage);
        if (totalPages <= 1) {
            pagiEl.style.display = 'none';
            return;
        }
        pagiEl.style.display = 'flex';
        var start = (state.page - 1) * state.perPage + 1;
        var end = Math.min(state.page * state.perPage, total);
        infoEl.textContent = start + '-' + end + ' of ' + total;

        btnsEl.innerHTML = '';

        var prevBtn = document.createElement('button');
        prevBtn.className = 'ac-pagi-btn';
        prevBtn.textContent = 'Prev';
        prevBtn.disabled = state.page <= 1;
        prevBtn.addEventListener('click', function() {
            if (state.page > 1) { state.page--; renderCompaniesTable(state.allCompanies); }
        });
        btnsEl.appendChild(prevBtn);

        var maxVisible = 5;
        var ps = Math.max(1, state.page - Math.floor(maxVisible / 2));
        var pe = Math.min(totalPages, ps + maxVisible - 1);
        if (pe - ps < maxVisible - 1) ps = Math.max(1, pe - maxVisible + 1);

        if (ps > 1) {
            var firstBtn = document.createElement('button');
            firstBtn.className = 'ac-pagi-btn';
            firstBtn.textContent = '1';
            firstBtn.addEventListener('click', function() { state.page = 1; renderCompaniesTable(state.allCompanies); });
            btnsEl.appendChild(firstBtn);
            if (ps > 2) {
                var dots = document.createElement('span');
                dots.className = 'ac-pagi-btn';
                dots.style.border = 'none';
                dots.style.cursor = 'default';
                dots.textContent = '...';
                btnsEl.appendChild(dots);
            }
        }

        for (var i = ps; i <= pe; i++) {
            (function(pageNum) {
                var btn = document.createElement('button');
                btn.className = 'ac-pagi-btn' + (pageNum === state.page ? ' active' : '');
                btn.textContent = pageNum;
                btn.addEventListener('click', function() { state.page = pageNum; renderCompaniesTable(state.allCompanies); });
                btnsEl.appendChild(btn);
            })(i);
        }

        if (pe < totalPages) {
            if (pe < totalPages - 1) {
                var dots2 = document.createElement('span');
                dots2.className = 'ac-pagi-btn';
                dots2.style.border = 'none';
                dots2.style.cursor = 'default';
                dots2.textContent = '...';
                btnsEl.appendChild(dots2);
            }
            var lastBtn = document.createElement('button');
            lastBtn.className = 'ac-pagi-btn';
            lastBtn.textContent = totalPages;
            lastBtn.addEventListener('click', function() { state.page = totalPages; renderCompaniesTable(state.allCompanies); });
            btnsEl.appendChild(lastBtn);
        }

        var nextBtn = document.createElement('button');
        nextBtn.className = 'ac-pagi-btn';
        nextBtn.textContent = 'Next';
        nextBtn.disabled = state.page >= totalPages;
        nextBtn.addEventListener('click', function() {
            if (state.page < totalPages) { state.page++; renderCompaniesTable(state.allCompanies); }
        });
        btnsEl.appendChild(nextBtn);
    }

    function showLoading(show) {
        var loadingEl = document.getElementById('ac-loading');
        if (loadingEl) loadingEl.style.display = show ? 'flex' : 'none';
    }

    function showError(msg) {
        var errEl = document.getElementById('ac-error');
        var errText = document.getElementById('ac-error-text');
        if (errText) errText.textContent = msg || 'Failed to load companies.';
        if (errEl) errEl.style.display = 'flex';
    }

    function hideError() {
        var errEl = document.getElementById('ac-error');
        if (errEl) errEl.style.display = 'none';
    }

    function removeCompanyFromUI(companyId) {
        var pendingItem = document.querySelector('.ac-pending-item[data-company-id="' + companyId + '"]');
        if (pendingItem && pendingItem.remove) pendingItem.remove();

        var tableRow = document.querySelector('#ac-table-body tr[data-company-id="' + companyId + '"]');
        if (tableRow) {
            var statusBadge = tableRow.querySelector('.ac-status-badge');
            if (statusBadge) {
                statusBadge.className = 'ac-status-badge verified';
                statusBadge.textContent = 'Verified';
            }
            var actionsDiv = tableRow.querySelector('.ac-row-actions');
            if (actionsDiv) {
                actionsDiv.innerHTML = '<a href="/admin/companies/' + companyId + '" class="ac-btn-sm ac-btn-sm-ghost">View</a>';
            }
        }

        var pendingCount = document.getElementById('ac-pending-count');
        if (pendingCount) {
            var remaining = document.querySelectorAll('.ac-pending-item').length;
            pendingCount.textContent = remaining;
        }

        var pendingList = document.getElementById('ac-pending-list');
        if (pendingList && !pendingList.querySelector('.ac-pending-item')) {
            pendingList.innerHTML = '<div class="ac-empty"><p class="ac-empty-text">No pending verification requests.</p></div>';
        }
    }

    async function handleApprove(companyId, btn) {
        try {
            if (btn) { btn.disabled = true; btn.textContent = 'Processing...'; }
            await approveCompany(companyId);
            showToast('Company verified successfully!', 'success');
            removeCompanyFromUI(companyId);
        } catch (error) {
            showToast(error.message || 'Failed to approve company', 'error');
        } finally {
            if (btn) { btn.disabled = false; btn.textContent = 'Approve'; }
        }
    }

    function openRejectModal(companyId) {
        state.pendingCompanyId = companyId;
        var modal = document.getElementById('ac-reject-modal');
        var textarea = document.getElementById('ac-reject-reason');
        if (textarea) textarea.value = '';
        if (modal) modal.style.display = 'flex';
    }

    async function submitReject() {
        if (!state.pendingCompanyId) return;
        var reason = document.getElementById('ac-reject-reason').value.trim();
        var confirmBtn = document.getElementById('ac-modal-confirm');

        if (!reason) {
            showToast('Please provide a reason for rejection', 'warning');
            return;
        }

        try {
            if (confirmBtn) { confirmBtn.disabled = true; confirmBtn.textContent = 'Processing...'; }
            await rejectCompany(state.pendingCompanyId, reason);
            showToast('Company verification rejected', 'success');

            var tableRow = document.querySelector('#ac-table-body tr[data-company-id="' + state.pendingCompanyId + '"]');
            if (tableRow) {
                var statusBadge = tableRow.querySelector('.ac-status-badge');
                if (statusBadge) {
                    statusBadge.className = 'ac-status-badge rejected';
                    statusBadge.textContent = 'Rejected';
                }
            }

            var pendingItem = document.querySelector('.ac-pending-item[data-company-id="' + state.pendingCompanyId + '"]');
            if (pendingItem && pendingItem.remove) pendingItem.remove();

            var pendingCount = document.getElementById('ac-pending-count');
            if (pendingCount) {
                var remaining = document.querySelectorAll('.ac-pending-item').length;
                pendingCount.textContent = remaining;
            }

            closeModal();
        } catch (error) {
            showToast(error.message || 'Failed to reject company', 'error');
        } finally {
            if (confirmBtn) { confirmBtn.disabled = false; confirmBtn.textContent = 'Reject'; }
            state.pendingCompanyId = null;
        }
    }

    function closeModal() {
        var modal = document.getElementById('ac-reject-modal');
        if (modal) modal.style.display = 'none';
        state.pendingCompanyId = null;
    }

    function applyFilters() {
        var query = document.getElementById('ac-search').value.trim().toLowerCase();
        var status = document.getElementById('ac-status-filter').value;
        var rows = document.querySelectorAll('#ac-table-body tr[data-company-id]');
        rows.forEach(function(row) {
            var text = row.textContent.toLowerCase();
            var statusCell = row.querySelector('.ac-status-badge');
            var rowStatus = statusCell ? statusCell.textContent.trim().toLowerCase() : '';
            var matchSearch = !query || text.indexOf(query) !== -1;
            var matchStatus = !status || rowStatus === status;
            row.style.display = matchSearch && matchStatus ? '' : 'none';
        });
    }

    async function loadCompanies() {
        showLoading(true);
        hideError();
        state.page = 1;

        try {
            state.allCompanies = await fetchCompanies();
            if (!state.allCompanies || !Array.isArray(state.allCompanies)) {
                throw new Error('Invalid response format');
            }
            renderPendingVerifications(state.allCompanies);
            renderCompaniesTable(state.allCompanies);
        } catch (error) {
            showError(error.message || 'Failed to load companies');
            var pendingList = document.getElementById('ac-pending-list');
            if (pendingList) {
                pendingList.innerHTML = '<div class="ac-empty"><p class="ac-empty-text">Failed to load companies. Please refresh the page.</p>'
                    + '<button class="ac-btn ac-btn-primary" onclick="location.reload()">Retry</button></div>';
            }
            var tableBody = document.getElementById('ac-table-body');
            if (tableBody) {
                tableBody.innerHTML = '<tr><td colspan="6"><div class="ac-empty"><p class="ac-empty-text">Failed to load companies. Please refresh the page.</p></div></td></tr>';
            }
        } finally {
            showLoading(false);
        }
    }

    function init() {
        loadCompanies();

        var reloadBtn = document.querySelector('[data-action="reload"]');
        if (reloadBtn) {
            reloadBtn.addEventListener('click', function(e) {
                e.preventDefault();
                loadCompanies();
            });
        }

        var retryBtn = document.querySelector('[data-action="retry"]');
        if (retryBtn) {
            retryBtn.addEventListener('click', function(e) {
                e.preventDefault();
                loadCompanies();
            });
        }

        document.addEventListener('click', function(e) {
            var btn = e.target.closest('[data-action]');
            if (!btn) return;
            var action = btn.getAttribute('data-action');
            var companyId = btn.getAttribute('data-company-id');
            if (action === 'approve' && companyId) {
                e.preventDefault();
                handleApprove(companyId, btn);
            } else if (action === 'reject' && companyId) {
                e.preventDefault();
                openRejectModal(companyId);
            }
        });

        var confirmBtn = document.getElementById('ac-modal-confirm');
        if (confirmBtn) confirmBtn.addEventListener('click', submitReject);

        var closeBtn = document.getElementById('ac-modal-close');
        if (closeBtn) closeBtn.addEventListener('click', closeModal);

        var cancelBtn = document.getElementById('ac-modal-cancel');
        if (cancelBtn) cancelBtn.addEventListener('click', closeModal);

        var modal = document.getElementById('ac-reject-modal');
        if (modal) {
            modal.addEventListener('click', function(e) {
                if (e.target === this) closeModal();
            });
        }

        var filterBtn = document.getElementById('ac-apply-filter');
        if (filterBtn) filterBtn.addEventListener('click', applyFilters);

        var searchInput = document.getElementById('ac-search');
        if (searchInput) {
            searchInput.addEventListener('keypress', function(e) {
                if (e.key === 'Enter') applyFilters();
            });
        }

        var statusFilter = document.getElementById('ac-status-filter');
        if (statusFilter) statusFilter.addEventListener('change', applyFilters);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
