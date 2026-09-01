package sqlserver

import (
	"blueprint/cli/spinner"
	sqlservermodels "blueprint/database/discovery/SQLServer/models"
	queries "blueprint/database/discovery/SQLServer/queries"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

func GenerateRunOrder(db *gorm.DB, outputDirectory string, databaseName string) {
	runOrderSpinner := spinner.New("Run Order", "Creating run order")
	if err := generateRunOrder(db, outputDirectory, databaseName); err != nil {
		runOrderSpinner.Stop("Failed")
		panic(err)
	}
	runOrderSpinner.Stop("Run order created")
}

func generateRunOrder(db *gorm.DB, outputDirectory string, databaseName string) error {
	dependencyRows, err := queryObjectDependencies(db)
	if err != nil {
		return err
	}

	objectGraph := BuildObjectGraph(dependencyRows)
	if err := AddFileDependencies(outputDirectory, objectGraph); err != nil {
		return fmt.Errorf("add SQL Server file dependencies: %w", err)
	}

	orderedObjects, err := BuildCreationOrder(objectGraph)
	if err != nil {
		return fmt.Errorf("build SQL Server creation order: %w", err)
	}

	runOrder, err := BuildRunOrderFile(outputDirectory, orderedObjects, objectGraph)
	if err != nil {
		return fmt.Errorf("build SQL Server run order file: %w", err)
	}
	if err := ValidateRunOrder(runOrder); err != nil {
		return fmt.Errorf("validate SQL Server run order: %w", err)
	}

	runOrderPath := filepath.Join(outputDirectory, "RunOrder.json")
	if err := WriteRunOrder(runOrderPath, runOrder); err != nil {
		return fmt.Errorf("write SQL Server run order: %w", err)
	}

	return nil
}

func queryObjectDependencies(db *gorm.DB) ([]sqlservermodels.DependencyRow, error) {
	dependencyRows, err := queries.SqlServerDependencies(db)
	if err != nil {
		return nil, fmt.Errorf("query SQL Server programmable object dependencies: %w", err)
	}

	return dependencyRows, nil
}

func BuildObjectGraph(rows []sqlservermodels.DependencyRow) map[int]*sqlservermodels.DatabaseObject {
	return BuildObjectDAG(rows)
}

func AddFileDependencies(outputDirectory string, objects map[int]*sqlservermodels.DatabaseObject) error {
	exportedFiles, err := exportedRunOrderFiles(outputDirectory)
	if err != nil {
		return err
	}

	objectsByFile := make(map[string]*sqlservermodels.DatabaseObject, len(objects))
	for _, object := range objects {
		objectsByFile[runOrderFilePath(*object)] = object
	}

	// SQL Server metadata can omit valid references, especially for unresolved
	// or unqualified names. Add exported objects so file matching can recover
	// those dependency edges.
	nextSyntheticID := -1
	for _, directory := range runOrderDirectories() {
		if directory == "Schemas" || directory == "DataTypes" {
			continue
		}

		for _, file := range exportedFiles[directory] {
			if _, exists := objectsByFile[file]; exists {
				continue
			}

			fileObject := runOrderFileObjectFromPath(file)
			if fileObject.Schema == "" || fileObject.Name == "" {
				continue
			}

			for {
				if _, exists := objects[nextSyntheticID]; !exists {
					break
				}
				nextSyntheticID--
			}

			object := &sqlservermodels.DatabaseObject{
				ID:        nextSyntheticID,
				Schema:    fileObject.Schema,
				Name:      fileObject.Name,
				TypeCode:  runOrderTypeCode(directory),
				Type:      directory,
				DependsOn: make(map[int]struct{}),
			}
			objects[object.ID] = object
			objectsByFile[file] = object
			nextSyntheticID--
		}
	}

	matchers := make(map[int]sqlObjectMatcher, len(objects))
	for _, object := range objects {
		matcher, err := newSQLObjectMatcher(object.Schema, object.Name)
		if err != nil {
			return err
		}
		matchers[object.ID] = matcher
	}

	for file, object := range objectsByFile {
		contents, err := os.ReadFile(filepath.Join(outputDirectory, file))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		text := normalizeSQLText(string(contents))
		for dependencyID, matcher := range matchers {
			if dependencyID == object.ID {
				continue
			}

			if matcher.matches(text) {
				object.DependsOn[dependencyID] = struct{}{}
			}
		}
	}

	return nil
}

func runOrderTypeCode(directory string) string {
	switch directory {
	case "Tables":
		return "U"
	case "ForeignKeys":
		return "F"
	case "Views":
		return "V"
	case "Functions":
		return "IF"
	case "Procedures":
		return "P"
	case "TableTypes":
		return "TT"
	case "Sequences":
		return "SO"
	case "Synonyms":
		return "SN"
	case "Triggers":
		return "TR"
	default:
		return ""
	}
}

func BuildObjectDAG(rows []sqlservermodels.DependencyRow) map[int]*sqlservermodels.DatabaseObject {
	objects := make(map[int]*sqlservermodels.DatabaseObject)

	for _, row := range rows {
		object, exists := objects[row.ReferencingObjectID]
		if !exists {
			object = &sqlservermodels.DatabaseObject{
				ID:        row.ReferencingObjectID,
				Schema:    row.ReferencingSchema,
				Name:      row.ReferencingObject,
				TypeCode:  row.ReferencingTypeCode,
				Type:      row.ReferencingType,
				DependsOn: make(map[int]struct{}),
			}

			objects[row.ReferencingObjectID] = object
		}

		if row.ReferencedObjectID == nil {
			continue
		}

		dependencyID := *row.ReferencedObjectID

		// Ignore direct self-references.
		if dependencyID == object.ID {
			continue
		}

		object.DependsOn[dependencyID] = struct{}{}
	}

	return objects
}

func BuildCreationOrder(
	objects map[int]*sqlservermodels.DatabaseObject,
) ([]sqlservermodels.OrderedObject, error) {
	unresolvedDependencies := make(map[int]int, len(objects))
	dependants := make(map[int][]int)

	for objectID := range objects {
		unresolvedDependencies[objectID] = 0
	}

	for objectID, object := range objects {
		for dependencyID := range object.DependsOn {
			// Ignore dependencies outside this graph.
			if _, exists := objects[dependencyID]; !exists {
				continue
			}

			unresolvedDependencies[objectID]++

			dependants[dependencyID] = append(
				dependants[dependencyID],
				objectID,
			)
		}
	}

	creationLevel := make(map[int]int, len(objects))
	ready := make([]int, 0)

	for objectID, count := range unresolvedDependencies {
		if count == 0 {
			creationLevel[objectID] = 1
			ready = append(ready, objectID)
		}
	}

	sort.Slice(ready, func(i, j int) bool {
		return runOrderObjectLess(objects[ready[i]], objects[ready[j]])
	})

	ordered := make([]sqlservermodels.OrderedObject, 0, len(objects))

	for len(ready) > 0 {
		objectID := ready[0]
		ready = ready[1:]

		object := objects[objectID]

		ordered = append(ordered, sqlservermodels.OrderedObject{
			DatabaseObject: *object,
			CreationLevel:  creationLevel[objectID],
			CreationOrder:  len(ordered) + 1,
		})

		for _, dependantID := range dependants[objectID] {
			unresolvedDependencies[dependantID]--

			nextLevel := creationLevel[objectID] + 1
			if nextLevel > creationLevel[dependantID] {
				creationLevel[dependantID] = nextLevel
			}

			if unresolvedDependencies[dependantID] == 0 {
				if creationLevel[dependantID] == 0 {
					creationLevel[dependantID] = 1
				}

				ready = append(ready, dependantID)
			}
		}

		sort.Slice(ready, func(i, j int) bool {
			if creationLevel[ready[i]] != creationLevel[ready[j]] {
				return creationLevel[ready[i]] <
					creationLevel[ready[j]]
			}

			return runOrderObjectLess(objects[ready[i]], objects[ready[j]])
		})
	}

	if len(ordered) != len(objects) {
		cyclicObjects := make([]string, 0)

		for objectID, count := range unresolvedDependencies {
			if count > 0 {
				object := objects[objectID]

				cyclicObjects = append(
					cyclicObjects,
					fmt.Sprintf(
						"[%s].[%s]",
						object.Schema,
						object.Name,
					),
				)
			}
		}

		sort.Strings(cyclicObjects)

		return ordered, fmt.Errorf(
			"circular dependencies detected: %s",
			strings.Join(cyclicObjects, ", "),
		)
	}

	return ordered, nil
}

func BuildRunOrderFile(outputDirectory string, orderedObjects []sqlservermodels.OrderedObject, objectGraph map[int]*sqlservermodels.DatabaseObject) (sqlservermodels.RunOrderFile, error) {
	exportedFiles, err := exportedRunOrderFiles(outputDirectory)
	if err != nil {
		return nil, err
	}

	runOrder := make(sqlservermodels.RunOrderFile, 0, len(exportedFiles))
	addedFiles := make(map[string]struct{}, len(exportedFiles))
	for _, directory := range []string{"Schemas", "DataTypes"} {
		for _, file := range exportedFiles[directory] {
			runOrder = append(runOrder, runOrderObjectFromFile(directory, file, len(runOrder)+1))
			addedFiles[file] = struct{}{}
		}
	}

	exportedFileSet := make(map[string]struct{}, len(exportedFiles))
	for _, files := range exportedFiles {
		for _, file := range files {
			exportedFileSet[file] = struct{}{}
		}
	}

	for _, orderedObject := range orderedObjects {
		file := runOrderFilePath(orderedObject.DatabaseObject)
		if _, exists := addedFiles[file]; exists {
			continue
		}
		if _, exists := exportedFileSet[file]; !exists {
			continue
		}

		runOrder = append(runOrder, runOrderObject(orderedObject.DatabaseObject, file, len(runOrder)+1))
		addedFiles[file] = struct{}{}
	}

	// Keep exported files that are not represented in the dependency graph.
	for _, directory := range runOrderDirectories() {
		for _, file := range exportedFiles[directory] {
			if _, exists := addedFiles[file]; exists {
				continue
			}

			runOrder = append(runOrder, runOrderObjectFromFile(directory, file, len(runOrder)+1))
			addedFiles[file] = struct{}{}
		}
	}

	return runOrder, nil
}

func ValidateRunOrder(runOrder sqlservermodels.RunOrderFile) error {
	seenFiles := make(map[string]int, len(runOrder))

	for index, object := range runOrder {
		expectedOrder := index + 1
		if object.Order != expectedOrder {
			return fmt.Errorf("invalid run order for %s: expected order %d, got %d", object.File, expectedOrder, object.Order)
		}

		if object.File == "" {
			return fmt.Errorf("invalid run order at order %d: file is empty", object.Order)
		}

		if previousOrder, exists := seenFiles[object.File]; exists {
			return fmt.Errorf("duplicate file in run order: %s appears at order %d and %d", object.File, previousOrder, object.Order)
		}

		seenFiles[object.File] = object.Order
	}

	return nil
}

func WriteRunOrder(path string, runOrder sqlservermodels.RunOrderFile) error {
	contents, err := json.MarshalIndent(runOrder, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal run order JSON: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create run order directory: %w", err)
	}

	if err := os.WriteFile(path, append(contents, '\n'), 0o644); err != nil {
		return fmt.Errorf("write run order %s: %w", path, err)
	}

	return nil
}

func exportedRunOrderFiles(outputDirectory string) (map[string][]string, error) {
	filesByDirectory := make(map[string][]string)

	for _, directory := range runOrderDirectories() {
		directoryPath := filepath.Join(outputDirectory, directory)
		entries, err := os.ReadDir(directoryPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s scripts: %w", directory, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
				continue
			}

			filesByDirectory[directory] = append(filesByDirectory[directory], filepath.ToSlash(filepath.Join(directory, entry.Name())))
		}

		sort.Strings(filesByDirectory[directory])
	}

	return filesByDirectory, nil
}

