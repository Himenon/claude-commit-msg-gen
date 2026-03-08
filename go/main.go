package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "[claude-commit-msg-gen] Usage: claude-commit-msg-gen <commit-msg-file> [commit-source]")
		os.Exit(0)
	}

	commitMsgFile := os.Args[1]
	commitSource := ""
	if len(os.Args) >= 3 {
		commitSource = os.Args[2]
	}

	// Merge commit はスキップ
	if commitSource == "merge" {
		os.Exit(0)
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

	// プロンプトファイルのパスを解決する
	// 相対パスが渡された場合はリポジトリルートを基準に絶対パスへ変換する
	repoRoot, err := getRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[claude-commit-msg-gen] リポジトリルートの取得に失敗しました: %v\n", err)
		os.Exit(0)
	}

	promptFilePath := os.Getenv("COMMIT_PROMPT_FILE")
	if promptFilePath == "" {
		promptFilePath = filepath.Join(repoRoot, "scripts", "commit-prompt.txt")
	} else if !filepath.IsAbs(promptFilePath) {
		promptFilePath = filepath.Join(repoRoot, promptFilePath)
	}

	promptBytes, err := os.ReadFile(promptFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[claude-commit-msg-gen] プロンプトファイルが見つかりません: %s\n", promptFilePath)
		os.Exit(0)
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
	commitMsg, err := generateCommitMessage(apiURL, apiKey, model, maxTokens, string(promptBytes), diff)
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
			return strings.TrimSpace(block.Text), nil
		}
	}

	return "", fmt.Errorf("レスポンスにテキストブロックが含まれていません")
}
