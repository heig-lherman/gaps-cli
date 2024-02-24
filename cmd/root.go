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
	defaultViper     = viper.New()
	credentialsViper = viper.New()

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

	defaultViper.BindPFlag(UrlViperKey, rootCmd.PersistentFlags().Lookup(UrlViperKey))
	defaultViper.SetDefault(UrlViperKey, "https://gaps.heig-vd.ch")
}

func getConfigDirectory() string {
	configDir, err := os.UserConfigDir()
	log.Debugf("host user config dir: %s", configDir)
	util.CheckErr(err)

	log.Debugf("creating config dir %s", configDir+"/gaps-cli")
	err = os.MkdirAll(configDir+"/gaps-cli", 0755)
	util.CheckErr(err)

	return configDir
}

func initViper(v *viper.Viper, name string) {
	v.AddConfigPath(getConfigDirectory())
	v.SetConfigType("yaml")
	v.SetConfigName("gaps-cli/" + name)
}

func bootstrapConfigFile(v *viper.Viper) {
	log.Debugf("writing config file %s", v.ConfigFileUsed())
	if err := v.SafeWriteConfig(); err != nil {
		util.CheckErrExcept(err, viper.ConfigFileAlreadyExistsError(""))
	}

	if err := v.ReadInConfig(); err == nil {
		log.WithField("file", v.ConfigFileUsed()).Infof("Reading global config file")
	}
}

func writeConfig() {
	defaultViper.WriteConfig()
	credentialsViper.WriteConfig()
}

func initConfig() {
	if loggerLevel != "" {
		level, err := log.ParseLevel(loggerLevel)
		util.CheckErr(err)
		log.SetLevel(level)
		log.Tracef("log level set to %s", level)
	}

	if cfgFile != "" {
		defaultViper.SetConfigFile(cfgFile)
	} else {
		initViper(defaultViper, "gaps")
	}

	initViper(credentialsViper, "credentials")

	defaultViper.SetEnvPrefix("gaps")
	defaultViper.AutomaticEnv()

	credentialsViper.SetEnvPrefix("gaps")
	credentialsViper.AutomaticEnv()

	bootstrapConfigFile(defaultViper)
	bootstrapConfigFile(credentialsViper)
}
