package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"lutonite.dev/gaps-cli/util"
	"os"
)

const (
	UrlViperKey = "url"
)

var (
	cfgFile     string
	loggerLevel string

	rootCmd = &cobra.Command{
		Use:   "gaps-cli",
		Short: "CLI for GAPS (Gaps is an Academical Planification System)",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "auth config file (default is $HOME/.config/gaps-cli/gaps.yaml)")
	rootCmd.PersistentFlags().StringVar(&loggerLevel, "log-level", "error", "logging level")
	rootCmd.PersistentFlags().String(UrlViperKey, "", "GAPS URL (default is https://gaps.heig-vd.ch/)")

	viper.BindPFlag(UrlViperKey, rootCmd.PersistentFlags().Lookup(UrlViperKey))
	viper.SetDefault(UrlViperKey, "https://gaps.heig-vd.ch")
}

func initConfig() {
	if loggerLevel != "" {
		level, err := log.ParseLevel(loggerLevel)
		util.CheckErr(err)
		log.SetLevel(level)
		log.Tracef("log level set to %s", level)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := os.UserConfigDir()
		log.Debugf("host user config dir: %s", configDir)
		util.CheckErr(err)

		log.Debugf("creating config dir %s", configDir+"/gaps-cli")
		err = os.MkdirAll(configDir+"/gaps-cli", 0755)
		util.CheckErr(err)

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("gaps-cli/gaps")
	}

	viper.SetEnvPrefix("gaps")
	viper.AutomaticEnv()

	log.Debugf("writing config file %s", viper.ConfigFileUsed())
	if err := viper.SafeWriteConfig(); err != nil {
		util.CheckErrExcept(err, viper.ConfigFileAlreadyExistsError(""))
	}

	if err := viper.ReadInConfig(); err == nil {
		log.WithField("file", viper.ConfigFileUsed()).Infof("Reading global config file")
	}
}
