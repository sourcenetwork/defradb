package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	queryStr string
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "A brief description of your command",
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

		if len(args) != 1 {
			log.Fatal("needs a single query argument")
		}
		query := args[0]
		if query == "" {
			log.Error("missing query")
			return
		}
		endpointStr := fmt.Sprintf("%s/graphql", dbaddr)
		endpoint, err := url.Parse(endpointStr)
		if err != nil {
			log.Fatal(err)
		}

		p := url.Values{}
		p.Add("query", query)
		endpoint.RawQuery = p.Encode()

		res, err := http.Get(endpoint.String())
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

		fmt.Printf("Response: %s", string(buf))
	},
}

func init() {
	clientCmd.AddCommand(queryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queryCmd.Flags().StringVar(&queryStr, "query", "", "Query to run on the database")
}
