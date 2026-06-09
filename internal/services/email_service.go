package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/resend/resend-go/v2"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gopkg.in/gomail.v2"

	"github.com/C9b3rD3vi1/Angazia/internal/config"
	"github.com/C9b3rD3vi1/Angazia/internal/models"
	"github.com/C9b3rD3vi1/Angazia/internal/repository"
)

type EmailProvider string

const (
	ProviderSendGrid EmailProvider = "sendgrid"
	ProviderResend   EmailProvider = "resend"
	ProviderSMTP     EmailProvider = "smtp"
)

type EmailService interface {
	// Account related emails
	SendVerificationEmail(to, token, userID, email string)
	SendPasswordResetEmail(to, token, userID, email string)
	SendPasswordChangedEmail(to, email string)
	SendWelcomeEmail(to, name, email string)
	SendEmployerWelcomeEmail(to, companyName, email string)

	// Application related emails
	SendApplicationConfirmation(to, jobTitle, companyName, applicationID, email string)
	SendNewApplicationNotification(to, jobTitle, candidateName, applicationID, email string)
	SendApplicationStatusUpdate(to, jobTitle, companyName, status, email string)
	SendInterviewInvitation(to, jobTitle, companyName string, interviewDate time.Time, interviewType, applicationID, email string)
	SendHiredNotification(to, jobTitle, companyName, email string)

	// Job alert emails
	SendJobAlertEmail(to string, jobs []JobAlert, email string)

	// Company verification emails
	SendAdminVerificationNotification(adminEmail, companyID, registrationNumber, taxID, verificationID string)
	SendVerificationApprovedEmail(to, companyName string)
	SendVerificationRejectedEmail(to, companyName, reason string)
	
	// Team invitation emails
	SendTeamInvitationExistingUser(to, companyName, role, token string)
	SendTeamInvitationNewUser(to, companyName, role, token string)

	// Billing and payment emails
	SendPaymentConfirmationEmail(to, planName, amount, currency, invoiceNumber, email string)
	SendPaymentFailedEmail(to, planName, amount, currency, reason, email string)
	SendSubscriptionCancelledEmail(to, planName, endDate, email string)
	SendSubscriptionReactivatedEmail(to, planName, email string)
	SendRenewalReminderEmail(to string, planName string, amount string, currency string, daysUntilRenewal int, email string)
	SendInvoiceAvailableEmail(to, invoiceNumber, amount, currency, invoiceURL, email string)

	// 2FA emails
	SendTwoFACode(to, code string) error
	SendTwoFABackupCodes(to string, codes []string) error
	SendTwoFAEnabled(to string) error
	SendTwoFADisabled(to string) error
	SendTwoFARecoveryEmail(to, recoveryLink string) error

	// Notification emails
	SendNotificationEmail(to, subject, htmlBody, textBody, email string) error
}

type EmailServiceImpl struct {
	cfg             *config.Config
	provider        EmailProvider
	sendGridClient  *sendgrid.Client
	resendClient    *resend.Client
	smtpDialer      *gomail.Dialer
	templateCache   map[string]*template.Template
	unsubscribeRepo repository.UnsubscribeRepository
	tokenService    TokenService
	mu              sync.RWMutex
}

type JobAlert struct {
	Title       string
	Company     string
	Location    string
	SalaryRange string
	Skills      string
	JobURL      string
}

type EmailData struct {
	// Common fields
	AppName          string
	AppURL           string
	AppDomain        string
	Year             int
	Email            string
	Subject          string
	HeaderTitle      string
	HeaderColor      string
	ButtonColor      string
	ButtonHoverColor string
	LinkColor        string
	StatusColor      string
	
	// Company verification fields
	CompanyID          string
	RegistrationNumber string
	TaxID              string
	VerificationID     string
	CompanyName        string
	Reason             string
	
	// Team invitation fields
	Role             string
	Token            string
	
	// Template-specific content
	BodyContent      template.HTML
	VerificationURL  string
	ResetURL         string
	Name             string
	DashboardURL     string
	HelpURL          string
	PostJobURL       string
	OnboardingURL    string
	JobTitle         string
	Company          string
	Status           string
	CandidateName    string
	ApplicationID    string
	ApplicationURL   string
	InterviewDate    string
	InterviewType    string
	Jobs             []JobAlert
	Count            int
	UnsubscribeToken string

	// 2FA fields
	Code         string   `json:"code,omitempty"`
	Codes        []string `json:"codes,omitempty"`
	RecoveryLink string   `json:"recovery_link,omitempty"`
}

