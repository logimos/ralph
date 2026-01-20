// Package baseline provides codebase analysis and familiarization capabilities.
// It allows Ralph to understand existing codebases by scanning files, detecting
// patterns, and creating a knowledge base for context-aware feature development.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// DefaultBaselineFile is the default baseline file name
	DefaultBaselineFile = "baseline.json"
)

// FileType represents the category of a file
type FileType string

const (
	FileTypeSource FileType = "source"
	FileTypeTest   FileType = "test"
	FileTypeConfig FileType = "config"
	FileTypeDocs   FileType = "docs"
	FileTypeAsset  FileType = "asset"
	FileTypeOther  FileType = "other"
)

// FileInfo represents information about a single file in the codebase
type FileInfo struct {
	Path         string   `json:"path"`
	Type         FileType `json:"type"`
	Language     string   `json:"language,omitempty"`
	Size         int64    `json:"size"`
	LineCount    int      `json:"line_count,omitempty"`
	LastModified time.Time `json:"last_modified"`
}

// PackageInfo represents a package or module in the codebase
type PackageInfo struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Type         string   `json:"type"` // go, npm, python, etc.
	Dependencies []string `json:"dependencies,omitempty"`
	EntryPoints  []string `json:"entry_points,omitempty"`
}

// TechStack represents the detected technology stack
type TechStack struct {
	Languages     []string          `json:"languages"`
	Frameworks    []string          `json:"frameworks,omitempty"`
	BuildTools    []string          `json:"build_tools,omitempty"`
	TestFramework string            `json:"test_framework,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// CodebaseStructure represents the overall structure of the codebase
type CodebaseStructure struct {
	RootPath     string        `json:"root_path"`
	Packages     []PackageInfo `json:"packages"`
	Directories  []string      `json:"directories"`
	EntryPoints  []string      `json:"entry_points,omitempty"`
	TestDirs     []string      `json:"test_dirs,omitempty"`
	ConfigFiles  []string      `json:"config_files,omitempty"`
}

// Baseline represents the complete baseline knowledge of a codebase
type Baseline struct {
	Version       string            `json:"version"`
	GeneratedAt   time.Time         `json:"generated_at"`
	RootPath      string            `json:"root_path"`
	TechStack     TechStack         `json:"tech_stack"`
	Structure     CodebaseStructure `json:"structure"`
	Files         []FileInfo        `json:"files"`
	FileCounts    map[FileType]int  `json:"file_counts"`
	TotalFiles    int               `json:"total_files"`
	TotalLines    int               `json:"total_lines"`
	Conventions   []string          `json:"conventions,omitempty"`
	Patterns      []string          `json:"patterns,omitempty"`
}

// Scanner handles codebase scanning and analysis
type Scanner struct {
	rootPath       string
	ignoreDirs     []string
	ignorePatterns []string
}

// NewScanner creates a new codebase scanner
func NewScanner(rootPath string) *Scanner {
	return &Scanner{
		rootPath: rootPath,
		ignoreDirs: []string{
			".git", "node_modules", "vendor", "__pycache__",
			".cache", ".venv", "venv", "dist", "build",
			".next", ".nuxt", "coverage", ".nyc_output",
			"target", "bin", "obj", ".idea", ".vscode",
		},
		ignorePatterns: []string{
			"*.pyc", "*.pyo", "*.class", "*.o", "*.a",
			"*.so", "*.dylib", "*.dll", "*.exe",
			"*.log", "*.tmp", "*.swp", "*.swo",
			".DS_Store", "Thumbs.db",
		},
	}
}

// SetIgnoreDirs sets custom directories to ignore
func (s *Scanner) SetIgnoreDirs(dirs []string) {
	s.ignoreDirs = dirs
}

// AddIgnoreDirs adds directories to the ignore list
func (s *Scanner) AddIgnoreDirs(dirs []string) {
	s.ignoreDirs = append(s.ignoreDirs, dirs...)
}

// Scan performs a full codebase scan and returns a Baseline
func (s *Scanner) Scan() (*Baseline, error) {
	absPath, err := filepath.Abs(s.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	baseline := &Baseline{
		Version:     "1.0",
		GeneratedAt: time.Now(),
		RootPath:    absPath,
		FileCounts:  make(map[FileType]int),
	}

	// Scan files
	files, err := s.scanFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}
	baseline.Files = files
	baseline.TotalFiles = len(files)

	// Count files by type and calculate total lines
	for _, f := range files {
		baseline.FileCounts[f.Type]++
		baseline.TotalLines += f.LineCount
	}

	// Detect tech stack
	baseline.TechStack = s.detectTechStack(files)

	// Analyze structure
	baseline.Structure = s.analyzeStructure(files)
	baseline.Structure.RootPath = absPath

	// Detect conventions and patterns
	baseline.Conventions = s.detectConventions(files)
	baseline.Patterns = s.detectPatterns(files)

	return baseline, nil
}

// scanFiles walks the directory tree and catalogs all files
func (s *Scanner) scanFiles() ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip ignored directories
		if info.IsDir() {
			if s.shouldIgnoreDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip ignored file patterns
		if s.shouldIgnoreFile(info.Name()) {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(s.rootPath, path)
		if err != nil {
			relPath = path
		}

		// Create file info
		fileInfo := FileInfo{
			Path:         relPath,
			Type:         detectFileType(relPath),
			Language:     detectLanguage(relPath),
			Size:         info.Size(),
			LastModified: info.ModTime(),
		}

		// Count lines for source files (skip large files)
		if fileInfo.Type == FileTypeSource || fileInfo.Type == FileTypeTest {
			if info.Size() < 1024*1024 { // Skip files > 1MB
				if lines, err := countLines(path); err == nil {
					fileInfo.LineCount = lines
				}
			}
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// shouldIgnoreDir checks if a directory should be ignored
func (s *Scanner) shouldIgnoreDir(name string) bool {
	for _, ignore := range s.ignoreDirs {
		if name == ignore {
			return true
		}
	}
	return false
}

// shouldIgnoreFile checks if a file should be ignored
func (s *Scanner) shouldIgnoreFile(name string) bool {
	for _, pattern := range s.ignorePatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

// detectFileType determines the type of a file based on its path and name
func detectFileType(path string) FileType {
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	ext := strings.ToLower(filepath.Ext(path))
	nameLower := strings.ToLower(name)

	// Test files
	testPatterns := []string{"_test.", ".test.", ".spec.", "_spec.", "test_", "spec_"}
	for _, pattern := range testPatterns {
		if strings.Contains(nameLower, pattern) {
			return FileTypeTest
		}
	}
	testDirs := []string{"test", "tests", "__tests__", "spec", "specs"}
	for _, td := range testDirs {
		if strings.Contains(strings.ToLower(dir), td) {
			return FileTypeTest
		}
	}

	// Config files
	configNames := []string{
		"package.json", "go.mod", "go.sum", "cargo.toml", "requirements.txt",
		"setup.py", "pyproject.toml", "pom.xml", "build.gradle", "makefile",
		".gitignore", ".dockerignore", "dockerfile", "docker-compose",
		"tsconfig.json", "jest.config", "webpack.config", "vite.config",
		".eslintrc", ".prettierrc", ".babelrc", ".editorconfig",
		"renovate.json", "dependabot.yml", ".github",
	}
	for _, cn := range configNames {
		if strings.Contains(nameLower, cn) {
			return FileTypeConfig
		}
	}
	configExts := []string{".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf"}
	for _, ce := range configExts {
		if ext == ce {
			return FileTypeConfig
		}
	}

	// Documentation
	docNames := []string{"readme", "changelog", "contributing", "license", "authors"}
	for _, dn := range docNames {
		if strings.Contains(nameLower, dn) {
			return FileTypeDocs
		}
	}
	docExts := []string{".md", ".rst", ".txt", ".adoc"}
	for _, de := range docExts {
		if ext == de {
			return FileTypeDocs
		}
	}
	docDirs := []string{"docs", "doc", "documentation"}
	for _, dd := range docDirs {
		if strings.Contains(strings.ToLower(dir), dd) {
			return FileTypeDocs
		}
	}

	// Assets
	assetExts := []string{
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp",
		".woff", ".woff2", ".ttf", ".eot",
		".mp3", ".mp4", ".wav", ".webm",
		".pdf", ".zip", ".tar", ".gz",
	}
	for _, ae := range assetExts {
		if ext == ae {
			return FileTypeAsset
		}
	}

	// Source code
	sourceExts := []string{
		".go", ".py", ".js", ".ts", ".jsx", ".tsx",
		".java", ".c", ".cpp", ".h", ".hpp", ".cs",
		".rb", ".rs", ".swift", ".kt", ".scala",
		".php", ".pl", ".sh", ".bash", ".zsh",
		".vue", ".svelte", ".html", ".css", ".scss", ".sass", ".less",
		".sql", ".graphql", ".proto",
	}
	for _, se := range sourceExts {
		if ext == se {
			return FileTypeSource
		}
	}

	return FileTypeOther
}

// detectLanguage determines the programming language of a file
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	
	languages := map[string]string{
		".go":      "Go",
		".py":      "Python",
		".js":      "JavaScript",
		".ts":      "TypeScript",
		".jsx":     "JavaScript (React)",
		".tsx":     "TypeScript (React)",
		".java":    "Java",
		".c":       "C",
		".cpp":     "C++",
		".h":       "C/C++ Header",
		".hpp":     "C++ Header",
		".cs":      "C#",
		".rb":      "Ruby",
		".rs":      "Rust",
		".swift":   "Swift",
		".kt":      "Kotlin",
		".scala":   "Scala",
		".php":     "PHP",
		".pl":      "Perl",
		".sh":      "Shell",
		".bash":    "Bash",
		".zsh":     "Zsh",
		".vue":     "Vue",
		".svelte":  "Svelte",
		".html":    "HTML",
		".css":     "CSS",
		".scss":    "SCSS",
		".sass":    "Sass",
		".less":    "Less",
		".sql":     "SQL",
		".graphql": "GraphQL",
		".proto":   "Protocol Buffers",
		".yaml":    "YAML",
		".yml":     "YAML",
		".json":    "JSON",
		".xml":     "XML",
		".md":      "Markdown",
	}

	if lang, ok := languages[ext]; ok {
		return lang
	}
	return ""
}

// countLines counts the number of lines in a file
func countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	
	lines := 1
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	return lines, nil
}

// detectTechStack analyzes files to determine the technology stack
func (s *Scanner) detectTechStack(files []FileInfo) TechStack {
	stack := TechStack{
		Metadata: make(map[string]string),
	}

	// Count languages
	langCounts := make(map[string]int)
	for _, f := range files {
		if f.Language != "" && f.Type == FileTypeSource {
			langCounts[f.Language]++
		}
	}

	// Sort languages by count
	type langCount struct {
		lang  string
		count int
	}
	var sorted []langCount
	for lang, count := range langCounts {
		sorted = append(sorted, langCount{lang, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})
	for _, lc := range sorted {
		stack.Languages = append(stack.Languages, lc.lang)
	}

	// Detect build tools and frameworks from config files
	for _, f := range files {
		name := strings.ToLower(filepath.Base(f.Path))
		
		switch {
		case name == "go.mod":
			stack.BuildTools = appendUnique(stack.BuildTools, "Go Modules")
		case name == "package.json":
			stack.BuildTools = appendUnique(stack.BuildTools, "npm/Node.js")
			// Could parse package.json for frameworks
		case name == "cargo.toml":
			stack.BuildTools = appendUnique(stack.BuildTools, "Cargo")
		case name == "requirements.txt" || name == "setup.py" || name == "pyproject.toml":
			stack.BuildTools = appendUnique(stack.BuildTools, "pip/Python")
		case name == "pom.xml":
			stack.BuildTools = appendUnique(stack.BuildTools, "Maven")
		case name == "build.gradle" || name == "build.gradle.kts":
			stack.BuildTools = appendUnique(stack.BuildTools, "Gradle")
		case name == "makefile" || name == "gnumakefile":
			stack.BuildTools = appendUnique(stack.BuildTools, "Make")
		case name == "dockerfile" || strings.HasPrefix(name, "dockerfile."):
			stack.BuildTools = appendUnique(stack.BuildTools, "Docker")
		}

		// Detect test frameworks
		switch {
		case name == "jest.config.js" || name == "jest.config.ts" || name == "jest.config.json":
			stack.TestFramework = "Jest"
		case name == "pytest.ini" || name == "conftest.py":
			stack.TestFramework = "pytest"
		case name == "karma.conf.js":
			stack.TestFramework = "Karma"
		case name == "mocha.opts" || name == ".mocharc.js":
			stack.TestFramework = "Mocha"
		}
	}

	// If Go, test framework is built-in
	for _, lang := range stack.Languages {
		if lang == "Go" && stack.TestFramework == "" {
			stack.TestFramework = "Go testing"
		}
	}

	return stack
}

// analyzeStructure analyzes the codebase structure
func (s *Scanner) analyzeStructure(files []FileInfo) CodebaseStructure {
	structure := CodebaseStructure{}

	// Collect unique directories
	dirSet := make(map[string]bool)
	testDirSet := make(map[string]bool)
	configFiles := []string{}

	for _, f := range files {
		dir := filepath.Dir(f.Path)
		if dir != "." {
			dirSet[dir] = true
		}

		// Track test directories
		if f.Type == FileTypeTest {
			testDirSet[dir] = true
		}

		// Track config files
		if f.Type == FileTypeConfig {
			configFiles = append(configFiles, f.Path)
		}

		// Detect entry points
		name := strings.ToLower(filepath.Base(f.Path))
		if name == "main.go" || name == "main.py" || name == "index.js" || name == "index.ts" ||
			name == "app.go" || name == "app.py" || name == "app.js" || name == "app.ts" {
			structure.EntryPoints = append(structure.EntryPoints, f.Path)
		}
	}

	// Convert sets to sorted slices
	for dir := range dirSet {
		structure.Directories = append(structure.Directories, dir)
	}
	sort.Strings(structure.Directories)

	for dir := range testDirSet {
		structure.TestDirs = append(structure.TestDirs, dir)
	}
	sort.Strings(structure.TestDirs)

	structure.ConfigFiles = configFiles
	sort.Strings(structure.ConfigFiles)

	// Detect packages
	structure.Packages = s.detectPackages(files)

	return structure
}

// detectPackages identifies packages/modules in the codebase
func (s *Scanner) detectPackages(files []FileInfo) []PackageInfo {
	var packages []PackageInfo
	pkgSet := make(map[string]bool)

	for _, f := range files {
		name := strings.ToLower(filepath.Base(f.Path))
		dir := filepath.Dir(f.Path)
		if dir == "." {
			dir = ""
		}

		switch name {
		case "go.mod":
			if !pkgSet[dir] {
				pkg := PackageInfo{
					Path: dir,
					Type: "go",
				}
				// Could parse go.mod for module name and dependencies
				packages = append(packages, pkg)
				pkgSet[dir] = true
			}
		case "package.json":
			if !pkgSet[dir] {
				pkg := PackageInfo{
					Path: dir,
					Type: "npm",
				}
				packages = append(packages, pkg)
				pkgSet[dir] = true
			}
		case "cargo.toml":
			if !pkgSet[dir] {
				pkg := PackageInfo{
					Path: dir,
					Type: "cargo",
				}
				packages = append(packages, pkg)
				pkgSet[dir] = true
			}
		case "setup.py", "pyproject.toml":
			if !pkgSet[dir] {
				pkg := PackageInfo{
					Path: dir,
					Type: "python",
				}
				packages = append(packages, pkg)
				pkgSet[dir] = true
			}
		case "pom.xml":
			if !pkgSet[dir] {
				pkg := PackageInfo{
					Path: dir,
					Type: "maven",
				}
				packages = append(packages, pkg)
				pkgSet[dir] = true
			}
		}
	}

	return packages
}

// detectConventions identifies coding conventions from the codebase
func (s *Scanner) detectConventions(files []FileInfo) []string {
	conventions := []string{}

	// Check for standard Go project layout
	hasCmd := false
	hasInternal := false
	hasPkg := false
	for _, f := range files {
		dir := strings.Split(f.Path, string(filepath.Separator))
		if len(dir) > 0 {
			switch dir[0] {
			case "cmd":
				hasCmd = true
			case "internal":
				hasInternal = true
			case "pkg":
				hasPkg = true
			}
		}
	}
	if hasCmd || hasInternal || hasPkg {
		conventions = append(conventions, "Standard Go project layout (cmd/, internal/, pkg/)")
	}

	// Check for src/ directory (common in many projects)
	for _, f := range files {
		if strings.HasPrefix(f.Path, "src"+string(filepath.Separator)) {
			conventions = append(conventions, "Source code in src/ directory")
			break
		}
	}

	// Check for test naming conventions
	hasGoTests := false
	hasJSTests := false
	for _, f := range files {
		if f.Type == FileTypeTest {
			if strings.HasSuffix(f.Path, "_test.go") {
				hasGoTests = true
			}
			if strings.Contains(f.Path, ".test.") || strings.Contains(f.Path, ".spec.") {
				hasJSTests = true
			}
		}
	}
	if hasGoTests {
		conventions = append(conventions, "Go test files with _test.go suffix")
	}
	if hasJSTests {
		conventions = append(conventions, "JavaScript/TypeScript test files with .test./.spec. naming")
	}

	return conventions
}

// detectPatterns identifies common patterns in the codebase
func (s *Scanner) detectPatterns(files []FileInfo) []string {
	patterns := []string{}

	// Check for common architectural patterns
	dirNames := make(map[string]bool)
	for _, f := range files {
		parts := strings.Split(f.Path, string(filepath.Separator))
		for _, part := range parts {
			dirNames[strings.ToLower(part)] = true
		}
	}

	// MVC pattern
	if dirNames["models"] && dirNames["views"] && dirNames["controllers"] {
		patterns = append(patterns, "MVC (Model-View-Controller) architecture")
	}

	// Clean/Hexagonal architecture
	if (dirNames["domain"] || dirNames["entities"]) && (dirNames["usecases"] || dirNames["services"]) {
		patterns = append(patterns, "Clean/Domain-driven architecture")
	}

	// Repository pattern
	if dirNames["repository"] || dirNames["repositories"] {
		patterns = append(patterns, "Repository pattern for data access")
	}

	// API structure
	if dirNames["api"] || dirNames["handlers"] || dirNames["routes"] {
		patterns = append(patterns, "API/Handler-based structure")
	}

	// Component-based (frontend)
	if dirNames["components"] {
		patterns = append(patterns, "Component-based architecture (frontend)")
	}

	return patterns
}

// appendUnique appends a string to a slice if it's not already present
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

// Save writes the baseline to a JSON file
func (b *Baseline) Save(path string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write baseline file: %w", err)
	}

	return nil
}

// Load reads a baseline from a JSON file
func Load(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline file: %w", err)
	}

	return &baseline, nil
}

// Summary returns a human-readable summary of the baseline
func (b *Baseline) Summary() string {
	var sb strings.Builder

	sb.WriteString("=== Codebase Baseline ===\n\n")
	sb.WriteString(fmt.Sprintf("Root: %s\n", b.RootPath))
	sb.WriteString(fmt.Sprintf("Generated: %s\n", b.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Total files: %d\n", b.TotalFiles))
	sb.WriteString(fmt.Sprintf("Total lines: %d\n\n", b.TotalLines))

	// File type breakdown
	sb.WriteString("File Types:\n")
	for fileType, count := range b.FileCounts {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", fileType, count))
	}
	sb.WriteString("\n")

	// Tech stack
	sb.WriteString("Tech Stack:\n")
	if len(b.TechStack.Languages) > 0 {
		sb.WriteString(fmt.Sprintf("  Languages: %s\n", strings.Join(b.TechStack.Languages, ", ")))
	}
	if len(b.TechStack.BuildTools) > 0 {
		sb.WriteString(fmt.Sprintf("  Build tools: %s\n", strings.Join(b.TechStack.BuildTools, ", ")))
	}
	if b.TechStack.TestFramework != "" {
		sb.WriteString(fmt.Sprintf("  Test framework: %s\n", b.TechStack.TestFramework))
	}
	sb.WriteString("\n")

	// Structure
	sb.WriteString("Structure:\n")
	sb.WriteString(fmt.Sprintf("  Directories: %d\n", len(b.Structure.Directories)))
	sb.WriteString(fmt.Sprintf("  Packages: %d\n", len(b.Structure.Packages)))
	if len(b.Structure.EntryPoints) > 0 {
		sb.WriteString(fmt.Sprintf("  Entry points: %s\n", strings.Join(b.Structure.EntryPoints, ", ")))
	}
	sb.WriteString("\n")

	// Conventions
	if len(b.Conventions) > 0 {
		sb.WriteString("Conventions:\n")
		for _, c := range b.Conventions {
			sb.WriteString(fmt.Sprintf("  - %s\n", c))
		}
		sb.WriteString("\n")
	}

	// Patterns
	if len(b.Patterns) > 0 {
		sb.WriteString("Patterns:\n")
		for _, p := range b.Patterns {
			sb.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}

	return sb.String()
}

// BuildPromptContext creates a formatted string of baseline knowledge to inject into prompts
func (b *Baseline) BuildPromptContext() string {
	var sb strings.Builder

	sb.WriteString("\n[CODEBASE CONTEXT - Knowledge about the existing codebase:]\n\n")

	// Tech stack
	sb.WriteString("Technology Stack:\n")
	if len(b.TechStack.Languages) > 0 {
		sb.WriteString(fmt.Sprintf("- Primary languages: %s\n", strings.Join(b.TechStack.Languages, ", ")))
	}
	if len(b.TechStack.BuildTools) > 0 {
		sb.WriteString(fmt.Sprintf("- Build tools: %s\n", strings.Join(b.TechStack.BuildTools, ", ")))
	}
	if b.TechStack.TestFramework != "" {
		sb.WriteString(fmt.Sprintf("- Test framework: %s\n", b.TechStack.TestFramework))
	}
	sb.WriteString("\n")

	// Project structure
	sb.WriteString("Project Structure:\n")
	sb.WriteString(fmt.Sprintf("- Total files: %d (%d source, %d test, %d config)\n",
		b.TotalFiles, b.FileCounts[FileTypeSource], b.FileCounts[FileTypeTest], b.FileCounts[FileTypeConfig]))
	if len(b.Structure.EntryPoints) > 0 {
		sb.WriteString(fmt.Sprintf("- Entry points: %s\n", strings.Join(b.Structure.EntryPoints, ", ")))
	}
	if len(b.Structure.TestDirs) > 0 && len(b.Structure.TestDirs) <= 5 {
		sb.WriteString(fmt.Sprintf("- Test directories: %s\n", strings.Join(b.Structure.TestDirs, ", ")))
	}
	sb.WriteString("\n")

	// Conventions to follow
	if len(b.Conventions) > 0 {
		sb.WriteString("Conventions to follow:\n")
		for _, c := range b.Conventions {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
		sb.WriteString("\n")
	}

	// Patterns used
	if len(b.Patterns) > 0 {
		sb.WriteString("Patterns in use:\n")
		for _, p := range b.Patterns {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
		sb.WriteString("\n")
	}

	// Key directories (first 10)
	if len(b.Structure.Directories) > 0 {
		sb.WriteString("Key directories:\n")
		limit := 10
		if len(b.Structure.Directories) < limit {
			limit = len(b.Structure.Directories)
		}
		for i := 0; i < limit; i++ {
			sb.WriteString(fmt.Sprintf("- %s/\n", b.Structure.Directories[i]))
		}
		if len(b.Structure.Directories) > limit {
			sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(b.Structure.Directories)-limit))
		}
	}

	sb.WriteString("\n[END CODEBASE CONTEXT]\n")
	sb.WriteString("\nPlease follow the existing conventions and patterns when implementing new features.\n")

	return sb.String()
}
