# claude-commit-msg-gen

[Claude API](https://www.anthropic.com/) と [Lefthook](https://lefthook.dev/) を組み合わせ、`git commit` 時に Conventional Commits 形式のコミットメッセージを自動生成するツールです。

構造・設計の詳細は [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) を参照してください。

## セットアップ

```sh
# 1. 本ライブラリのインストール
# Without Node.js
curl -fsSL https://raw.githubusercontent.com/Himenon/claude-commit-msg-gen/main/scripts/install.sh | sh
# With Node.js
pnpm install -g @himenon/

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

1. `lefthook.yml`に以下の内容を記述する
    ```yaml
    prepare-commit-msg:
      jobs:
        - name: auto-commit-message
          run: claude-commit-msg-gen
    ```
2. `lefthook install`を実行する。

### lefthookをプロジェクトに含めたくない場合もしくは、プロジェクトのlefthookを汚染したくない場合

1. `.gitignore`に`lefthook-local.yml`を追加する
2. **lefthook-local.yml**に[lefthookのセットアップ](#lefthookのセットアップ)の内容を記述する
3. `lefthook install`を実行する。


