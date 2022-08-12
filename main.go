package main

import (
	"fmt"
	"github.com/RyanSusana/archstats/snippets"
	"github.com/RyanSusana/archstats/views"
	"github.com/jessevdk/go-flags"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
)

type GeneralOptions struct {
	Args struct {
		RootDir string `description:"Root directory of the project" required:"true" positional-arg-name:"<project-directory>"`
	} `positional-args:"true" required:"true"`

	View     string `short:"v" long:"view" default:"directories-recursive" description:"Type of view to show" required:"true"`
	AllViews bool   `long:"all-views" description:"Show all views in JSON format."`

	Snippets []string `short:"s" long:"snippet" description:"Regular Expression to match snippet types. Snippet types are named by using regex named groups(?P<typeName>). For example, if you want to match a JavaScript function, you can use the regex 'function (?P<function>.*)'"`

	Extensions []string `short:"e" long:"extension"  description:"This option adds support for additional extensions. The value of this option is a comma separated list of extensions. The supported extensions are: php"`

	Columns []string `short:"c" long:"column" description:"When this option is present, it will only show columns in the comma-separated list of columns."`

	NoHeader bool `long:"no-header" description:"No header (only applicable for csv, tsv, table)"`

	SortedBy string `long:"sorted-by" description:"Sorted by column name. For number based columns, this is in descending order."`

	OutputFormat string `short:"o" long:"output-format" choice:"table" choice:"ndjson" choice:"json" choice:"csv" choice:"tsv" description:"Output format"`

	Profile struct {
		Cpu string `long:"cpu" description:"File to write CPU profile to"`
		Mem string `long:"mem" description:"File to write memory profile to"`
	} `group:"Profiling" hidden:"true" namespace:"profile"`
}

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	output, err := RunArchstats(os.Args[1:])
	if err != nil {
		exitCode = printError(err)
	} else {
		fmt.Println(output)
	}
}

func RunArchstats(args []string) (string, error) {
	generalOptions, err := getOptions(args)

	if err != nil {
		return "", err
	}
	// Enable cpu profiling if requested.
	if generalOptions.Profile.Cpu != "" {
		f, err := os.Create(generalOptions.Profile.Cpu)
		if err != nil {
			return "", err
		}
		defer f.Close() // TODO handle error
		if err := pprof.StartCPUProfile(f); err != nil {
			return "", err
		}
		defer pprof.StopCPUProfile()
	}

	output, err := runArchStats(generalOptions)

	// Enable memory profiling if requested.
	if generalOptions.Profile.Mem != "" {
		f, err := os.Create(generalOptions.Profile.Mem)
		if err != nil {
			return "", err
		}
		defer f.Close() // TODO handle error
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			return "", err
		}
	}
	if err != nil {
		return "", err
	} else {
		return output, nil
	}
}

func getOptions(args []string) (*GeneralOptions, error) {
	generalOptions := &GeneralOptions{}
	_, err := flags.NewParser(generalOptions, flags.Default|flags.IgnoreUnknown).ParseArgs(args)
	return generalOptions, err
}
func runArchStats(generalOptions *GeneralOptions) (string, error) {
	generalOptions.Args.RootDir, _ = filepath.Abs(generalOptions.Args.RootDir)
	var extensions []snippets.Extension
	for _, extension := range generalOptions.Extensions {
		provider, err := getExtension(extension)
		if err != nil {
			return "", err
		}
		extensions = append(extensions, provider)
	}

	extensions = append(extensions,
		&snippets.RegexBasedSnippetsProvider{
			Patterns: parseRegexes(generalOptions.Snippets),
		},
	)
	settings := &snippets.AnalysisSettings{RootPath: generalOptions.Args.RootDir,
		Extensions: extensions}

	allResults, err := snippets.Analyze(settings)
	if err != nil {
		return "", err
	}

	if generalOptions.AllViews {
		allViews := views.GetAllViews(allResults)
		return printAllViews(allViews), nil
	} else {
		resultsFromCommand, err := views.GetView(generalOptions.View, allResults)
		if err != nil {
			return "", err
		}
		sortRows(generalOptions.SortedBy, resultsFromCommand)
		return printRows(resultsFromCommand, generalOptions), nil
	}
}

func parseRegexes(input []string) []*regexp.Regexp {
	var toReturn []*regexp.Regexp
	for _, s := range input {
		toReturn = append(toReturn, regexp.MustCompile(s))
	}
	return toReturn
}
func printError(err error) int {
	fmt.Printf("Error: %s", err)
	return 1
}
