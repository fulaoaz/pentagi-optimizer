package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/hardening"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard"
	"pentagi/pkg/version"
)

type Config struct {
	envPath     string
	showVersion bool
}

func main() {
	config := parseFlags(os.Args)

	if config.showVersion {
		fmt.Println(version.GetBinaryVersion())
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	envPath, err := validateEnvPath(config.envPath)
	if err != nil {
		log.Fatalf("错误：%v", err)
	}

	appState, err := initializeState(envPath)
	if err != nil {
		log.Fatalf("初始化状态失败：%v", err)
	}

	if err := hardening.DoMigrateSettings(appState); err != nil {
		log.Fatalf("迁移设置失败：%v", err)
	}

	if err := hardening.DoSyncNetworkSettings(appState); err != nil {
		log.Fatalf("同步网络设置失败：%v", err)
	}

	checkResult, err := gatherSystemFacts(ctx, appState)
	if err != nil {
		log.Fatalf("收集系统信息失败：%v", err)
	}

	printStartupInfo(envPath, checkResult)

	if err := hardening.DoHardening(appState, checkResult); err != nil {
		log.Fatalf("执行安全加固失败：%v", err)
	}

	if err := runApplication(ctx, appState, checkResult); err != nil {
		log.Fatalf("应用程序错误：%v", err)
	}

	cleanup(appState)
}

func parseFlags(args []string) Config {
	var config Config

	name := "installer"
	if len(args) > 0 {
		args, name = args[1:], filepath.Base(args[0])
	}

	flagSet := flag.NewFlagSet(name, flag.ContinueOnError)
	flagSet.BoolVar(&config.showVersion, "v", false, "显示版本信息")
	flagSet.StringVar(&config.envPath, "e", ".env", "环境配置文件路径")
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "PentAGI 安装程序 v%s\n\n", version.GetBinaryVersion())
		fmt.Fprintf(os.Stderr, "用法：%s [选项]\n\n", name)
		fmt.Fprintf(os.Stderr, "选项：\n")
		flagSet.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例：\n")
		fmt.Fprintf(os.Stderr, "  %s                    # 使用默认 .env 文件\n", name)
		fmt.Fprintf(os.Stderr, "  %s -e config/.env     # 使用指定的环境配置文件\n", name)
		fmt.Fprintf(os.Stderr, "  %s -v                 # 显示版本\n", name)
	}

	flagSet.Parse(args)
	return config
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("收到信号 %v，正在正常退出...", sig)
		cancel()
	}()
}

func validateEnvPath(envPath string) (string, error) {
	// convert to absolute path
	absPath, err := filepath.Abs(envPath)
	if err != nil {
		return "", fmt.Errorf("路径 '%s' 无效：%w", envPath, err)
	}

	// check if file exists
	if info, err := os.Stat(absPath); os.IsNotExist(err) {
		// file doesn't exist, check if we can create it in the directory
		dir := filepath.Dir(absPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("目录 '%s' 创建失败：%w", dir, err)
			}
		} else if err != nil {
			return "", fmt.Errorf("目录 '%s' 访问失败：%w", dir, err)
		}

		// try to create initial env file
		if err := createInitialEnvFile(absPath); err != nil {
			return "", fmt.Errorf("环境配置文件 '%s' 创建失败：%w", absPath, err)
		}
	} else if info.IsDir() {
		return "", fmt.Errorf("'%s' 是一个目录", absPath)
	} else if err != nil {
		return "", fmt.Errorf("文件 '%s' 访问失败：%w", absPath, err)
	}

	return absPath, nil
}

func createInitialEnvFile(path string) error {
	f := files.NewFiles()

	content, err := f.GetContent(".env")
	if err != nil {
		return fmt.Errorf(".env 文件读取失败：%w", err)
	}

	content = fmt.Appendf(nil, `# PentAGI 环境配置
# 由 PentAGI 安装程序 v%s 生成
#
# 本文件包含 PentAGI 的环境变量配置。
# 你可以通过安装程序界面修改这些值。
#
%s`, version.GetBinaryVersion(), string(content))

	if err := os.WriteFile(path, content, 0600); err != nil {
		return fmt.Errorf(".env 文件写入失败：%w", err)
	}

	return nil
}

func initializeState(envPath string) (state.State, error) {
	appState, err := state.NewState(envPath)
	if err != nil {
		return nil, fmt.Errorf("状态管理器创建失败：%w", err)
	}

	return appState, nil
}

func gatherSystemFacts(ctx context.Context, appState state.State) (checker.CheckResult, error) {
	result, err := checker.Gather(ctx, appState)
	if err != nil {
		return result, fmt.Errorf("系统信息收集失败：%w", err)
	}

	return result, nil
}

func printStartupInfo(envPath string, checkResult checker.CheckResult) {
	fmt.Printf("PentAGI 安装程序 v%s\n", version.GetBinaryVersion())
	fmt.Printf("环境配置文件：%s\n", envPath)

	if !checkResult.IsReadyToContinue() {
		fmt.Println("⚠️  系统尚未就绪，请先解决上述问题。")
	} else {
		fmt.Println("✅ 系统已就绪，可以继续。")
	}
}

func runApplication(ctx context.Context, appState state.State, checkResult checker.CheckResult) error {
	return wizard.Run(ctx, appState, checkResult, files.NewFiles())
}

func cleanup(appState state.State) {
	if appState.IsDirty() {
		fmt.Println("你有尚未应用的更改。")
		fmt.Println("再次运行安装程序可继续处理，或提交现有更改。")
	}
}
