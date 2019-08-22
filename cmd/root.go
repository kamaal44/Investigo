package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/agrison/go-tablib"
	color "github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	"github.com/qor/admin"
	"github.com/spf13/cobra"

	"github.com/lucmski/Investigo/config"
	"github.com/lucmski/Investigo/model"
	"github.com/lucmski/Investigo/service"
)

var options config.Options
var configuration *config.Config
var DB *gorm.DB

//
var DBook *tablib.Databook

var (
	guard     = make(chan int, config.MaxGoroutines)
	waitGroup = &sync.WaitGroup{}
	logger    = log.New(color.Output, "", 0)
	siteData  = map[string]model.SiteData{}
	options2  struct {
		noColor         bool
		updateBeforeRun bool
		withTor         bool
		withTorAddress  string
		withAdmin       bool
		withExport      string
		withFormat      string
		verbose         bool
		checkForUpdate  bool
	}
)

// RootCmd is the root command for limo
var RootCmd = &cobra.Command{
	Use:     "investigo",
	Short:   "investigos",
	Long:    `investigo.`,
	Version: config.Version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			query := args[0]
			pp.Println("usernames: ", query)
		} else {
			log.Fatal("You must specify a username at least...")
		}

		// Loads site data from sherlock database and assign to a variable.
		initializeSiteData(options.CheckForUpdate)

		// Loads extra site data
		initializeExtraSiteData()

		if options.WithAdmin {
			DB, _ = gorm.Open("sqlite3", "investigo.db")
			DB.AutoMigrate(&model.Result{})
			// Initalize
			Admin = admin.New(&admin.AdminConfig{DB: DB})

			// Allow to use Admin to manage User, Product
			Admin.AddResource(&model.Result{})
			// Admin.AddResource(&SiteData{})
		}

		if options.WithExport != "" {
			DBook = tablib.NewDatabook()
		}

		for _, username := range args {
			var output *tablib.Dataset
			if options.WithExport != "" {
				output = tablib.NewDataset([]string{"Status", "Username", "Site", "Info"})
			}
			if options.NoColor {
				fmt.Printf("Investigating %s on:\n", username)
			} else {
				fmt.Fprintf(color.Output, "Investigating %s on:\n", color.HiGreenString(username))
			}
			waitGroup.Add(len(siteData))
			for site := range siteData {
				guard <- 1
				go func(site string) {
					defer waitGroup.Done()
					res := service.Lookup(username, site, siteData[site], options)
					//if !options.withExport {
					WriteResult(res)
					//}
					if options.WithExport != "" {
						if res.Exist || res.Err {
							output.AppendValues(res.Exist, username, site, res.ErrMsg)
						}
					}
					<-guard
				}(site)
			}
			waitGroup.Wait()
			if options.WithExport != "" {
				DBook.AddSheet(username, output)
			}
		}
		if options.WithExport != "" {
			// fmt.Println(DBook.YAML())
			for name := range DBook.Sheets() {
				ods := DBook.Sheet(name).Dataset().Tabular("markdown" /* tablib.TabularMarkdown */)
				fmt.Println(ods)
			}
		}
		if options.WithAdmin {
			// initalize an HTTP request multiplexer
			mux := http.NewServeMux()

			// Mount admin interface to mux
			Admin.MountTo("/admin", mux)
			fmt.Println("Listening on: 9000")
			http.ListenAndServe(":9000", mux)
		}

	},
}

// WriteResult writes investigation result to stdout and file
func WriteResult(result model.Result) {

	if options.NoColor {
		if result.Exist {
			logger.Printf("[%s] %s: %s\n", ("+"), result.Site, result.Link)
		} else {
			if result.Err {
				logger.Printf("[%s] %s: ERROR: %s", ("!"), result.Site, (result.ErrMsg))
			} else if options.Verbose {
				logger.Printf("[%s] %s: %s", ("-"), result.Site, ("Not Found!"))
			}
		}
	} else {
		if result.Exist {
			logger.Printf("[%s] %s: %s\n", color.HiGreenString("+"), color.HiWhiteString(result.Site), result.Link)
			if options.WithAdmin {
				if err := DB.Create(&result).Error; err != nil {
					fmt.Println(err)
					return
				}
			}
		} else {
			if result.Err {
				logger.Printf("[%s] %s: %s: %s", color.HiRedString("!"), result.Site, color.HiMagentaString("ERROR"), color.HiRedString(result.ErrMsg))
				if options.WithAdmin {
					if err := DB.Create(&result).Error; err != nil {
						fmt.Println(err)
						return
					}
				}
			} else if options.Verbose {
				logger.Printf("[%s] %s: %s", color.HiRedString("-"), result.Site, color.HiYellowString("Not Found!"))
			}
		}
	}

	return
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	flags := RootCmd.PersistentFlags()
	flags.BoolVarP(&options.NoColor, "no-color", "n", false, "no color")
	flags.BoolVarP(&options.WithTor, "tor", "t", false, "use tor proxy")
	flags.StringVarP(&options.WithTorAddress, "tor-address", "p", "", "use tor proxy")
	flags.StringVarP(&options.WithExport, "export", "e", "exportfile", "export file base name")
	flags.StringVarP(&options.WithFormat, "format", "f", "yaml", "export format")
	flags.BoolVarP(&options.WithAdmin, "webui", "i", false, "webui interface")
	flags.BoolVarP(&options.CheckForUpdate, "update", "u", false, "check for updates")
	flags.BoolVarP(&options.Verbose, "verbose", "v", false, "verbose output")
}
