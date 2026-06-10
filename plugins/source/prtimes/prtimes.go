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
	baseURL      = "https://prtimes.jp"
	searchFormat = "%s/topics/keywords/%s"
	crawlWait    = time.Second
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
		log.Printf("[prtimes] %q: %d results", kw, len(results))
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
	searchURL := fmt.Sprintf(searchFormat, baseURL, url.PathEscape(keyword))

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

// parseSearchPage finds all press-release links on a PR TIMES keyword page.
// Links match /main/html/rd/p/{digits}.{digits}.html
func parseSearchPage(r io.Reader) ([]core.Item, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var items []core.Item
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			if item, ok := extractFromAnchor(n); ok {
				items = append(items, item)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return items, nil
}

// extractFromAnchor builds an Item from a press-release <a> tag.
func extractFromAnchor(n *html.Node) (core.Item, bool) {
	href := attr(n, "href")
	if !isArticleHref(href) {
		return core.Item{}, false
	}
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}

	title := strings.TrimSpace(textContent(n))
	if title == "" {
		return core.Item{}, false
	}

	// Walk up to find a container with more context (date, summary).
	container := parentN(n, 4)
	summary := closestText(container, "p")
	dateStr := closestAttr(container, "time", "datetime")
	published := parseDate(dateStr)

	return core.Item{
		ID:          itemID(href),
		Title:       title,
		Body:        summary,
		URL:         href,
		PublishedAt: published,
	}, true
}

func isArticleHref(href string) bool {
	// e.g. /main/html/rd/p/000000030.000138935.html
	return strings.Contains(href, "/main/html/rd/p/")
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func parentN(n *html.Node, levels int) *html.Node {
	for i := 0; i < levels && n != nil; i++ {
		n = n.Parent
	}
	return n
}

func closestText(n *html.Node, tag string) string {
	if n == nil {
		return ""
	}
	var walk func(*html.Node) string
	walk = func(node *html.Node) string {
		if node.Type == html.ElementNode && node.Data == tag {
			if t := strings.TrimSpace(textContent(node)); t != "" {
				return t
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if t := walk(c); t != "" {
				return t
			}
		}
		return ""
	}
	return walk(n)
}

func closestAttr(n *html.Node, tag, key string) string {
	if n == nil {
		return ""
	}
	var walk func(*html.Node) string
	walk = func(node *html.Node) string {
		if node.Type == html.ElementNode && node.Data == tag {
			if v := attr(node, key); v != "" {
				return v
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if v := walk(c); v != "" {
				return v
			}
		}
		return ""
	}
	return walk(n)
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
	for _, layout := range []string{"2006-01-02T15:04:05Z07:00", "2006-01-02", "2006年1月2日"} {
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
