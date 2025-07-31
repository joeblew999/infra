package pocketbase

import (
	"context"
	
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pocketbase",
	Short: "Start PocketBase server",
	Long:  `Start an embedded PocketBase server for database and API management`,
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")
		port := config.GetPocketBasePort()
		dataPath := config.GetPocketBaseDataPath()
		
		log.Info("Starting PocketBase server", 
			"port", port, 
			"env", env, 
			"data_path", dataPath)
		
		server := NewServer(env)
		server.SetDataDir(dataPath)
		
		ctx := context.Background()
		if err := server.Start(ctx); err != nil {
			log.Error("Failed to start PocketBase server", "error", err)
			panic(err)
		}
	},
}

func init() {
	Cmd.Flags().StringP("env", "e", "production", "Environment (development/production)")
}