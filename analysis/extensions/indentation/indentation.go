package indentation

import (
	"bufio"
	"bytes"
	"github.com/RyanSusana/archstats/analysis"
	"strings"
)

const (
	Max   = "indentation_max"
	Count = "indentation_count"
	Avg   = "indentation_avg"
)

type Analyzer struct{}

func (i *Analyzer) typeAssertions() (analysis.Initializable, analysis.FileAnalyzer) {
	return i, i
}

func (i *Analyzer) Init(settings analysis.Settings) {
	settings.SetStatAccumulator(Max, maxAccumulator)
	settings.SetStatAccumulator(Avg, avgAccumulator)
}

func maxAccumulator(indentations []interface{}) interface{} {
	curMax := 0
	for _, indentation := range indentations {
		if indentation.(int) > curMax {
			curMax = indentation.(int)
		}
	}
	return curMax
}

func avgAccumulator(indentations []interface{}) interface{} {
	allIndentations := 0.0
	allLines := 0.0
	for _, indentation := range indentations {
		stat := indentation.(*indentationStat)
		allIndentations += float64(stat.indentation)
		allLines += float64(stat.lines)
	}
	return allIndentations / allLines
}

func (i *Analyzer) AnalyzeFile(file analysis.File) *analysis.FileResults {
	bytesReader := bytes.NewReader(file.Content())

	fileReader := bufio.NewReader(bytesReader)

	fileReader.ReadBytes('\n')

	var maxIndentations int
	var totalIndentation int
	var lineCount int
	for {
		line, err := fileReader.ReadBytes('\n')
		lineCount++
		if err != nil {
			break
		}
		indentation := getLeadingIndentation(line)
		totalIndentation += indentation
		if indentation > maxIndentations {
			maxIndentations = indentation
		}
	}

	return &analysis.FileResults{
		Stats: []*analysis.StatRecord{
			{
				StatType: Max,
				Value:    maxIndentations,
			},
			{
				StatType: Count,
				Value:    totalIndentation,
			},
			{
				StatType: Avg,
				Value: &indentationStat{
					indentation: totalIndentation,
					lines:       lineCount,
				},
			},
		},
	}
}

type indentationStat struct {
	indentation int
	lines       int
}

func getLeadingIndentation(line []byte) int {
	lineTabs := strings.ReplaceAll(string(line), "    ", "\t")

	indentation := 0
	for _, char := range lineTabs {
		if char == '\t' {
			indentation++
		} else {
			break
		}
	}

	return indentation
}
