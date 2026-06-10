.PHONY: build run run-v0 run-v1 test tidy lint clean

build:
	go build ./...

# デフォルト実行 (plego.yaml = v0)
run:
	go run ./cmd/plego -config plego.yaml

run-v0:
	go run ./cmd/plego -config plego.v0.yaml

run-v1:
	GMAIL_USER=$(GMAIL_USER) GMAIL_APP_PASSWORD=$(GMAIL_APP_PASSWORD) \
		go run ./cmd/plego -config plego.v1.yaml

test:
	go test ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

clean:
	rm -f docs/calendar.ics docs/feed.json

# 手動で iCal と JSON を確認
show:
	@echo "--- calendar.ics ---"
	@cat docs/calendar.ics 2>/dev/null || echo "(not yet generated)"
	@echo "--- feed.json ---"
	@cat docs/feed.json 2>/dev/null || echo "(not yet generated)"
