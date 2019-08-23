package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
var requestIP string

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

/*
	// If no filename given as argument, read from stdin. Allows use as piped tool.
	flag.Parse()
	var data []byte
	var err error
	switch flag.NArg() {
	case 0:
		data, err = ioutil.ReadAll(os.Stdin)
		check(err)
		fmt.Printf("stdin data: %v\n", string(data))
		break
	case 1:
		data, err = ioutil.ReadFile(flag.Arg(0))
		check(err)
		fmt.Printf("file data: %v\n", string(data))
		break
	default:
		fmt.Printf("input must be from stdin or file\n")
		os.Exit(1)
	}

*/

// RootCmd is the root command for limo
var RootCmd = &cobra.Command{
	Use:     "investigo",
	Short:   "investigos",
	Long:    `investigo.`,
	Version: config.Version,
	Run: func(cmd *cobra.Command, args []string) {

		// cat accounts.txt | go run main.go -t -f all
		// https://gist.github.com/mashbridge/4365101
		// pipe... cf upper

		if len(args) > 0 {
			// query := args[0]
			// pp.Println("usernames: ", query)
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

		requestIP = getContent("https://api.ipify.org")
		options.RequestIP = requestIP
		if options.WithTor {
			pp.Println("Proxy IP: ", requestIP)
		}

		for _, username := range args {
			var output *tablib.Dataset
			if options.WithExport != "" {
				output = tablib.NewDataset([]string{"Status", "Username", "Site", "Info", "RequestIP"})
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
					WriteResult(res)
					if options.WithExport != "" {
						if res.Exist || res.Err {
							output.AppendValues(res.Exist, username, site, res.ErrMsg, requestIP)
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
				prefixPath := filepath.Join("data", "collector", name)
				err := os.MkdirAll(prefixPath, 0700)
				if err != nil {
					return
				}

				var ods *tablib.Exportable
				// var err error
				ds := DBook.Sheet(name).Dataset()
				var validOutput bool
				if options.WithFormat == "all" {
					options.WithFormat = "json,yaml,xml,xlsx,csv,tsv,html,tabularmarkdown,tabulargrid,tabularcondensed,tabularsimple,postgres,mysql"
				}
				// pp.Println("options.WithFormat", options.WithFormat)
				formats := strings.Split(options.WithFormat, ",")
				for _, format := range formats {
					// format := strings.ToLower(options.WithFormat)
					outputFile := name + "." + format
					if strings.HasPrefix(format, "tabular") {
						outputFile = name + "." + strings.ToLower(format) + ".md"
					}
					outputPath := filepath.Join(prefixPath, outputFile)
					// pp.Println("prefixPath:", prefixPath)
					// pp.Println("format:", format)
					// pp.Println("outputPath:", outputPath)
					switch strings.ToLower(format) {
					// missing: csv, tsv, XLSXL, html
					case "yaml":
						ods, err = ds.YAML()
						if err != nil {
							continue
						}
						validOutput = true
					case "json":
						ods, err = ds.JSON()
						if err != nil {
							continue
						}
						validOutput = true
					case "xml":
						ods, err = ds.XML()
						if err != nil {
							continue
						}
						validOutput = true
					case "xlsx":
						ods, err = ds.XLSX()
						if err != nil {
							continue
						}
					case "csv":
						ods, err = ds.CSV()
						if err != nil {
							continue
						}
					case "tsv":
						ods, err = ds.TSV()
						if err != nil {
							continue
						}
					case "html":
						ods = ds.HTML()
						validOutput = true

					case "tabularmarkdown":
						ods = ds.Tabular("markdown" /* tablib.TabularMarkdown */)
						validOutput = true
					case "tabulargrid":
						ods = ds.Tabular("grid" /* tablib.TabularMarkdown */)
						validOutput = true
					case "tabularcondensed":
						ods = ds.Tabular("condensed" /* tablib.TabularMarkdown */)
						validOutput = true
					case "tabularsimple":
						ods = ds.Tabular("simple" /* tablib.TabularMarkdown */)
						validOutput = true
					case "mysql":
						ods = ds.MySQL("investigo")
						validOutput = true
					case "postgres":
						ods = ds.Postgres("investigo")
						validOutput = true
					}
					if validOutput {
						ods.WriteFile(outputPath, 0600)
					}
				}
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

/*
func WriteToFile(filepath, filename, format, data string) error {
	err := os.MkdirAll(filepath, 0700)
	if err != nil {
		return err
	}
	outputFile := fmt.Printf("%s.%s", filename, format)
	ioutil.WriteFile(outputFile, data, 0600)
	return nil
}
*/

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

// https://github.com/gjbae1212/go-module/tree/master/ip
// https://github.com/twexler/gd-ddns-client/blob/master/ipifyapi.go
func getContent(url string) string {
	r, err := service.Request(url, options)
	if err != nil || r.StatusCode != 200 {
		fmt.Println("errpr:", err)
		panic("Failed to connect to Investigo repository.")
	} else {
		defer r.Body.Close()
	}
	// pp.Println(service.ReadResponseBody(r))
	return service.ReadResponseBody(r)
}

func init() {
	flags := RootCmd.PersistentFlags()
	flags.BoolVarP(&options.NoColor, "no-color", "n", false, "no color")
	flags.BoolVarP(&options.WithTor, "tor", "t", false, "use tor as a proxy")
	flags.StringVarP(&options.WithTorAddress, "tor-address", "p", "", "tor proxy address")
	flags.StringVarP(&options.WithExport, "export", "e", "data/{username}", "export file base name")
	flags.StringVarP(&options.WithFormat, "format", "f", "yaml", "export format (available: JSON, YAML, CSV, TSV, XLSX, Postgres, MySQL, TabularMarkdown, TabularGrid, TabularSimple, TabularCondensed.")
	flags.BoolVarP(&options.WithFormatAll, "format-all", "a", false, "export to all available formats.")
	flags.BoolVarP(&options.WithHttpCache, "cache", "c", false, "use http cache module")
	flags.BoolVarP(&options.WithAdmin, "webui", "i", false, "webui interface")
	flags.BoolVarP(&options.CheckForUpdate, "update", "u", false, "check for updates")
	flags.BoolVarP(&options.Verbose, "verbose", "v", false, "verbose output")
}
