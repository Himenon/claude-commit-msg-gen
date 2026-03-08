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
      run: claude-commit-msg-gen
      env:
      COMMIT_PROMPT: >
        以下のgit diffを分析し、Conventional Commits形式のコミットメッセージを1行だけ生成してください。

        形式: type(scope): subject

        typeの選択基準:
        - feat: 新機能の追加
        - fix: バグ修正
        - docs: ドキュメントのみの変更
        - style: コードの動作に影響しない変更（フォーマット、空白など）
        - refactor: バグ修正でも機能追加でもないコード変更
        - test: テストの追加・修正
        - chore: ビルドプロセスや補助ツールの変更
        - ci: CI設定ファイルやスクリプトの変更
        - perf: パフォーマンス改善

        ルール:
        - subjectは日本語で記述すること
        - subjectは50文字以内にすること
        - 末尾に句読点を付けないこと
        - scopeは変更対象のモジュールやファイル名（省略可）
        - コミットメッセージの1行のみを出力すること。説明文や前後の文章は不要

        文章ルール:
        - 「データ」「正しい」「通常」「異常」「大量」「少量」などの抽象的表現を使わないこと。変更対象・処理内容・状態を具体的な技術用語で表現すること
        - 主語を具体的に書くこと。「これ」「それ」「あれ」などの指示詞を使わないこと
        - 「大きい」「小さい」「多い」「少ない」のような比較表現を使う場合は比較対象を明示すること（例: 「リクエスト数が毎秒100件を超える場合」）
```


### lefthookをプロジェクトに含めたくない場合もしくは、プロジェクトのlefthookを汚染したくない場合

1. `.gitignore`に`lefthook-local.yml`を追加する
2. **lefthook-local.yml**に[lefthookのセットアップ](#lefthookのセットアップ)の内容を記述する
3. `lefthook install`を実行する。
