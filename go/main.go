package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// version はビルド時に -ldflags "-X main.version=vX.Y.Z" で注入される
var version = "dev"

const (
	defaultBaseURL      = "https://api.anthropic.com"
	anthropicAPIVersion = "2023-06-01"

	// defaultModel: claude-haiku を使用
	// 理由: commit message 生成という単純なタスクに十分な性能で、
	//       git commit のワークフローを中断しない高速レスポンスを実現する
	defaultModel = "claude-haiku-4-5-20251001"

	// defaultMaxTokens: 150 トークンに制限
	// 理由: commit message は1行(50文字前後)で十分なため、
	//       生成トークン数を抑えることで実行速度を向上し API コストを削減する
	defaultMaxTokens = 150

	// defaultCommitPrompt: COMMIT_PROMPT 未設定時のデフォルトプロンプト
	defaultCommitPrompt = `以下のgit diffを分析し、Conventional Commits形式のコミットメッセージを1行だけ生成してください。

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
- 「大きい」「小さい」「多い」「少ない」のような比較表現を使う場合は比較対象を明示すること（例: 「リクエスト数が毎秒100件を超える場合」）`

	// apiTimeout: API リクエストのタイムアウト時間
	// 理由: ネットワーク障害時に無限に待ち続けず、commit ワークフローをブロックしない
	apiTimeout = 30 * time.Second

	// maxDiffSize: git diff の最大サイズ（バイト数）
	// 理由: 巨大な diff を API に送信して APIエラーやコスト増加を防ぐ
	maxDiffSize = 50000
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requestBody struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responseBody struct {
	Content []contentBlock `json:"content"`
}

type apiError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func main() {
	// エラーハンドリングポリシー:
	// このツールは常に exit 0 で終了し、git commit を止めない設計
	// 理由: API障害や設定ミスがあっても commit ワークフローを中断させない
	//       エラー内容は stderr に出力し、ユーザーに通知する

	if len(os.Args) >= 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(version)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "[claude-commit-msg-gen] Usage: claude-commit-msg-gen <commit-msg-file> [commit-source]")
		os.Exit(0)
	}

	commitMsgFile := os.Args[1]
	commitSource := ""
	if len(os.Args) >= 3 {
		commitSource = os.Args[2]
	}

	// Merge commit はスキップ（commitSource による判定）
	if commitSource == "merge" {
		os.Exit(0)
	}

	// Merge commit はスキップ（.git/MERGE_HEAD ファイルの存在による判定）
	// 理由: commitSource が渡されない環境でも merge 中を確実に検知する
	if repoRoot, err := getRepoRoot(); err == nil {
		if _, err := os.Stat(repoRoot + "/.git/MERGE_HEAD"); err == nil {
			os.Exit(0)
		}
	}

	// -m オプション等でメッセージが直接指定された場合はスキップ
	if commitSource == "message" {
		os.Exit(0)
	}

	// API キー: ANTHROPIC_API_KEY を優先し、ANTHROPIC_AUTH_TOKEN にフォールバック
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_AUTH_TOKEN")
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "[claude-commit-msg-gen] ANTHROPIC_API_KEY が未設定のためスキップします")
		os.Exit(0)
	}

	// 使用モデル
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = defaultModel
	}

	// 最大トークン数
	maxTokens := defaultMaxTokens
	if v := os.Getenv("CLAUDE_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxTokens = n
		}
	}

	commitPrompt := os.Getenv("COMMIT_PROMPT")
	if commitPrompt == "" {
		commitPrompt = defaultCommitPrompt
	}

	// ステージングされた差分を取得する
	diff, err := getStagedDiff()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[claude-commit-msg-gen] git diff の取得に失敗しました: %v\n", err)
		os.Exit(0)
	}
	if diff == "" {
		os.Exit(0)
	}
	if len(diff) > maxDiffSize {
		diff = diff[:maxDiffSize]
	}

	// API エンドポイント
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	apiURL := baseURL + "/v1/messages"

	// commit message を生成する
	commitMsg, err := generateCommitMessage(apiURL, apiKey, model, maxTokens, commitPrompt, diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[claude-commit-msg-gen] コミットメッセージの生成に失敗しました: %v\n", err)
		os.Exit(0)
	}

	// 既存のコミットメッセージファイルの内容（git のコメント行）を保持しつつ先頭に挿入する
	existingContent, _ := os.ReadFile(commitMsgFile)
	newContent := commitMsg + "\n\n" + string(existingContent)

	if err := os.WriteFile(commitMsgFile, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "[claude-commit-msg-gen] ファイルへの書き込みに失敗しました: %v\n", err)
		os.Exit(0)
	}

	fmt.Fprintf(os.Stderr, "[claude-commit-msg-gen] 生成されたコミットメッセージ: %s\n", commitMsg)
}

// cleanMessage はAPIレスポンスからコミットメッセージとして不適切な文字列を除去する。
// バッククォートを含む行（コードブロックマーカー等）を除去し、
// Conventional Commits形式（type(scope): subject）の行を優先して返す。
func cleanMessage(raw string) string {
	var candidates []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "`") {
			continue
		}
		candidates = append(candidates, line)
	}

	// Conventional Commits形式の行を優先する
	for _, line := range candidates {
		if isConventionalCommit(line) {
			return line
		}
	}

	// 該当行がなければ最初の空でない行を返す
	if len(candidates) > 0 {
		return candidates[0]
	}

	return strings.TrimSpace(raw)
}

// isConventionalCommit は type(scope): subject 形式かどうかを判定する
func isConventionalCommit(s string) bool {
	prefix, _, ok := strings.Cut(s, ":")
	if !ok {
		return false
	}
	// type または type(scope) の形式であることを確認する
	// type は小文字アルファベットのみ
	typeStr, _, _ := strings.Cut(prefix, "(")
	if len(typeStr) == 0 {
		return false
	}
	for _, c := range typeStr {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}

func getRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func getStagedDiff() (string, error) {
	out, err := exec.Command("git", "diff", "--cached", "--no-color").Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func generateCommitMessage(apiURL, apiKey, model string, maxTokens int, promptTemplate, diff string) (string, error) {
	prompt := strings.TrimSpace(promptTemplate) + "\n\n---\n" + diff

	reqBody := requestBody{
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("リクエストのシリアライズに失敗: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP リクエストの生成に失敗: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	client := &http.Client{Timeout: apiTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API リクエストに失敗: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンスの読み取りに失敗: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error.Message != "" {
			return "", fmt.Errorf("API エラー (%d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return "", fmt.Errorf("API エラー (%d): %s", resp.StatusCode, string(body))
	}

	var respBody responseBody
	if err := json.Unmarshal(body, &respBody); err != nil {
		return "", fmt.Errorf("レスポンスのパースに失敗: %w", err)
	}

	for _, block := range respBody.Content {
		if block.Type == "text" {
			return cleanMessage(block.Text), nil
		}
	}

	return "", fmt.Errorf("レスポンスにテキストブロックが含まれていません")
}
