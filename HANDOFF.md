# Plego 引き継ぎ文書

## プロジェクト概要

PR TIMES をソースに、飲食店・カフェ・ホテルのイベント（試食会・発表会等）を自動収集し、
iCal フィード配信およびメール通知を行う Go 製パイプラインツール。

## リポジトリ

- GitHub: https://github.com/bonsai/plego
- GitHub Pages (iCal): https://bonsai.github.io/plego/calendar.ics
- module path: `github.com/dance/plego`

## バージョン構成

| バージョン | 設定ファイル | 機能 | Secrets |
|---------|------------|------|--------|
| **v0** (現在アクティブ) | `recipe/plego.v0.yaml` | iCal + JSON | 不要 |
| v1 | `recipe/plego.v1.yaml` | iCal + 日次メール | `GMAIL_USER` `GMAIL_APP_PASSWORD` |
| v2 | `recipe/plego.v2.yaml` | iCal + メール + Google Calendar API | + `GOOGLE_CALENDAR_CREDENTIALS` `GOOGLE_CALENDAR_ID` |

`recipe/plego.yaml` は現在 v0 と同内容。バージョンアップ時は `cp recipe/plego.vX.yaml recipe/plego.yaml` して commit。

## アーキテクチャ

```
main.go                   エントリポイント、プラグインをワイヤリング
core/plugin.go             Item 構造体、Source / Output / Flusher インターフェース
core/pipeline.go           Source → dedup → Outputs → Flush のフロー
config/config.go           YAML 設定から構造体へのマッピング
state/sqlite.go            SQLite で済み ID を永続化（重複配信防止）

plugins/source/
  prtimes/prtimes.go       PUB: PR TIMES キーワード検索 + 業種フィルタ
  filesystem/filesystem.go PUB: ファイルシステム（デバッグ・テスト用）

plugins/output/
  ical/ical.go             SUB: docs/calendar.ics 生成 → GitHub Pages
  smtp/smtp.go             SUB: SMTP 日次ダイジェストメール
  gmail/gmail.go           SUB: Gmail API OAuth2（レガシー・ v2 向け）
```

## プラグイン追加の手順

### 新しいソース（PUB）
1. `plugins/source/{name}/{name}.go` を作成
2. `core.Source` インターフェースを実装し、`Name()` と `Items()` を定義
3. `main.go` の `buildSource()` に case 追加
4. `config/config.go` の `SourceConfig` に必要なフィールドを追加

### 新しい出力（SUB）
1. `plugins/output/{name}/{name}.go` を作成
2. `core.Output` を実装。バッチ送信なら `core.Flusher` も実装
3. `main.go` の `buildOutput()` に case 追加

## GitHub Actions

| ワークフロー | スケジュール | 用途 |
|------------|----------|------|
| `digest-v0.yml` | 毎日 JST 07:00 | iCal + JSON (アクティブ) |
| `digest-v1.yml` | workflow_dispatch | v1 テスト・手動起動 |

v1 を定期実行にする時は `digest-v1.yml` の `on:` を v0 と同様に変更し、`digest-v0.yml` の schedule を削除する。

## 必要な GitHub Secrets

### v1 に必要
1. `GMAIL_USER` — 送信元 Gmail アドレス
2. `GMAIL_APP_PASSWORD` — Gmail アプリパスワード
   - 取得: Google アカウント → セキュリティ → 2段階認証有効化 → 「アプリパスワード」

### v2 追加分
3. `GOOGLE_CALENDAR_CREDENTIALS` — サービスアカウント JSON を base64 エンコードしたもの
4. `GOOGLE_CALENDAR_ID` — 登録先カレンダー ID（個人の場合 `primary`）

## iCal 購読方法（ユーザー向け）

1. Google カレンダーを開く
2. 左サイドバー「他のカレンダー」の `+` → 「URL」
3. `https://bonsai.github.io/plego/calendar.ics` を入力
4. 以後自動更新（最大 24h 遅延）

## 既知の課題（TODO）

- `prtimes.go` の HTML パーサーは PR TIMES の実際の HTML 構造で検証要。`<article>` タグがない場合は `findLink` の URL パターンの変更が必要。
- `go mod tidy` で `golang.org/x/net` が直接依存に追加される。CI が自動実行する。
- v2 `gcal` プラグインは未実装。実装時は `plugins/output/gcal/gcal.go` を作成し `buildOutput` に case 追加。
- キーワード・業種フィルタの上書き: `recipe/plego.yaml` を編集するか、将来的に設定 UI を追加。

## ローカル実行

```bash
# 初回のみ
 go mod tidy

# v0 テスト
go run . -config recipe/plego.v0.yaml

# 生成された iCal
cat docs/calendar.ics
```
