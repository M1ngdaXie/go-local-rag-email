package cli

import (
	"fmt"
	"strings"

	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/email"
	"github.com/spf13/cobra"
)
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "列出本地数据库中的邮件",
    RunE: func(cmd *cobra.Command, args []string) error {
        repo := email.NewSQLiteRepository(application.SQLiteDB(), application.Logger())
        
        // 使用 Cobra 自带的 Context
        ctx := cmd.Context()

        // 使用你定义的 Pagination 获取前 20 封
        emails, err := repo.List(ctx, email.Filter{}, email.Pagination{Limit: 20})
        if err != nil {
            return fmt.Errorf("读取数据库失败: %w", err)
        }

        if len(emails) == 0 {
            fmt.Println("📭 数据库中没有邮件。请先运行 'sync' 命令。")
            return nil
        }

        fmt.Printf("--- 本地数据库最近 %d 封邮件 ---\n\n", len(emails))
        for _, e := range emails {
            // 打印核心元数据
            fmt.Printf("ID:      %s\n", e.ID)
            fmt.Printf("Subject: %s\n", e.Subject)
            fmt.Printf("From:    %s\n", e.From)
            fmt.Printf("Date:    %s\n", e.Date.Format("2006-01-02 15:04"))
            
            // 关键：看看 RAG 用的 BodyText 是否成功解析了
            // 打印前 100 个字符进行预览
            bodyPreview := e.BodyText
            if len(bodyPreview) > 100 {
                bodyPreview = bodyPreview[:100] + "..."
            }
            fmt.Printf("Content: %s\n", bodyPreview)
            fmt.Println(strings.Repeat("-", 40))
        }
        
        return nil
    },
}


func init() {
    rootCmd.AddCommand(listCmd)
}