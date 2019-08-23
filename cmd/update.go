package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	color "github.com/fatih/color"
	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"

	"github.com/lucmski/Investigo/config"
	"github.com/lucmski/Investigo/service"
)

var UpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update ",
	Long:    "Update your local database.",
	Example: fmt.Sprintf("  %s update", config.ProgramName),
	Run: func(cmd *cobra.Command, args []string) {
		pp.Println("Updating your local database")
		initializeSiteData(true)
	},
}

func initializeSiteData(forceUpdate bool) {
	jsonFile, err := os.Open(config.DataFileName)
	if err != nil || forceUpdate {
		if options.NoColor {
			fmt.Printf(
				"%s Failed to read %s from current directory. %s",
				("->"),
				config.DataFileName,
				("Downloading..."),
			)
		} else {
			fmt.Fprintf(
				color.Output,
				"%s Failed to read %s from current directory. %s",
				color.HiRedString("->"),
				config.DataFileName,
				color.HiYellowString("Downloading..."),
			)
		}

		if forceUpdate {
			jsonFile.Close()
		}

		r, err := service.Request("https://raw.githubusercontent.com/sherlock-project/sherlock/master/data.json", options)
		if err != nil || r.StatusCode != 200 {
			if options.NoColor {
				fmt.Printf(" [%s]\n", ("Failed"))
			} else {
				fmt.Fprintf(color.Output, " [%s]\n", color.HiRedString("Failed"))
			}
			panic("Failed to connect to Investigo repository.")
		} else {
			defer r.Body.Close()
		}
		if _, err := os.Stat(config.DataFileName); !os.IsNotExist(err) {
			if err = os.Remove(config.DataFileName); err != nil {
				panic(err)
			}
		}
		_updateFile, _ := os.OpenFile(config.DataFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if _, err := _updateFile.WriteString(service.ReadResponseBody(r)); err != nil {
			if options.NoColor {
				fmt.Printf("Failed to update data.\n")
			} else {
				fmt.Fprintf(color.Output, color.RedString("Failed to update data.\n"))
			}
			panic(err)
		}

		_updateFile.Close()
		jsonFile, _ = os.Open(config.DataFileName)

		fmt.Println(" [Done]")
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic("Error while read " + config.DataFileName)
	} else {
		json.Unmarshal([]byte(byteValue), &siteData)
	}
	return
}

func init() {
	// RootCmd.AddCommand(UpdateCmd)
}
