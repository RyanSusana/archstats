package component

import (
	"github.com/archstats/archstats/core/file"
)

// Connection is a connection between two components.
type Connection struct {
	From string
	To   string
	//The file in which the connection is made. The from side.
	File  string
	Begin *file.Position
	End   *file.Position
}

func (c *Connection) String() string {
	return c.From + " -> " + c.To + " in " + c.File + " [ " + c.Begin.String() + " - " + c.End.String() + " ]"
}

func GetConnections(snippetsByType file.SnippetGroup, snippetsByComponent file.SnippetGroup) []*Connection {
	var toReturn []*Connection
	from := snippetsByType[file.ComponentImport]
	for _, snippet := range from {
		if _, componentExistsInCodebase := snippetsByComponent[snippet.Value]; componentExistsInCodebase {
			toReturn = append(toReturn, &Connection{
				From:  snippet.Component,
				To:    snippet.Value,
				File:  snippet.File,
				Begin: snippet.Begin,
				End:   snippet.End,
			})
		}
	}
	return toReturn
}