func orderPhaseFiles(outputDirectory string, files []string) ([]string, error) {
	if len(files) < 2 {
		return files, nil
	}

	phaseObjects := make([]runOrderFileObject, 0, len(files))
	matchers := make([]sqlObjectMatcher, 0, len(files))
	for _, file := range files {
		object := runOrderFileObjectFromPath(file)
		phaseObjects = append(phaseObjects, object)
		matcher, err := newSQLObjectMatcher(object.Schema, object.Name)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, matcher)
	}

	fileSet := make(map[string]struct{}, len(files))
	dependants := make(map[string][]string)
	unresolvedDependencies := make(map[string]int, len(files))
	for _, file := range files {
		fileSet[file] = struct{}{}
		unresolvedDependencies[file] = 0
	}

	for _, file := range files {
		contents, err := os.ReadFile(filepath.Join(outputDirectory, file))
		if err != nil {
			return nil, err
		}
		text := normalizeSQLText(string(contents))

		for index, dependency := range phaseObjects {
			dependencyFile := dependency.File
			if dependencyFile == file {
				continue
			}
			if _, exists := fileSet[dependencyFile]; !exists {
				continue
			}

			if matchers[index].matches(text) {
				unresolvedDependencies[file]++
				dependants[dependencyFile] = append(dependants[dependencyFile], file)
			}
		}
	}

	ready := make([]string, 0)
	for _, file := range files {
		if unresolvedDependencies[file] == 0 {
			ready = append(ready, file)
		}
	}
	sort.Strings(ready)

	ordered := make([]string, 0, len(files))
	for len(ready) > 0 {
		file := ready[0]
		ready = ready[1:]
		ordered = append(ordered, file)

		for _, dependant := range dependants[file] {
			unresolvedDependencies[dependant]--
			if unresolvedDependencies[dependant] == 0 {
				ready = append(ready, dependant)
			}
		}
		sort.Strings(ready)
	}

	if len(ordered) != len(files) {
		return files, nil
	}

	return ordered, nil
}

