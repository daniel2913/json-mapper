/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands

var rootCmd = &cobra.Command{
	Use:   "jschem",
	Short: "Get Go struct from json file",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one argument")
		}
		return nil
	},
	Long: `It is long`,

	Run: func(cmd *cobra.Command, args []string) {
		// var await sync.WaitGroup
		paths, err := producePaths(args[0])
		if err != nil {
			panic(err.Error())
		}
		schema := ConcSchema[any]{Schema: make(map[string]any, 30)}
		sideKicks := ConcSchema[any]{Schema: make(map[string]any)}
		for _, targPath := range paths {
			handler, err := os.Open(targPath)
			if err != nil {
				cmd.ErrOrStderr().Write([]byte(err.Error()))
			}
			// await.Add(1)
			// go func() {
			// 	defer await.Done()
			collectFields(handler, &schema, &sideKicks)
			// }()
		}
		// await.Wait()
		for key, value := range schema.Schema {
			fmt.Println(key, value)
		}
		for key, value := range sideKicks.Schema {
			fmt.Println(key, value)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jschem.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
