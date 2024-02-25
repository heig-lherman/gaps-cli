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

type LoginCmdOpts struct {
	changePassword bool
}

var (
	loginOpts = &LoginCmdOpts{}
	loginCmd  = &cobra.Command{
		Use:   "login",
		Short: "Allows to login to GAPS for future commands",
		Run: func(cmd *cobra.Command, args []string) {
			if credentialsViper.GetString(TokenValueViperKey.Key()) != "" && !loginOpts.changePassword {
				// Default session duration is 6 hours on GAPS
				if !isTokenExpired() {
					log.Info("User already logged in, keeping existing token")
					return
				}

				log.Info("Token expired, attempting refresh")
			}

			var username string
			var password string

			if defaultViper.GetString(UsernameViperKey.Key()) == "" {
				fmt.Print("Enter your HEIG-VD einet AAI username: ")
				reader := bufio.NewReader(os.Stdin)
				un, err := reader.ReadString('\n')
				username = un[:len(un)-1]
				util.CheckErr(err)
			} else {
				username = defaultViper.GetString(UsernameViperKey.Key())
			}

			if credentialsViper.GetString(PasswordViperKey.Key()) == "" || loginOpts.changePassword {
				fmt.Print("Enter your HEIG-VD einet AAI password: ")
				passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				password = string(passwordBytes)
				fmt.Println("ok")
				util.CheckErr(err)
			} else {
				password = credentialsViper.GetString(PasswordViperKey.Key())
			}

			refreshToken(username, password)
		},
	}
)

func init() {
	loginCmd.Flags().BoolVar(&loginOpts.changePassword, "clear-password", false, "reset the password stored in the config file (if any)")

	loginCmd.Flags().StringP(UsernameViperKey.Flag(), "u", "", "einet aai username (if not provided, you will be prompted to enter it)")
	defaultViper.BindPFlag(UsernameViperKey.Key(), loginCmd.Flags().Lookup(UsernameViperKey.Flag()))

	loginCmd.Flags().String(PasswordViperKey.Flag(), "", "einet aai password (if not provided, you will be prompted to enter it)")
	credentialsViper.BindPFlag(PasswordViperKey.Key(), loginCmd.Flags().Lookup(PasswordViperKey.Flag()))

	credentialsViper.SetDefault(TokenValueViperKey.Key(), "")
	defaultViper.SetDefault(TokenStudentIdViperKey.Key(), -1)
	defaultViper.SetDefault(TokenDateValueViperKey.Key(), time.Now().UnixMilli())

	rootCmd.AddCommand(loginCmd)
}

func isTokenExpired() bool {
	return time.Now().UnixMilli()-defaultViper.GetInt64(TokenDateValueViperKey.Key()) > 6*60*60*1000
}

func refreshToken(username string, password string) {
	defaultViper.Set(UsernameViperKey.Key(), username)
	credentialsViper.Set(PasswordViperKey.Key(), password)

	cfg := new(gaps.ClientConfiguration)
	cfg.Init(defaultViper.GetString(UrlViperKey.Key()))

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

	credentialsViper.Set(TokenValueViperKey.Key(), token)
	defaultViper.Set(TokenStudentIdViperKey.Key(), studentId)
	defaultViper.Set(TokenDateValueViperKey.Key(), time.Now().UnixMilli())

	log.Debug("saving config")
	writeConfig()
}

func buildTokenClientConfiguration() *gaps.TokenClientConfiguration {
	if credentialsViper.GetString(TokenValueViperKey.Key()) == "" {
		log.Fatal("No token found, please login first")
	}

	// if token is expired, refresh it
	if isTokenExpired() {
		log.Info("Token expired, attempting refresh")
		refreshToken(defaultViper.GetString(UsernameViperKey.Key()), credentialsViper.GetString(PasswordViperKey.Key()))
	}

	cfg := new(gaps.TokenClientConfiguration)
	cfg.InitToken(
		defaultViper.GetString(UrlViperKey.Key()),
		credentialsViper.GetString(TokenValueViperKey.Key()),
		defaultViper.GetUint(TokenStudentIdViperKey.Key()),
	)

	return cfg
}
