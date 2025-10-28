package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// ChangeType represents the type of change
type ChangeType int

const (
	Addition ChangeType = iota
	Deletion
	Modification
)

// Change represents a single change in the diff
type Change struct {
	Type     ChangeType
	Path     string
	OldValue interface{}
	NewValue interface{}
}

// isSliceOfDictsWithIds checks if a slice contains dictionaries with identifier fields
func isSliceOfDictsWithIds(slice []interface{}) bool {
	if len(slice) == 0 {
		return false
	}

	for _, item := range slice {
		if reflect.TypeOf(item).Kind() != reflect.Map {
			return false
		}
		m := item.(map[interface{}]interface{})
		// Check for common identifier fields
		if _, hasName := m["name"]; hasName {
			return true
		}
		if _, hasKey := m["key"]; hasKey {
			return true
		}
		if _, hasId := m["id"]; hasId {
			return true
		}
	}
	return false
}

// diffSliceOfDicts compares slices of dictionaries by matching on identifier fields
func diffSliceOfDicts(oldSlice, newSlice []interface{}, path string) []Change {
	var changes []Change

	// Group by identifier
	oldMap := make(map[string]interface{})
	newMap := make(map[string]interface{})

	for _, item := range oldSlice {
		if m, ok := item.(map[interface{}]interface{}); ok {
			if name, hasName := m["name"]; hasName {
				oldMap[fmt.Sprintf("%v", name)] = item
			} else if key, hasKey := m["key"]; hasKey {
				oldMap[fmt.Sprintf("%v", key)] = item
			} else if id, hasId := m["id"]; hasId {
				oldMap[fmt.Sprintf("%v", id)] = item
			}
		}
	}

	for _, item := range newSlice {
		if m, ok := item.(map[interface{}]interface{}); ok {
			if name, hasName := m["name"]; hasName {
				newMap[fmt.Sprintf("%v", name)] = item
			} else if key, hasKey := m["key"]; hasKey {
				newMap[fmt.Sprintf("%v", key)] = item
			} else if id, hasId := m["id"]; hasId {
				newMap[fmt.Sprintf("%v", id)] = item
			}
		}
	}

	// Find matches and differences
	for key, oldItem := range oldMap {
		if newItem, exists := newMap[key]; exists {
			// Both exist, diff them
			subChanges := diffValues(oldItem, newItem, path+"["+key+"]")
			changes = append(changes, subChanges...)
		} else {
			// Only in old, it's a deletion
			changes = append(changes, Change{
				Type:     Deletion,
				Path:     path + "[" + key + "]",
				OldValue: oldItem,
				NewValue: nil,
			})
		}
	}

	for key, newItem := range newMap {
		if _, exists := oldMap[key]; !exists {
			// Only in new, it's an addition
			changes = append(changes, Change{
				Type:     Addition,
				Path:     path + "[" + key + "]",
				OldValue: nil,
				NewValue: newItem,
			})
		}
	}

	return changes
}

// generateColoredDiff generates a colored diff showing only changed items
func generateColoredDiff(changes []Change) string {
	if len(changes) == 0 {
		return "No changes found.\n"
	}

	// Sort changes alphabetically by path for consistency
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})

	var result strings.Builder
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	for _, change := range changes {
		switch change.Type {
		case Addition:
			coloredPrefix := green.Sprint("+ ")
			result.WriteString(coloredPrefix)
			result.WriteString(change.Path)
			result.WriteString(": ")
			formattedValue := formatValue(change.NewValue)
			if strings.Contains(formattedValue, "\n") {
				// Complex value - add newline and prefix subsequent lines
				result.WriteString("\n")
				result.WriteString(prefixLinesComplex(formattedValue, coloredPrefix))
			} else {
				// Simple value - show on same line
				result.WriteString(formattedValue)
				result.WriteString("\n")
			}
		case Deletion:
			coloredPrefix := red.Sprint("- ")
			result.WriteString(coloredPrefix)
			result.WriteString(change.Path)
			result.WriteString(": ")
			formattedValue := formatValue(change.OldValue)
			if strings.Contains(formattedValue, "\n") {
				// Complex value - add newline and prefix subsequent lines
				result.WriteString("\n")
				result.WriteString(prefixLinesComplex(formattedValue, coloredPrefix))
			} else {
				// Simple value - show on same line
				result.WriteString(formattedValue)
				result.WriteString("\n")
			}
		case Modification:
			result.WriteString(yellow.Sprint("~ "))
			result.WriteString(change.Path)
			result.WriteString(": ")
			oldStr := formatValue(change.OldValue)
			newStr := formatValue(change.NewValue)

			// For string values, show character-level differences
			if isStringValue(change.OldValue) && isStringValue(change.NewValue) {
				oldStrColored, newStrColored := colorStringDiff(change.OldValue.(string), change.NewValue.(string))
				result.WriteString(fmt.Sprintf("%s → %s\n", oldStrColored, newStrColored))
			} else {
				result.WriteString(fmt.Sprintf("%s → %s\n", oldStr, newStr))
			}
		}
	}

	return result.String()
}

