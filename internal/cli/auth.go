package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/M1ngdaXie/go-local-rag-email/pkg/oauth"
	"github.com/spf13/cobra"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)


var authCmd = &cobra.Command{
    Use:   "auth",
    Short: "Authenticate with Gmail (OAuth2)",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 获取 Cobra 命令自带的 context
        ctx := cmd.Context() 
        cfg := application.Config()

        // 1. 建议修改 oauth.GetClient 支持传入 ctx (如果内部有网络操作)
        client, err := oauth.GetClient(cfg.Gmail.CredentialsPath, cfg.Gmail.TokenPath)
        if err != nil {
             return fmt.Errorf("authentication failed: %w", err)
        }

        fmt.Println("✅ Authentication successful! Token saved to:", cfg.Gmail.TokenPath)

        // 2. 显式传递 ctx
        verifyToken(ctx, client) 
        return nil 
    },
}

// 修改函数签名
func verifyToken(ctx context.Context, client *http.Client) {
    // 使用传入的 ctx，并加上额外的超时保护
    vCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    srv, err := gmail.NewService(vCtx, option.WithHTTPClient(client))
    if err != nil {
        log.Fatalf("无法初始化 Gmail 服务: %v", err)
    }

    profile, err := srv.Users.GetProfile("me").Do()
    if err != nil {
        log.Fatalf("Token 验证失败: %v", err)
    }

    fmt.Printf("✅ 验证成功！\n当前用户: %s\n", profile.EmailAddress)
}
func init() {
	rootCmd.AddCommand(authCmd)
}
