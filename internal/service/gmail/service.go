package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/M1ngdaXie/go-local-rag-email/internal/domain"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Service struct {
	client *gmail.Service
}

func New(ctx context.Context, httpClient *http.Client) (*Service, error) {
	svc, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("error creating gmail service: %w", err)
	}
	return &Service{client: svc}, nil
}

// FetchEmails 抓取最近的邮件
func (s *Service) FetchEmails(ctx context.Context, maxResults int64) ([]*domain.Email, error) {
	resp, err := s.client.Users.Messages.List("me").MaxResults(maxResults).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list messages: %w", err)
	}

	var emails []*domain.Email
	for _, m := range resp.Messages {
		// 获取完整内容（包括 Headers 和 Payload）
		msg, err := s.client.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			continue // 生产环境建议记录日志
		}
		emails = append(emails, parseMessage(msg))
	}

	return emails, nil
}

func parseMessage(msg *gmail.Message) *domain.Email {
	email := &domain.Email{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Snippet:  msg.Snippet,
	}

	var dateStr string

	// 1. 提取 Headers
	for _, h := range msg.Payload.Headers {
		switch h.Name {
		case "From":
			email.From = h.Value
		case "Subject":
			email.Subject = h.Value
		case "Date":
			dateStr = h.Value
		}
	}

	// 2. 解析日期 (使用 InternalDate 作为 fallback)
	email.Date = parseEmailDateWithFallback(dateStr, msg.InternalDate)

	// 3. 提取正文 (RAG 的核心数据)
	email.BodyText = getBodyText(msg.Payload)

	return email
}

func getBodyText(payload *gmail.MessagePart) string {
	// Step 1: Try to find text/plain first (preferred for RAG)
	if plainText := findPlainText(payload); plainText != "" {
		return plainText
	}

	// Step 2: Fallback to HTML with tags stripped
	if htmlText := findHTMLText(payload); htmlText != "" {
		return stripHTMLTags(htmlText)
	}

	return ""
}

// findPlainText recursively searches for text/plain content
func findPlainText(payload *gmail.MessagePart) string {
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		if decoded, err := decodeBase64Body(payload.Body.Data); err == nil {
			return decoded
		}
	}

	for _, part := range payload.Parts {
		if text := findPlainText(part); text != "" {
			return text
		}
	}
	return ""
}

// findHTMLText recursively searches for text/html content
func findHTMLText(payload *gmail.MessagePart) string {
	if payload.MimeType == "text/html" && payload.Body != nil && payload.Body.Data != ""{
		if decoded, err := decodeBase64Body(payload.Body.Data); err == nil {
			return decoded
		}
	}
	for _, part := range payload.Parts{
		if text := findHTMLText(part); text != ""{
			return text
		}
	}

	return ""
}

func decodeBase64Body(data string) (string, error) {
	var decoded []byte
	var err error

	// 1. 首先尝试最标准的 URL 安全解码
	decoded, err = base64.URLEncoding.DecodeString(data)
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(data)
	}
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(data)
	}

	if err != nil {
		return "", fmt.Errorf("all decoding attempts failed: %v", err)
	}

	return string(decoded), nil
}

// stripHTMLTags removes HTML tags and decodes entities for RAG processing
func stripHTMLTags(input string) string {
	content := input

	// 0. 先删除 <style> 和 <script> 块（包括内容）
	// 这些标签的内容不是可读文本，必须整块移除
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	content = reStyle.ReplaceAllString(content, "")

	reScript := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	content = reScript.ReplaceAllString(content, "")

	// 1. 预处理：将块级标签替换为换行符，保留基本的段落结构
	// 把 <br>, <p>, </div> 等替换为换行，防止文字粘连
	re := regexp.MustCompile(`(?i)<(br|p|/p|/div|/h[1-6])\s*/?>`)
	content = re.ReplaceAllString(content, "\n")

	// 2. 核心：删除所有 HTML 标签
	// 匹配 < 符号开始，到 > 符号结束的所有内容
	reTags := regexp.MustCompile(`<[^>]*>`)
	content = reTags.ReplaceAllString(content, " ")

	// 3. 处理 HTML 实体字符 (Entity Decoding)
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&nbsp;", " ",
		"&quot;", "\"",
		"&#39;", "'",
	)
	content = replacer.Replace(content)

	// 4. 清理空白字符：将多个连续空格压缩为一个，去掉首尾空格
	// 使用正则表达式 \s+ 匹配所有空白符（空格、制表符等）
	reSpace := regexp.MustCompile(`\s+`)
	content = reSpace.ReplaceAllString(content, " ")
	
	// 5. 去掉首尾空格并返回
	return strings.TrimSpace(content)
}

