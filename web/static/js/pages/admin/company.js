(function() {
    'use strict';

    var state = {
        companies: [],
        page: 1,
        perPage: 20,
        total: 0,
        totalPages: 0,
        pendingCompanyId: null,
        loading: false,
        searchQuery: '',
        verificationStatus: '',
        hasDocuments: '',
        activeTab: 'all',
    };

    function showToast(msg, type) {
        type = type || 'info';
        if (window.AngaziaApp && AngaziaApp.showToast) {
            AngaziaApp.showToast(msg, type);
        } else {
            console.log('[' + type.toUpperCase() + '] ' + msg);
            if (type === 'error') alert('Error: ' + msg);
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

    function fetchStats() {
        AngaziaAPI.admin.userStats()
            .then(function(res) {
                var data = res && res.data ? res.data : res;
                var el = function(id) { return document.getElementById(id); };
                if (data) {
                    var total = (data.role_employer || 0);
                    if (el('ac-stat-total')) el('ac-stat-total').textContent = total;
                    if (el('ac-stat-verified')) el('ac-stat-verified').textContent = data.verification_verified || 0;
                    if (el('ac-stat-pending')) el('ac-stat-pending').textContent = data.verification_pending || 0;
                    if (el('ac-stat-rejected')) el('ac-stat-rejected').textContent = data.verification_rejected || 0;
                }
            })
            .catch(function() {});
    }

    function fetchCompanies() {
        state.loading = true;
        showLoading(true);
        hideError();

        var params = {
            page: state.page,
            limit: state.perPage,
        };
        if (state.searchQuery) params.search = state.searchQuery;
        if (state.activeTab !== 'all') {
            params.verification_status = state.activeTab;
        }
        if (state.hasDocuments) {
            params.has_documents = state.hasDocuments;
        }

        AngaziaAPI.admin.companies(params)
            .then(function(res) {
                var data = res && res.data ? res.data : res;
                state.companies = (data && data.companies) || [];
                state.total = data ? data.total : 0;
                state.totalPages = data ? data.total_pages : 0;

                showLoading(false);
                render();
                if (data && data.stats) renderStats(data.stats);
            })
            .catch(function(err) {
                state.loading = false;
                showLoading(false);
                showError(err.message || 'Failed to load companies');
            });
    }

    function render() {
        var tbody = document.getElementById('ac-table-body');
        var countEl = document.getElementById('ac-total-count');
        var emptyEl = document.getElementById('ac-empty');
        var table = document.getElementById('ac-table');

        if (!state.companies.length) {
            if (emptyEl) emptyEl.style.display = 'flex';
            if (table) table.style.display = 'none';
            if (countEl) countEl.textContent = '0';
            hidePagination();
            return;
        }

        if (emptyEl) emptyEl.style.display = 'none';
        if (table) table.style.display = '';
        if (countEl) countEl.textContent = state.total;

        var html = '';
        for (var i = 0; i < state.companies.length; i++) {
            var company = state.companies[i];
            var vs = company.verification_status || 'unverified';
            var logoHtml = company.logo
                ? '<img src="' + escapeHtml(company.logo) + '" alt="" class="ac-table-logo">'
                : '<span class="ac-table-logo-placeholder">' + getInitials(company.name) + '</span>';
            var actionsHtml = '';
            if (vs !== 'verified') {
                actionsHtml += '<button class="ac-btn-sm ac-btn-sm-accept" data-action="approve" data-company-id="' + company.id + '">Approve</button>';
                actionsHtml += '<button class="ac-btn-sm ac-btn-sm-reject" data-action="reject" data-company-id="' + company.id + '">Reject</button>';
            }
            actionsHtml += '<a href="/admin/companies/' + company.id + '" class="ac-btn-sm ac-btn-sm-ghost">View</a>';

            html += '<tr data-company-id="' + company.id + '">'
                + '<td><a href="/admin/companies/' + company.id + '" class="ac-table-link">'
                + logoHtml
                + '<span class="ac-table-name">' + escapeHtml(company.name) + '</span></a></td>'
                + '<td><span class="ac-table-muted">' + escapeHtml(company.email) + '</span></td>'
                + '<td><span class="ac-status-badge ' + getVerificationClass(vs) + '">'
                + getVerificationText(vs) + '</span></td>'
                + '<td><span class="ac-table-number">' + (company.jobs_count || 0) + '</span></td>'
                + '<td><span class="ac-table-number">' + (company.document_count || 0) + '</span></td>'
                + '<td><span class="ac-table-muted">' + formatDate(company.created_at) + '</span></td>'
                + '<td><div class="ac-row-actions">' + actionsHtml + '</div></td>'
                + '</tr>';
        }

        tbody.innerHTML = html;
        updatePagination();
    }

    function showLoading(show) {
        var loadingEl = document.getElementById('ac-loading');
        if (loadingEl) loadingEl.style.display = show ? 'flex' : 'none';
    }

    function showError(msg) {
        var errEl = document.getElementById('ac-error');
        var errText = document.getElementById('ac-error-text');
        var table = document.getElementById('ac-table');
        var emptyEl = document.getElementById('ac-empty');
        if (errText) errText.textContent = msg || 'Failed to load companies.';
        if (errEl) errEl.style.display = 'flex';
        if (table) table.style.display = 'none';
        if (emptyEl) emptyEl.style.display = 'none';
        hidePagination();
    }

    function hideError() {
        var errEl = document.getElementById('ac-error');
        if (errEl) errEl.style.display = 'none';
    }

    function hidePagination() {
        var pagiEl = document.getElementById('ac-pagination');
        if (pagiEl) pagiEl.style.display = 'none';
    }

    function updatePagination() {
        var pagiEl = document.getElementById('ac-pagination');
        var infoEl = document.getElementById('ac-pagi-info');
        var btnsEl = document.getElementById('ac-pagi-btns');
        if (!pagiEl || !infoEl || !btnsEl) return;

        if (state.totalPages <= 1) {
            pagiEl.style.display = 'none';
            return;
        }
        pagiEl.style.display = 'flex';

        var start = (state.page - 1) * state.perPage + 1;
        var end = Math.min(state.page * state.perPage, state.total);
        infoEl.textContent = start + '-' + end + ' of ' + state.total;

        btnsEl.innerHTML = '';

        var prevBtn = document.createElement('button');
        prevBtn.className = 'ac-pagi-btn';
        prevBtn.textContent = 'Prev';
        prevBtn.disabled = state.page <= 1;
        prevBtn.addEventListener('click', function() {
            if (state.page > 1) { state.page--; fetchCompanies(); }
        });
        btnsEl.appendChild(prevBtn);

        var maxVisible = 5;
        var ps = Math.max(1, state.page - Math.floor(maxVisible / 2));
        var pe = Math.min(state.totalPages, ps + maxVisible - 1);
        if (pe - ps < maxVisible - 1) ps = Math.max(1, pe - maxVisible + 1);

        if (ps > 1) {
            var firstBtn = document.createElement('button');
            firstBtn.className = 'ac-pagi-btn';
            firstBtn.textContent = '1';
            firstBtn.addEventListener('click', function() { state.page = 1; fetchCompanies(); });
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
                btn.addEventListener('click', function() { state.page = pageNum; fetchCompanies(); });
                btnsEl.appendChild(btn);
            })(i);
        }

        if (pe < state.totalPages) {
            if (pe < state.totalPages - 1) {
                var dots2 = document.createElement('span');
                dots2.className = 'ac-pagi-btn';
                dots2.style.border = 'none';
                dots2.style.cursor = 'default';
                dots2.textContent = '...';
                btnsEl.appendChild(dots2);
            }
            var lastBtn = document.createElement('button');
            lastBtn.className = 'ac-pagi-btn';
            lastBtn.textContent = state.totalPages;
            lastBtn.addEventListener('click', function() { state.page = state.totalPages; fetchCompanies(); });
            btnsEl.appendChild(lastBtn);
        }

        var nextBtn = document.createElement('button');
        nextBtn.className = 'ac-pagi-btn';
        nextBtn.textContent = 'Next';
        nextBtn.disabled = state.page >= state.totalPages;
        nextBtn.addEventListener('click', function() {
            if (state.page < state.totalPages) { state.page++; fetchCompanies(); }
        });
        btnsEl.appendChild(nextBtn);
    }

    function renderStats(stats) {
        var el = function(id) { return document.getElementById(id); };
        if (!stats) return;
        if (el('ac-stat-total')) el('ac-stat-total').textContent = stats.verified + stats.pending + stats.rejected + stats.unverified;
        if (el('ac-stat-verified')) el('ac-stat-verified').textContent = stats.verified || 0;
        if (el('ac-stat-pending')) el('ac-stat-pending').textContent = stats.pending || 0;
        if (el('ac-stat-rejected')) el('ac-stat-rejected').textContent = stats.rejected || 0;
    }

    function openApproveModal(companyId) {
        state.pendingCompanyId = companyId;
        var modal = document.getElementById('ac-approve-modal');
        if (modal) modal.style.display = 'flex';
    }

    async function submitApprove() {
        if (!state.pendingCompanyId) return;
        var confirmBtn = document.getElementById('ac-approve-confirm');
        try {
            if (confirmBtn) { confirmBtn.disabled = true; confirmBtn.textContent = 'Processing...'; }
            await AngaziaAPI.admin.verifyCompany(state.pendingCompanyId);
            showToast('Company verified successfully!', 'success');
            closeApproveModal();
            fetchCompanies();
        } catch (error) {
            showToast(error.message || 'Failed to approve company', 'error');
        } finally {
            if (confirmBtn) { confirmBtn.disabled = false; confirmBtn.textContent = 'Approve'; }
            state.pendingCompanyId = null;
        }
    }

    function closeApproveModal() {
        var modal = document.getElementById('ac-approve-modal');
        if (modal) modal.style.display = 'none';
        state.pendingCompanyId = null;
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
            await AngaziaAPI.admin.rejectCompany(state.pendingCompanyId, { reason: reason });
            showToast('Company verification rejected', 'success');

            var tableRow = document.querySelector('#ac-table-body tr[data-company-id="' + state.pendingCompanyId + '"]');
            if (tableRow) {
                var statusBadge = tableRow.querySelector('.ac-status-badge');
                if (statusBadge) {
                    statusBadge.className = 'ac-status-badge rejected';
                    statusBadge.textContent = 'Rejected';
                }
            }

            var pendingCount = document.getElementById('ac-stat-pending');
            if (pendingCount) {
                var current = parseInt(pendingCount.textContent, 10);
                pendingCount.textContent = Math.max(0, current - 1);
            }

            closeModal();
            fetchCompanies();
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

    function switchTab(tab) {
        state.activeTab = tab;
        state.page = 1;
        var tabs = document.querySelectorAll('.ac-tab');
        tabs.forEach(function(t) {
            t.classList.toggle('active', t.getAttribute('data-tab') === tab);
        });
        fetchCompanies();
    }

    function applyFilters() {
        state.page = 1;
        state.searchQuery = document.getElementById('ac-search').value.trim();
        state.verificationStatus = document.getElementById('ac-status-filter').value;
        state.hasDocuments = document.getElementById('ac-docs-filter').value;
        fetchCompanies();
    }

    function init() {
        fetchCompanies();

        var reloadBtn = document.querySelector('[data-action="reload"]');
        if (reloadBtn) {
            reloadBtn.addEventListener('click', function(e) {
                e.preventDefault();
                fetchCompanies();
            });
        }

        var retryBtn = document.querySelector('[data-action="retry"]');
        if (retryBtn) {
            retryBtn.addEventListener('click', function(e) {
                e.preventDefault();
                fetchCompanies();
            });
        }

        document.addEventListener('click', function(e) {
            var btn = e.target.closest('[data-action]');
            if (!btn) return;
            var action = btn.getAttribute('data-action');
            var companyId = btn.getAttribute('data-company-id');
            if (action === 'approve' && companyId) {
                e.preventDefault();
                openApproveModal(companyId);
            } else if (action === 'reject' && companyId) {
                e.preventDefault();
                openRejectModal(companyId);
            }
        });

        var tabs = document.querySelectorAll('.ac-tab');
        tabs.forEach(function(tab) {
            tab.addEventListener('click', function() {
                switchTab(this.getAttribute('data-tab'));
            });
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

        var approveConfirmBtn = document.getElementById('ac-approve-confirm');
        if (approveConfirmBtn) approveConfirmBtn.addEventListener('click', submitApprove);

        var approveCloseBtn = document.getElementById('ac-approve-close');
        if (approveCloseBtn) approveCloseBtn.addEventListener('click', closeApproveModal);

        var approveCancelBtn = document.getElementById('ac-approve-cancel');
        if (approveCancelBtn) approveCancelBtn.addEventListener('click', closeApproveModal);

        var approveModal = document.getElementById('ac-approve-modal');
        if (approveModal) {
            approveModal.addEventListener('click', function(e) {
                if (e.target === this) closeApproveModal();
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

        var docsFilter = document.getElementById('ac-docs-filter');
        if (docsFilter) docsFilter.addEventListener('change', applyFilters);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
