package main

import (
	"archstats/snippets"
	"archstats/walker"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sync"
)

type GeneralOptions struct {
	View    string `positional-args:"0" description:"Type of view to show" required:"true"`
	RootDir string `positional-args:"1" description:"Root directory" required:"true"`

	RegexStats []string `short:"r" long:"snippet-type" description:""`

	Language string `short:"l" long:"language" description:"Programming language. This flag adds language-specific support for components, packages, functions, etc. (Supported: php)"`

	NoHeader bool `long:"no-header" description:"No header (only applicable csv, tsv, table)"`

	SortedBy string `long:"sorted-by" short:"s" description:"Sorted by column name. For number based columns, this is in descending order."`

	OutputFormat string `short:"o" long:"output-format" description:"Output format: table, ndjson, json, csv (default: table)"`

	CpuProfile string `long:"cpu-profile" description:"Write cpu profile to file"`
	MemProfile string `long:"mem-profile" description:"Write memory profile to file"`
}

func main() {
	generalOptions := &GeneralOptions{}
	args, err := flags.Parse(generalOptions)

	if err != nil {
		return
	}

	// Enable cpu profiling if requested.
	if generalOptions.CpuProfile != "" {
		f, err := os.Create(generalOptions.CpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // TODO handle error
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	runArchStats(args, generalOptions)

	// Enable memory profiling if requested.
	if generalOptions.MemProfile != "" {
		f, err := os.Create(generalOptions.MemProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // TODO handle error
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

func runArchStats(args []string, generalOptions *GeneralOptions) error {

	extensions := getLanguageExtensions(generalOptions.Language)

	extensions = append(extensions,
		&snippets.RegexBasedSnippetsProvider{
			Patterns: parseRegexes(generalOptions.RegexStats),
		},
	)
	settings := snippets.AnalysisSettings{SnippetProviders: extensions}
	allResults, err := Analyze(generalOptions.RootDir, settings)
	if err != nil {
		return err
	}
	resultsFromCommand, err := getRowsFromResults(generalOptions.View, allResults)
	if err != nil {
		return err
	}
	sortRows(generalOptions.SortedBy, resultsFromCommand)

	printRows(resultsFromCommand, generalOptions)
	return nil
}
func Analyze(rootPath string, settings snippets.AnalysisSettings) (*snippets.Results, error) {

	var allSnippets []*snippets.Snippet
	lock := sync.Mutex{}

	walker.WalkDirectoryConcurrently(rootPath, func(file walker.OpenedFile) {
		var foundSnippets []*snippets.Snippet
		for _, provider := range settings.SnippetProviders {
			foundSnippets = append(foundSnippets, provider.GetSnippetsFromFile(file)...)
		}
		lock.Lock()
		allSnippets = append(allSnippets, foundSnippets...)
		lock.Unlock()
	})
	return snippets.CalculateResults(allSnippets), nil
}

func parseRegexes(input []string) []*regexp.Regexp {
	var toReturn []*regexp.Regexp
	for _, s := range input {
		toReturn = append(toReturn, regexp.MustCompile(s))
	}
	return toReturn
}
