package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/gomail.v2"
)

// EmailService provides email functionality
type EmailService struct {
	Host     string
	Port     int
	Username string
	Password string
	FromName string
	FromAddr string
	Debug    bool
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject string
	HTML    string
	Text    string
}

// NewEmailService creates a new email service with configuration from environment variables
func NewEmailService() (*EmailService, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}

	portStr := os.Getenv("SMTP_PORT")
	if portStr == "" {
		portStr = "587"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP port: %v", err)
	}

	username := os.Getenv("SMTP_USERNAME")
	if username == "" {
		return nil, fmt.Errorf("SMTP_USERNAME environment variable is required")
	}

	password := os.Getenv("SMTP_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD environment variable is required")
	}

	fromName := os.Getenv("EMAIL_FROM_NAME")
	if fromName == "" {
		fromName = "BlogCommerce"
	}

	fromAddr := os.Getenv("EMAIL_FROM_ADDRESS")
	if fromAddr == "" {
		fromAddr = username
	}

	debug := os.Getenv("EMAIL_DEBUG") == "true"

	return &EmailService{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		FromName: fromName,
		FromAddr: fromAddr,
		Debug:    debug,
	}, nil
}

// SendEmail sends an email
func (s *EmailService) SendEmail(to, subject, htmlBody, textBody string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.FromAddr, s.FromName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	
	// Set HTML body
	m.SetBody("text/html", htmlBody)
	
	// Set text body as alternative
	if textBody != "" {
		m.AddAlternative("text/plain", textBody)
	}

	// Create dialer
	d := gomail.NewDialer(s.Host, s.Port, s.Username, s.Password)

	// Send email
	if s.Debug {
		fmt.Printf("DEBUG: Sending email to %s with subject: %s\n", to, subject)
		fmt.Printf("DEBUG: HTML Body: %s\n", htmlBody)
		fmt.Printf("DEBUG: Text Body: %s\n", textBody)
		return nil
	}

	return d.DialAndSend(m)
}

// SendTemplateEmail sends an email using a template and data
func (s *EmailService) SendTemplateEmail(to string, template EmailTemplate, data interface{}) error {
	// Parse HTML template
	htmlTmpl, err := s.parseTemplate(template.HTML, data)
	if err != nil {
		return fmt.Errorf("error parsing HTML template: %v", err)
	}

	// Parse text template
	var textTmpl string
	if template.Text != "" {
		textTmpl, err = s.parseTemplate(template.Text, data)
		if err != nil {
			return fmt.Errorf("error parsing text template: %v", err)
		}
	}

	// Parse subject template
	subject, err := s.parseTemplate(template.Subject, data)
	if err != nil {
		return fmt.Errorf("error parsing subject template: %v", err)
	}

	// Send email
	return s.SendEmail(to, subject, htmlTmpl, textTmpl)
}

