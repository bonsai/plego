# PRD: plego-agent

## なにこれ

GUI から plego の pipeline をポチポチ操作できるようにするやつ。

- Source / Filter / Output を Web 上で設定
- 新しいプラグインも GUI から AI 生成
- 実行ログをリアルタイム表示
- 生成物 (iCal/JSON) を Preview → ワンクリック公開

## 構成

```
Vercel (Next.js) ←→WS→ Agent (Go) → plego core
                              ↘ GitHub commit → Pages
```

## やること

| Phase | やること |
|-------|---------|
| 1 | Agent サーバー建てる + Next.js 生やす + Vercel→GHA→Agent 起動 |
| 2 | Pipeline の CRUD、実行・ログ、OAuth、Preview |
| 3 | プラグインをテンプレート+AI で自動生成、自動統合 |
| 4 | commit→Pages 公開、スケジュール実行、Fly.io 常駐 |

## 関連

- https://github.com/bonsai/plego
- https://bonsai.github.io/plego/