func NewEmailService(cfg *config.Config, unsubscribeRepo repository.UnsubscribeRepository, tokenService TokenService) EmailService {
	service := &EmailServiceImpl{
		cfg:             cfg,
		unsubscribeRepo: unsubscribeRepo,
		tokenService:    tokenService,
		templateCache:   make(map[string]*template.Template),
	}

	if err := service.loadTemplates(); err != nil {
		fmt.Printf("Warning: Failed to load email templates: %v\n", err)
	}

	switch cfg.EmailProvider {
	case "sendgrid":
		service.provider = ProviderSendGrid
		service.sendGridClient = sendgrid.NewSendClient(cfg.SendGridAPIKey)
		fmt.Println("📧 Email service initialized with SendGrid")
	case "resend":
		service.provider = ProviderResend
		service.resendClient = resend.NewClient(cfg.ResendAPIKey)
		fmt.Println("📧 Email service initialized with Resend")
	default:
		service.provider = ProviderSMTP
		smtpPort, _ := strconv.Atoi(cfg.SMTPPort)
		if smtpPort == 0 {
			smtpPort = 587
		}
		service.smtpDialer = gomail.NewDialer(
			cfg.SMTPHost,
			smtpPort,
			cfg.SMTPUser,
			cfg.SMTPPassword,
		)
		fmt.Println("📧 Email service initialized with SMTP")
	}

	return service
}

func (s *EmailServiceImpl) loadTemplates() error {
	templateDir := "web/templates/emails"
	baseTemplate := filepath.Join(templateDir, "base.html")

	templates := map[string]string{
		"verification":                 filepath.Join(templateDir, "verification.html"),
		"password_reset":               filepath.Join(templateDir, "password_reset.html"),
		"password_changed":             filepath.Join(templateDir, "password_changed.html"),
		"welcome":                      filepath.Join(templateDir, "welcome.html"),
		"employer_welcome":             filepath.Join(templateDir, "employer_welcome.html"),
		"application_confirmation":     filepath.Join(templateDir, "application_confirmation.html"),
		"new_application_notification": filepath.Join(templateDir, "new_application_notification.html"),
		"application_status_update":    filepath.Join(templateDir, "application_status_update.html"),
		"interview_invitation":         filepath.Join(templateDir, "interview_invitation.html"),
		"hired_notification":           filepath.Join(templateDir, "hired_notification.html"),
		"job_alert":                    filepath.Join(templateDir, "job_alert.html"),
		// New templates
		"admin_verification":           filepath.Join(templateDir, "admin_verification.html"),
		"verification_approved":        filepath.Join(templateDir, "verification_approved.html"),
		"verification_rejected":        filepath.Join(templateDir, "verification_rejected.html"),
		"team_invitation_existing":     filepath.Join(templateDir, "team_invitation_existing.html"),
		"team_invitation_new":          filepath.Join(templateDir, "team_invitation_new.html"),
		// 2FA templates
		"twofa_code":                   filepath.Join(templateDir, "twofa_code.html"),
		"twofa_backup_codes":           filepath.Join(templateDir, "twofa_backup_codes.html"),
		"twofa_enabled":                filepath.Join(templateDir, "twofa_enabled.html"),
		"twofa_disabled":               filepath.Join(templateDir, "twofa_disabled.html"),
		"twofa_recovery":               filepath.Join(templateDir, "twofa_recovery.html"),
	}
	
	for name, tmplPath := range templates {
		tmpl, err := template.ParseFiles(baseTemplate, tmplPath)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		s.mu.Lock()
		s.templateCache[name] = tmpl
		s.mu.Unlock()
	}

	fmt.Printf("📧 Loaded %d email templates\n", len(templates))
	return nil
}

