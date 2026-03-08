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

## API キーをシェル環境に書きたくない場合

`lefthook-local.yml` に API キーを記述する方法があります。このファイルは `.gitignore` 対象のため、リポジトリに混入しません。

```yaml
# lefthook-local.yml（.gitignore 対象）
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      env:
        ANTHROPIC_API_KEY: "sk-ant-..."
```

`lefthook-local.yml` は `lefthook.yml` の設定を上書き・マージします。記述した `env` のみが上書きされ、他の設定は `lefthook.yml` の値が引き続き使われます。

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
