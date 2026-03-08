# claude-commit-msg-gen

[Claude Code](https://claude.ai/code) と [Lefthook](https://lefthook.dev/) を組み合わせ、`git commit` 実行時にステージングされた差分を自動解析してコミットメッセージを生成するサンプルリポジトリです。

生成されるコミットメッセージは [Conventional Commits](https://www.conventionalcommits.org/) 形式（`type(scope): subject`）に準拠します。

## 動作の仕組み

```
git commit
  └─ lefthook: prepare-commit-msg フック
       └─ scripts/auto-commit-msg.sh
            ├─ git diff --cached でステージング差分を取得
            ├─ scripts/commit-prompt.txt のプロンプトと組み合わせて Claude に送信
            └─ 生成されたメッセージをコミットメッセージファイルに書き込む
```

- Merge Commit（`git pull` 等によるマージ）は自動生成をスキップします
- `-m` オプションでメッセージを直接指定した場合もスキップします
- Claude の呼び出しに失敗した場合は `git commit` を止めず、空のメッセージで続行します

## 前提条件

| ツール | インストール方法 |
|---|---|
| [Claude Code](https://claude.ai/code) | `npm install -g @anthropic-ai/claude-code` |
| [Lefthook](https://lefthook.dev/) | 後述のインストール方法を参照 |

Claude Code のセットアップ（APIキーの設定等）が完了していることを確認してください。

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
git clone https://github.com/your-org/claude-commit-msg-gen.git
cd claude-commit-msg-gen
```

### 3. Lefthook フックのインストール

```sh
lefthook install
```

これで `.git/hooks/prepare-commit-msg` が設定され、次回の `git commit` から自動生成が有効になります。

### 既存プロジェクトへの導入（curl 一発セットアップ）

既存の Git リポジトリに本プロジェクトのファイルを取り込む場合は、以下のコマンドで必要なファイルを一括ダウンロードできます。

```sh
REPO_URL="https://raw.githubusercontent.com/your-org/claude-commit-msg-gen/main"

# スクリプトとプロンプトをダウンロード
mkdir -p scripts
curl -fsSL "$REPO_URL/scripts/auto-commit-msg.sh" -o scripts/auto-commit-msg.sh
curl -fsSL "$REPO_URL/scripts/commit-prompt.txt"  -o scripts/commit-prompt.txt
curl -fsSL "$REPO_URL/lefthook.yml"               -o lefthook.yml
chmod +x scripts/auto-commit-msg.sh

# フックをインストール
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
        CLAUDE_MAX_TOKENS: "150"                 # 最大トークン数（プロンプト指示として渡される）
        COMMIT_PROMPT_FILE: scripts/commit-prompt.txt  # プロンプトファイルのパス
```

### 環境変数による一時的な上書き

`lefthook.yml` を変更せずに、コミット単位で設定を変えることもできます。

```sh
# より高精度なモデルを使用する
CLAUDE_MODEL=claude-sonnet-4-6 git commit

# トークン数を増やす
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
├── lefthook.yml                  # Lefthook 設定ファイル
└── scripts/
    ├── auto-commit-msg.sh        # コミットメッセージ自動生成スクリプト
    └── commit-prompt.txt         # Claude へのプロンプトテンプレート
```

## トラブルシューティング

**`lefthook install` を実行しても自動生成されない**

`lefthook install` が実行されているか確認してください。

```sh
lefthook install
```

**`Can't find lefthook in PATH` と表示される**

Lefthook がインストールされているか確認してください。

```sh
lefthook --version
```

**Claude の呼び出しがスキップされる**

Claude Code が正しくインストールされ、認証済みであることを確認してください。

```sh
claude --version
claude  # インタラクティブセッションで認証状態を確認
```

**自動生成を一時的に無効にしたい**

```sh
LEFTHOOK=0 git commit
```
