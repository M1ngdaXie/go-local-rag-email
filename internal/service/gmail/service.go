package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
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

	// 1. 提取 Headers
	for _, h := range msg.Payload.Headers {
		switch h.Name {
		case "From":
			email.From = h.Value
		case "Subject":
			email.Subject = h.Value
		case "Date":
			// 邮件日期格式多变，RFC1123Z 是最常见的
			t, err := time.Parse(time.RFC1123Z, h.Value)
			if err == nil {
				email.Date = t
			}
		}
	}

	// 2. 提取正文 (RAG 的核心数据)
	email.BodyText = getBodyText(msg.Payload)

	return email
}

func getBodyText(payload *gmail.MessagePart) string {
	// 逻辑 A: 如果当前 Part 就是纯文本，直接解码
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	// 逻辑 B: 如果有子部件 (Multipart)，递归寻找 text/plain
	// 优先寻找纯文本，因为 HTML 里的标签会污染 RAG 的 Embedding 效果
	for _, part := range payload.Parts {
		if text := getBodyText(part); text != "" {
			return text
		}
	}

	return ""
}