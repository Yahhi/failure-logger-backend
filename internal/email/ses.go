package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/yourorg/failure-uploader/internal/logging"
)

// Sender handles email sending via SES
type Sender struct {
	client *ses.Client
	from   string
	to     string
}

// NewSender creates a new SES email sender
func NewSender(ctx context.Context, region, from, to string) (*Sender, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := ses.NewFromConfig(cfg)

	return &Sender{
		client: client,
		from:   from,
		to:     to,
	}, nil
}

// FailureNotification contains data for the failure notification email
type FailureNotification struct {
	FailureID   string
	Project     string
	Env         string
	Method      string
	URL         string
	AppVersion  string
	Platform    string
	EnvelopeURL string
}

// SendFailureNotification sends an email notification about a completed failure upload
func (s *Sender) SendFailureNotification(ctx context.Context, notif FailureNotification) error {
	subject := fmt.Sprintf("[%s/%s] Failed Request Captured: %s", notif.Project, notif.Env, notif.FailureID)

	body := fmt.Sprintf(`A failed network request has been captured and uploaded.

Failure ID: %s
Project: %s
Environment: %s

Request Details:
- Method: %s
- URL: %s

Client:
- App Version: %s
- Platform: %s

Download envelope:
%s

---
This is an automated notification from failure-uploader.
`,
		notif.FailureID,
		notif.Project,
		notif.Env,
		notif.Method,
		notif.URL,
		notif.AppVersion,
		notif.Platform,
		notif.EnvelopeURL,
	)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
.container { max-width: 600px; margin: 0 auto; padding: 20px; }
.header { background: #f44336; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
.content { background: #f9f9f9; padding: 20px; border-radius: 0 0 8px 8px; }
.field { margin-bottom: 10px; }
.label { font-weight: bold; color: #666; }
.value { color: #333; }
.button { display: inline-block; background: #2196F3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin-top: 15px; }
.footer { margin-top: 20px; font-size: 12px; color: #999; }
</style></head>
<body>
<div class="container">
<div class="header">
<h2 style="margin:0;">Failed Request Captured</h2>
<p style="margin:5px 0 0 0;">%s / %s</p>
</div>
<div class="content">
<div class="field"><span class="label">Failure ID:</span> <span class="value">%s</span></div>
<div class="field"><span class="label">Project:</span> <span class="value">%s</span></div>
<div class="field"><span class="label">Environment:</span> <span class="value">%s</span></div>
<h3>Request Details</h3>
<div class="field"><span class="label">Method:</span> <span class="value">%s</span></div>
<div class="field"><span class="label">URL:</span> <span class="value">%s</span></div>
<h3>Client</h3>
<div class="field"><span class="label">App Version:</span> <span class="value">%s</span></div>
<div class="field"><span class="label">Platform:</span> <span class="value">%s</span></div>
<a href="%s" class="button">Download Envelope</a>
</div>
<div class="footer">This is an automated notification from failure-uploader.</div>
</div>
</body>
</html>`,
		notif.Project, notif.Env,
		notif.FailureID,
		notif.Project,
		notif.Env,
		notif.Method,
		notif.URL,
		notif.AppVersion,
		notif.Platform,
		notif.EnvelopeURL,
	)

	input := &ses.SendEmailInput{
		Source: aws.String(s.from),
		Destination: &types.Destination{
			ToAddresses: []string{s.to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data:    aws.String(body),
					Charset: aws.String("UTF-8"),
				},
				Html: &types.Content{
					Data:    aws.String(htmlBody),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		logging.Error().Err(err).Str("failureId", notif.FailureID).Msg("failed to send email notification")
		return err
	}

	logging.Info().Str("failureId", notif.FailureID).Str("to", s.to).Msg("email notification sent")
	return nil
}
