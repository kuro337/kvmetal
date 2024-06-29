package fpath

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"strings"
)

func NewEnvPath(path string) *EnvPathResolver {
	return &EnvPathResolver{
		unique:   &[]string{},
		dupes:    &[]string{},
		currPath: TrimPathTrailingSlash(path),
	}
}

type EnvPathResolver struct {
	EnvPath  string
	Shell    string
	Dest     string
	currPath string

	present bool
	unique  *[]string
	dupes   *[]string
}

func (e *EnvPathResolver) GenerateNewPath() string {
	if e.EnvPath == "" {
		e.GetENV()
	}

	if e.unique == nil {
		e.unique = &[]string{}
	}

	if e.dupes == nil {
		e.dupes = &[]string{}
	}

	if len(*e.unique) == 0 {
		e.CheckDuplicates()
	}

	fmt.Printf("Length of Uniques %d\n", len(*e.unique))
	fmt.Printf("Length of Dupes %d\n", len(*e.dupes))

	PrintPaths(*e.unique)

	if !e.PathExists(e.currPath) {
		*e.unique = append(*e.unique, e.currPath)
	}

	newPath := strings.Join(*e.unique, ":")
	return newPath
}

func (e *EnvPathResolver) GetAliased() string {
	newPath := strings.Join(*e.unique, ":")
	return fmt.Sprintf("export PATH=%s", ReplaceHOME(newPath))
}

func ReplaceHOME(path string) string {
	home := os.Getenv("HOME")
	if home != "" {
		return strings.ReplaceAll(path, home, "$HOME")
	}
	return path
}

func (e *EnvPathResolver) GetENV() error {
	if e.currPath == "" {
		e.CurrDirPath()
	}
	envPath := os.Getenv("PATH")
	e.EnvPath = envPath
	return nil
}

func (e *EnvPathResolver) Results() {
	sort.Strings(*e.unique)
	log.Print("Unique Paths:")
	for _, uq := range *e.unique {
		log.Print(uq)
	}

	log.Print("Duplicate Paths:")
	sort.Strings(*e.dupes)
	for _, dupe := range *e.dupes {
		log.Print(dupe)
	}
}

func (e *EnvPathResolver) CurrDirPath() error {
	path, err := CwdNoTrailing()
	if err != nil {
		return err
	}
	e.currPath = path
	return nil
}

/* Checks Duplicates from Path */
func (e *EnvPathResolver) CheckDuplicates() ([]string, []string, error) {
	if e.EnvPath == "" {
		if err := e.GetENV(); err != nil {
			return nil, nil, err
		}
	}
	unique, dupes := CheckPath(e.EnvPath)
	log.Printf("%d unique, %d duplicates", len(unique), len(dupes))
	e.unique = &unique
	e.dupes = &dupes
	return unique, dupes, nil
}

func (e *EnvPathResolver) PathExists(path string) bool {
	if len(*e.unique) == 0 {
		e.CheckDuplicates()
	}

	if slices.Contains(*e.unique, path) {
		e.present = true
		log.Printf("Current Dir Path already in ENV PATH.")
		return true
	}
	e.present = false
	log.Printf("Curr Dir was not in ENV Path")

	return false
}

/* Checks Duplicates from Path */
func CheckPath(currPath string) ([]string, []string) {
	paths := strings.Split(currPath, ":")

	var filteredPaths []string
	for _, path := range paths {
		if path != "" {
			filteredPaths = append(filteredPaths, path)
		}
	}

	sort.Strings(filteredPaths)

	duplicatePaths := make([]string, 0)
	pathCount := make(map[string]int)
	for _, path := range filteredPaths {
		pathCount[path]++
		if pathCount[path] == 2 {
			duplicatePaths = append(duplicatePaths, path)
		}
	}

	// Compact the sorted, filtered paths to remove adjacent duplicates
	compactedPaths := slices.Compact(filteredPaths)

	// Note: compactedPaths is now a slice with a potentially reduced length, with duplicates removed
	return compactedPaths, duplicatePaths
}

func GenerateNewPath(paths []string) string {
	// export PATH="/path/to/directory:$PATH"
	return fmt.Sprintf(`export PATH="%s"`, strings.Join(paths, ":"))
}

func GetPATHS() (string, error) {
	currEnv := os.Getenv("PATH")
	spl := strings.Split(currEnv, ":")
	sorted := sort.StringSlice(spl)
	j := strings.Join(sorted, "\n")
	return j, nil
}

func ExistsInPATH(PATH []string, path string) bool {
	return slices.Contains(PATH, path)
}

func CwdNoTrailing() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed cwd() ERROR:%s", err)
		return "", err
	}

	if len(cwd) > 0 && cwd[len(cwd)-1] == '/' {
		cwd = cwd[:len(cwd)-1]
	}
	return cwd, nil
}

func TrimPathTrailingSlash(path string) string {
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

func PrintPaths(paths []string) {
	sort.Strings(paths)
	for _, path := range paths {
		log.Print(path)
	}
}

func GetUniquePATHS() ([]string, error) {
	currEnv := os.Getenv("PATH")
	spl := strings.Split(currEnv, ":")

	fmt.Println("Paths Initial:")
	PrintPaths(spl)
	initialLen := len(spl)

	uniquePaths := make(map[string]bool)
	for _, path := range spl {
		uniquePaths[path] = true
	}

	var uniqueSlice []string
	for path := range uniquePaths {
		uniqueSlice = append(uniqueSlice, path)
	}

	sort.Strings(uniqueSlice)

	fmt.Printf("Remaining %d unique.\n", len(uniqueSlice))
	fmt.Printf("Cleaned up %d duplicates.\n", initialLen-len(uniqueSlice))

	fmt.Println("Paths after Cleanup:")
	PrintPaths(uniqueSlice)

	return uniqueSlice, nil

	// j := strings.Join(uniqueSlice, "\n")
	// return j, nil
}

func GetUniqueCompacted() ([]string, error) {
	currEnv := os.Getenv("PATH")
	spl := strings.Split(currEnv, ":")

	initialLen := len(spl)

	slices.Sort(spl)

	spl = slices.Compact(spl)

	fmt.Printf("Compacted: Cleaned up %d duplicates.\n", initialLen-len(spl))

	PrintPaths(spl)

	return spl, nil
}

func AddToPath(paths []string, path string) string {
	// export PATH="/path/to/directory:$PATH"
	return fmt.Sprintf(`export PATH="%s"`, strings.Join(append(paths, path), ":"))
}
