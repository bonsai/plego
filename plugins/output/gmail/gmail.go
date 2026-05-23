package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dance/plego/auth"
	"github.com/dance/plego/core"
	"golang.org/x/oauth2"
)

var gmailScopes = []string{
	"https://www.googleapis.com/auth/gmail.compose",
}

type Output struct {
	To              string
	CredentialsFile string
	TokenFile       string

	// oauth is initialized during InitAuth
	oauth *auth.OAuthHandler
	cfg   *oauth2.Config
}

func (o *Output) Name() string { return "gmail" }

func (o *Output) InitAuth(ctx context.Context) error {
	handler, err := auth.NewGoogleOAuth(o.CredentialsFile, o.TokenFile, o.Name(), gmailScopes)
	if err != nil {
		return err
	}
	o.oauth = handler
	o.cfg = handler.Config
	_, err = handler.Token(ctx)
	return err
}

func (o *Output) Publish(ctx context.Context, item core.Item) error {
	client, err := o.httpClient(ctx)
	if err != nil {
		return err
	}

	raw := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", o.To, item.Title, item.Body)
	encoded := base64.URLEncoding.EncodeToString([]byte(raw))

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

func (o *Output) httpClient(ctx context.Context) (*http.Client, error) {
	tok, err := o.oauth.Token(ctx)
	if err != nil {
		return nil, err
	}
	return o.cfg.Client(ctx, tok), nil
}

var _ core.Output = (*Output)(nil)
var _ core.Authorizer = (*Output)(nil)
