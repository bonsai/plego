// Package yotei keeps only items that announce a scheduled event.
// It matches keywords that indicate a specific date/venue is planned.
package yotei

import (
	"strings"

	"github.com/dance/plego/core"
)

// keywords that signal an upcoming scheduled event.
var keywords = []string{
	"開催",
	"予定",
	"実施",
	"来場",
	"参加",
	"イベント",
	"試食会",
	"発表会",
	"披露",
	"お披露目",
}

type Filter struct{}

func (f *Filter) Name() string { return "yotei" }

func (f *Filter) Keep(item core.Item) bool {
	// Always keep items where the scraper already found an explicit event date.
	if item.EventAt != nil {
		return true
	}
	text := item.Title + " " + item.Body
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

var _ core.Filter = (*Filter)(nil)