func runOrderDirectories() []string {
	return []string{
		"Schemas",
		"DataTypes",
		"TableTypes",
		"Sequences",
		"Synonyms",
		"Functions",
		"Tables",
		"ForeignKeys",
		"Views",
		"Procedures",
		"Triggers",
	}
}

func runOrderObjectLess(left *sqlservermodels.DatabaseObject, right *sqlservermodels.DatabaseObject) bool {
	leftPriority := runOrderTypePriority(left.TypeCode)
	rightPriority := runOrderTypePriority(right.TypeCode)
	if leftPriority != rightPriority {
		return leftPriority < rightPriority
	}

	return left.Schema+"."+left.Name < right.Schema+"."+right.Name
}

func runOrderTypePriority(typeCode string) int {
	switch typeCode {
	case "TT":
		return 10
	case "SO":
		return 20
	case "SN":
		return 30
	case "FN", "IF", "TF":
		return 40
	case "U":
		return 50
	case "F":
		return 60
	case "V":
		return 70
	case "P":
		return 80
	case "TR":
		return 90
	default:
		return 100
	}
}

func sqlTextReferencesObject(text string, schemaName string, objectName string) bool {
	matcher, err := newSQLObjectMatcher(schemaName, objectName)
	return err == nil && matcher.matches(normalizeSQLText(text))
}

