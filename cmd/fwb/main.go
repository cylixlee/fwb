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
	"github.com/cylixlee/fwb/internal/handler"
	"github.com/cylixlee/fwb/internal/infra"
	"github.com/cylixlee/fwb/internal/infra/flutter"
	"github.com/cylixlee/fwb/internal/srv"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	exitCode := consts.ExitFailure

	app := fx.New(
		fx.NopLogger,
		fx.Provide(
			infra.NewConfigStore,
			infra.NewWindowsHost,
			flutter.NewCLI,
			infra.NewDownloader,
			flutter.NewReleaseSource,
			infra.NewSDKStore,
			infra.NewProjectMirrorFactory,
			infra.NewDeviceRepository,
			srv.NewBridge,
			handler.NewCLI,
		),
		fx.Invoke(func(h *handler.CLI) {
			exitCode = h.Execute(ctx, os.Args[1:])
		}),
	)
	defer app.Stop(ctx)

	if err := app.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, consts.ErrorOutputFormat, err)
		os.Exit(consts.ExitFailure)
	}
	if exitCode != consts.ExitSuccess {
		os.Exit(exitCode)
	}
}
