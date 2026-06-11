# PRD: plego-agent

## 1. 概要

plego のパイプライン（Source / Filter / Output）を Web GUI から動的に拡張・制御するための
**エージェント基盤** と **Vercel フロントエンド** を提供する。

## 2. 背景・課題

- 現在の plego は YAML 編集 + Go コード追加が必要で、非エンジニアが pipeline を拡張できない
- プラグイン追加ごとに PR → review → merge → deploy のサイクルが遅い
- pipeline の状態（実行状況・エラー・生成物）をリアルタイムに確認できない
- 複数の pipeline を並行管理する手段がない

## 3. ゴール

- Vercel 上の Next.js GUI から pipeline の作成・編集・実行・監視ができる
- Agent が WebSocket 経由で Vercel からの指示を受け取り、plego を動的に拡張する
- プラグインの追加が GUI 操作だけで完結する（コード生成 AI 支援）
- 生成物（iCal / JSON / メール等）のプレビューと公開状態を GUI で確認できる

## 4. アーキテクチャ概要

```
┌─────────────────┐     WebSocket      ┌──────────────────┐
│  Vercel (Next.js) │ ◄──────────────► │  plego-agent      │
│  - Pipeline GUI   │                   │  - API Server     │
│  - Plugin Builder │   GitHub API      │  - WS Handler     │
│  - Preview        │ ◄──────────────► │  - Plugin Gen     │
│  - Auth (OAuth)   │                   │  - Pipeline Exec  │
└─────────────────┘                   └────────┬─────────┘
                                               │
                                        ┌──────▼──────┐
                                        │  plego core  │
                                        │  (Source→    │
                                        │   Filter→    │
                                        │   Output)    │
                                        └─────────────┘
```

## 5. コンポーネント

### 5.1 plego-agent (Go)

| 機能 | 説明 |
|------|------|
| API Server | REST + WebSocket サーバー。Vercel からの接続を受付 |
| WS Handler | 双方向通信。Vercel → Agent に指示、 Agent → Vercel に状態通知 |
| Plugin Generator | テンプレート + AI で Source/Filter/Output の Go コードを生成 |
| Pipeline Executor | plego pipeline をプログラムから実行・制御 |
| Config Manager | recipe/ 以下の YAML を CRUD 操作 |
| State Watcher | SQLite state DB の変更を監視し Vercel に push |
| GitHub Deployer | GitHub API 経由で commit → push → Pages deploy を実行 |

### 5.2 Vercel Frontend (Next.js)

| 機能 | 説明 |
|------|------|
| Pipeline List | 登録済み pipeline の一覧・状態表示 |
| Pipeline Editor | Source/Filter/Output のビジュアル設定 |
| Plugin Marketplace | 既存プラグインの一覧・有効化 |
| Plugin Creator | フォーム入力 + AI で新プラグインを生成 |
| Execution Log | pipeline 実行のリアルタイムログ表示 |
| Preview Pane | 生成された iCal/JSON/メールのプレビュー |
| OAuth Settings | Gmail 等の認証情報管理 |

### 5.3 Agent 起動モード

```
GitHub Actions (dispatch):
  Vercel → GitHub API (workflow_dispatch) → Agent 起動 → WS接続確立

常駐サーバー (Fly.io / Railway):
  Agent が常時起動、WS 待受 → Vercel から直接 WS 接続
```

## 6. ユーザーフロー

1. User が Vercel GUI にログイン (GitHub OAuth)
2. GUI で pipeline 一覧を確認
3. "New Pipeline" で Source / Filter / Output を選択・設定
4. "Generate Plugin" で既存にないプラグインを AI 生成
5. "Run Pipeline" で Agent に実行指示 → WS でリアルタイムログ
6. 生成物を Preview で確認 → "Deploy" で Pages 公開
7. スケジュール実行の設定も GUI から可能

## 7. 非機能要件

- Agent 起動〜WS 接続確立: < 5秒
- プラグインコード生成: < 30秒
- WebSocket 再接続: 自動 (exponential backoff)
- 認証: GitHub OAuth (Vercel 側) + API Token (Agent 側)
- エラーハンドリング: 全 API にエラーレスポンス定義、WS は error イベント

## 8. フェーズ計画

### Phase 1: 基盤
- plego-agent API Server + WS ハンドラ
- Next.js プロジェクト作成、基本レイアウト
- Vercel → GitHub Actions → Agent 起動フロー

### Phase 2: Pipeline 管理
- Pipeline CRUD (GUI + API)
- Config Manager (recipe/ YAML 操作)
- Pipeline 実行・ログ表示

### Phase 3: Plugin 拡張
- Plugin Generator (テンプレートベース)
- AI 連携によるコード生成
- Plugin Marketplace 表示

### Phase 4: デプロイ・運用
- GitHub Deployer (commit → Pages)
- スケジュール管理
- 監視・アラート

## 9. 関連リンク

- GitHub: https://github.com/bonsai/plego
- Pages: https://bonsai.github.io/plego/