// prefixLinesComplex prefixes each line of a complex (multi-line) value with the given prefix and extra indentation
func prefixLinesComplex(s, prefix string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return ""
	}

	var result strings.Builder
	for i, line := range lines {
		if i > 0 || line != "" { // Skip empty first line if any
			result.WriteString(prefix)
			// Add extra indentation (3 spaces) for better visual presentation
			if strings.TrimSpace(line) != "" {
				result.WriteString("   ")
			}
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// isStringValue checks if a value is a string
func isStringValue(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

// colorStringDiff colors entire strings for better readability
func colorStringDiff(oldStr, newStr string) (string, string) {
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)

	return red.Sprint(oldStr), green.Sprint(newStr)
}

// formatValue formats a value for display, using YAML formatting for complex values
func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Map, reflect.Slice:
		// Format complex values as YAML with 3-space indentation
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(3) // 3-space indentation
		if err := encoder.Encode(v); err != nil {
			return fmt.Sprintf("%v", v) // fallback to default formatting
		}
		encoder.Close()

		// Return the YAML string as-is
		return strings.TrimSuffix(buf.String(), "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// diffValues compares two normalized values and returns a list of changes
func diffValues(oldVal, newVal interface{}, path string) []Change {
	var changes []Change

	if reflect.DeepEqual(oldVal, newVal) {
		return changes
	}

	oldType := reflect.TypeOf(oldVal)
	newType := reflect.TypeOf(newVal)

	// If types are different, it's a modification
	if oldType != newType && oldVal != nil && newVal != nil {
		changes = append(changes, Change{
			Type:     Modification,
			Path:     path,
			OldValue: oldVal,
			NewValue: newVal,
		})
		return changes
	}

	// Handle nil values
	if oldVal == nil && newVal != nil {
		changes = append(changes, Change{
			Type:     Addition,
			Path:     path,
			OldValue: nil,
			NewValue: newVal,
		})
		return changes
	}
	if oldVal != nil && newVal == nil {
		changes = append(changes, Change{
			Type:     Deletion,
			Path:     path,
			OldValue: oldVal,
			NewValue: nil,
		})
		return changes
	}

	switch oldType.Kind() {
	case reflect.Map:
		oldMap := oldVal.(map[interface{}]interface{})
		newMap := newVal.(map[interface{}]interface{})

		// Check for deletions and modifications
		for key, oldValue := range oldMap {
			keyStr := fmt.Sprintf("%v", key)
			newValue, exists := newMap[key]
			if !exists {
				changes = append(changes, Change{
					Type:     Deletion,
					Path:     path + "." + keyStr,
					OldValue: oldValue,
					NewValue: nil,
				})
			} else {
				subChanges := diffValues(oldValue, newValue, path+"."+keyStr)
				changes = append(changes, subChanges...)
			}
		}

		// Check for additions
		for key, newValue := range newMap {
			keyStr := fmt.Sprintf("%v", key)
			if _, exists := oldMap[key]; !exists {
				changes = append(changes, Change{
					Type:     Addition,
					Path:     path + "." + keyStr,
					OldValue: nil,
					NewValue: newValue,
				})
			}
		}

	case reflect.Slice:
		oldSlice := oldVal.([]interface{})
		newSlice := newVal.([]interface{})

		// Check if this is a slice of dictionaries with identifier fields
		if isSliceOfDictsWithIds(oldSlice) && isSliceOfDictsWithIds(newSlice) {
			changes = append(changes, diffSliceOfDicts(oldSlice, newSlice, path)...)
		} else {
			// For slices, we compare element by element since they're sorted
			minLen := len(oldSlice)
			if len(newSlice) < minLen {
				minLen = len(newSlice)
			}

			for i := 0; i < minLen; i++ {
				subChanges := diffValues(oldSlice[i], newSlice[i], path+"["+strconv.Itoa(i)+"]")
				changes = append(changes, subChanges...)
			}

			// Handle extra elements
			if len(oldSlice) > len(newSlice) {
				for i := len(newSlice); i < len(oldSlice); i++ {
					changes = append(changes, Change{
						Type:     Deletion,
						Path:     path + "[" + strconv.Itoa(i) + "]",
						OldValue: oldSlice[i],
						NewValue: nil,
					})
				}
			} else if len(newSlice) > len(oldSlice) {
				for i := len(oldSlice); i < len(newSlice); i++ {
					changes = append(changes, Change{
						Type:     Addition,
						Path:     path + "[" + strconv.Itoa(i) + "]",
						OldValue: nil,
						NewValue: newSlice[i],
					})
				}
			}
		}

	default:
		// Primitive values - if they're different, it's a modification
		if !reflect.DeepEqual(oldVal, newVal) {
			changes = append(changes, Change{
				Type:     Modification,
				Path:     path,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}

	return changes
}

// normalizeValue recursively normalizes a YAML value by sorting maps and slices
func normalizeValue(v interface{}) interface{} {
	if v == nil {
		return v
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Map:
		// Sort map keys
		keys := make([]reflect.Value, 0, val.Len())
		for _, key := range val.MapKeys() {
			keys = append(keys, key)
		}

		// Sort keys by their string representation
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
		})

		// Create normalized map
		normalized := make(map[interface{}]interface{})
		for _, key := range keys {
			normalized[key.Interface()] = normalizeValue(val.MapIndex(key).Interface())
		}
		return normalized

	case reflect.Slice:
		// Sort slice elements
		elements := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			elements[i] = normalizeValue(val.Index(i).Interface())
		}

		// Only sort slices that are not lists of dictionaries with identifiers
		if !isSliceOfDictsWithIds(elements) {
			// Sort by string representation for consistency
			sort.Slice(elements, func(i, j int) bool {
				return fmt.Sprintf("%v", elements[i]) < fmt.Sprintf("%v", elements[j])
			})
		}
		return elements

	default:
		return v
	}
}

// YAMLDocument holds a document with its comments
type YAMLDocument struct {
	Data     interface{}
	Comments []string
}

// Global configuration flags
var disableComments bool
var noDocComment bool
var noColor bool

// printHelp displays the help message
func printHelp() {
	helpText := `ymldiff - A smart YAML diff tool with semantic comparison

USAGE:
    ymldiff [OPTIONS] <file1.yaml> <file2.yaml>

DESCRIPTION:
    ymldiff is an intelligent YAML comparison tool that goes beyond simple text
    diffs. It understands YAML structure and provides meaningful, colored output
    showing additions, deletions, and modifications.

OPTIONS:
    -h, --help              Show this help message and exit
    -c, --disable-comments  Disable display of YAML comments in output
    -d, --no-doc-comment    Disable document separator comments (--- # YAML Document: X/Y)
    -n, --no-color          Disable colored output

EXAMPLES:
    # Basic comparison
    ymldiff old.yaml new.yaml

    # Compare without showing comments
    ymldiff -c config1.yaml config2.yaml
    ymldiff --disable-comments config1.yaml config2.yaml

    # Compare without document separator comments
    ymldiff -d config1.yaml config2.yaml

    # Compare without colors (for piping to files or logs)
    ymldiff -n config1.yaml config2.yaml

    # Combine multiple options (short flags can be combined)
    ymldiff -cd config1.yaml config2.yaml
    ymldiff -cdn config1.yaml config2.yaml

AUTHOR:
    Marek Wajdzik <marek@jest.pro>

LICENSE:
    MIT License
`
	fmt.Print(helpText)
}

// parseYAML parses a YAML file and normalizes it, handling multiple documents and preserving comments
func parseYAML(filename string) ([]YAMLDocument, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var documents []YAMLDocument
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var node yaml.Node
		if err := decoder.Decode(&node); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Extract comments from the node
		comments := extractComments(&node)

		// Convert node to interface{}
		var doc interface{}
		if err := node.Decode(&doc); err != nil {
			return nil, err
		}

		documents = append(documents, YAMLDocument{
			Data:     normalizeValue(doc),
			Comments: comments,
		})
	}

	return documents, nil
}

// extractComments recursively extracts all comments from a YAML node
func extractComments(node *yaml.Node) []string {
	var comments []string

	if node.HeadComment != "" {
		lines := strings.Split(strings.TrimSpace(node.HeadComment), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				if !strings.HasPrefix(line, "#") {
					line = "# " + line
				}
				comments = append(comments, line)
			}
		}
	}

	if node.LineComment != "" {
		line := strings.TrimSpace(node.LineComment)
		if !strings.HasPrefix(line, "#") {
			line = "# " + line
		}
		comments = append(comments, line)
	}

	if node.FootComment != "" {
		lines := strings.Split(strings.TrimSpace(node.FootComment), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				if !strings.HasPrefix(line, "#") {
					line = "# " + line
				}
				comments = append(comments, line)
			}
		}
	}

	// Recursively extract from children
	for _, child := range node.Content {
		comments = append(comments, extractComments(child)...)
	}

	return comments
}

