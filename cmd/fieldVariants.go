/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// fieldVariantsCmd represents the fieldVariants command
var fieldVariantsCmd = &cobra.Command{
	Use:   "field-variants",
	Short: "Collect all type variants of specific field",
	Long:  "test",
	Run: func(cmd *cobra.Command, args []string) {

		variants := make(map[string]int, 0)
		field := args[1]
		paths, err := producePaths(args[0])
		if err != nil {
			cmd.ErrOrStderr().Write([]byte(err.Error()))
			panic(err.Error())
		}
		for _, targPath := range paths {
			handler, err := os.Open(targPath)
			if err != nil {
				cmd.ErrOrStderr().Write([]byte(err.Error()))
				continue
			}
			collectFieldVariants(handler, field, &variants)
		}
		fmt.Println(variants)
	},
}

func init() {
	rootCmd.AddCommand(fieldVariantsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fieldVariantsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fieldVariantsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
