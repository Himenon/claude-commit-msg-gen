# claude-commit-msg-gen

[Claude API](https://www.anthropic.com/) と [Lefthook](https://lefthook.dev/) を組み合わせ、`git commit` 時に Conventional Commits 形式のコミットメッセージを自動生成するツールです。

> `claude` CLI を使ったシェルスクリプトによる代替実装は [docs/shell-script-alternative.md](docs/shell-script-alternative.md) を参照してください。

## セットアップ

### 1. Anthropic API キーを設定

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 2. Lefthook をインストール

```sh
# Homebrew（macOS / Linux）
brew install lefthook

# pnpm
pnpm add -g lefthook

# npm
npm install -g lefthook
```

### 3. バイナリをインストール

```sh
# pnpm
pnpm install -g @himenon/claude-commit-msg-gen

# npm
npm install -g @himenon/claude-commit-msg-gen
```

### 4. フックを有効化

```sh
lefthook install
```

以上で `git commit` 時にコミットメッセージが自動生成されます。

---

## 設定

`lefthook.yml` の `env` セクションで動作を変更できます。

```yaml
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      run: claude-commit-msg-gen {1} {2}
      env:
        CLAUDE_MODEL: claude-haiku-4-5-20251001  # 使用モデル
        CLAUDE_MAX_TOKENS: "150"                 # Anthropic API の max_tokens に渡される値
        COMMIT_PROMPT_FILE: scripts/commit-prompt.txt  # プロンプトファイルのパス（プロジェクトルートからの相対パス）
      fail_text: "コミットメッセージの自動生成をスキップしました"
```

### 環境変数による一時的な上書き

```sh
CLAUDE_MODEL=claude-sonnet-4-6 git commit
CLAUDE_MAX_TOKENS=300 git commit
COMMIT_PROMPT_FILE=scripts/my-prompt.txt git commit
```

### 利用可能なモデル

| モデルID | 特徴 |
|---|---|
| `claude-haiku-4-5-20251001` | **デフォルト**。高速・低コスト |
| `claude-sonnet-4-6` | バランス型。精度が求められる場合に使用 |
| `claude-opus-4-6` | 最高精度。複雑な差分の解析に使用 |

### プロンプトのカスタマイズ

`scripts/commit-prompt.txt` を編集することで生成ルールを変更できます。

---

## 動作の仕組み

```
git commit
  └─ lefthook: prepare-commit-msg フック
       └─ claude-commit-msg-gen（Go バイナリ）
            ├─ git diff --cached でステージング差分を取得
            ├─ scripts/commit-prompt.txt のプロンプトと組み合わせて Claude API に送信
            └─ 生成されたメッセージをコミットメッセージファイルに書き込む
```

- Merge Commit（`git pull` 等によるマージ）は自動生成をスキップします
- `-m` オプションでメッセージを直接指定した場合もスキップします
- Claude API の呼び出しに失敗した場合は `git commit` を止めず続行します

## ファイル構成

```
.
├── .github/workflows/
│   ├── ci.yml                        # push 時にビルド検証する CI
│   └── release.yml                   # タグ push 時に npm へ publish する CI
├── docs/
│   └── shell-script-alternative.md  # シェルスクリプト版の代替実装
├── go/
│   ├── main.go                       # Go 実装（Anthropic API 直接呼び出し）
│   └── go.mod                        # Go 1.26、外部依存なし
├── scripts/
│   ├── auto-commit-msg.sh            # シェルスクリプト版（代替案）
│   ├── build.sh                      # クロスコンパイルスクリプト
│   └── commit-prompt.txt             # プロンプトテンプレート
├── bin/                              # ビルド済みバイナリ（.gitignore 対象）
├── lefthook.yml                      # Lefthook 設定ファイル
└── package.json                      # @himenon/claude-commit-msg-gen
```

## ソースからビルドする場合

```sh
git clone https://github.com/Himenon/claude-commit-msg-gen.git
cd claude-commit-msg-gen
pnpm run build
lefthook install
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

```sh
pnpm run build
```

**`ANTHROPIC_API_KEY が未設定` と表示される**

```sh
echo $ANTHROPIC_API_KEY
```

**自動生成を一時的に無効にしたい**

```sh
LEFTHOOK=0 git commit
```
