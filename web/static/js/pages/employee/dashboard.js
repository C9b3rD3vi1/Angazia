// Employee Dashboard - Interactive Features
(function () {
  'use strict';

  document.addEventListener('DOMContentLoaded', function () {
    setupEventListeners();
  });

  function setupEventListeners() {
    document.addEventListener('click', function (e) {
      var applyBtn = e.target.closest('[data-action="apply-job"]');
      if (applyBtn) {
        var jobId = applyBtn.getAttribute('data-id');
        if (jobId) window.applyToJob && window.applyToJob(jobId);
        return;
      }
    });
  }

  window.applyToJob = async function (jobId) {
    if (!jobId) return;
    try {
      await AngaziaAPI.applications.apply({ job_id: jobId });
      if (AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Application submitted successfully!', 'success');
      }
    } catch (error) {
      console.error('Apply error:', error);
      if (AngaziaApp && AngaziaApp.showToast) {
        AngaziaApp.showToast('Failed to submit application', 'error');
      }
    }
  };

  window.confirmInterview = function (interviewId) {
    if (AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast('Interview confirmation coming soon', 'info');
    }
  };

  window.rescheduleInterview = function (interviewId) {
    if (AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast('Rescheduling feature coming soon', 'info');
    }
  };

  window.addToCalendar = function (interviewId) {
    if (AngaziaApp && AngaziaApp.showToast) {
      AngaziaApp.showToast('Calendar integration coming soon', 'info');
    }
  };
})();
