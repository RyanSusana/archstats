package views

import (
	"github.com/RyanSusana/archstats/snippets"
	"golang.org/x/exp/slices"
	"math"
	"sort"
)

func GenericView(allColumns []string, group snippets.SnippetGroup) *View {
	var toReturn []*Row
	for groupItem, groupedSnippets := range group {
		if groupItem == "" {
			groupItem = "Unknown"
		}
		stats := snippetsToStats(allColumns, groupedSnippets)
		data := statsToRowData(groupItem, stats)
		addFileCount(data, groupedSnippets)
		addAbstractness(data, stats)
		toReturn = append(toReturn, &Row{
			Data: data,
		})
	}

	columnsToReturn := []string{"name", FileCount}
	if slices.Contains(allColumns, snippets.AbstractType) {
		columnsToReturn = append(columnsToReturn, "abstractness")
	}
	for _, column := range allColumns {
		columnsToReturn = append(columnsToReturn, column)
	}
	return &View{
		OrderedColumns: columnsToReturn,
		Rows:           toReturn,
	}
}

func addFileCount(data map[string]interface{}, groupedSnippets []*snippets.Snippet) {
	data[FileCount] = getDistinctCount(groupedSnippets, fileCount)
}
func addAbstractness(data map[string]interface{}, stats Stats) {
	if _, hasAbstractTypes := data[snippets.AbstractType]; hasAbstractTypes {
		abstractTypes, types := stats[snippets.AbstractType], stats[snippets.Type]
		abstractness := math.Max(0, math.Min(1, float64(abstractTypes)/float64(types)))
		data[Abstractness] = nanToZero(abstractness)
	}
}

func statsToRowData(name string, stats Stats) map[string]interface{} {
	toReturn := make(map[string]interface{}, len(stats)+1)
	toReturn["name"] = name
	for k, v := range stats {
		toReturn[k] = v
	}
	return toReturn
}

func snippetsToStats(allStats []string, allSnippets []*snippets.Snippet) Stats {
	stats := Stats{}
	all := snippets.GroupSnippetsBy(allSnippets, snippets.ByType)

	for _, stat := range allStats {
		snippetsForType := all[stat]
		statToAdd := Stats{stat: len(snippetsForType)}

		stats = stats.Merge(statToAdd)
	}
	return stats
}

func getStatsPerGroup(allStats []string, group snippets.SnippetGroup) map[string]Stats {
	toReturn := map[string]Stats{}
	for groupItem, snippets := range group {
		toReturn[groupItem] = snippetsToStats(allStats, snippets)
	}
	return toReturn
}

func getDistinctColumnsFromResults(results *snippets.Results) []string {
	var toReturn []string
	for theType, _ := range results.SnippetsByType {
		toReturn = append(toReturn, theType)
	}
	sort.Strings(toReturn)
	return toReturn
}

func fileCount(snippet *snippets.Snippet) interface{} {
	return snippet.File
}
func getDistinctCount(results []*snippets.Snippet, distinctFunc func(snippet *snippets.Snippet) interface{}) int {
	files := make(map[interface{}]bool)
	for _, snippet := range results {
		files[distinctFunc(snippet)] = true
	}
	return len(files)
}
