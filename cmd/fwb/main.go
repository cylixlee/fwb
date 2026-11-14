// Command fwb is a single-binary bridge that exposes a Windows desktop Flutter
// app as a Flutter custom device to a Flutter SDK running inside WSL.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/handler/cli"
	"github.com/cylixlee/fwb/internal/infra"
	"github.com/cylixlee/fwb/internal/infra/flutter"
	"github.com/cylixlee/fwb/internal/infra/mirror"
	"github.com/cylixlee/fwb/internal/infra/powershell"
	"github.com/cylixlee/fwb/internal/srv"
)

func main() {
	os.Exit(execute())
}

// execute runs one command and returns its exit status. The deferred teardown
// has to run before the process exits, so os.Exit stays in main.
func execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	exitCode := consts.ExitFailure

	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			infra.NewConfigStore,
			powershell.NewHost,
			flutter.NewCLI,
			infra.NewDownloader,
			flutter.NewReleaseSource,
			infra.NewSDKStore,
			mirror.NewFactory,
			infra.NewDeviceRepository,
			srv.NewSDKProvider,
			srv.NewLauncher,
			srv.NewDeviceHooks,
			srv.NewCleaner,
			srv.NewDiagnostics,
			cli.New,
		),
		fx.Invoke(func(h *cli.Struct) {
			exitCode = h.Execute(ctx, os.Args[1:])
		}),
	)
	defer app.Stop(ctx)

	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, consts.ErrorOutputFormat, err)
		return consts.ExitFailure
	}
	return exitCode
}
