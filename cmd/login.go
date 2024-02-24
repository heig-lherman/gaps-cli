package cmd

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"lutonite.dev/gaps-cli/gaps"
	"lutonite.dev/gaps-cli/util"
	"os"
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
			if credentialsViper.GetString(TokenValueKey) != "" && !loginOpts.changePassword {
				// Default session duration is 6 hours on GAPS
				if !isTokenExpired() {
					log.Info("User already logged in, keeping existing token")
					return
				}

				log.Info("Token expired, attempting refresh")
			}

			var username string
			var password string

			if defaultViper.GetString(UsernameViperKey) == "" {
				fmt.Print("Enter your HEIG-VD einet AAI username: ")
				reader := bufio.NewReader(os.Stdin)
				un, err := reader.ReadString('\n')
				username = un[:len(un)-1]
				util.CheckErr(err)
			} else {
				username = defaultViper.GetString(UsernameViperKey)
			}

			if credentialsViper.GetString(PasswordViperKey) == "" || loginOpts.changePassword {
				fmt.Print("Enter your HEIG-VD einet AAI password: ")
				passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				password = string(passwordBytes)
				fmt.Println("ok")
				util.CheckErr(err)
			} else {
				password = credentialsViper.GetString(PasswordViperKey)
			}

			refreshToken(username, password)
		},
	}
)

func init() {
	loginCmd.Flags().StringP("username", "u", "", "einet aai username (if not provided, you will be prompted to enter it)")
	loginCmd.Flags().String("password", "", "einet aai password (if not provided, you will be prompted to enter it)")
	loginCmd.Flags().BoolVar(&loginOpts.changePassword, "clear-password", false, "reset the password stored in the config file (if any)")

	defaultViper.BindPFlag(UsernameViperKey, loginCmd.Flags().Lookup("username"))
	credentialsViper.BindPFlag(PasswordViperKey, loginCmd.Flags().Lookup("password"))
	credentialsViper.SetDefault(TokenValueKey, "")
	defaultViper.SetDefault(TokenStudentIdKey, -1)
	defaultViper.SetDefault(TokenDateValueKey, time.Now().UnixMilli())

	rootCmd.AddCommand(loginCmd)
}

func isTokenExpired() bool {
	return time.Now().UnixMilli()-defaultViper.GetInt64(TokenDateValueKey) > 6*60*60*1000
}

func refreshToken(username string, password string) {
	defaultViper.Set(UsernameViperKey, username)
	credentialsViper.Set(PasswordViperKey, password)

	cfg := new(gaps.ClientConfiguration)
	cfg.Init(defaultViper.GetString(UrlViperKey))

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

	credentialsViper.Set(TokenValueKey, token)
	defaultViper.Set(TokenStudentIdKey, studentId)
	defaultViper.Set(TokenDateValueKey, time.Now().UnixMilli())

	log.Debug("saving config")
	writeConfig()
}

func buildTokenClientConfiguration() *gaps.TokenClientConfiguration {
	if credentialsViper.GetString(TokenValueKey) == "" {
		log.Fatal("No token found, please login first")
	}

	// if token is expired, refresh it
	if isTokenExpired() {
		log.Info("Token expired, attempting refresh")
		refreshToken(defaultViper.GetString(UsernameViperKey), credentialsViper.GetString(PasswordViperKey))
	}

	cfg := new(gaps.TokenClientConfiguration)
	cfg.InitToken(defaultViper.GetString(UrlViperKey), credentialsViper.GetString(TokenValueKey), defaultViper.GetUint(TokenStudentIdKey))

	return cfg
}
