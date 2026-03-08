# claude-commit-msg-gen

[Claude API](https://www.anthropic.com/) と [Lefthook](https://lefthook.dev/) を組み合わせ、`git commit` 実行時にステージングされた差分を自動解析してコミットメッセージを生成するサンプルリポジトリです。

生成されるコミットメッセージは [Conventional Commits](https://www.conventionalcommits.org/) 形式（`type(scope): subject`）に準拠します。

## 動作の仕組み

```
git commit
  └─ lefthook: prepare-commit-msg フック
       └─ bin/claude-commit-msg-gen（Goバイナリ）
            ├─ git diff --cached でステージング差分を取得
            ├─ scripts/commit-prompt.txt のプロンプトと組み合わせて Claude API に送信
            └─ 生成されたメッセージをコミットメッセージファイルに書き込む
```

- Merge Commit（`git pull` 等によるマージ）は自動生成をスキップします
- `-m` オプションでメッセージを直接指定した場合もスキップします
- Claude API の呼び出しに失敗した場合は `git commit` を止めず続行します

## 前提条件

| ツール | インストール方法 |
|---|---|
| [Lefthook](https://lefthook.dev/) | 後述のインストール方法を参照 |
| Go 1.23 以上 | [go.dev/dl](https://go.dev/dl/) またはローカルビルドには不要（pnpm でバイナリを取得） |

また、Anthropic API キーが必要です。以下の環境変数をシェルの設定ファイル（`~/.zshrc` 等）に追加してください。

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

## セットアップ

### 1. Lefthook のインストール

**Homebrew（macOS / Linux）**

```sh
brew install lefthook
```

**npm / pnpm / yarn**

```sh
# pnpm
pnpm add -g lefthook

# npm
npm install -g lefthook

# yarn
yarn global add lefthook
```

**curl で直接インストール（Linux / macOS）**

```sh
curl -fsSL https://raw.githubusercontent.com/evilmartians/lefthook/master/scripts/install.sh | bash
```

その他のインストール方法は [Lefthook 公式ドキュメント](https://lefthook.dev/installation/) を参照してください。

### 2. リポジトリのクローン

```sh
git clone https://github.com/Himenon/claude-commit-msg-gen.git
cd claude-commit-msg-gen
```

### 3. バイナリのセットアップ（どちらか一方）

**方法 A: pnpm でインストール（推奨）**

```sh
pnpm install
pnpm run build
```

**方法 B: Go でビルド**

```sh
go build -o bin/claude-commit-msg-gen-darwin-arm64 ./go
```

### 4. Lefthook フックのインストール

```sh
lefthook install
```

これで `.git/hooks/prepare-commit-msg` が設定され、次回の `git commit` から自動生成が有効になります。

### 既存プロジェクトへの導入（curl 一発セットアップ）

既存の Git リポジトリに本プロジェクトのファイルを取り込む場合は、以下のコマンドで必要なファイルを一括ダウンロードできます。

```sh
REPO_URL="https://raw.githubusercontent.com/Himenon/claude-commit-msg-gen/main"

# 設定ファイルとスクリプトをダウンロード
mkdir -p scripts
curl -fsSL "$REPO_URL/scripts/commit-prompt.txt" -o scripts/commit-prompt.txt
curl -fsSL "$REPO_URL/scripts/build.sh"          -o scripts/build.sh
curl -fsSL "$REPO_URL/lefthook.yml"              -o lefthook.yml
curl -fsSL "$REPO_URL/package.json"              -o package.json
chmod +x scripts/build.sh

# バイナリをビルドしてフックをインストール
pnpm run build
lefthook install
```

## 設定

`lefthook.yml` の `env` セクションで動作を変更できます。

```yaml
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      env:
        CLAUDE_MODEL: claude-haiku-4-5-20251001  # 使用モデル
        CLAUDE_MAX_TOKENS: "150"                 # Anthropic API の max_tokens に渡される値
        COMMIT_PROMPT_FILE: scripts/commit-prompt.txt  # プロンプトファイルのパス
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

## ファイル構成

```
.
├── go/
│   ├── main.go                       # コミットメッセージ生成の Go 実装
│   └── go.mod                        # Go モジュール定義
├── scripts/
│   ├── auto-commit-msg.sh            # シェルスクリプト版実装（参考用）
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

## トラブルシューティング

**`lefthook install` を実行しても自動生成されない**

`lefthook install` が実行されているか確認してください。

```sh
lefthook install
```

**`Binary not found` と表示される**

バイナリが未ビルドの可能性があります。ビルドを実行してください。

```sh
pnpm run build
```

**`ANTHROPIC_API_KEY が未設定` と表示される**

シェルの設定ファイルに API キーが設定されているか確認してください。

```sh
echo $ANTHROPIC_API_KEY
```

**自動生成を一時的に無効にしたい**

```sh
LEFTHOOK=0 git commit
```
