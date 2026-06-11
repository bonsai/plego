# ADR-003: Vercel Frontend Stack

## Status

Proposed

## Context

plego-agent を操作する GUI を構築する。要件:
- モダンな SPA で pipeline 管理ができる
- Vercel にデプロイする
- GitHub OAuth で認証する
- WebSocket で Agent と通信する
- コンポーネント数が多くなるため型安全が必須

## Decision

| 層 | 技術 | 理由 |
|----|------|------|
| Framework | Next.js 15 (App Router) | Vercel ネイティブ、SSR/SSG/WS 対応 |
| Language | TypeScript | 型安全、開発効率 |
| UI | Tailwind CSS + shadcn/ui | 高速な開発、デザインシステム不要 |
| State | Zustand | 軽量、WS との相性が良い |
| WS Client | use-websocket | React hooks ネイティブ、自動再接続 |
| Auth | NextAuth.js (GitHub Provider) | Vercel との親和性、OAuth 標準対応 |
| API Client | fetch + React Query | SWR 管理、キャッシュ・リトライ |

### ディレクトリ構成

```
frontend/
├── src/
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx              # Pipeline一覧
│   │   ├── pipelines/
│   │   │   ├── [id]/
│   │   │   │   ├── page.tsx      # Pipeline詳細・設定
│   │   │   │   └── runs/
│   │   │   │       └── page.tsx  # 実行履歴
│   │   │   └── new/
│   │   │       └── page.tsx      # Pipeline新規作成
│   │   ├── plugins/
│   │   │   ├── page.tsx          # Plugin Marketplace
│   │   │   └── create/
│   │   │       └── page.tsx      # Plugin AI生成
│   │   └── settings/
│   │       └── page.tsx          # OAuth設定
│   ├── components/
│   │   ├── pipeline/
│   │   ├── plugin/
│   │   ├── ws/                   # WebSocket Provider
│   │   └── ui/                   # shadcn/ui コンポーネント
│   ├── lib/
│   │   ├── api.ts                # REST API client
│   │   ├── ws.ts                 # WebSocket client
│   │   └── types.ts              # 共有型定義
│   └── auth.ts                   # NextAuth 設定
├── package.json
├── next.config.ts
├── tailwind.config.ts
└── tsconfig.json
```

## Consequences

Positive:
- Vercel へのデプロイが一瞬 (git push)
- TypeScript + shadcn/ui で型安全かつ見た目が整う
- NextAuth.js で OAuth がカスタマイズ不要

Negative:
- Next.js の App Router は比較的新しく知見が少ない
- WebSocket の永続接続は Serverless 環境では維持できない（Agent 常駐サーバー必須）
- shadcn/ui のバージョンアップ追従が必要

## Alternatives considered

1. **Nuxt 3 (Vue)**: 同程度の機能だがエコシステムが狭い
2. **Remix**: 独自のデータローディングモデルに学習コスト
3. **純粋な React + Vite**: Vercel デプロイの最適化が弱い

## WebSocket 接続戦略

Vercel Serverless では WS を直接張れないため、以下の二段構え:

1. **Vercel → Agent 起動**: GitHub API (workflow_dispatch) または Agent 常駐サーバーに REST で起動指示
2. **Agent ↔ Browser**: Agent 起動後、ブラウザから Agent の WS エンドポイントに直接接続する

```typescript
// フロントエンド側 WS 接続
const startAgentAndConnect = async (pipelineId: string) => {
    // 1. Agent 起動 API を叩く
    await fetch('/api/agent/start', { method: 'POST', body: JSON.stringify({ pipelineId }) });

    // 2. Agent の WS エンドポイントに接続 (polling で起動完了を待つ)
    const ws = new WebSocket(`wss://plego-agent.fly.io/ws?token=${token}`);
    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        // ログ表示・状態更新
    };
};
```
