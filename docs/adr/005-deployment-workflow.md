# ADR-005: Deployment & Workflow Architecture

## Status

Proposed

## Context

Vercel からの指示で plego-agent を起動し、pipeline を実行して結果を GitHub Pages に deploy する。
また、スケジュール実行と手動実行の両方をサポートする。

## Decision

### 1. トリガー経路

```
手動 (GUI):
  Vercel GUI → REST → GitHub API (dispatch) → GHA → Agent 起動
              ↘ WebSocket 直接 ← ← ← ← ← ← ← ← ← ← ↑

スケジュール:
  GitHub Actions schedule → Agent 起動 → pipeline 実行

常時稼働 (Phase 2+):
  Fly.io / Railway に Agent 常駐 → Vercel から直接 WS 接続
```

### 2. GitHub Actions Workflow 設計

```yaml
# .github/workflows/agent.yml
name: plego-agent

on:
  workflow_dispatch:
    inputs:
      pipeline:
        description: "Pipeline config file (recipe/*.yaml)"
        required: true
        default: recipe/plego.v0.yaml
      action:
        description: "Action: run | generate-plugin | deploy"
        required: true
        default: run
      ws_port:
        description: "WebSocket port for live log streaming"
        required: false
        default: "0"  # 0 = 不使用

jobs:
  agent:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Start Agent WebSocket
        if: inputs.ws_port != '0'
        run: |
          go run ./cmd/agent \
            --port ${{ inputs.ws_port }} \
            --pipeline ${{ inputs.pipeline }} \
            --token ${{ secrets.AGENT_WS_TOKEN }}
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}

      - name: Run Pipeline
        if: inputs.action == 'run'
        run: go run . -config ${{ inputs.pipeline }}

      - name: Generate Plugin
        if: inputs.action == 'generate-plugin'
        run: go run ./cmd/agent generate-plugin --config ${{ inputs.pipeline }}

      - name: Commit & Push
        run: |
          git config user.name "plego-bot"
          git add -A
          git diff --staged --quiet || git commit -m "chore: agent update [skip ci]"
          git push
```

### 3. WebSocket ログストリーム設計

Agent が pipeline 実行中に Vercel にリアルタイムログを送信:

```json
// Agent → Vercel
{ "type": "log",   "level": "info",  "message": ">> starting pipeline...", "timestamp": "..." }
{ "type": "log",   "level": "error", "message": "source: connection failed", "timestamp": "..." }
{ "type": "state", "phase": "source", "status": "running", "progress": "3/5" }
{ "type": "state", "phase": "output", "status": "done", "result": "calendar.ics" }
{ "type": "done",  "status": "success", "duration": "12.3s" }

// Vercel → Agent
{ "type": "cancel", "reason": "user cancelled" }
{ "type": "ping" }
```

### 4. GitHub Pages デプロイフロー

```
pipeline 実行 → docs/ に生成物書き込み
         ↓
git add docs/ recipe/ state/
         ↓
git commit -m "chore: agent deploy [skip ci]"
         ↓
git push → GitHub Pages 自動デプロイ (約1-2分)
         ↓
Vercel が Pages URL を表示・プレビュー
```

## Consequences

Positive:
- スケジュール・手動・GUI の全経路で同じ pipeline を実行可能
- WebSocket でリアルタイムフィードバック
- GitHub Pages へのデプロイが既存インフラで完結

Negative:
- GitHub Actions の実行時間制限 (6h) に注意
- WebSocket は GHA 内プロセスとして動作するため、GHA 終了とともに切断
- 常時稼働モードには別途 Fly.io / Railway 等のサーバーが必要

## Alternatives considered

1. **Vercel から直接 SSH で Agent 起動**: セキュリティリスク大
2. **Agent を Docker コンテナとして Fly.io 常駐**: Phase 2 で対応
3. **WebSocket の代わりに SSE (Server-Sent Events)**: 片方向のみで双方向不可
