# claude-commit-msg-gen

[Claude API](https://www.anthropic.com/) と [Lefthook](https://lefthook.dev/) を組み合わせ、`git commit` 時に Conventional Commits 形式のコミットメッセージを自動生成するツールです。

構造・設計の詳細は [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) を参照してください。

## セットアップ

### 1. Anthropic API キーを設定

```sh
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 2. Lefthook をインストール

```sh
brew install lefthook
pnpm add -g lefthook
```

### 3. バイナリをインストール

```sh
pnpm install -g @himenon/claude-commit-msg-gen
```

### 4. フックを有効化

```sh
lefthook install
```

以上で `git commit` 時にコミットメッセージが自動生成されます。

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

> `claude` CLI を使ったシェルスクリプトによる代替実装は [docs/shell-script-alternative.md](docs/shell-script-alternative.md) を参照してください。
