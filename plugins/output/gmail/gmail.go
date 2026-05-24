package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"unicode"

	"github.com/dance/plego/auth"
	"github.com/dance/plego/core"
	"golang.org/x/oauth2"
)

var gmailScopes = []string{
	"https://www.googleapis.com/auth/gmail.compose",
}

type Publish struct {
	To              string
	CredentialsFile string
	TokenFile       string

	oauth *auth.OAuthHandler
	cfg   *oauth2.Config
}

func (p *Publish) Name() string { return "Publish::Gmail" }

func (p *Publish) InitAuth(ctx context.Context) error {
	handler, err := auth.NewGoogleOAuth(p.CredentialsFile, p.TokenFile, p.Name(), gmailScopes)
	if err != nil {
		return err
	}
	p.oauth = handler
	p.cfg = handler.Config
	_, err = handler.Token(ctx)
	return err
}

func encodeSubject(subject string) string {
	for _, r := range subject {
		if r > unicode.MaxASCII {
			return mime.BEncoding.Encode("utf-8", subject)
		}
	}
	return subject
}

func (p *Publish) Publish(ctx context.Context, entry *core.Entry) error {
	client, err := p.httpClient(ctx)
	if err != nil {
		return err
	}

	subject := encodeSubject(entry.Title)
	raw := fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		p.To, subject, entry.Body)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(raw))

	body, _ := json.Marshal(map[string]any{
		"message": map[string]string{"raw": encoded},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST",
		"https://gmail.googleapis.com/gmail/v1/users/me/drafts",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gmail API %d: %s", resp.StatusCode, b)
	}
	return nil
}

func (p *Publish) httpClient(ctx context.Context) (*http.Client, error) {
	tok, err := p.oauth.Token(ctx)
	if err != nil {
		return nil, err
	}
	return p.cfg.Client(ctx, tok), nil
}

var _ core.Publish = (*Publish)(nil)
var _ core.Authorizer = (*Publish)(nil)
