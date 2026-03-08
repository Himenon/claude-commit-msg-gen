#!/bin/sh
# auto-commit-msg.sh
#
# prepare-commit-msg フック用スクリプト
# git diffを解析してClaudeにConventional Commits形式のコミットメッセージを生成させる
#
# 環境変数:
#   CLAUDE_MODEL      : 使用するClaudeモデル (デフォルト: claude-haiku-4-5-20251001)
#   CLAUDE_MAX_TOKENS : 最大トークン数 (デフォルト: 150)
#   COMMIT_PROMPT_FILE: プロンプトファイルのパス (デフォルト: scripts/commit-prompt.txt)

COMMIT_MSG_FILE="$1"
COMMIT_SOURCE="$2"

# Merge Commitの場合はスキップ
if [ "$COMMIT_SOURCE" = "merge" ]; then
  exit 0
fi

# 既にコミットメッセージが入力されている場合はスキップ（-m オプション等）
if [ "$COMMIT_SOURCE" = "message" ]; then
  exit 0
fi

# 設定値（環境変数で上書き可能）
MODEL="${CLAUDE_MODEL:-claude-haiku-4-5-20251001}"
# CLAUDE_MAX_TOKENS はプロンプトへの指示として使用する（APIオプションではない）
MAX_TOKENS="${CLAUDE_MAX_TOKENS:-150}"

# プロジェクトルートを取得
REPO_ROOT="$(git rev-parse --show-toplevel)"

# プロンプトファイルのパス
# COMMIT_PROMPT_FILEが相対パスの場合はプロジェクトルートを基準に解決する
if [ -n "$COMMIT_PROMPT_FILE" ]; then
  case "$COMMIT_PROMPT_FILE" in
    /*) PROMPT_FILE="$COMMIT_PROMPT_FILE" ;;
    *)  PROMPT_FILE="${REPO_ROOT}/${COMMIT_PROMPT_FILE}" ;;
  esac
else
  PROMPT_FILE="${REPO_ROOT}/scripts/commit-prompt.txt"
fi

# プロンプトファイルが存在しない場合はスキップ
if [ ! -f "$PROMPT_FILE" ]; then
  echo "[auto-commit-msg] プロンプトファイルが見つかりません: $PROMPT_FILE" >&2
  exit 0
fi

# ステージングされた差分を取得
DIFF="$(git diff --cached --no-color 2>/dev/null)"

# 差分がない場合はスキップ
if [ -z "$DIFF" ]; then
  exit 0
fi

# プロンプトを構築
# MAX_TOKENSをプロンプトに含めて出力量を制御する
PROMPT="$(cat "$PROMPT_FILE")
（出力は${MAX_TOKENS}トークン以内に収めること）

---
$(echo "$DIFF" | head -c 8000)
"

# Claudeにコミットメッセージを生成させる
# CLAUDECODE変数をunsetしてネストしたセッション制限を回避する
GENERATED_MSG="$(echo "$PROMPT" | env -u CLAUDECODE claude --print \
  --model "$MODEL" \
  --output-format text \
  --dangerously-skip-permissions \
  2>/dev/null)"

# 生成に失敗した場合はスキップ（git commitを止めない）
if [ $? -ne 0 ] || [ -z "$GENERATED_MSG" ]; then
  echo "[auto-commit-msg] コミットメッセージの生成に失敗しました（処理を継続します）" >&2
  exit 0
fi

# 生成されたメッセージからConventional Commits形式の行を抽出する
# バッククォートを含む行（コードブロックマーカー等）を除去し、type(scope): subject 形式の行を優先して取得する
CLEANED_MSG="$(echo "$GENERATED_MSG" \
  | grep -v '`' \
  | sed 's/^[[:space:]]*//' \
  | sed 's/[[:space:]]*$//' \
  | grep -E '^[a-z]+(\([^)]+\))?: .+' \
  | head -1)"

# Conventional Commits形式の行が見つからない場合は最初の空でない行を使用する
if [ -z "$CLEANED_MSG" ]; then
  CLEANED_MSG="$(echo "$GENERATED_MSG" \
    | grep -v '`' \
    | sed 's/^[[:space:]]*//' \
    | sed 's/[[:space:]]*$//' \
    | grep -v '^$' \
    | head -1)"
fi

# 残存するバッククォートを除去する（念のため）
CLEANED_MSG="$(echo "$CLEANED_MSG" | tr -d '`')"

# コミットメッセージファイルに書き込む
# 既存の内容（コメント行）を保持しつつ先頭に挿入
EXISTING_CONTENT="$(cat "$COMMIT_MSG_FILE")"
printf '%s\n\n%s\n' "$CLEANED_MSG" "$EXISTING_CONTENT" > "$COMMIT_MSG_FILE"

echo "[auto-commit-msg] コミットメッセージを生成しました: $CLEANED_MSG" >&2
