# ADR-004: Plugin Code Generation

## Status

Proposed

## Context

ユーザーが GUI から新しい Source/Filter/Output プラグインを追加できるようにする。
Go コードを自動生成し、build テストをパスした上で pipeline に組み込む必要がある。

## Decision

### 1. 三段階生成パイプライン

```
Phase 1: メタデータ作成
  GUI 入力 → Plugin メタデータ (name, type, config fields, description)
  ↓
Phase 2: コード生成
  text/template + AI (OpenAI/Claude) で Go コード生成
  ↓
Phase 3: 検証・統合
  go build → main.go に case 追加 → yaml に module 追加 → commit
```

### 2. テンプレート構造

各プラグインタイプに Go text/template を用意:

```
plugins/generator/templates/
├── source.go.tmpl        # Source インターフェース実装
├── filter.go.tmpl        # Filter インターフェース実装
├── output.go.tmpl        # Output インターフェース実装
└── test.go.tmpl          # テストファイル
```

テンプレート例 (source.go.tmpl):
```go
package {{.PackageName}}

import (
    "context"
    "{{.ModulePath}}/core"
)

type Source struct {
    {{range .ConfigFields}}
    {{.Name}} {{.Type}} `yaml:"{{.YamlTag}}"`
    {{end}}
}

func (s *Source) Name() string { return "{{.PluginName}}" }

func (s *Source) Items(ctx context.Context) ([]*core.Item, error) {
    // TODO: AI がこの関数の中身を生成
    {{.AICode}}
}
```

### 3. AI 生成プロンプト

```go
// AI に渡すプロンプトテンプレート
const promptTemplate = `
You are generating a Go plugin for plego (an RSS pipeline tool).
The plugin type is {{.Type}} and its name is {{.Name}}.

Interface to implement:
{{.Interface}}

Config fields available:
{{.ConfigFields}}

User's requirement:
{{.UserDescription}}

Generate ONLY the body of the {{.MethodName}} method.
Return compilable Go code. Use standard library only unless specified.
`
```

### 4. 検証フロー

1. `go build ./plugins/source/{name}/` でコンパイルチェック
2. `go vet ./plugins/source/{name}/` で静的解析
3. 自動生成コードに `// @generated` マーカーを付与
4. 人間のレビューを推奨 (AI コードは信頼できない)

## Consequences

Positive:
- GUI からのプラグイン追加が現実的な工数に
- テンプレート + AI で品質の下限が保証される
- `// @generated` マーカーで自動生成コードを明確化

Negative:
- AI 生成コードの品質はプロンプト次第
- ビルドが通ってもロジックが正しいとは限らない
- AI API のコストが発生する

## Alternatives considered

1. **Pluggable アーキテクチャ (WASM)**: 動的ロードは可能だが、AI 生成が困難
2. **手動コード追加**: 現状。拡張性に乏しい
3. **プラグインを別リポジトリ管理**: バージョン管理が複雑に
