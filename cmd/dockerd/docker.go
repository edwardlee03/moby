package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/docker/docker/cli"
	"github.com/docker/docker/daemon/config"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/pkg/reexec"
	"github.com/docker/docker/pkg/term"
	"github.com/moby/buildkit/util/apicaps"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// 新的守护进程命令
func newDaemonCommand() *cobra.Command {
	// 新的守护进程选项
	opts := newDaemonOptions(config.New())

	cmd := &cobra.Command{
		Use:           "dockerd [OPTIONS]",
		Short:         "A self-sufficient runtime for containers/容器的自足运行时环境.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cli.NoArgs,
		// Run函数
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.flags = cmd.Flags()
			// 运行守护进程
			return runDaemon(opts)
		},
		DisableFlagsInUseLine: true,
		Version:               fmt.Sprintf("%s, build %s", dockerversion.Version, dockerversion.GitCommit),
	}
	// 为根命令设置默认用法，帮助和错误处理
	cli.SetupRootCommand(cmd)

	flags := cmd.Flags()
	// docker version
	flags.BoolP("version", "v", false, "Print version information and quit")
	// 守护进程配置文件
	flags.StringVar(&opts.configFile, "config-file", defaultDaemonConfigFile, "Daemon configuration file")
	opts.InstallFlags(flags)
	installConfigFlags(opts.daemonConfig, flags)
	installServiceFlags(flags)

	return cmd
}

func init() {
	if dockerversion.ProductName != "" {
		apicaps.ExportedProduct = dockerversion.ProductName
	}
}

/**
 * 启动入口
 */
func main() {
	if reexec.Init() {
		return
	}

	// Set terminal emulation based on platform as required.
	// 基于平台设置终端仿真
	_, stdout, stderr := term.StdStreams()

	// @jhowardmsft - maybe there is a historic reason why on non-Windows, stderr is used
	// here. However, on Windows it makes no sense and there is no need.
	if runtime.GOOS == "windows" {
		logrus.SetOutput(stdout)
	} else {
		logrus.SetOutput(stderr)
	}

	// 新的守护进程命令
	cmd := newDaemonCommand()
	cmd.SetOutput(stdout)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(stderr, "%s\n", err)
		os.Exit(1)
	}
}
