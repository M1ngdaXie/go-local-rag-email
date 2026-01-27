package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// GetClient returns an authenticated HTTP client for Gmail API
func GetClient(credentialsPath, tokenPath string) (*http.Client, error) {
    ctx := context.Background()
    b, err := os.ReadFile(credentialsPath)
    if err != nil {
        return nil, fmt.Errorf("read credentials failed: %w", err)
    }

    config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
    if err != nil {
        return nil, fmt.Errorf("parse config failed: %w", err)
    }

    tok, err := tokenFromFile(tokenPath)
    if err != nil {
        // 找不到文件，走全自动授权
        fmt.Println("🔑 No local token found. Opening browser for authorization...", err)
        tok, err = GetTokenViaLoopback(config) 
        if err != nil {
            return nil, err
        }
        err = saveToken(tokenPath, tok) 
		if err != nil{
			fmt.Print("Error saving token into files", err)
		}
    } else {
        // --- 增加过期检测逻辑 ---
        timeLeft := time.Until(tok.Expiry)
        if timeLeft > 0 {
            fmt.Printf("ℹ️  Token is valid. Expires in: %.0f minutes\n", timeLeft.Minutes())
        } else {
            fmt.Println("⏳ Token expired, attempting silent refresh...")
            // 注意：这里不需要手动调用刷新，config.Client 会在第一次请求时根据 RefreshToken 自动刷新
            // 但如果 tok 里没有 RefreshToken，这里就该提示用户重新 Auth
        }
    }
	ts := config.TokenSource(ctx, tok)
	newToken, err := ts.Token()
	if err != nil {
        return nil, fmt.Errorf("failed to get token from source: %w", err)
    }
	if newToken.AccessToken != tok.AccessToken {
		fmt.Print("Found token is not the same \n")
        fmt.Println("🔄 Detected token refresh, saving new token to disk...")
        saveToken(tokenPath, newToken)
    }
    // 这个 client 是个“智能”客户端：
    // 1. 如果 AccessToken 没过期，直接用。
    // 2. 如果过期了但有 RefreshToken，它会偷偷换个新的并继续请求。
    return config.Client(ctx, newToken), nil
}
// getTokenFromWeb starts OAuth flow in browser
func GetTokenViaLoopback(config *oauth2.Config) (*oauth2.Token, error) {
	// 1. 定义一个用于接收 code 的通道
	codeChan := make(chan string)
	errChan := make(chan error)

	// 2. 启动一个临时服务器
	server := &http.Server{Addr: ":8081"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			w.Write([]byte("<h1>授权失败，未获取到 Code</h1>"))
			return
		}
		// 把 code 传回给主逻辑
		codeChan <- code
		w.Write([]byte("<h1>授权成功！</h1><p>您可以关闭此窗口并回到终端。</p>"))
	})

	go func() {
		fmt.Println("📡 本地授权服务器正在监听 :8081...")
    	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        // 如果端口被占，这里会打印 "address already in use"
        fmt.Printf("❌ 服务器启动失败: %v\n", err)
        errChan <- err
    }
	}()

	// 3. 构造授权 URL 并打开浏览器
	// 关键：RedirectURL 必须和临时服务器地址匹配
	config.RedirectURL = "http://localhost:8081"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	
	fmt.Printf("正在为您打开浏览器授权: %v\n", authURL)
	OpenBrowser(authURL) // 调用之前写的 OpenBrowser 函数

	// 4. 等待结果
	select {
	case code := <-codeChan:
		// 拿到 code 后立刻关闭服务器
		server.Shutdown(context.Background())
		return config.Exchange(context.Background(), code)
	case err := <-errChan:
		return nil, fmt.Errorf("服务器错误: %w", err)
	case <-time.After(2 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("授权超时")
	}
}

// tokenFromFile reads token from JSON file
func tokenFromFile(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open token file failed: %w", err)
	}
	defer f.Close()

	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, fmt.Errorf("decode token failed: %w", err)
	}
	return tok, nil
}

// saveToken saves token to JSON file
func saveToken(path string, token *oauth2.Token) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create token file failed: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("encode token failed: %w", err)
	}
	return nil
}

func OpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		// Windows 下需要特殊处理 cmd 参数
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default:
		// 如果是不支持的系统，就退回到手动模式，不报错
		return nil
	}
	return err
}