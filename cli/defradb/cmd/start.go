package cmd

import (
	"os"

	"github.com/sourcenetwork/defradb/db"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a DefraDB server ",
	Long: `Start a new instance of the DefraDB:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting DefraDB process...")
		db, err := db.NewDB(&config.Database)
		if err != nil {
			log.Error("Failed to initiate database:", err)
			os.Exit(1)
		}

		db.Listen()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