type sqlObjectMatcher struct {
	terms []string
}

func newSQLObjectMatcher(schemaName string, objectName string) (sqlObjectMatcher, error) {
	terms := []string{
		strings.ToLower(schemaName + "." + objectName),
		strings.ToLower(objectName),
	}

	return sqlObjectMatcher{terms: terms}, nil
}

func (matcher sqlObjectMatcher) matches(text string) bool {
	for _, term := range matcher.terms {
		for offset := strings.Index(text, term); offset >= 0; {
			end := offset + len(term)
			if sqlIdentifierBoundary(text, offset-1) && sqlIdentifierBoundary(text, end) {
				return true
			}

			next := strings.Index(text[end:], term)
			if next < 0 {
				break
			}
			offset = end + next
		}
	}

	return false
}

func normalizeSQLText(text string) string {
	return strings.ToLower(strings.NewReplacer("[", "", "]", "").Replace(text))
}

func sqlIdentifierBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	character := text[index]
	return !((character >= 'a' && character <= 'z') ||
		(character >= '0' && character <= '9') ||
		character == '_' || character == '#')
}

func runOrderObject(object sqlservermodels.DatabaseObject, file string, order int) sqlservermodels.RunOrderObject {
	return sqlservermodels.RunOrderObject{
		Order:    order,
		Name:     qualifiedObjectName(object.Schema, object.Name),
		File:     file,
		Type:     object.TypeCode,
		ObjectID: object.ID,
		Schema:   object.Schema,
	}
}

func runOrderObjectFromFile(directory string, file string, order int) sqlservermodels.RunOrderObject {
	name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	return sqlservermodels.RunOrderObject{
		Order: order,
		Name:  name,
		File:  file,
		Type:  directory,
	}
}

type runOrderFileObject struct {
	File   string
	Schema string
	Name   string
}

func runOrderFileObjectFromPath(file string) runOrderFileObject {
	name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return runOrderFileObject{File: file, Name: name}
	}

	return runOrderFileObject{File: file, Schema: parts[0], Name: parts[1]}
}

func runOrderFilePath(object sqlservermodels.DatabaseObject) string {
	directory := "Procedures"
	switch object.TypeCode {
	case "U":
		directory = "Tables"
	case "F":
		directory = "ForeignKeys"
	case "V":
		directory = "Views"
	case "FN", "IF", "TF":
		directory = "Functions"
	case "P":
		directory = "Procedures"
	case "TT":
		directory = "TableTypes"
	case "SO":
		directory = "Sequences"
	case "SN":
		directory = "Synonyms"
	case "TR":
		directory = "Triggers"
	}

	return fmt.Sprintf("%s/%s.sql", directory, qualifiedObjectName(object.Schema, object.Name))
}

func qualifiedObjectName(schemaName string, objectName string) string {
	return schemaName + "." + objectName
}