func (s *EmailServiceImpl) renderTemplate(templateName string, data *EmailData) (string, error) {
	s.mu.RLock()
	tmpl, ok := s.templateCache[templateName]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (s *EmailServiceImpl) SendVerificationEmail(to, token, userID, email string) {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s&user_id=%s", s.cfg.AppURL, token, userID)

	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          "Verify Your Email Address - " + s.cfg.AppName,
		HeaderTitle:      "Welcome to " + s.cfg.AppName + "!",
		HeaderColor:      "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor:      "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:        "#667eea",
		VerificationURL:  verificationURL,
	}

	htmlBody, err := s.renderTemplate("verification", data)
	if err != nil {
		fmt.Printf("Failed to render verification email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Please verify your email address by clicking this link: %s", verificationURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendPasswordResetEmail(to, token, userID, email string) {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s&user_id=%s", s.cfg.AppURL, token, userID)

	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          "Reset Your Password - " + s.cfg.AppName,
		HeaderTitle:      "Reset Your Password",
		HeaderColor:      "#ef4444",
		ButtonColor:      "#ef4444",
		ButtonHoverColor: "#dc2626",
		LinkColor:        "#ef4444",
		ResetURL:         resetURL,
	}

	htmlBody, err := s.renderTemplate("password_reset", data)
	if err != nil {
		fmt.Printf("Failed to render password reset email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Reset your password by clicking this link: %s", resetURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendPasswordChangedEmail(to, email string) {
	data := &EmailData{
		AppName:     s.cfg.AppName,
		AppURL:      s.cfg.AppURL,
		AppDomain:   s.cfg.AppDomain,
		Year:        time.Now().Year(),
		Email:       email,
		Subject:     "Password Changed Successfully - " + s.cfg.AppName,
		HeaderTitle: "Password Changed",
		HeaderColor: "#10b981",
		LinkColor:   "#10b981",
	}

	htmlBody, err := s.renderTemplate("password_changed", data)
	if err != nil {
		fmt.Printf("Failed to render password changed email: %v\n", err)
		return
	}

	textBody := "Your password has been changed successfully. If this wasn't you, please contact support immediately."
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendWelcomeEmail(to, name, email string) {
	if name == "" {
		name = "there"
	}

	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("Welcome to %s! 🚀", s.cfg.AppName),
		HeaderTitle:      fmt.Sprintf("Welcome to %s, %s! 🎉", s.cfg.AppName, name),
		HeaderColor:      "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor:      "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:        "#667eea",
		Name:             name,
		DashboardURL:     fmt.Sprintf("%s/dashboard", s.cfg.AppURL),
		HelpURL:          fmt.Sprintf("%s/help", s.cfg.AppURL),
	}

	htmlBody, err := s.renderTemplate("welcome", data)
	if err != nil {
		fmt.Printf("Failed to render welcome email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Welcome to %s, %s! Start your journey today at %s/dashboard", s.cfg.AppName, name, s.cfg.AppURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendEmployerWelcomeEmail(to, companyName, email string) {
	if companyName == "" {
		companyName = "your company"
	}

	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("Welcome to %s - Start Hiring! 🚀", s.cfg.AppName),
		HeaderTitle:      fmt.Sprintf("Welcome to %s, %s! 🚀", s.cfg.AppName, companyName),
		HeaderColor:      "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor:      "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:        "#667eea",
		CompanyName:      companyName,
		PostJobURL:       fmt.Sprintf("%s/employer/jobs/post", s.cfg.AppURL),
		HelpURL:          fmt.Sprintf("%s/employer/help", s.cfg.AppURL),
	}

	htmlBody, err := s.renderTemplate("employer_welcome", data)
	if err != nil {
		fmt.Printf("Failed to render employer welcome email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Welcome to %s! Post your first job and find top tech talent in Kenya. %s/employer/jobs/post", s.cfg.AppName, s.cfg.AppURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendApplicationConfirmation(to, jobTitle, companyName, applicationID, email string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("Application Confirmed: %s at %s", jobTitle, companyName),
		HeaderTitle:      "Application Submitted Successfully!",
		HeaderColor:      "linear-gradient(135deg, #10b981 0%, #059669 100%)",
		ButtonColor:      "#10b981",
		ButtonHoverColor: "#059669",
		LinkColor:        "#10b981",
		JobTitle:         jobTitle,
		Company:          companyName,
		ApplicationID:    applicationID,
		ApplicationURL:   fmt.Sprintf("%s/employee/applications/%s", s.cfg.AppURL, applicationID),
	}

	htmlBody, err := s.renderTemplate("application_confirmation", data)
	if err != nil {
		fmt.Printf("Failed to render application confirmation email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your application for %s at %s has been submitted successfully.\n\nView your application: %s", jobTitle, companyName, data.ApplicationURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendNewApplicationNotification(to, jobTitle, candidateName, applicationID, email string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("New Application: %s for %s", candidateName, jobTitle),
		HeaderTitle:      "New Job Application Received!",
		HeaderColor:      "linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)",
		ButtonColor:      "#3b82f6",
		ButtonHoverColor: "#2563eb",
		LinkColor:        "#3b82f6",
		JobTitle:         jobTitle,
		CandidateName:    candidateName,
		ApplicationID:    applicationID,
		ApplicationURL:   fmt.Sprintf("%s/employer/applications/%s", s.cfg.AppURL, applicationID),
	}

	htmlBody, err := s.renderTemplate("new_application_notification", data)
	if err != nil {
		fmt.Printf("Failed to render new application notification email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("New application received for %s from %s.\n\nView application: %s", jobTitle, candidateName, data.ApplicationURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendApplicationStatusUpdate(to, jobTitle, companyName, status, email string) {
	statusDisplay := "Under Review"
	statusColor := "#f59e0b"

	switch status {
	case "shortlisted":
		statusDisplay = "Shortlisted 🎉"
		statusColor = "#10b981"
	case "rejected":
		statusDisplay = "Not Selected"
		statusColor = "#ef4444"
	case "hired":
		statusDisplay = "Hired! 🎊"
		statusColor = "#10b981"
	case "interview":
		statusDisplay = "Interview Scheduled"
		statusColor = "#3b82f6"
	case "viewed":
		statusDisplay = "Application Viewed"
		statusColor = "#8b5cf6"
	}

	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("Application Status Update - %s at %s", jobTitle, companyName),
		HeaderTitle:      "Application Status Updated",
		HeaderColor:      statusColor,
		ButtonColor:      "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:        "#667eea",
		StatusColor:      statusColor,
		JobTitle:         jobTitle,
		Company:          companyName,
		Status:           statusDisplay,
	}

	htmlBody, err := s.renderTemplate("application_status_update", data)
	if err != nil {
		fmt.Printf("Failed to render application status update email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your application for %s at %s has been %s", jobTitle, companyName, statusDisplay)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendInterviewInvitation(to, jobTitle, companyName string, interviewDate time.Time, interviewType, applicationID, email string) {
	dateStr := interviewDate.Format("Monday, January 2, 2006 at 3:04 PM")

	typeDisplay := "Interview"
	switch interviewType {
	case "phone":
		typeDisplay = "Phone Screening"
	case "technical":
		typeDisplay = "Technical Interview"
	case "onsite":
		typeDisplay = "On-site Interview"
	case "final":
		typeDisplay = "Final Interview"
	}

	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("Interview Invitation: %s at %s", jobTitle, companyName),
		HeaderTitle:      "Interview Invitation",
		HeaderColor:      "linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)",
		ButtonColor:      "#8b5cf6",
		ButtonHoverColor: "#7c3aed",
		LinkColor:        "#8b5cf6",
		JobTitle:         jobTitle,
		Company:          companyName,
		ApplicationID:    applicationID,
		InterviewDate:    dateStr,
		InterviewType:    typeDisplay,
	}

	htmlBody, err := s.renderTemplate("interview_invitation", data)
	if err != nil {
		fmt.Printf("Failed to render interview invitation email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("You have been invited for a %s for %s at %s on %s", typeDisplay, jobTitle, companyName, dateStr)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendHiredNotification(to, jobTitle, companyName, email string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("Congratulations! You've been hired for %s at %s", jobTitle, companyName),
		HeaderTitle:      "Congratulations! 🎉",
		HeaderColor:      "linear-gradient(135deg, #f59e0b 0%, #d97706 100%)",
		ButtonColor:      "#f59e0b",
		ButtonHoverColor: "#d97706",
		LinkColor:        "#f59e0b",
		JobTitle:         jobTitle,
		Company:          companyName,
		OnboardingURL:    fmt.Sprintf("%s/onboarding", s.cfg.AppURL),
	}

	htmlBody, err := s.renderTemplate("hired_notification", data)
	if err != nil {
		fmt.Printf("Failed to render hired notification email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Congratulations! You've been hired for %s at %s. Please complete your onboarding at %s", jobTitle, companyName, data.OnboardingURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendJobAlertEmail(to string, jobs []JobAlert, email string) {
	unsubscribeToken, err := s.generateAndStoreUnsubscribeToken(email, "")
	if err != nil {
		fmt.Printf("Failed to generate unsubscribe token for %s: %v\n", email, err)
		unsubscribeToken = ""
	}

	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            email,
		Subject:          fmt.Sprintf("New Job Matches for You! 🎯 (%d jobs)", len(jobs)),
		HeaderTitle:      "New Job Matches! 🎯",
		HeaderColor:      "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor:      "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:        "#667eea",
		Jobs:             jobs,
		Count:            len(jobs),
		UnsubscribeToken: unsubscribeToken,
	}

	htmlBody, err := s.renderTemplate("job_alert", data)
	if err != nil {
		fmt.Printf("Failed to render job alert email: %v\n", err)
		return
	}

	var jobsText string
	for _, job := range jobs {
		jobsText += fmt.Sprintf("- %s at %s (%s)\n", job.Title, job.Company, job.Location)
	}
	textBody := fmt.Sprintf("Found %d new jobs matching your profile:\n\n%s\n\nView all jobs: %s/jobs\n\nTo unsubscribe: %s/unsubscribe?token=%s",
		len(jobs), jobsText, s.cfg.AppURL, s.cfg.AppURL, unsubscribeToken)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) generateAndStoreUnsubscribeToken(email, userID string) (string, error) {
	existingToken, err := s.unsubscribeRepo.GetActiveTokenByEmail(context.Background(), email)
	if err == nil && existingToken != nil {
		return existingToken.Token, nil
	}

	token, err := s.tokenService.GenerateUnsubscribeToken(email, userID)
	if err != nil {
		return "", err
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	unsubscribeToken := &models.UnsubscribeToken{
		ID:        uuid.New().String(),
		Email:     email,
		Token:     token,
		UserID:    userIDPtr,
		IsActive:  true,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().AddDate(1, 0, 0),
	}

	if err := s.unsubscribeRepo.CreateToken(context.Background(), unsubscribeToken); err != nil {
		return "", err
	}

	return token, nil
}

func (s *EmailServiceImpl) sendEmail(to, subject, htmlBody, textBody, email string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic while sending email to %s: %v\n", to, r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error

		switch s.provider {
		case ProviderSendGrid:
			from := mail.NewEmail(s.cfg.SMTPFromName, s.cfg.SMTPFromEmail)
			toEmail := mail.NewEmail(to, email)
			message := mail.NewSingleEmail(from, subject, toEmail, textBody, htmlBody)
			response, err := s.sendGridClient.Send(message)
			if err != nil {
				err = fmt.Errorf("sendgrid error: %w", err)
			} else if response.StatusCode >= 400 {
				err = fmt.Errorf("sendgrid returned status: %d, body: %s", response.StatusCode, response.Body)
			}
		case ProviderResend:
			params := &resend.SendEmailRequest{
				From:    fmt.Sprintf("%s <%s>", s.cfg.SMTPFromName, s.cfg.SMTPFromEmail),
				To:      []string{email},
				Subject: subject,
				Html:    htmlBody,
				Text:    textBody,
			}
			_, err = s.resendClient.Emails.SendWithContext(ctx, params)
			if err != nil {
				err = fmt.Errorf("resend error: %w", err)
			}
		default:
			m := gomail.NewMessage()
			m.SetHeader("From", fmt.Sprintf("%s <%s>", s.cfg.SMTPFromName, s.cfg.SMTPFromEmail))
			m.SetHeader("To", to)
			m.SetHeader("Subject", subject)
			m.SetBody("text/plain", textBody)
			m.AddAlternative("text/html", htmlBody)
			err = s.smtpDialer.DialAndSend(m)
			if err != nil {
				err = fmt.Errorf("SMTP error: %w", err)
			}
		}

		if err != nil {
			fmt.Printf("Failed to send email to %s: %v\n", to, err)
		} else {
			fmt.Printf("Email sent successfully to %s: %s\n", to, subject)
		}
	}()
}


func (s *EmailServiceImpl) SendAdminVerificationNotification(adminEmail, companyID, registrationNumber, taxID, verificationID string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            adminEmail,
		Subject:          "New Company Verification Request - " + s.cfg.AppName,
		HeaderTitle:      "New Verification Request",
		HeaderColor:      "linear-gradient(135deg, #f59e0b 0%, #d97706 100%)",
		ButtonColor:      "#f59e0b",
		ButtonHoverColor: "#d97706",
		LinkColor:        "#f59e0b",
		CompanyID:        companyID,
		RegistrationNumber: registrationNumber,
		TaxID:            taxID,
		VerificationID:   verificationID,
	}

	htmlBody, err := s.renderTemplate("admin_verification", data)
	if err != nil {
		fmt.Printf("Failed to render admin verification email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("New verification request from company %s\nRegistration: %s\nTax ID: %s", companyID, registrationNumber, taxID)
	s.sendEmail(adminEmail, data.Subject, htmlBody, textBody, adminEmail)
}

func (s *EmailServiceImpl) SendVerificationApprovedEmail(to, companyName string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            to,
		Subject:          "Your Company Has Been Verified - " + s.cfg.AppName,
		HeaderTitle:      "Verification Approved! 🎉",
		HeaderColor:      "linear-gradient(135deg, #10b981 0%, #059669 100%)",
		ButtonColor:      "#10b981",
		ButtonHoverColor: "#059669",
		LinkColor:        "#10b981",
		CompanyName:      companyName,
	}

	htmlBody, err := s.renderTemplate("verification_approved", data)
	if err != nil {
		fmt.Printf("Failed to render verification approved email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Congratulations! Your company %s has been verified.", companyName)
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
}

func (s *EmailServiceImpl) SendVerificationRejectedEmail(to, companyName, reason string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            to,
		Subject:          "Company Verification Update - " + s.cfg.AppName,
		HeaderTitle:      "Verification Status",
		HeaderColor:      "linear-gradient(135deg, #ef4444 0%, #dc2626 100%)",
		ButtonColor:      "#ef4444",
		ButtonHoverColor: "#dc2626",
		LinkColor:        "#ef4444",
		CompanyName:      companyName,
		Reason:           reason,
	}

	htmlBody, err := s.renderTemplate("verification_rejected", data)
	if err != nil {
		fmt.Printf("Failed to render verification rejected email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your company %s verification was rejected. Reason: %s", companyName, reason)
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
}

func (s *EmailServiceImpl) SendTeamInvitationExistingUser(to, companyName, role, token string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            to,
		Subject:          fmt.Sprintf("You've been invited to join %s on %s", companyName, s.cfg.AppName),
		HeaderTitle:      "Team Invitation",
		HeaderColor:      "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor:      "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:        "#667eea",
		CompanyName:      companyName,
		Role:             role,
		Token:            token,
	}

	htmlBody, err := s.renderTemplate("team_invitation_existing", data)
	if err != nil {
		fmt.Printf("Failed to render team invitation email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("You've been invited to join %s as %s. Accept your invitation: %s/api/v1/invitations/%s/accept", companyName, role, s.cfg.AppURL, token)
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
}

func (s *EmailServiceImpl) SendTeamInvitationNewUser(to, companyName, role, token string) {
	data := &EmailData{
		AppName:          s.cfg.AppName,
		AppURL:           s.cfg.AppURL,
		AppDomain:        s.cfg.AppDomain,
		Year:             time.Now().Year(),
		Email:            to,
		Subject:          fmt.Sprintf("You've been invited to join %s on %s", companyName, s.cfg.AppName),
		HeaderTitle:      "Welcome to " + s.cfg.AppName + "!",
		HeaderColor:      "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor:      "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:        "#667eea",
		CompanyName:      companyName,
		Role:             role,
		Token:            token,
	}

	htmlBody, err := s.renderTemplate("team_invitation_new", data)
	if err != nil {
		fmt.Printf("Failed to render team invitation email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("You've been invited to join %s as %s. Create your account: %s/register?email=%s&invite=%s", companyName, role, s.cfg.AppURL, to, token)
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
}

func (s *EmailServiceImpl) SendPaymentConfirmationEmail(to, planName, amount, currency, invoiceNumber, email string) {
	data := &EmailData{
		AppName:     s.cfg.AppName,
		AppURL:      s.cfg.AppURL,
		AppDomain:   s.cfg.AppDomain,
		Year:        time.Now().Year(),
		Email:       email,
		Subject:     fmt.Sprintf("Payment Confirmed - %s", s.cfg.AppName),
		HeaderTitle: "Payment Successful!",
		HeaderColor: "linear-gradient(135deg, #10b981 0%, #059669 100%)",
		ButtonColor: "#10b981",
		Name:        planName,
		Company:     fmt.Sprintf("%s %s", amount, currency),
		Status:      invoiceNumber,
	}

	htmlBody, err := s.renderTemplate("payment_confirmation", data)
	if err != nil {
		fmt.Printf("Failed to render payment confirmation email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your payment of %s %s for %s has been confirmed. Invoice: %s.", amount, currency, planName, invoiceNumber)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendPaymentFailedEmail(to, planName, amount, currency, reason, email string) {
	data := &EmailData{
		AppName:     s.cfg.AppName,
		AppURL:      s.cfg.AppURL,
		AppDomain:   s.cfg.AppDomain,
		Year:        time.Now().Year(),
		Email:       email,
		Subject:     fmt.Sprintf("Payment Failed - %s", s.cfg.AppName),
		HeaderTitle: "Payment Failed",
		HeaderColor: "linear-gradient(135deg, #ef4444 0%, #dc2626 100%)",
		ButtonColor: "#ef4444",
		LinkColor:   "#ef4444",
		Name:        planName,
		Company:     fmt.Sprintf("%s %s", amount, currency),
		Reason:      reason,
	}

	htmlBody, err := s.renderTemplate("payment_failed", data)
	if err != nil {
		fmt.Printf("Failed to render payment failed email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your payment of %s %s for %s has failed. Reason: %s. Please update your payment method.", amount, currency, planName, reason)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendSubscriptionCancelledEmail(to, planName, endDate, email string) {
	data := &EmailData{
		AppName:     s.cfg.AppName,
		AppURL:      s.cfg.AppURL,
		AppDomain:   s.cfg.AppDomain,
		Year:        time.Now().Year(),
		Email:       email,
		Subject:     fmt.Sprintf("Subscription Cancelled - %s", s.cfg.AppName),
		HeaderTitle: "Subscription Cancelled",
		HeaderColor: "linear-gradient(135deg, #f59e0b 0%, #d97706 100%)",
		ButtonColor: "#f59e0b",
		Name:        planName,
		Status:      endDate,
	}

	htmlBody, err := s.renderTemplate("subscription_cancelled", data)
	if err != nil {
		fmt.Printf("Failed to render subscription cancelled email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your %s subscription has been cancelled. You will have access until %s.", planName, endDate)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendSubscriptionReactivatedEmail(to, planName, email string) {
	data := &EmailData{
		AppName:     s.cfg.AppName,
		AppURL:      s.cfg.AppURL,
		AppDomain:   s.cfg.AppDomain,
		Year:        time.Now().Year(),
		Email:       email,
		Subject:     fmt.Sprintf("Subscription Reactivated - %s", s.cfg.AppName),
		HeaderTitle: "Welcome Back!",
		HeaderColor: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor: "#667eea",
		Name:        planName,
	}

	htmlBody, err := s.renderTemplate("subscription_reactivated", data)
	if err != nil {
		fmt.Printf("Failed to render subscription reactivated email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your %s subscription has been reactivated.", planName)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendRenewalReminderEmail(to, planName, amount, currency string, daysUntilRenewal int, email string) {
	data := &EmailData{
		AppName:     s.cfg.AppName,
		AppURL:      s.cfg.AppURL,
		AppDomain:   s.cfg.AppDomain,
		Year:        time.Now().Year(),
		Email:       email,
		Subject:     fmt.Sprintf("Subscription Renewal Reminder - %s", s.cfg.AppName),
		HeaderTitle: "Renewal Reminder",
		HeaderColor: "linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)",
		ButtonColor: "#3b82f6",
		Name:        planName,
		Company:     fmt.Sprintf("%s %s", amount, currency),
		Count:       daysUntilRenewal,
	}

	htmlBody, err := s.renderTemplate("renewal_reminder", data)
	if err != nil {
		fmt.Printf("Failed to render renewal reminder email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Your %s subscription will renew in %d days for %s %s.", planName, daysUntilRenewal, amount, currency)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}

func (s *EmailServiceImpl) SendInvoiceAvailableEmail(to, invoiceNumber, amount, currency, invoiceURL, email string) {
	data := &EmailData{
		AppName:     s.cfg.AppName,
		AppURL:      s.cfg.AppURL,
		AppDomain:   s.cfg.AppDomain,
		Year:        time.Now().Year(),
		Email:       email,
		Subject:     fmt.Sprintf("Invoice Available - %s", invoiceNumber),
		HeaderTitle: "Invoice Available",
		HeaderColor: "linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)",
		ButtonColor: "#8b5cf6",
		LinkColor:   "#8b5cf6",
		Name:        invoiceNumber,
		Company:     fmt.Sprintf("%s %s", amount, currency),
		VerificationURL: invoiceURL,
	}

	htmlBody, err := s.renderTemplate("invoice_available", data)
	if err != nil {
		fmt.Printf("Failed to render invoice email: %v\n", err)
		return
	}

	textBody := fmt.Sprintf("Invoice %s for %s %s is now available. View: %s", invoiceNumber, amount, currency, invoiceURL)
	s.sendEmail(to, data.Subject, htmlBody, textBody, email)
}


func (s *EmailServiceImpl) SendTwoFACode(to, code string) error {
	data := &EmailData{
		AppName:      s.cfg.AppName,
		AppURL:       s.cfg.AppURL,
		AppDomain:    s.cfg.AppDomain,
		Year:         time.Now().Year(),
		Email:        to,
		Name:         to,
		Subject:      "Your Two-Factor Authentication Code - " + s.cfg.AppName,
		HeaderTitle:  "Verification Code",
		HeaderColor:  "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		ButtonColor:  "#667eea",
		ButtonHoverColor: "#5a67d8",
		LinkColor:    "#667eea",
		Code:         code,
	}

	htmlBody, err := s.renderTemplate("twofa_code", data)
	if err != nil {
		return fmt.Errorf("failed to render 2FA code template: %w", err)
	}
	textBody := fmt.Sprintf("Your verification code is: %s", code)
	
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
	return nil
}

func (s *EmailServiceImpl) SendTwoFABackupCodes(to string, codes []string) error {
	data := &EmailData{
		AppName:      s.cfg.AppName,
		AppURL:       s.cfg.AppURL,
		AppDomain:    s.cfg.AppDomain,
		Year:         time.Now().Year(),
		Email:        to,
		Name:         to,
		Subject:      "Your Backup Codes - " + s.cfg.AppName,
		HeaderTitle:  "Backup Codes",
		HeaderColor:  "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
		Codes:        codes,
	}

	htmlBody, err := s.renderTemplate("twofa_backup_codes", data)
	if err != nil {
		return fmt.Errorf("failed to render backup codes template: %w", err)
	}
	textBody := "Your backup codes have been generated. Please store them securely."
	
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
	return nil
}

func (s *EmailServiceImpl) SendTwoFAEnabled(to string) error {
	data := &EmailData{
		AppName:      s.cfg.AppName,
		AppURL:       s.cfg.AppURL,
		AppDomain:    s.cfg.AppDomain,
		Year:         time.Now().Year(),
		Email:        to,
		Name:         to,
		Subject:      "Two-Factor Authentication Enabled - " + s.cfg.AppName,
		HeaderTitle:  "2FA Enabled",
		HeaderColor:  "linear-gradient(135deg, #10b981 0%, #059669 100%)",
	}

	htmlBody, err := s.renderTemplate("twofa_enabled", data)
	if err != nil {
		return fmt.Errorf("failed to render 2FA enabled template: %w", err)
	}
	textBody := "Two-factor authentication has been enabled on your account."
	
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
	return nil
}

func (s *EmailServiceImpl) SendTwoFADisabled(to string) error {
	data := &EmailData{
		AppName:      s.cfg.AppName,
		AppURL:       s.cfg.AppURL,
		AppDomain:    s.cfg.AppDomain,
		Year:         time.Now().Year(),
		Email:        to,
		Name:         to,
		Subject:      "Two-Factor Authentication Disabled - " + s.cfg.AppName,
		HeaderTitle:  "2FA Disabled",
		HeaderColor:  "linear-gradient(135deg, #ef4444 0%, #dc2626 100%)",
	}

	htmlBody, err := s.renderTemplate("twofa_disabled", data)
	if err != nil {
		return fmt.Errorf("failed to render 2FA disabled template: %w", err)
	}
	textBody := "Two-factor authentication has been disabled on your account."
	
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
	return nil
}

func (s *EmailServiceImpl) SendTwoFARecoveryEmail(to, recoveryLink string) error {
	data := &EmailData{
		AppName:      s.cfg.AppName,
		AppURL:       s.cfg.AppURL,
		AppDomain:    s.cfg.AppDomain,
		Year:         time.Now().Year(),
		Email:        to,
		Name:         to,
		Subject:      "Two-Factor Authentication Recovery - " + s.cfg.AppName,
		HeaderTitle:  "2FA Recovery",
		HeaderColor:  "linear-gradient(135deg, #f59e0b 0%, #d97706 100%)",
		RecoveryLink: recoveryLink,
	}

	htmlBody, err := s.renderTemplate("twofa_recovery", data)
	if err != nil {
		return fmt.Errorf("failed to render 2FA recovery template: %w", err)
	}
	textBody := fmt.Sprintf("Click this link to recover your account: %s\n\nThis link expires in 15 minutes.", recoveryLink)
	s.sendEmail(to, data.Subject, htmlBody, textBody, to)
	return nil
}

func (s *EmailServiceImpl) SendNotificationEmail(to, subject, htmlBody, textBody, email string) error {
	s.sendEmail(to, subject, htmlBody, textBody, email)
	return nil
}