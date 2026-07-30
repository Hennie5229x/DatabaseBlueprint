package sqlserver

import (
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
	if err := generateRunOrder(db, outputDirectory, databaseName); err != nil {
		panic(err)
	}
}

func generateRunOrder(db *gorm.DB, outputDirectory string, databaseName string) error {
	dependencyRows, err := queryObjectDependencies(db)
	if err != nil {
		return err
	}

	objectGraph := BuildObjectGraph(dependencyRows)

	orderedObjects, err := BuildCreationOrder(objectGraph)
	if err != nil {
		return fmt.Errorf("build SQL Server creation order: %w", err)
	}

	runOrder := BuildRunOrderFile(databaseName, orderedObjects, objectGraph)
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
		left := objects[ready[i]]
		right := objects[ready[j]]

		return left.Schema+"."+left.Name <
			right.Schema+"."+right.Name
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
			left := objects[ready[i]]
			right := objects[ready[j]]

			if creationLevel[ready[i]] != creationLevel[ready[j]] {
				return creationLevel[ready[i]] <
					creationLevel[ready[j]]
			}

			return left.Schema+"."+left.Name <
				right.Schema+"."+right.Name
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

func BuildRunOrderFile(databaseName string, orderedObjects []sqlservermodels.OrderedObject, objectGraph map[int]*sqlservermodels.DatabaseObject) sqlservermodels.RunOrderFile {
	runOrder := make(sqlservermodels.RunOrderFile, 0, len(orderedObjects))

	for _, orderedObject := range orderedObjects {
		dependencies := make([]sqlservermodels.Dependency, 0, len(orderedObject.DependsOn))
		dependencyNames := make([]string, 0, len(orderedObject.DependsOn))
		for dependencyID := range orderedObject.DependsOn {
			dependency, exists := objectGraph[dependencyID]
			if !exists {
				continue
			}

			dependencies = append(dependencies, sqlservermodels.Dependency{
				ObjectID: dependency.ID,
				Schema:   dependency.Schema,
				Name:     dependency.Name,
				Type:     dependency.TypeCode,
				File:     runOrderFilePath(*dependency),
			})
		}

		sort.Slice(dependencies, func(i, j int) bool {
			if dependencies[i].Schema != dependencies[j].Schema {
				return dependencies[i].Schema < dependencies[j].Schema
			}

			return dependencies[i].Name < dependencies[j].Name
		})
		for _, dependency := range dependencies {
			dependencyNames = append(dependencyNames, qualifiedObjectName(dependency.Schema, dependency.Name))
		}

		runOrder = append(runOrder, sqlservermodels.RunOrderObject{
			Type:           orderedObject.TypeCode,
			Name:           qualifiedObjectName(orderedObject.Schema, orderedObject.Name),
			File:           runOrderFilePath(orderedObject.DatabaseObject),
			DependsOnNames: dependencyNames,
			CreationOrder:  orderedObject.CreationOrder,
			CreationLevel:  orderedObject.CreationLevel,
			ObjectID:       orderedObject.ID,
			Schema:         orderedObject.Schema,
			DependsOn:      dependencies,
		})
	}

	return runOrder
}

func ValidateRunOrder(runOrder sqlservermodels.RunOrderFile) error {
	seenObjects := make(map[int]int, len(runOrder))

	for index, object := range runOrder {
		expectedOrder := index + 1
		if object.CreationOrder != expectedOrder {
			return fmt.Errorf("invalid run order for [%s].[%s]: expected creationOrder %d, got %d", object.Schema, object.Name, expectedOrder, object.CreationOrder)
		}

		if previousOrder, exists := seenObjects[object.ObjectID]; exists {
			return fmt.Errorf("duplicate object in run order: objectId %d appears at creationOrder %d and %d", object.ObjectID, previousOrder, object.CreationOrder)
		}

		for _, dependency := range object.DependsOn {
			dependencyOrder, exists := seenObjects[dependency.ObjectID]
			if !exists {
				return fmt.Errorf("invalid dependency order for [%s].[%s]: dependency [%s].[%s] must appear earlier", object.Schema, object.Name, dependency.Schema, dependency.Name)
			}

			if dependencyOrder >= object.CreationOrder {
				return fmt.Errorf("invalid dependency order for [%s].[%s]: dependency [%s].[%s] appears at creationOrder %d", object.Schema, object.Name, dependency.Schema, dependency.Name, dependencyOrder)
			}
		}

		seenObjects[object.ObjectID] = object.CreationOrder
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

func runOrderFilePath(object sqlservermodels.DatabaseObject) string {
	directory := "Procedures"
	switch object.TypeCode {
	case "V":
		directory = "Views"
	case "FN", "IF", "TF":
		directory = "Functions"
	case "P":
		directory = "Procedures"
	}

	return fmt.Sprintf("%s/%s.sql", directory, qualifiedObjectName(object.Schema, object.Name))
}

func qualifiedObjectName(schemaName string, objectName string) string {
	return schemaName + "." + objectName
}
