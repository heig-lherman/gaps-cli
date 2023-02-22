package cmd

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"lutonite.dev/gaps-cli/gaps"
	"lutonite.dev/gaps-cli/util"
	"os"
	"syscall"
	"time"
)

const (
	UsernameViperKey  = "login.username"
	PasswordViperKey  = "login.password"
	TokenValueKey     = "login.token.value"
	TokenStudentIdKey = "login.token.studentId"
	TokenDateValueKey = "login.token.generatedAt"
)

type LoginCmdOpts struct {
	changePassword bool
}

var (
	loginOpts = &LoginCmdOpts{}
	loginCmd  = &cobra.Command{
		Use:   "login",
		Short: "Allows to login to GAPS for future commands",
		Run: func(cmd *cobra.Command, args []string) {
			if viper.GetString(TokenValueKey) != "" {
				// Default session duration is 6 hours on GAPS
				if time.Now().UnixMilli()-viper.GetInt64(TokenDateValueKey) < 6*60*60*1000 {
					log.Info("User already logged in, keeping existing token")
					return
				}

				log.Info("Token expired, attempting refresh")
			}

			var username string
			var password string

			if viper.GetString(UsernameViperKey) == "" {
				fmt.Print("Enter your HEIG-VD einet AAI username: ")
				reader := bufio.NewReader(os.Stdin)
				un, err := reader.ReadString('\n')
				username = un[:len(un)-1]
				util.CheckErr(err)
			} else {
				username = viper.GetString(UsernameViperKey)
			}

			if viper.GetString(PasswordViperKey) == "" || loginOpts.changePassword {
				fmt.Print("Enter your HEIG-VD einet AAI password: ")
				passwordBytes, err := term.ReadPassword(syscall.Stdin)
				password = string(passwordBytes)
				fmt.Println("ok")
				util.CheckErr(err)
			} else {
				password = viper.GetString(PasswordViperKey)
			}

			viper.Set(UsernameViperKey, username)
			viper.Set(PasswordViperKey, password)

			cfg := new(gaps.ClientConfiguration)
			cfg.Init(viper.GetString(UrlViperKey))

			log.Debug("fetching token...")
			login := gaps.NewLoginAction(cfg, username, password)
			token, err := login.FetchToken()
			util.CheckErr(err)

			log.Debug("fetching student id...")
			studentId, err := login.FetchStudentId(token)
			util.CheckErr(err)

			log.Info("Successfully logged in")
			log.Tracef("Token: %s", token)
			log.Tracef("Student Id: %d", studentId)

			viper.Set(TokenValueKey, token)
			viper.Set(TokenStudentIdKey, studentId)
			viper.Set(TokenDateValueKey, time.Now().UnixMilli())

			log.Debug("saving config")
			viper.WriteConfig()
		},
	}
)

func init() {
	loginCmd.Flags().StringP("username", "u", "", "einet aai username (if not provided, you will be prompted to enter it)")
	loginCmd.Flags().String("password", "", "einet aai password (if not provided, you will be prompted to enter it)")
	loginCmd.Flags().BoolVar(&loginOpts.changePassword, "clear-password", false, "reset the password stored in the config file (if any)")

	viper.BindPFlag(UsernameViperKey, loginCmd.Flags().Lookup("username"))
	viper.BindPFlag(PasswordViperKey, loginCmd.Flags().Lookup("password"))
	viper.SetDefault(TokenValueKey, "")
	viper.SetDefault(TokenStudentIdKey, -1)
	viper.SetDefault(TokenDateValueKey, time.Now().UnixMilli())

	rootCmd.AddCommand(loginCmd)
}

func buildTokenClientConfiguration() *gaps.TokenClientConfiguration {
	cfg := new(gaps.TokenClientConfiguration)
	cfg.InitToken(viper.GetString(UrlViperKey), viper.GetString(TokenValueKey), viper.GetUint(TokenStudentIdKey))
	return cfg
}
