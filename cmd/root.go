package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"lutonite.dev/gaps-cli/util"
	"os"
	"strings"
)

type ViperKey string

func viperKey(key string, flag string) ViperKey {
	k := ViperKey(key)
	flagMapping[flag] = k
	return k
}

func (k ViperKey) Key() string {
	return string(k)
}
func (k ViperKey) Flag() string {
	for flag, v := range flagMapping {
		if v == k {
			return flag
		}
	}

	return ""
}

const (
	envPrefix = "GAPS"
)

var (
	UrlViperKey               = viperKey("url", "url")
	GradesHistoryFileViperKey = viperKey("history.grades.file", "history")
	UsernameViperKey          = viperKey("login.username", "username")
	PasswordViperKey          = viperKey("login.password", "password")
	ScraperApiUrlViperKey     = viperKey("scraper.api.url", "api-url")
	ScraperApiKeyViperKey     = viperKey("scraper.api.key", "api-key")
	TokenValueViperKey        = viperKey("login.token.value", "")
	TokenStudentIdViperKey    = viperKey("login.token.studentId", "")
	TokenDateValueViperKey    = viperKey("login.token.generatedAt", "")

	flagMapping = make(map[string]ViperKey)
)

var (
	defaultViper     = viper.New()
	credentialsViper = viper.New()

	cfgFile     string
	credsFile   string
	loggerLevel string

	rootCmd = &cobra.Command{
		Use:   "gaps-cli",
		Short: "CLI for GAPS (Gaps is an Academical Planification System)",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initializeConfig(cmd)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			writeConfig()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "auth config file (default is $HOME/.config/gaps-cli/gaps.yaml)")
	rootCmd.PersistentFlags().StringVar(&credsFile, "credentials", "", "credentials config file (default is $HOME/.config/gaps-cli/credentials.yaml)")
	rootCmd.PersistentFlags().StringVar(&loggerLevel, "log-level", "error", "logging level")
	rootCmd.PersistentFlags().String(UrlViperKey.Flag(), "", "GAPS URL (default is https://gaps.heig-vd.ch/)")

	defaultViper.BindPFlag(UrlViperKey.Key(), rootCmd.PersistentFlags().Lookup(UrlViperKey.Flag()))
	defaultViper.SetDefault(UrlViperKey.Key(), "https://gaps.heig-vd.ch")
}

func initializeConfig(cmd *cobra.Command) {
	if loggerLevel != "" {
		level, err := log.ParseLevel(loggerLevel)
		util.CheckErr(err)
		log.SetLevel(level)
		log.Tracef("log level set to %s", level)
	}

	configDir := getConfigDirectory()
	initViper(cmd, defaultViper, "gaps", configDir, cfgFile)
	initViper(cmd, credentialsViper, "credentials", configDir, credsFile)
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

func initViper(cmd *cobra.Command, v *viper.Viper, name string, configDir string, path string) {
	if path != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(configDir)
		v.SetConfigType("yaml")
		v.SetConfigName("gaps-cli/" + name)
	}

	log.Debugf("writing config file %s", v.ConfigFileUsed())
	if err := v.SafeWriteConfig(); err != nil {
		util.CheckErrExcept(err, viper.ConfigFileAlreadyExistsError(""))
	}

	if err := v.ReadInConfig(); err == nil {
		log.WithField("file", v.ConfigFileUsed()).Infof("Reading global config file")
	}

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	bindFlags(cmd, v)
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := flagMapping[f.Name]
		if !f.Changed && v.IsSet(configName.Key()) {
			val := v.Get(configName.Key())
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func writeConfig() {
	defaultViper.WriteConfig()
	credentialsViper.WriteConfig()
}
