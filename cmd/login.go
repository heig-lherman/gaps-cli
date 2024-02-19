package cmd

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
	"lutonite.dev/gaps-cli/gaps"
	"lutonite.dev/gaps-cli/util"
	"os"
	"time"
)

const (
	ServiceKey        = "gaps-cli"
	KeyringTokenKey   = "gaps.token"
	UsernameViperKey  = "login.username"
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
			if token, _ := getKeyringValue(KeyringTokenKey); token != "" && !loginOpts.changePassword {
				// Default session duration is 6 hours on GAPS
				if !isTokenExpired() {
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

			if pwd, _ := getKeyringValue(username); pwd == "" || loginOpts.changePassword {
				fmt.Print("Enter your HEIG-VD einet AAI password: ")
				passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				password = string(passwordBytes)
				fmt.Println("ok")
				util.CheckErr(err)
			} else {
				var err error
				password, err = getKeyringValue(username)
				util.CheckErr(err)
			}

			refreshToken(username, password)
		},
	}
)

func init() {
	loginCmd.Flags().StringP("username", "u", "", "einet aai username (if not provided, you will be prompted to enter it)")
	loginCmd.Flags().String("password", "", "einet aai password (if not provided, you will be prompted to enter it)")
	loginCmd.Flags().BoolVar(&loginOpts.changePassword, "clear-password", false, "reset the password stored in the config file (if any)")

	viper.BindPFlag(UsernameViperKey, loginCmd.Flags().Lookup("username"))
	viper.SetDefault(TokenStudentIdKey, -1)
	viper.SetDefault(TokenDateValueKey, time.Now().UnixMilli())

	rootCmd.AddCommand(loginCmd)
}

func isTokenExpired() bool {
	return time.Now().UnixMilli()-viper.GetInt64(TokenDateValueKey) > 6*60*60*1000
}

func refreshToken(username string, password string) {
	viper.Set(UsernameViperKey, username)
	err := keyring.Set(ServiceKey, username, password)
	util.CheckErr(err)

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

	err = keyring.Set(ServiceKey, KeyringTokenKey, token)
	util.CheckErr(err)

	viper.Set(TokenStudentIdKey, studentId)
	viper.Set(TokenDateValueKey, time.Now().UnixMilli())

	log.Debug("saving config")
	viper.WriteConfig()
}

func buildTokenClientConfiguration() *gaps.TokenClientConfiguration {
	token, _ := getKeyringValue(KeyringTokenKey)
	if token == "" {
		log.Fatal("No token found, please login first")
	}

	// if token is expired, refresh it
	if isTokenExpired() {
		log.Info("Token expired, attempting refresh")

		pwd, err := keyring.Get(ServiceKey, viper.GetString(UsernameViperKey))
		util.CheckErr(err)

		refreshToken(viper.GetString(UsernameViperKey), pwd)
	}

	cfg := new(gaps.TokenClientConfiguration)
	cfg.InitToken(viper.GetString(UrlViperKey), token, viper.GetUint(TokenStudentIdKey))

	return cfg
}

func getKeyringValue(key string) (string, error) {
	secret, err := keyring.Get(ServiceKey, key)

	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return secret, nil
}