// parseTemplate parses a template with data
func (s *EmailService) parseTemplate(content string, data interface{}) (string, error) {
	t, err := template.New("email").Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// LoadTemplateFromFile loads an email template from file
func (s *EmailService) LoadTemplateFromFile(templateDir, templateName string) (EmailTemplate, error) {
	var template EmailTemplate

	// Load subject
	subjectFile := filepath.Join(templateDir, templateName, "subject.txt")
	subjectBytes, err := os.ReadFile(subjectFile)
	if err != nil {
		return template, fmt.Errorf("error loading subject file: %v", err)
	}
	template.Subject = string(subjectBytes)

	// Load HTML body
	htmlFile := filepath.Join(templateDir, templateName, "body.html")
	htmlBytes, err := os.ReadFile(htmlFile)
	if err != nil {
		return template, fmt.Errorf("error loading HTML file: %v", err)
	}
	template.HTML = string(htmlBytes)

	// Load text body if exists
	textFile := filepath.Join(templateDir, templateName, "body.txt")
	textBytes, err := os.ReadFile(textFile)
	if err == nil {
		template.Text = string(textBytes)
	}

	return template, nil
}

// Common email templates and functions

// SendWelcomeEmail sends a welcome email to a new user
func (s *EmailService) SendWelcomeEmail(to, name string) error {
	subject := "Welcome to BlogCommerce"
	
	htmlBody := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Welcome to BlogCommerce</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.header { text-align: center; padding: 20px 0; }
			.content { padding: 20px; background-color: #f9f9f9; border-radius: 5px; }
			.footer { text-align: center; margin-top: 20px; font-size: 12px; color: #999; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>Welcome to BlogCommerce!</h1>
			</div>
			<div class="content">
				<p>Hello ` + name + `,</p>
				<p>Thank you for registering with BlogCommerce. We're excited to have you join our community!</p>
				<p>With your new account, you can:</p>
				<ul>
					<li>Shop our latest products</li>
					<li>Read and comment on blog posts</li>
					<li>Track your orders</li>
					<li>And much more!</li>
				</ul>
				<p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>
				<p>Happy shopping!</p>
			</div>
			<div class="footer">
				<p>&copy; ` + strconv.Itoa(time.Now().Year()) + ` BlogCommerce. All rights reserved.</p>
				<p>This email was sent to ` + to + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	
	textBody := `Welcome to BlogCommerce!

Hello ` + name + `,

Thank you for registering with BlogCommerce. We're excited to have you join our community!

With your new account, you can:
- Shop our latest products
- Read and comment on blog posts
- Track your orders
- And much more!

If you have any questions or need assistance, please don't hesitate to contact our support team.

Happy shopping!

© ` + strconv.Itoa(time.Now().Year()) + ` BlogCommerce. All rights reserved.
This email was sent to ` + to
	
	return s.SendEmail(to, subject, htmlBody, textBody)
}

// SendOrderConfirmationEmail sends an order confirmation email
func (s *EmailService) SendOrderConfirmationEmail(to, name, orderNumber string, orderDetails string, totalAmount float64) error {
	subject := "Order Confirmation #" + orderNumber
	
	htmlBody := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Order Confirmation</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.header { text-align: center; padding: 20px 0; }
			.content { padding: 20px; background-color: #f9f9f9; border-radius: 5px; }
			.order-details { margin: 20px 0; }
			.total { font-weight: bold; text-align: right; }
			.footer { text-align: center; margin-top: 20px; font-size: 12px; color: #999; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>Order Confirmation</h1>
			</div>
			<div class="content">
				<p>Hello ` + name + `,</p>
				<p>Thank you for your order! We're processing it now and will notify you when it ships.</p>
				<p><strong>Order Number:</strong> ` + orderNumber + `</p>
				
				<div class="order-details">
					` + orderDetails + `
				</div>
				
				<p class="total">Total: $` + fmt.Sprintf("%.2f", totalAmount) + `</p>
				
				<p>If you have any questions about your order, please contact our customer service.</p>
			</div>
			<div class="footer">
				<p>&copy; ` + strconv.Itoa(time.Now().Year()) + ` BlogCommerce. All rights reserved.</p>
				<p>This email was sent to ` + to + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	
	textBody := `Order Confirmation

Hello ` + name + `,

Thank you for your order! We're processing it now and will notify you when it ships.

Order Number: ` + orderNumber + `

` + orderDetails + `

Total: $` + fmt.Sprintf("%.2f", totalAmount) + `

If you have any questions about your order, please contact our customer service.

© ` + strconv.Itoa(time.Now().Year()) + ` BlogCommerce. All rights reserved.
This email was sent to ` + to
	
	return s.SendEmail(to, subject, htmlBody, textBody)
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to, name, resetToken, resetURL string) error {
	subject := "Password Reset Request"
	
	htmlBody := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Password Reset</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.header { text-align: center; padding: 20px 0; }
			.content { padding: 20px; background-color: #f9f9f9; border-radius: 5px; }
			.button { display: inline-block; padding: 10px 20px; background-color: #0066cc; color: white; text-decoration: none; border-radius: 5px; }
			.footer { text-align: center; margin-top: 20px; font-size: 12px; color: #999; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>Password Reset</h1>
			</div>
			<div class="content">
				<p>Hello ` + name + `,</p>
				<p>We received a request to reset your password. If you didn't make this request, you can ignore this email.</p>
				<p>To reset your password, click the button below:</p>
				<p style="text-align: center;">
					<a href="` + resetURL + `?token=` + resetToken + `" class="button">Reset Password</a>
				</p>
				<p>Or copy and paste this URL into your browser:</p>
				<p>` + resetURL + `?token=` + resetToken + `</p>
				<p>This link will expire in 1 hour.</p>
			</div>
			<div class="footer">
				<p>&copy; ` + strconv.Itoa(time.Now().Year()) + ` BlogCommerce. All rights reserved.</p>
				<p>This email was sent to ` + to + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	
	textBody := `Password Reset

Hello ` + name + `,

We received a request to reset your password. If you didn't make this request, you can ignore this email.

To reset your password, visit this link:
` + resetURL + `?token=` + resetToken + `

This link will expire in 1 hour.

© ` + strconv.Itoa(time.Now().Year()) + ` BlogCommerce. All rights reserved.
This email was sent to ` + to
	
	return s.SendEmail(to, subject, htmlBody, textBody)
}

// SendContactFormEmail sends a contact form submission to site admin
func (s *EmailService) SendContactFormEmail(name, email, subject, message string) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = s.FromAddr
	}
	
	emailSubject := "Contact Form: " + subject
	
	htmlBody := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<title>Contact Form Submission</title>
		<style>
			body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
			.container { max-width: 600px; margin: 0 auto; padding: 20px; }
			.header { padding: 10px 0; border-bottom: 1px solid #eee; }
			.content { padding: 20px 0; }
			.field { margin-bottom: 15px; }
			.label { font-weight: bold; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h2>New Contact Form Submission</h2>
			</div>
			<div class="content">
				<div class="field">
					<div class="label">Name:</div>
					<div>` + name + `</div>
				</div>
				<div class="field">
					<div class="label">Email:</div>
					<div>` + email + `</div>
				</div>
				<div class="field">
					<div class="label">Subject:</div>
					<div>` + subject + `</div>
				</div>
				<div class="field">
					<div class="label">Message:</div>
					<div>` + message + `</div>
				</div>
			</div>
		</div>
	</body>
	</html>
	`
	
	textBody := `Contact Form Submission

Name: ` + name + `
Email: ` + email + `
Subject: ` + subject + `
Message: ` + message
	
	return s.SendEmail(adminEmail, emailSubject, htmlBody, textBody)
}