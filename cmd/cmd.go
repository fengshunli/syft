package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/anchore/stereoscope"
	"github.com/anchore/syft/internal/config"
	"github.com/anchore/syft/internal/log"
	"github.com/anchore/syft/internal/logger"
	"github.com/anchore/syft/syft"
	"github.com/gookit/color"
	"github.com/spf13/viper"
	"github.com/wagoodman/go-partybus"
)

var (
	appConfig         *config.Application
	eventBus          *partybus.Bus
	eventSubscription *partybus.Subscription
)

func init() {
	cobra.OnInitialize(
		initCmdAliasBindings,
		initAppConfig,
		initLogging,
		logAppConfig,
		initEventBus,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.Red.Sprint(err.Error()))
		os.Exit(1)
	}
}

// we must setup the config-cli bindings first before the application configuration is parsed. However, this cannot
// be done without determining what the primary command that the config options should be bound to since there are
// shared concerns (the root-packages alias).
func initCmdAliasBindings() {
	activeCmd, _, err := rootCmd.Find(os.Args[1:])
	if err != nil {
		panic(err)
	}

	// enable all cataloger by default if power-user command is run
	if activeCmd == powerUserCmd {
		config.PowerUserCatalogerEnabledDefault()
	}

	switch activeCmd {
	case packagesCmd, rootCmd:
		// note: we need to lazily bind config options since they are shared between both the root command
		// and the packages command. Otherwise there will be global viper state that is in contention.
		// See for more details: https://github.com/spf13/viper/issues/233 . Additionally, the bindings must occur BEFORE
		// reading the application configuration, which implies that it must be an initializer (or rewrite the command
		// initialization structure against typical patterns used with cobra, which is somewhat extreme for a
		// temporary alias)
		if err = bindPackagesConfigOptions(activeCmd.Flags()); err != nil {
			panic(err)
		}
	default:
		// even though the root command or packages command is NOT being run, we still need default bindings
		// such that application config parsing passes.
		if err = bindPackagesConfigOptions(packagesCmd.Flags()); err != nil {
			panic(err)
		}
	}
}

func initAppConfig() {
	cfg, err := config.LoadApplicationConfig(viper.GetViper(), persistentOpts)
	if err != nil {
		fmt.Printf("failed to load application config: \n\t%+v\n", err)
		os.Exit(1)
	}

	appConfig = cfg
}

func initLogging() {
	cfg := logger.LogrusConfig{
		EnableConsole: (appConfig.Log.FileLocation == "" || appConfig.CliOptions.Verbosity > 0) && !appConfig.Quiet,
		EnableFile:    appConfig.Log.FileLocation != "",
		Level:         appConfig.Log.LevelOpt,
		Structured:    appConfig.Log.Structured,
		FileLocation:  appConfig.Log.FileLocation,
	}

	logWrapper := logger.NewLogrusLogger(cfg)
	syft.SetLogger(logWrapper)
	stereoscope.SetLogger(&logger.LogrusNestedLogger{
		Logger: logWrapper.Logger.WithField("from-lib", "stereoscope"),
	})
}

func logAppConfig() {
	log.Debugf("application config:\n%+v", color.Magenta.Sprint(appConfig.String()))
}

func initEventBus() {
	eventBus = partybus.NewBus()
	eventSubscription = eventBus.Subscribe()

	stereoscope.SetBus(eventBus)
	syft.SetBus(eventBus)
}
