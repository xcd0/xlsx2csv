//go:build ignore

package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	const debug_mode = false
	if debug_mode {
		log.SetFlags(log.Ltime | log.Lshortfile)
	} else {
		log.SetOutput(io.Discard)
	}
	files := listAllFiles(".")
	patterns := readGitignore(".gitignore")
	files = applyFilters(files, patterns)
	generateEmbedFile(files, "embedded_files.go")
}

// GitignorePattern は .gitignore のパターンを表す構造体です
type GitignorePattern struct {
	Pattern string
	Type    string
}

// readGitignore は .gitignore ファイルを読み込み、パターンを GitignorePattern のスライスとして返します
func readGitignore(filePath string) []GitignorePattern {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("readGitignore error: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	var patterns []GitignorePattern
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patternType := "wildcard"
		if strings.HasPrefix(line, "/") && !strings.Contains(line, "*") {
			patternType = "exact"
		} else if strings.HasPrefix(line, "*.") || strings.HasPrefix(line, "/*.") {
			patternType = "extension"
			line = strings.TrimPrefix(line, "/")
		}
		patterns = append(patterns, GitignorePattern{Pattern: line, Type: patternType})
	}
	log.Printf("readGitignore patterns:\n%v", patterns)
	return patterns
}

// listAllFiles は指定されたディレクトリ以下の全てのファイルをリストアップします
func listAllFiles(root string) []string {
	var allFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			allFiles = append(allFiles, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		log.Fatalf("listAllFiles error: %v", err)
	}
	log.Printf("listAllFiles:\n%v", allFiles)
	return allFiles
}

// applyFilters は与えられたファイルリストに対して、パターンに基づいてフィルタを適用します
func applyFilters(files []string, patterns []GitignorePattern) []string {
	filteredFiles := files
	for _, pattern := range patterns {
		var filter func([]string, GitignorePattern) []string
		switch pattern.Type {
		case "exact":
			filter = filterFilesByExactPattern
		case "extension":
			filter = filterFilesByExtension
		default:
			filter = filterFilesByPattern
		}
		filteredFiles = filter(filteredFiles, pattern)
	}
	return filteredFiles
}

// filterFilesByPattern はワイルドカードパターンで指定されたファイルをフィルタリングします
func filterFilesByPattern(files []string, pattern GitignorePattern) []string {
	var filteredFiles []string
	normalizedPattern := normalizePattern(pattern.Pattern)
	for _, file := range files {
		match, err := filepath.Match(normalizedPattern, filepath.ToSlash(file))
		if err != nil {
			log.Fatalf("filepath.Match error: %v", err)
		}
		if !match {
			filteredFiles = append(filteredFiles, file)
		} else {
			log.Printf("Ignoring file %s due to pattern %s", file, pattern.Pattern)
		}
	}
	log.Printf("filterFilesByPattern:\n%v", filteredFiles)
	return filteredFiles
}

// generateEmbedFile はフィルタリングされたファイルリストを使用して embedded_files.go ファイルを生成します
func generateEmbedFile(files []string, outputPath string) {
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("generateEmbedFile error: %v", err)
	}
	defer f.Close()

	f.WriteString("package main\n\n")
	f.WriteString("// Code generated by go:generate; DO NOT EDIT.\n")
	f.WriteString("import \"embed\"\n\n")
	f.WriteString("//go:embed ")
	for _, file := range files {
		f.WriteString(strings.TrimSpace(file) + " ")
	}
	f.WriteString(`
var embedded embed.FS

func init() {
	embeddedFiles = &embedded
}
`)

}

// normalizePattern は .gitignore のパターンを正規化します
func normalizePattern(pattern string) string {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	if pattern[0] == '/' {
		pattern = "." + pattern // ルートディレクトリからの相対パスとして扱う
	} else if !strings.HasPrefix(pattern, "**/") {
		pattern = "**/" + pattern // 再帰的なマッチングを有効にする
	}
	log.Printf("normalizePattern: %s -> %s", pattern, pattern)
	return pattern
}

// filterFilesByExactPattern は / で始まり、* が含まれていないパターンにマッチするファイルをフィルタリングします
func filterFilesByExactPattern(files []string, pattern GitignorePattern) []string {
	var filteredFiles []string
	for _, file := range files {
		if !strings.HasPrefix(filepath.ToSlash(file), pattern.Pattern[1:]) {
			filteredFiles = append(filteredFiles, file)
		} else {
			log.Printf("Ignoring file %s due to exact pattern %s", file, pattern.Pattern)
		}
	}
	log.Printf("filterFilesByExactPattern:\n%v", filteredFiles)
	return filteredFiles
}

// filterFilesByExtension は拡張子パターンにマッチするファイルをフィルタリングします
func filterFilesByExtension(files []string, pattern GitignorePattern) []string {
	var filteredFiles []string
	for _, file := range files {
		if !strings.HasSuffix(file, strings.TrimPrefix(pattern.Pattern, "*.")) {
			filteredFiles = append(filteredFiles, file)
		} else {
			log.Printf("Ignoring file %s due to extension pattern %s", file, pattern.Pattern)
		}
	}
	log.Printf("filterFilesByExtension:\n%v", filteredFiles)
	return filteredFiles
}
