# CLAUDE.md

プロジェクトの構造・コンポーネント設計は **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** を参照してください。

## 作業ルール

- ファイルを変更したら必ず `git commit` を実行すること

---

## プロジェクト概要

`claude-commit-msg-gen` — Claude API と Lefthook を組み合わせ、`git commit` 時に Conventional Commits 形式のコミットメッセージを自動生成するツール。

npm パッケージ名: `@himenon/claude-commit-msg-gen`

---

## 構築手順のまとめ

### 1. Lefthook フックの設定（シェルスクリプト版）

- `lefthook.yml` に `prepare-commit-msg` フックを追加
- `scripts/auto-commit-msg.sh` を作成
  - `git diff --cached` で差分取得
  - `env -u CLAUDECODE claude --print` で Claude CLI を呼び出し
  - `{1}` = commit message ファイル、`{2}` = commit source（`merge` / `message` 時はスキップ）
  - エラー時は `exit 0` で git commit を止めない
- `scripts/commit-prompt.txt` にプロンプトを外部化（後から変更可能）

### 2. 設定の外部化

`lefthook.yml` の `env` セクションに以下を明示:

| 環境変数 | デフォルト値 | 用途 |
|---|---|---|
| `CLAUDE_MODEL` | `claude-haiku-4-5-20251001` | 使用モデル |
| `CLAUDE_MAX_TOKENS` | `"150"` | 最大トークン数 |
| `COMMIT_PROMPT_FILE` | `scripts/commit-prompt.txt` | プロンプトファイルパス（相対パス可） |

### 3. Go バイナリ実装への移行

シェルスクリプト版は `max_tokens` をプロンプト指示として渡すため制御が不確実。Go 実装で Anthropic API を直接呼び出すことで厳密に制御。

- `go/main.go` — Anthropic API を直接呼び出す実装（外部依存なし・標準ライブラリのみ）
- `go/go.mod` — Go 1.26
- API キーは `ANTHROPIC_API_KEY`（`ANTHROPIC_AUTH_TOKEN` にフォールバック）
- `scripts/auto-commit-msg.sh` はそのまま残す（代替案として保持）

### 4. npm パッケージとしての配布

- `package.json` — `@himenon/claude-commit-msg-gen` として定義
- `scripts/build.sh` — 4プラットフォーム向けクロスコンパイル
  - `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`
  - `bin/claude-commit-msg-gen` にプラットフォーム自動判定ラッパーを生成
- `bin/` は `.gitignore` 対象だが `package.json` の `files` に含めることで npm publish 時に配布される

### 5. GitHub Actions

- `.github/workflows/ci.yml` — 全ブランチ push と pull_request でビルド検証
- `.github/workflows/release.yml` — `v*` タグ push 時に自動ビルド・publish
- 全 action は SHA pinning 済み
- `NPM_TOKEN` を GitHub Secrets に登録が必要

### 6. ドキュメント整理

- `README.md` — Go バイナリ（メイン）の説明のみ
- `docs/ARCHITECTURE.md` — プロジェクト構造・コンポーネント設計
- `docs/shell-script-alternative.md` — シェルスクリプト版の代替案を分離

---

## ローカルでのビルド・テスト

```sh
pnpm run build   # bin/ にバイナリ生成
lefthook install # git フック有効化
git add <file>
git commit       # フックが起動し自動生成される
```

## publish

```sh
git tag v1.0.0
git push origin v1.0.0  # GitHub Actions が自動でビルド・publish
```
