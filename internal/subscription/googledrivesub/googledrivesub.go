package googledrivesub

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/dance/plego/auth"
	"github.com/dance/plego/core"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var driveScopes = []string{"https://www.googleapis.com/auth/drive.file"}

type Subscription struct {
	fileId       string
	credentials  string
	oauth        *auth.OAuthHandler
	cfg          *oauth2.Config
}

func New(fileId, credentials string) (*Subscription, error) {
	if fileId == "" {
		return nil, fmt.Errorf("Subscription::GoogleDrive: fileId is required")
	}
	if credentials == "" {
		return nil, fmt.Errorf("Subscription::GoogleDrive: credentials is required")
	}
	return &Subscription{
		fileId:      fileId,
		credentials: credentials,
	}, nil
}

func (s *Subscription) Name() string { return "Subscription::GoogleDrive" }

func (s *Subscription) InitAuth(ctx context.Context) error {
	handler, err := auth.NewGoogleOAuth(s.credentials, "", s.Name(), driveScopes)
	if err != nil {
		return fmt.Errorf("init auth: %w", err)
	}
	s.oauth = handler
	s.cfg = handler.Config

	tok, err := handler.Token(ctx)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	_ = tok
	return nil
}

func (s *Subscription) Fetch(ctx context.Context) ([]*core.Feed, error) {
	client := s.oauth.Config.Client(ctx, nil)
	tok, err := s.oauth.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	client = s.cfg.Client(ctx, tok)

	svc, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("create drive service: %w", err)
	}

	file, err := svc.Files.Get(s.fileId).Fields("id, name, mimeType, modifiedTime").Do()
	if err != nil {
		return nil, fmt.Errorf("get file info: %w", err)
	}

	log.Printf("[%s] downloading: %s (mime: %s)", s.Name(), file.Name, file.MimeType)

	resp, err := svc.Files.Get(s.fileId).Download()
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	guid := fmt.Sprintf("googledrive:%s:%s", s.fileId, file.ModifiedTime)

	entries := []*core.Entry{{
		GUID:  guid,
		Title: file.Name,
		Body:  string(body),
		URL:   fmt.Sprintf("https://drive.google.com/file/d/%s", s.fileId),
	}}

	return []*core.Feed{{
		GUID:    guid,
		Title:   "GoogleDrive: " + file.Name,
		Content: string(body),
		Entries: entries,
	}}, nil
}

var _ core.Subscription = (*Subscription)(nil)
var _ core.Authorizer = (*Subscription)(nil)
var _ = time.Now
