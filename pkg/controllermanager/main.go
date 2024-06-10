/*
 * SPDX-FileCopyrightText: 2019 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package controllermanager

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/gardener/controller-manager-library/pkg/config"
	"github.com/gardener/controller-manager-library/pkg/configmain"
	areacfg "github.com/gardener/controller-manager-library/pkg/controllermanager/config"
	"github.com/gardener/controller-manager-library/pkg/ctxutil"
	"github.com/gardener/controller-manager-library/pkg/logger"
	"github.com/gardener/controller-manager-library/pkg/run"
)

const DeletionActivity = "DeletionActivity"

var Version = "dev-version"

func Start(use, short, long string, modifiers ...ConfigurationModifier) {
	c := PrepareStart(use, long)
	for _, m := range modifiers {
		c = m(c)
	}
	c.Start(use, short)
}

func PrepareStart(use, long string) Configuration {
	args := strings.Split(use, " ")
	return Configure(args[0], long, nil).ByDefault()
}

func (this Configuration) Start(use, short string) {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	def := this.Definition()
	long := def.GetDescription()
	var (
		cctx = ctxutil.CancelContext(ctxutil.WaitGroupContext(context.Background(), "main"))
		ctx  = ctxutil.TickContext(cctx, DeletionActivity)
		c    = make(chan os.Signal, 2)
		t    = make(chan os.Signal, 2)
	)

	signal.Notify(t, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		cnt := 0
	loop:
		for {
			select {
			case <-c:
				cnt++
				if cnt == 2 {
					break loop
				}
				logger.Infof("process is being terminated")
				ctxutil.Cancel(ctx)
			case <-t:
				cnt++
				if cnt == 2 {
					break loop
				}
				grace := areacfg.GracePeriod
				if grace > 0 {
					logger.Infof("process is being terminated with grace period for cleanup")
					go ctxutil.CancelAfterInactivity(ctx, DeletionActivity, grace)
				} else {
					logger.Infof("process is being terminated without grace period")
					ctxutil.Cancel(ctx)
				}
			}
		}
		logger.Infof("process is aborted immediately")
		os.Exit(0)
	}()

	//	if err := plugins.HandleCommandLine("--plugin-file", os.Args); err != nil {
	//		panic(err)
	//	}

	command := NewCommand(ctx, use, short, long, def)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}

	var gracePeriod = 120 * time.Second
	logger.Infof("waiting for everything to shutdown (max. %d seconds)", gracePeriod/time.Second)
	ctxutil.WaitGroupWait(ctx, gracePeriod, "main")
	logger.Infof("%s exits.", use)
}

func NewCommand(ctx context.Context, use, short, long string, def Definition) *cobra.Command {
	ctx, cfg := configmain.WithConfig(ctx, nil)
	def.ExtendConfig(cfg)
	fileName := ""
	cmd := &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Version: Version,
	}
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		if fileName != "" {
			logger.Infof("reading config from file %q", fileName)
			if err := config.MergeConfigFile(fileName, cmd.Flags(), false); err != nil {
				return fmt.Errorf("invalid config file %q; %s", fileName, err)
			}
		}
		if err := runCM(ctx, def); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return nil
	}

	cfg.AddToCommand(cmd)
	cmd.Flags().StringVarP(&fileName, "config", "", "", "config file")
	return cmd
}

func runCM(ctx context.Context, def Definition) error {
	return run.Run(ctx, func() error {
		logger.Infof("starting controller manager")
		controllerManager, err := NewControllerManager(ctx, def)
		if err != nil {
			return err
		}
		return controllerManager.Run()
	})
}
