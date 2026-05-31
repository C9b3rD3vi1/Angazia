# Email Templates

This directory contains HTML email templates for the application.

## Structure

- `base.html` - Base template with common structure and styles
- `verification.html` - Email verification template
- `password_reset.html` - Password reset template
- `password_changed.html` - Password changed confirmation
- `welcome.html` - Welcome email for new users
- `application_status.html` - Application status updates
- `job_alert.html` - Job alert notifications
- `employer_welcome.html` - Welcome email for employers

## Styling

All emails use inline CSS for maximum email client compatibility. The base template includes:
- Responsive design for mobile devices
- Dark mode support
- Consistent branding colors
- Accessible typography

## Adding New Templates

1. Create a new HTML file in this directory
2. Define a template with `{{define "name"}}` block
3. Add to `loadTemplates()` in email service
4. Create a new method in email service to use the template

## Testing

Test emails using:
- Email testing services (Mailtrap, Ethereal)
- Real email clients (Gmail, Outlook, Apple Mail)
- Mobile devices

## Variables

Available variables in all templates:
- `{{.AppName}}` - Application name
- `{{.AppURL}}` - Application base URL
- `{{.Year}}` - Current year
- `{{.Email}}` - Recipient's email
- `{{.Subject}}` - Email subject line