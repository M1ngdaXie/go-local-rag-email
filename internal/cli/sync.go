package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/email"
	"github.com/M1ngdaXie/go-local-rag-email/internal/service/gmail"
	"github.com/M1ngdaXie/go-local-rag-email/pkg/oauth"
	"github.com/spf13/cobra"
)


var maxEmails int64

var syncCmd = &cobra.Command{
    Use:   "sync",
    Short: "Sync emails from Gmail",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 使用带有超时控制的 Context，防止同步任务无限期挂起
        ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
        defer cancel()

        cfg := application.Config()

        // 1. 获取认证客户端
        httpClient, err := oauth.GetClient(cfg.Gmail.CredentialsPath, cfg.Gmail.TokenPath)
        if err != nil {
            return fmt.Errorf("auth failed: %w", err)
        }

        // 2. 初始化 Gmail Service
        gmailSvc, err := gmail.New(ctx, httpClient)
        if err != nil {
            return fmt.Errorf("init gmail service failed: %w", err)
        }

        // 3. 拉取邮件
        fmt.Printf("Fetching up to %d emails from Gmail...\n", maxEmails)
        emails, err := gmailSvc.FetchEmails(ctx, maxEmails)
        if err != nil {
            return fmt.Errorf("fetch failed: %w", err)
        }
        fmt.Printf("Successfully fetched %d emails\n", len(emails))

        // 4. 保存到 SQLite
        repo := email.NewSQLiteRepository(application.SQLiteDB(), application.Logger())
        
        var newCount int
        for _, e := range emails {
            // 在 Repository 层建议实现 "Upsert" 逻辑或在 Create 时检查 ID 冲突
            err := repo.Create(ctx, e)
            if err != nil {
                // 如果是唯一的 ID 冲突（邮件已存在），我们可以忽略它继续执行
                // 假设你的 repo 返回特定的 Duplicate 错误
                continue 
            }
            newCount++
        }

        // 5. 打印总结报告
        fmt.Printf("✅ Sync complete: %d new emails saved to database.\n", newCount)

        return nil
    },
}

func init() {
	syncCmd.Flags().Int64Var(&maxEmails, "max", 50, "Max emails to sync")
	rootCmd.AddCommand(syncCmd)
}
