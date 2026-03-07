package resend

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v3"
)

type VerificationEmailSender struct {
	client *resend.Client
	from   string
}

func NewVerificationEmailSEnder(apiKey, from string) *VerificationEmailSender {
	return &VerificationEmailSender{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

func (s *VerificationEmailSender) SendVerificationEmail(ctx context.Context, toEmail, code string) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verify your email</title>
</head>
<body style="margin: 0; padding: 0; background-color: #f3f4f6; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background-color: #f3f4f6; padding: 40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width: 480px; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07); overflow: hidden;">
          <tr>
            <td style="padding: 32px 40px 24px; text-align: center; border-bottom: 1px solid #f3f4f6;">
              <h1 style="margin: 0; font-size: 22px; font-weight: 600; color: #111827;">Verify your email</h1>
            </td>
          </tr>
          <tr>
            <td style="padding: 32px 40px;">
              <p style="margin: 0 0 20px; font-size: 15px; line-height: 1.6; color: #4b5563;">Use the code below to verify your account. It expires in 15 minutes.</p>
              <table role="presentation" width="100%%" cellspacing="0" cellpadding="0">
                <tr>
                  <td align="center" style="padding: 16px 0 24px;">
                    <span style="display: inline-block; padding: 14px 28px; background-color: #615fff; color: #ffffff; font-size: 28px; font-weight: 700; letter-spacing: 6px; border-radius: 8px;">%s</span>
                  </td>
                </tr>
              </table>
              <p style="margin: 0; font-size: 14px; line-height: 1.5; color: #6b7280;">If you didn’t request this, you can safely ignore this email.</p>
            </td>
          </tr>
          <tr>
            <td style="padding: 20px 40px 32px; text-align: center; border-top: 1px solid #f3f4f6;">
              <p style="margin: 0; font-size: 12px; color: #9ca3af;">Money Mate</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, code)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{toEmail},
		Subject: "Verify your email",
		Html:    html,
	})
	return err
}
