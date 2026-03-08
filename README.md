# claude-commit-msg-gen

[Claude API](https://www.anthropic.com/) と [Lefthook](https://lefthook.dev/) を組み合わせ、`git commit` 実行時にステージングされた差分を自動解析してコミットメッセージを生成するツールです。

生成されるコミットメッセージは [Conventional Commits](https://www.conventionalcommits.org/) 形式（`type(scope): subject`）に準拠します。

## 動作の仕組み

```
git commit
  └─ lefthook: prepare-commit-msg フック
       └─ claude-commit-msg-gen（バイナリ または シェルスクリプト）
            ├─ git diff --cached でステージング差分を取得
            ├─ scripts/commit-prompt.txt のプロンプトと組み合わせて Claude API に送信
            └─ 生成されたメッセージをコミットメッセージファイルに書き込む
```

- Merge Commit（`git pull` 等によるマージ）は自動生成をスキップします
- `-m` オプションでメッセージを直接指定した場合もスキップします
- Claude API の呼び出しに失敗した場合は `git commit` を止めず続行します

---

## 方法 A: Go バイナリ（推奨）

Go でビルドされたバイナリを npm / pnpm 経由で配布します。`max_tokens` を API パラメータとして厳密に制御できます。

### 前提条件

- [Lefthook](https://lefthook.dev/)
- Anthropic API キー

シェルの設定ファイル（`~/.zshrc` 等）に API キーを追加してください。

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

### インストール

**グローバルインストール**

```sh
# pnpm
pnpm install -g @himenon/claude-commit-msg-gen

# npm
npm install -g @himenon/claude-commit-msg-gen
```

**プロジェクトローカルインストール**

```sh
pnpm add -D @himenon/claude-commit-msg-gen
```

インストール後、`lefthook.yml` の `run` を以下のように設定してください。

```yaml
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      run: claude-commit-msg-gen {1} {2}
      env:
        CLAUDE_MODEL: claude-haiku-4-5-20251001
        CLAUDE_MAX_TOKENS: "150"
        COMMIT_PROMPT_FILE: scripts/commit-prompt.txt
      fail_text: "コミットメッセージの自動生成をスキップしました"
```

### ソースからビルドする場合

```sh
git clone https://github.com/Himenon/claude-commit-msg-gen.git
cd claude-commit-msg-gen
pnpm run build   # bin/ 以下にプラットフォーム別バイナリを生成
lefthook install
```

---

## 方法 B: シェルスクリプト（代替案）

Go や npm 環境がない場合の代替手段です。`claude` CLI を利用してコミットメッセージを生成します。`claude` CLI がインストール済みであれば API キーの設定は不要です。

> **注意:** シェルスクリプト版は `max_tokens` をプロンプトへの指示として渡すため、バイナリ版と比べてトークン数の制御が不確実です。

### 前提条件

- [Lefthook](https://lefthook.dev/)
- [Claude Code](https://claude.ai/code)（`claude` CLI）

### セットアップ

```sh
git clone https://github.com/Himenon/claude-commit-msg-gen.git
cd claude-commit-msg-gen
lefthook install
```

`lefthook.yml` の `run` を以下のように設定してください。

```yaml
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      run: sh "$(git rev-parse --show-toplevel)/scripts/auto-commit-msg.sh" {1} {2}
      env:
        CLAUDE_MODEL: claude-haiku-4-5-20251001
        CLAUDE_MAX_TOKENS: "150"
        COMMIT_PROMPT_FILE: scripts/commit-prompt.txt
      fail_text: "コミットメッセージの自動生成をスキップしました"
```

---

## 方法 A / B の比較

| 項目 | Go バイナリ（方法 A） | シェルスクリプト（方法 B） |
|---|---|---|
| API 呼び出し | Anthropic API を直接呼び出す | `claude` CLI 経由 |
| `max_tokens` の制御 | API パラメータとして厳密に指定 | プロンプトへの指示（不確実） |
| API キー | `ANTHROPIC_API_KEY` が必要 | 不要（`claude` CLI が保持） |
| 依存関係 | pnpm / npm のみ | `claude` CLI |
| 実行速度 | 高速（ネイティブバイナリ） | `claude` CLI 起動コストあり |

---

## 設定

`lefthook.yml` の `env` セクションで動作を変更できます。

```yaml
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      env:
        CLAUDE_MODEL: claude-haiku-4-5-20251001  # 使用モデル
        CLAUDE_MAX_TOKENS: "150"                 # 最大トークン数
        COMMIT_PROMPT_FILE: scripts/commit-prompt.txt  # プロンプトファイルのパス（プロジェクトルートからの相対パス）
```

### 環境変数による一時的な上書き

`lefthook.yml` を変更せずに、コミット単位で設定を変えることもできます。

```sh
# より高精度なモデルを使用する
CLAUDE_MODEL=claude-sonnet-4-6 git commit

# 生成トークン数を増やす
CLAUDE_MAX_TOKENS=300 git commit

# 別のプロンプトファイルを指定する
COMMIT_PROMPT_FILE=scripts/my-prompt.txt git commit
```

### 利用可能なモデル

| モデルID | 特徴 |
|---|---|
| `claude-haiku-4-5-20251001` | **デフォルト**。高速・低コスト |
| `claude-sonnet-4-6` | バランス型。精度が求められる場合に使用 |
| `claude-opus-4-6` | 最高精度。複雑な差分の解析に使用 |

### プロンプトのカスタマイズ

`scripts/commit-prompt.txt` を編集することで、コミットメッセージの生成ルールを変更できます。

プロジェクト固有のルールを追加する例：

```
# scripts/commit-prompt.txt に追記

追加ルール:
- チケット番号が関連する場合は subject の末尾に [#123] 形式で付記すること
- マイグレーションファイルの変更は必ず db(migration) を scope に使うこと
```

## Lefthook のインストール

```sh
# Homebrew（macOS / Linux）
brew install lefthook

# pnpm
pnpm add -g lefthook

# npm
npm install -g lefthook

# curl（Linux / macOS）
curl -fsSL https://raw.githubusercontent.com/evilmartians/lefthook/master/scripts/install.sh | bash
```

その他のインストール方法は [Lefthook 公式ドキュメント](https://lefthook.dev/installation/) を参照してください。

## ファイル構成

```
.
├── .github/
│   └── workflows/
│       └── release.yml               # タグ push 時に npm へ publish する CI
├── go/
│   ├── main.go                       # コミットメッセージ生成の Go 実装
│   └── go.mod                        # Go モジュール定義（Go 1.26）
├── scripts/
│   ├── auto-commit-msg.sh            # シェルスクリプト版実装（方法 B）
│   ├── build.sh                      # Go バイナリのビルドスクリプト
│   └── commit-prompt.txt             # Claude API へのプロンプトテンプレート
├── bin/                              # ビルド済みバイナリ（.gitignore 対象）
│   ├── claude-commit-msg-gen         # プラットフォーム自動判定ラッパー
│   ├── claude-commit-msg-gen-darwin-arm64
│   ├── claude-commit-msg-gen-darwin-amd64
│   ├── claude-commit-msg-gen-linux-amd64
│   └── claude-commit-msg-gen-linux-arm64
├── lefthook.yml                      # Lefthook 設定ファイル
└── package.json                      # npm パッケージ定義（@himenon/claude-commit-msg-gen）
```

## npm への publish 手順

`v` プレフィックスのタグを push すると GitHub Actions が自動でビルド・publish します。

```sh
git tag v1.0.0
git push origin v1.0.0
```

リポジトリの Secrets に `NPM_TOKEN` を登録してください。

## トラブルシューティング

**`Binary not found` と表示される**

バイナリが未ビルドの可能性があります。

```sh
pnpm run build
```

**`ANTHROPIC_API_KEY が未設定` と表示される（方法 A）**

```sh
echo $ANTHROPIC_API_KEY
```

**自動生成を一時的に無効にしたい**

```sh
LEFTHOOK=0 git commit
```
