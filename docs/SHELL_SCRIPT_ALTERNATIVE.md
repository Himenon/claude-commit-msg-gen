# シェルスクリプトによる代替実装

Go や npm 環境がない場合の代替手段です。`claude` CLI を利用してコミットメッセージを生成します。`claude` CLI がインストール済みであれば API キーの設定は不要です。

> **注意:** シェルスクリプト版は `max_tokens` をプロンプトへの指示として渡すため、Go バイナリ版と比べてトークン数の制御が不確実です。

## バイナリ版との比較

| 項目                | Go バイナリ（メイン）          | シェルスクリプト（本ドキュメント） |
| ------------------- | ------------------------------ | ---------------------------------- |
| API 呼び出し        | Anthropic API を直接呼び出す   | `claude` CLI 経由                  |
| `max_tokens` の制御 | API パラメータとして厳密に指定 | プロンプトへの指示（不確実）       |
| API キー            | `ANTHROPIC_API_KEY` が必要     | 不要（`claude` CLI が保持）        |
| 依存関係            | pnpm / npm のみ                | `claude` CLI                       |
| 実行速度            | 高速（ネイティブバイナリ）     | `claude` CLI 起動コストあり        |

## 前提条件

- [Lefthook](https://lefthook.dev/)
- [Claude Code](https://claude.ai/code)（`claude` CLI）

## セットアップ

```sh
git clone https://github.com/Himenon/claude-commit-msg-gen.git
cd claude-commit-msg-gen
lefthook install
```

`lefthook.yml` の `run` を以下のように変更してください。

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

## 実装ファイル

- `scripts/auto-commit-msg.sh` — シェルスクリプト本体
- `scripts/commit-prompt.txt` — プロンプトテンプレート（バイナリ版と共通）
