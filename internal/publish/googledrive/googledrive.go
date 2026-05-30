package googledrive

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dance/plego/auth"
	"github.com/dance/plego/core"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

var driveScopes = []string{"https://www.googleapis.com/auth/drive.file"}

type Publish struct {
	folderId        string
	fileNamePrefix  string
	timestampFormat string
	extension       string
	credentials     string
	oauth           *auth.OAuthHandler
	cfg             *oauth2.Config
}

func New(folderId, fileNamePrefix, timestampFormat, extension, credentials string) (*Publish, error) {
	if folderId == "" {
		return nil, fmt.Errorf("Publish::GoogleDrive: folderId is required")
	}
	if credentials == "" {
		return nil, fmt.Errorf("Publish::GoogleDrive: credentials is required")
	}
	if fileNamePrefix == "" {
		fileNamePrefix = "bookmark-"
	}
	if timestampFormat == "" {
		timestampFormat = "20060102-150405"
	}
	if extension == "" {
		extension = ".txt"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return &Publish{
		folderId:        folderId,
		fileNamePrefix:  fileNamePrefix,
		timestampFormat: timestampFormat,
		extension:       extension,
		credentials:     credentials,
	}, nil
}

func (p *Publish) Name() string { return "Publish::GoogleDrive" }

func (p *Publish) InitAuth(ctx context.Context) error {
	handler, err := auth.NewGoogleOAuth(p.credentials, "", p.Name(), driveScopes)
	if err != nil {
		return fmt.Errorf("init auth: %w", err)
	}
	p.oauth = handler
	p.cfg = handler.Config

	_, err = handler.Token(ctx)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}
	return nil
}

func (p *Publish) Publish(ctx context.Context, entry *core.Entry) error {
	tok, err := p.oauth.Token(ctx)
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	client := p.cfg.Client(ctx, tok)
	svc, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("create drive service: %w", err)
	}

	fileName := p.fileNamePrefix + time.Now().Format(p.timestampFormat) + p.extension

	f := &drive.File{
		Name:    fileName,
		MimeType: "text/plain",
		Parents: []string{p.folderId},
	}

	created, err := svc.Files.Create(f).
		Media(bytes.NewReader([]byte(entry.Body))).
		Fields("id, name, webViewLink").
		Do()
	if err != nil {
		return fmt.Errorf("create file in drive: %w", err)
	}

	log.Printf("[%s] created: %s (id: %s)", p.Name(), created.Name, created.Id)
	return nil
}

var _ core.Publish = (*Publish)(nil)
var _ core.Authorizer = (*Publish)(nil)