func main() {
	// Define flags with pflag (supports POSIX-style flag combining like -cd)
	helpFlag := flag.BoolP("help", "h", false, "Show help message")
	disableCommentsFlag := flag.BoolP("disable-comments", "c", false, "Disable display of YAML comments")
	noDocCommentFlag := flag.BoolP("no-doc-comment", "d", false, "Disable document separator comments")
	noColorFlag := flag.BoolP("no-color", "n", false, "Disable colored output")

	// Custom usage function
	flag.Usage = func() {
		printHelp()
	}

	// Parse flags
	flag.Parse()

	// Check for help flags
	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	// Set global flags
	disableComments = *disableCommentsFlag
	noDocComment = *noDocCommentFlag
	noColor = *noColorFlag

	// Disable colors globally if flag is set
	if noColor {
		color.NoColor = true
	}

	// Get remaining arguments (file names)
	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: Expected exactly 2 YAML files to compare\n\n")
		printHelp()
		os.Exit(1)
	}

	file1 := args[0]
	file2 := args[1]

	documents1, err := parseYAML(file1)
	if err != nil {
		log.Fatalf("Error parsing %s: %v", file1, err)
	}

	documents2, err := parseYAML(file2)
	if err != nil {
		log.Fatalf("Error parsing %s: %v", file2, err)
	}

	// Compare documents by index
	maxDocs := len(documents1)
	if len(documents2) > maxDocs {
		maxDocs = len(documents2)
	}

	blue := color.New(color.FgBlue)

	// Determine total document count for the header
	totalDocs := maxDocs

	for i := 0; i < maxDocs; i++ {
		var doc1Data, doc2Data interface{}
		var comments []string

		if i < len(documents1) {
			doc1Data = documents1[i].Data
			comments = documents1[i].Comments
		}
		if i < len(documents2) {
			doc2Data = documents2[i].Data
			// Merge comments from both documents, preferring doc2
			if len(documents2[i].Comments) > 0 {
				comments = documents2[i].Comments
			}
		}

		// Skip if both documents are nil
		if doc1Data == nil && doc2Data == nil {
			continue
		}

		changes := diffValues(doc1Data, doc2Data, "")

		// Skip documents with no changes
		if len(changes) == 0 {
			continue
		}

		// Output document separator with inline comment
		if noDocComment {
			blue.Println("---")
		} else {
			blue.Printf("--- # YAML Document: %d/%d\n", i+1, totalDocs)
		}

		// Output all comments from the document (unless disabled)
		if !disableComments {
			for _, comment := range comments {
				blue.Println(comment)
			}
		}

		// Generate colored diff output showing only changes
		coloredDiff := generateColoredDiff(changes)
		fmt.Print(coloredDiff)
		fmt.Println() // Add blank line between documents
	}
}
