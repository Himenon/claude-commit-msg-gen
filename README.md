# claude-commit-msg-gen

[Claude API](https://www.anthropic.com/) と [Lefthook](https://lefthook.dev/) を組み合わせ、`git commit` 時に Conventional Commits 形式のコミットメッセージを自動生成するツールです。

構造・設計の詳細は [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) を参照してください。

## セットアップ

```sh
# 1. 本ライブラリのインストール
# Without Node.js. バージョン更新の際もこのスクリプトを一度実行する
curl -fsSL https://raw.githubusercontent.com/Himenon/claude-commit-msg-gen/main/scripts/install.sh | sh
# With Node.js
pnpm install -g @himenon/claude-commit-msg-gen

# Installできているか確認する
claude-commit-msg-gen --version

## 2. lefthookのインストール
brew install lefthook
# または
pnpm add -g lefthook

# 3. Anthropic API キーを設定
# .bashrc, .zshrc, .fish/config.fish
export ANTHROPIC_API_KEY="sk-ant-..."
```

## lefthookのセットアップ

`lefthook.yml`に以下の内容を記述し、`lefthook install`を実行する。

```yaml
prepare-commit-msg:
  jobs:
    - name: auto-commit-message
      run: claude-commit-msg-gen {1} {2}
      env:
        CLAUDE_MODEL: claude-haiku-4-5-20251001
        CLAUDE_MAX_TOKENS: "150"
```

`COMMIT_PROMPT` は省略可能です。省略した場合、Conventional Commits 形式のコミットメッセージを生成するデフォルトプロンプトが使用されます。

`COMMIT_LANGUAGE` でコミットメッセージの言語を切り替えられます。

| 値 | 言語 |
| -- | ---- |
| `ja`（デフォルト） | 日本語（subject は50文字以内） |
| `en` | 英語（subject は72文字以内） |

```yaml
      env:
        CLAUDE_MODEL: claude-haiku-4-5-20251001
        CLAUDE_MAX_TOKENS: "150"
        COMMIT_LANGUAGE: en  # 英語で生成する場合
```

プロジェクト固有のルールを追加したい場合は `COMMIT_PROMPT` で上書きできます。`COMMIT_PROMPT` を指定した場合、`COMMIT_LANGUAGE` は無視されます。

```yaml
      env:
        COMMIT_PROMPT: >
          以下のgit diffを分析し、Conventional Commits形式のコミットメッセージを1行だけ生成してください。
          （プロジェクト独自のルールをここに追加）
```


### lefthookをプロジェクトに含めたくない場合もしくは、プロジェクトのlefthookを汚染したくない場合

1. `.gitignore`に`lefthook-local.yml`を追加する
2. **lefthook-local.yml**に[lefthookのセットアップ](#lefthookのセットアップ)の内容を記述する
3. `lefthook install`を実行する。
