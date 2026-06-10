package prtimes

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/dance/plego/core"
)

const (
	baseURL   = "https://prtimes.jp"
	crawlWait = time.Second
)

type Source struct {
	Keywords   []string
	Industries []string
}

func (s *Source) Name() string { return "prtimes" }

func (s *Source) Items(ctx context.Context) ([]core.Item, error) {
	seen := map[string]bool{}
	var items []core.Item

	for i, kw := range s.Keywords {
		if i > 0 {
			select {
			case <-time.After(crawlWait):
			case <-ctx.Done():
				return items, ctx.Err()
			}
		}
		results, err := s.search(ctx, kw)
		if err != nil {
			log.Printf("[prtimes] search %q: %v", kw, err)
			continue
		}
		for _, item := range results {
			if seen[item.ID] {
				continue
			}
			if s.matchIndustry(item) {
				seen[item.ID] = true
				items = append(items, item)
			}
		}
	}
	return items, nil
}

func (s *Source) search(ctx context.Context, keyword string) ([]core.Item, error) {
	searchURL := fmt.Sprintf("%s/main/html/searchrlp/key/%s",
		baseURL, url.PathEscape(keyword))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Plego/1.0; +https://github.com/dance/plego)")
	req.Header.Set("Accept-Language", "ja,en;q=0.9")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %q", resp.StatusCode, keyword)
	}
	return parseSearchPage(resp.Body)
}

func (s *Source) matchIndustry(item core.Item) bool {
	if len(s.Industries) == 0 {
		return true
	}
	text := strings.ToLower(item.Title + " " + item.Body + " " + item.Location)
	for _, ind := range s.Industries {
		if strings.Contains(text, strings.ToLower(ind)) {
			return true
		}
	}
	return false
}

// parseSearchPage extracts articles from a PR TIMES HTML search result page.
func parseSearchPage(r io.Reader) ([]core.Item, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var items []core.Item
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "article" {
			if item, ok := extractItem(n); ok {
				items = append(items, item)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return items, nil
}

// extractItem pulls title/url/summary/date from an <article> node.
func extractItem(n *html.Node) (core.Item, bool) {
	link, title := findLink(n)
	if link == "" {
		return core.Item{}, false
	}
	// Resolve relative URLs.
	if strings.HasPrefix(link, "/") {
		link = baseURL + link
	}

	summary := findText(n, "p")
	dateStr := findAttr(n, "time", "datetime")
	published := parseDate(dateStr)

	id := itemID(link)
	return core.Item{
		ID:          id,
		Title:       title,
		Body:        summary,
		URL:         link,
		PublishedAt: published,
	}, true
}

func findLink(n *html.Node) (href, text string) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" && strings.Contains(a.Val, "/main/html/rd/") {
				return a.Val, strings.TrimSpace(textContent(n))
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if h, t := findLink(c); h != "" {
			return h, t
		}
	}
	return "", ""
}

func findText(n *html.Node, tag string) string {
	if n.Type == html.ElementNode && n.Data == tag {
		return strings.TrimSpace(textContent(n))
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := findText(c, tag); t != "" {
			return t
		}
	}
	return ""
}

func findAttr(n *html.Node, tag, attr string) string {
	if n.Type == html.ElementNode && n.Data == tag {
		for _, a := range n.Attr {
			if a.Key == attr {
				return a.Val
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if v := findAttr(c, tag, attr); v != "" {
			return v
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func parseDate(s string) time.Time {
	for _, layout := range []string{"2006-01-02T15:04:05Z07:00", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Now()
}

func itemID(rawURL string) string {
	h := sha256.Sum256([]byte(rawURL))
	return fmt.Sprintf("%x", h[:8])
}
