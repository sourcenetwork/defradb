package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping defradb to test an API connection",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbaddr := viper.GetString("database.address")
		if dbaddr == "" {
			log.Error("No database url provided")
		}
		if !strings.HasPrefix(dbaddr, "http") {
			dbaddr = "http://" + dbaddr
		}

		log.Info("Sending ping...")
		res, err := http.Get(fmt.Sprintf("%s/ping", dbaddr))
		if err != nil {
			log.Error("request failed: ", err)
			return
		}
		defer res.Body.Close()
		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Error("request failed: ", err)
			return
		}
		if string(buf) == "pong" {
			log.Info("Success!")
		} else {
			log.Error("Unexpected result: ", string(buf))
		}
	},
}

func init() {
	clientCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
