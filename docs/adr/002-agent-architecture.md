# ADR-002: plego-agent アーキテクチャ

## Status

Proposed

## Context

plego の pipeline を動的に拡張・制御するには、従来の YAML + Go コード編集では不十分。
Web GUI からの操作と、リアルタイムな pipeline 制御を可能にする Agent 基盤が必要。

## Decision

### 1. Agent は Go で実装する

plego 本体が Go であり、同じ runtime 上で pipeline をライブラリとして直接実行できる。
別言語を挟むよりレイテンシが低く、型・インターフェースを再利用可能。

```go
// plego-agent 内部で plego をライブラリとして利用
import "github.com/dance/plego/core"
import "github.com/dance/plego/config"

func runPipeline(cfg *config.Config) error {
    pipeline := &core.Pipeline{
        Source:  src,
        Filters: filters,
        Outputs: outputs,
        State:   store,
    }
    return pipeline.Run(ctx)
}
```

### 2. 通信プロトコルは WebSocket を主とする

| 方式 | 用途 | 理由 |
|------|------|------|
| WebSocket | 双方向リアルタイム通信 | ログストリーム・状態通知を Vercel に push |
| REST | CRUD 操作・管理 API | シンプルな操作は REST、GET はキャッシュ可能 |
| GitHub API | workflow_dispatch | Vercel → GHA 経由で Agent を起動 |

### 3. Agent の起動方法は二段階

```
Vercel                        GitHub                       Agent (Fly.io/GHA)
  │                             │                             │
  ├── POST /api/agent/start ──► │                             │
  │                             ├── workflow_dispatch ──────► │
  │                             │                             ├── WS :8080 起動
  │◄──────────────────────── WS connect ──────────────────────┤
  │                             │                             │
  ├── WS: {type: "run", pipeline: "v0"} ───────────────────► │
  │                             │                             ├── plego実行
  │◄── WS: {type: "log", line: "..."} ───────────────────────┤
  │◄── WS: {type: "done", result: "ok"} ─────────────────────┤
```

Agent が常時稼働の場合は WS 接続を直接張り、起動ステップをスキップする。

### 4. プラグイン生成はテンプレート + AI の二段構え

```
テンプレート (Go text/template):
  固定構造 (インターフェース実装、config 構造体) はテンプレートから生成

AI 生成 (OpenAI API / Claude API):
  ロジック部分 (スクレイピング・フィルタ条件・フォーマット変換) を AI が生成
  ユーザーが自然言語で「◯◯からデータを取得して××でフィルタして△△に出力」と指定
```

## Consequences

Positive:
- plego の型・インターフェースを完全に再利用できる
- WebSocket により Vercel 側で即時フィードバックが可能
- テンプレート + AI でプラグイン追加の敷居が大幅に下がる

Negative:
- Agent の常時稼働には Fly.io / Railway 等のサーバーが必要
- AI コード生成には品質検証 (build test) が必須
- WebSocket の再接続・状態管理の実装が必要

## Alternatives considered

1. **Cloudflare Workers + Durable Objects**: WS は可能だが Go が使えない、plego の再利用不可
2. **Node.js で Agent 実装**: plego を child_process 実行 → 型共有不可・オーバーヘッド大
3. **GitHub Actions だけで完結**: 常時待受不可・WS 非対応・GUI との連携に GitHub API が必要で複雑
