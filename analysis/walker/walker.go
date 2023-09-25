package walker

import (
	"github.com/RyanSusana/archstats/analysis/file"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

func WalkDirectoryConcurrently(dirAbsolutePath string, visitor func(file file.File)) {
	dirFS := os.DirFS(dirAbsolutePath).(fs.ReadFileFS)

	allFiles := GetAllFiles(dirAbsolutePath)
	WalkFiles(dirFS, allFiles, visitor)
}

func WalkFiles(fileSystem fs.ReadFileFS, allFiles []PathToFile, visitor func(file file.File)) {
	wg := &sync.WaitGroup{}
	wg.Add(len(allFiles))
	for _, theFile := range allFiles {
		go func(file PathToFile, group *sync.WaitGroup) {

			//TODO cleanup error handling
			content, _ := fileSystem.ReadFile(file.Path())
			openedFile := &openedFile{
				path: file.Path(),

				content: content,
			}

			visitor(openedFile)
			group.Done()
		}(theFile, wg)
	}
	wg.Wait()
}

func GetAllFiles(dirAbsolutePath string) []PathToFile {
	return getAllFiles(os.DirFS(dirAbsolutePath).(fs.ReadDirFS), ".", 0, ignoreContext{}, nil)
}

func getAllFiles(fileSystem fs.ReadDirFS, dirAbsolutePath string, depth int, ignoreCtx ignoreContext, shouldAdd func(info fs.DirEntry) bool) []PathToFile {
	separator := string(filepath.Separator)

	dirAbsolutePath = filepath.Clean(dirAbsolutePath)
	var fileDescriptions []PathToFile

	files, err := fileSystem.ReadDir(dirAbsolutePath)
	if err != nil {
		panic(err)
	}
	ignoreCtx.addIgnoreLines(fileSystem, dirAbsolutePath, files)

	gitIgnore := ignoreCtx.getGitIgnore()
	for _, entry := range files {
		path := dirAbsolutePath + separator + entry.Name()
		if shouldIgnore(path, gitIgnore) {
			continue
		}

		if entry.IsDir() {
			path += separator
			fileDescriptions = append(fileDescriptions, getAllFiles(fileSystem, path, depth+1, ignoreCtx, shouldAdd)...)
		} else if shouldAdd == nil || shouldAdd(entry) {
			info, err := entry.Info()
			// What could go wrong :D
			if err == nil {
				fileDescriptions = append(fileDescriptions, &pathToFile{
					path: path,
					info: info,
				})
			}
		}
	}
	return fileDescriptions
}

type PathToFile interface {
	Path() string
	File() fs.FileInfo
}
type pathToFile struct {
	path string
	info fs.FileInfo
}

func (f *pathToFile) File() fs.FileInfo {
	return f.info
}

func (f *pathToFile) Path() string {
	return f.path
}
