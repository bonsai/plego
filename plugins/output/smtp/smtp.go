package smtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	gosmtp "net/smtp"
	"os"
	"strings"
	"time"

	"github.com/dance/plego/core"
)

var digestTmpl = template.Must(template.New("digest").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><style>
body{font-family:sans-serif;max-width:700px;margin:0 auto;padding:16px;color:#333}
h1{font-size:18px;border-bottom:2px solid #e60033;padding-bottom:8px}
.item{margin:20px 0;padding:16px;border:1px solid #ddd;border-radius:4px}
.item h2{font-size:15px;margin:0 0 6px}
.item h2 a{color:#e60033;text-decoration:none}
.item p{margin:4px 0;font-size:13px;color:#555}
.item .meta{font-size:12px;color:#999;margin-top:8px}
</style></head><body>
<h1>飲食イベント新着 — {{.Date}}</h1>
{{range .Items}}
<div class="item">
  <h2><a href="{{.URL}}">{{.Title}}</a></h2>
  <p>{{.Body}}</p>
  <div class="meta">{{.PublishedAt.Format "2006-01-02"}}
  {{- if .Location}} &nbsp;|&nbsp; {{.Location}}{{end}}</div>
</div>
{{end}}
<hr><p style="font-size:11px;color:#aaa">Plego Digest — <a href="https://bonsai.github.io/plego/calendar.ics">iCal 購読</a></p>
</body></html>
`))

type Output struct {
	From     string
	Password string
	To       []string
	BCC      []string
	Subject  string

	pending []core.Item
}

func (o *Output) Name() string { return "smtp" }

func (o *Output) Publish(_ context.Context, item core.Item) error {
	o.pending = append(o.pending, item)
	return nil
}

func (o *Output) Flush(ctx context.Context) error {
	if len(o.pending) == 0 {
		log.Println("[smtp] no new items, skipping")
		return nil
	}

	from := envOrVal(o.From, "GMAIL_USER")
	password := envOrVal(o.Password, "GMAIL_APP_PASSWORD")
	if from == "" || password == "" {
		return fmt.Errorf("smtp: GMAIL_USER / GMAIL_APP_PASSWORD not set")
	}

	body, err := buildHTML(o.pending)
	if err != nil {
		return fmt.Errorf("smtp: build body: %w", err)
	}

	subject := o.Subject
	if subject == "" {
		subject = fmt.Sprintf("[Plego] 飲食イベント新着 %s (%d件)",
			time.Now().Format("2006-01-02"), len(o.pending))
	}

	allTo := append(append([]string{}, o.To...), o.BCC...)
	msg := buildMIME(from, o.To, o.BCC, subject, body)

	auth := gosmtp.PlainAuth("", from, password, "smtp.gmail.com")
	if err := gosmtp.SendMail("smtp.gmail.com:587", auth, from, allTo, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	log.Printf("[smtp] sent digest (%d items) to %s", len(o.pending), strings.Join(o.To, ","))
	return nil
}

func buildHTML(items []core.Item) (string, error) {
	var buf bytes.Buffer
	data := struct {
		Date  string
		Items []core.Item
	}{
		Date:  time.Now().Format("2006年01月02日"),
		Items: items,
	}
	if err := digestTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func buildMIME(from string, to, bcc []string, subject, htmlBody string) []byte {
	var b bytes.Buffer
	boundary := "plego-boundary-001"

	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString(fmt.Sprintf("From: Plego Digest <%s>\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	if len(bcc) > 0 {
		b.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(bcc, ", ")))
	}
	enc := base64.StdEncoding.EncodeToString([]byte(subject))
	b.WriteString(fmt.Sprintf("Subject: =?UTF-8?B?%s?=\r\n", enc))
	b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString("このメールは HTML 対応クライアントでご確認ください。\r\n")

	b.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	b.WriteString(htmlBody)
	b.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	return b.Bytes()
}

func envOrVal(val, envKey string) string {
	if val != "" {
		return val
	}
	return os.Getenv(envKey)
}

var _ core.Output = (*Output)(nil)
var _ core.Flusher = (*Output)(nil)
