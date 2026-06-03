(function () {
  'use strict';

  function init() {
    initMatchActions();
    initInterviewQuestions();
  }

  window.__matchesInit = init;
  window.__initMatchActions = initMatchActions;
  window.__initInterviewQuestions = initInterviewQuestions;

  function initMatchActions() {
    document.querySelectorAll('.emp-match-action').forEach(function (btn) {
      btn.addEventListener('click', function (e) {
        e.preventDefault();
        var candidateId = this.dataset.candidateId;
        var jobId = this.dataset.jobId;
        var action = this.dataset.action;
        if (!candidateId || typeof AngaziaAPI === 'undefined') return;

        switch (action) {
          case 'view-profile':
            window.location.href = '/employer/candidates/' + candidateId;
            break;

          case 'add-to-pool':
            addToTalentPool(candidateId, jobId, btn);
            break;

          case 'schedule-interview':
            var date = prompt('Interview date (YYYY-MM-DD):');
            if (!date) return;
            AngaziaAPI.applications.interview(candidateId, { scheduled_at: date, job_id: jobId })
              .then(function () {
                showToast('Interview scheduled', 'success');
              })
              .catch(function (err) {
                showToast(err.message || 'Failed to schedule', 'error');
              });
            break;

          case 'view-analysis':
            loadAnalysis(candidateId, jobId);
            break;
        }
      });
    });
  }

  function addToTalentPool(candidateId, jobId, btn) {
    if (typeof AngaziaAPI === 'undefined') return;
    var poolId = prompt('Enter Talent Pool ID (leave blank for default):');
    var data = { candidate_id: candidateId, job_id: jobId };
    if (poolId) data.pool_id = poolId;

    var fn = poolId
      ? AngaziaAPI.talentPools.addCandidate(poolId, data)
      : (AngaziaAPI.talentPools.list({ limit: 1 }).then(function (pools) {
          var poolsList = pools.pools || pools.data || pools || [];
          if (poolsList.length) return AngaziaAPI.talentPools.addCandidate(poolsList[0].id, data);
          return AngaziaAPI.talentPools.create({ name: 'AI Matches' }).then(function (p) {
            return AngaziaAPI.talentPools.addCandidate(p.id || p.data.id, data);
          });
        }));

    fn.then(function () {
      showToast('Candidate added to talent pool', 'success');
      if (btn) { btn.textContent = 'Added'; btn.disabled = true; }
    })
    .catch(function (err) {
      showToast(err.message || 'Failed to add to pool', 'error');
    });
  }

  function loadAnalysis(candidateId, jobId) {
    if (typeof AngaziaAPI === 'undefined' || typeof AngaziaModal === 'undefined') return;
    AngaziaAPI.matches.analysis(jobId, candidateId)
      .then(function (data) {
        var analysis = data.analysis || data || {};
        var html = '<div style="font-family:var(--fm,sans-serif);font-size:13px;line-height:1.6;">';
        if (analysis.skills_match) {
          html += '<h4 style="margin:0 0 8px;font-size:14px;">Skills Match</h4><p>' + (analysis.skills_match.percentage || 'N/A') + '% match</p>';
          if (analysis.skills_match.matching && analysis.skills_match.matching.length) {
            html += '<div style="display:flex;flex-wrap:wrap;gap:4px;margin-bottom:12px;">'
              + analysis.skills_match.matching.map(function (s) { return '<span style="background:rgba(0,229,160,0.1);color:#00e5a0;padding:2px 8px;border-radius:4px;font-size:11px;">' + escapeHtml(s) + '</span>'; }).join('')
              + '</div>';
          }
        }
        if (analysis.experience) {
          html += '<h4 style="margin:12px 0 8px;font-size:14px;">Experience Fit</h4><p>' + escapeHtml(analysis.experience.summary || '') + '</p>';
        }
        if (analysis.recommendation) {
          html += '<h4 style="margin:12px 0 8px;font-size:14px;">Recommendation</h4><p>' + escapeHtml(analysis.recommendation) + '</p>';
        }
        html += '</div>';
        AngaziaModal.open(html, { title: 'Match Analysis', maxWidth: '600px' });
      })
      .catch(function () {
        showToast('Failed to load analysis', 'error');
      });
  }

  function initInterviewQuestions() {
    document.querySelectorAll('.emp-gen-questions').forEach(function (btn) {
      btn.addEventListener('click', function (e) {
        e.preventDefault();
        var jobId = this.dataset.jobId;
        if (!jobId || typeof AngaziaAPI === 'undefined') return;

        if (typeof AngaziaModal !== 'undefined') {
          AngaziaModal.open('<div style="text-align:center;padding:24px;">Generating interview questions...</div>', {
            title: 'Interview Questions',
            maxWidth: '600px',
          });
          AngaziaModal.setLoading(true);
        }

        AngaziaAPI.matches.interviewQuestions(jobId)
          .then(function (data) {
            var questions = data.questions || data.data || data || [];
            if (typeof AngaziaModal !== 'undefined') {
              AngaziaModal.setLoading(false);
              var html = '<div style="font-family:var(--fm,sans-serif);font-size:13px;line-height:1.6;">';
              if (questions.length) {
                html += '<ol style="padding-left:20px;margin:0;">';
                questions.forEach(function (q) {
                  html += '<li style="padding:8px 0;border-bottom:1px solid var(--border,#e5e7eb);">' + escapeHtml(typeof q === 'string' ? q : q.question || q.text || '') + '</li>';
                });
                html += '</ol>';
              } else {
                html += '<p style="color:var(--muted,#6b7280);">No questions generated. Try again with more job details.</p>';
              }
              html += '</div>';
              AngaziaModal.open(html, { title: 'Interview Questions', maxWidth: '600px' });
            }
          })
          .catch(function () {
            if (typeof AngaziaModal !== 'undefined') {
              AngaziaModal.setLoading(false);
              AngaziaModal.open('<p style="color:var(--muted,#6b7280);">Failed to generate questions.</p>', { title: 'Error' });
            }
          });
      });
    });
  }

  function showToast(msg, type) {
    if (typeof AngaziaApp !== 'undefined' && AngaziaApp.showToast) {
      AngaziaApp.showToast(msg, type);
      return;
    }
    var c = document.getElementById('toast-container');
    if (!c) { c = document.createElement('div'); c.id = 'toast-container'; c.style.cssText = 'position:fixed;bottom:16px;right:16px;z-index:9999;display:flex;flex-direction:column;gap:8px;'; document.body.appendChild(c); }
    var t = document.createElement('div');
    var bg = type === 'success' ? '#00e5a0' : type === 'error' ? '#ef4444' : '#3b82f6';
    t.style.cssText = 'background:' + bg + ';color:#fff;padding:12px 20px;border-radius:10px;font-size:13px;font-family:var(--fm,sans-serif);box-shadow:0 4px 16px rgba(0,0,0,0.15);';
    t.textContent = msg;
    c.appendChild(t);
    setTimeout(function () { t.style.opacity = '0'; setTimeout(function () { t.remove(); }, 200); }, 3500);
  }

  function escapeHtml(t) {
    if (!t) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(t));
    return d.innerHTML;
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