// parseEmailDate tries multiple date formats commonly used in email headers
func parseEmailDate(dateStr string) time.Time {
	// 预处理：去掉末尾可能存在的 (UTC) 或 (PST) 等非标准后缀
	// 邮件头有时会出现 "Mon, 02 Jan 2006 15:04:05 -0700 (UTC)"
	dateStr = regexp.MustCompile(`\s\([A-Z]{3}\)$`).ReplaceAllString(dateStr, "")

	formats := []string{
		time.RFC1123Z,                      // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,                       // "Mon, 02 Jan 2006 15:04:05 MST"
		"Mon, 2 Jan 2006 15:04:05 -0700",   // 单位数字日期
		"02 Jan 2006 15:04:05 -0700",       // 无星期
		"2 Jan 2006 15:04:05 -0700",        // 无星期且单位数字
		time.RFC3339,                       // ISO 格式
		"2006-01-02 15:04:05",              // 简单格式
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t
		}
	}

	// 如果所有格式都失败了，记录原始字符串方便以后 Debug 增加格式
	// fmt.Printf("⚠️ 无法解析日期字符串: %s\n", dateStr)
	return time.Time{}
}
// parseEmailDateWithFallback uses InternalDate (Unix ms) as reliable fallback
func parseEmailDateWithFallback(dateStr string, internalDateMs int64) time.Time {
	// 1. 优先尝试解析 Header 里的 Date 字符串
	t := parseEmailDate(dateStr)

	// 2. 如果解析失败（返回了零值），或者 Date 字符串本身为空
	if t.IsZero() {
		// 使用 Gmail 提供的内部时间戳（毫秒转为 time.Time）
		return time.UnixMilli(internalDateMs)
	}

	return t
}

// ============================================================================
// Test Utilities (Exported for CLI testing)
// ============================================================================

// ParseMessage is the exported version of parseMessage for testing
func ParseMessage(msg *gmail.Message) *domain.Email {
	return parseMessage(msg)
}

// GetFilthyEmail returns a test email with messy HTML content for parser testing
func GetFilthyEmail() *gmail.Message {
	// 模拟一个极其混乱的 HTML 正文
	// 包含：样式标签、多层 div、各种 HTML 实体、无意义空格
	dirtyHTML := `
		<html>
			<head><style>.bad { color: red; }</style></head>
			<body>
				<div class="wrapper">
					<h1>Interview&nbsp;Update!!</h1>
					<p>Dear Candidate,<br/>We are &lt;excited&gt; to invite you.</p>
					<div style="display:none">Click here to unsubscribe from &amp;all emails.</div>
					<a href="https://glassdoor.com/test">Check details</a>
					&quot;Best of Luck&quot; &#39;2026&#39;
				</div>
			</body>
		</html>`

	return &gmail.Message{
		Id:           "filthy_test_001",
		InternalDate: 1737934684000, // 2026-01-26 约 15:38 UTC
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "   [Urgent] Re: Job Application & Status   "},
				{Name: "From", Value: "Recruiter <hr-noreply@glassdoor.com>"},
				// 故意加上带括号后缀的非标准日期
				{Name: "Date", Value: "Mon, 26 Jan 2026 13:31:54 -0800 (PST)"},
			},
			MimeType: "multipart/alternative",
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/html",
					Body: &gmail.MessagePartBody{
						// 将上面的脏 HTML 转为 Base64 (URL Safe)
						Data: base64.URLEncoding.EncodeToString([]byte(dirtyHTML)),
					},
				},
			},
		},
	}
}