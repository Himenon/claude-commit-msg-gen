# Architecture

## 概要

`claude-commit-msg-gen` は Lefthook の `prepare-commit-msg` フックから Go バイナリを呼び出し、Anthropic API に git diff を送信してコミットメッセージを生成する。

```
git commit
  └─ lefthook: prepare-commit-msg
       └─ bin/claude-commit-msg-gen（ラッパー）
            └─ bin/claude-commit-msg-gen-{os}-{arch}（Go バイナリ）
                 ├─ git diff --cached
                 ├─ scripts/commit-prompt.txt + diff → Anthropic API
                 └─ 生成メッセージを COMMIT_EDITMSG に書き込む
```

## ファイル構成

```
.
├── .github/workflows/
│   ├── ci.yml                        # push 時にビルド検証する CI
│   └── release.yml                   # v* タグ push → npm publish CI
├── docs/
│   ├── ARCHITECTURE.md               # 本ドキュメント
│   └── shell-script-alternative.md  # シェルスクリプト版の代替実装
├── go/
│   ├── main.go                       # Anthropic API を直接呼び出す実装
│   └── go.mod                        # Go 1.26、外部依存なし
├── scripts/
│   ├── auto-commit-msg.sh            # シェルスクリプト版（代替案）
│   ├── build.sh                      # クロスコンパイルスクリプト
│   └── commit-prompt.txt             # プロンプトテンプレート（変更可能）
├── bin/                              # ビルド済みバイナリ（.gitignore 対象）
│   ├── claude-commit-msg-gen         # OS/ARCH 自動判定ラッパースクリプト
│   ├── claude-commit-msg-gen-darwin-arm64
│   ├── claude-commit-msg-gen-darwin-amd64
│   ├── claude-commit-msg-gen-linux-amd64
│   └── claude-commit-msg-gen-linux-arm64
├── lefthook.yml                      # prepare-commit-msg フック設定
└── package.json                      # @himenon/claude-commit-msg-gen
```

## 主要コンポーネント

### `go/main.go`

Anthropic API（`POST /v1/messages`）を標準ライブラリのみで直接呼び出す。外部依存なし。

**引数:**

| 位置 | 内容                                                       |
| ---- | ---------------------------------------------------------- |
| `$1` | コミットメッセージファイルのパス（`COMMIT_EDITMSG`）       |
| `$2` | コミットソース（`merge` / `message` のとき処理をスキップ） |

**環境変数:**

| 変数名               | デフォルト値                | 説明                                                                 |
| -------------------- | --------------------------- | -------------------------------------------------------------------- |
| `ANTHROPIC_API_KEY`  | —                           | Anthropic API キー（必須）。未設定時は `ANTHROPIC_AUTH_TOKEN` を参照 |
| `CLAUDE_MODEL`       | `claude-haiku-4-5-20251001` | 使用モデル                                                           |
| `CLAUDE_MAX_TOKENS`  | `150`                       | Anthropic API の `max_tokens` に直接渡される                         |
| `COMMIT_PROMPT_FILE` | `scripts/commit-prompt.txt` | プロンプトファイルのパス（相対パスはリポジトリルート基準）           |
| `ANTHROPIC_BASE_URL` | `https://api.anthropic.com` | API エンドポイント（プロキシ利用時に変更）                           |

**エラーハンドリング方針:** 全エラーで `exit 0`。API 障害・設定ミスがあっても `git commit` を止めない。

### `scripts/commit-prompt.txt`

Claude API に渡すプロンプトのテンプレート。このファイルを編集することでプロジェクト固有のルールを追加できる。

### `bin/claude-commit-msg-gen`（ラッパー）

`scripts/build.sh` が生成するシェルスクリプト。`uname` で OS / ARCH を判定し、対応するプラットフォーム固有バイナリに `exec` する。npm の `bin` エントリはこのファイルを参照する。

### `lefthook.yml`

`prepare-commit-msg` フックに `claude-commit-msg-gen {1} {2}` を登録。`fail_text` により、バイナリが見つからない場合も `git commit` は続行される。

## npm 配布の仕組み

```
pnpm run build
  └─ scripts/build.sh
       ├─ go build × 4プラットフォーム → bin/claude-commit-msg-gen-{os}-{arch}
       └─ bin/claude-commit-msg-gen（ラッパー）を生成

pnpm publish（または git tag v* push）
  └─ bin/ と scripts/commit-prompt.txt を同梱して npm へ公開
     （bin/ は .gitignore 対象だが package.json の files に含まれるため配布される）
```

## GitHub Actions

`.github/workflows/ci.yml` — 全ブランチの push と pull_request でビルド検証。

`.github/workflows/release.yml` — `v*` タグ push 時に4プラットフォームを並列ビルドし npm publish。全 action は SHA pinning 済み。
