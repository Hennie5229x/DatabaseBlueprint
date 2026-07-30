package models

type Column struct {
	ColumnID           int
	ColumnName         string
	DataType           string
	MaxLength          int
	Precision          int
	Scale              int
	IsNullable         bool
	IsIdentity         bool
	IdentitySeed       string
	IdentityIncrement  string
	ComputedDefinition string
	IsPersisted        bool
}

type DefaultConstraint struct {
	ColumnID        int
	ColumnName      string
	ConstraintName  string
	ConstraintValue string
}

type PrimaryKeyColumn struct {
	ConstraintObjectID int
	IndexType          string
	ColumnName         string
	KeyOrdinal         int
	IsDescending       bool
}

type UniqueConstraintColumn struct {
	ConstraintObjectID int
	IndexType          string
	ColumnName         string
	KeyOrdinal         int
	IsDescending       bool
}

type ForeignKeyColumn struct {
	ForeignKeyName     string
	ForeignKeyObjectID int
	ParentSchema       string
	ParentTable        string
	ColumnName         string
	KeyOrdinal         int
	ReferencedSchema   string
	ReferencedTable    string
	ReferencedColumn   string
	DeleteAction       string
	UpdateAction       string
}

type CheckConstraint struct {
	ConstraintObjectID int
	Definition         string
}

type IndexColumn struct {
	IndexID           int
	IndexName         string
	IndexType         string
	IsUnique          bool
	ColumnName        string
	KeyOrdinal        int
	IsDescending      bool
	IsIncluded        bool
	IncludeOrder      int
	HasFilter         bool
	FilterDefinition  string
	IsUserDefinedName bool
}

type Views struct {
	Schema     string
	View       string
	Definition string
}

type Functions struct {
	Schema     string
	Name       string
	Definition string
}

type Procedures struct {
	Schema     string
	Name       string
	Definition string
}

type DependencyRow struct {
	ReferencingObjectID int    `gorm:"column:ReferencingObjectId"`
	ReferencingSchema   string `gorm:"column:ReferencingSchema"`
	ReferencingObject   string `gorm:"column:ReferencingObject"`
	ReferencingTypeCode string `gorm:"column:ReferencingTypeCode"`
	ReferencingType     string `gorm:"column:ReferencingType"`

	ReferencedObjectID *int    `gorm:"column:ReferencedObjectId"`
	ReferencedSchema   *string `gorm:"column:ReferencedSchema"`
	ReferencedObject   *string `gorm:"column:ReferencedObject"`
	ReferencedTypeCode *string `gorm:"column:ReferencedTypeCode"`
	ReferencedType     *string `gorm:"column:ReferencedType"`
}

type DatabaseObject struct {
	ID       int
	Schema   string
	Name     string
	TypeCode string
	Type     string

	DependsOn map[int]struct{}
}

type OrderedObject struct {
	DatabaseObject
	CreationLevel int
	CreationOrder int
}

type RunOrderFile []RunOrderObject

type RunOrderObject struct {
	Type           string   `json:"type"`
	Name           string   `json:"name"`
	File           string   `json:"file"`
	DependsOnNames []string `json:"dependsOn,omitempty"`

	CreationOrder int          `json:"-"`
	CreationLevel int          `json:"-"`
	ObjectID      int          `json:"-"`
	Schema        string       `json:"-"`
	DependsOn     []Dependency `json:"-"`
}

type Dependency struct {
	ObjectID int    `json:"-"`
	Schema   string `json:"-"`
	Name     string `json:"-"`
	Type     string `json:"-"`
	File     string `json:"-"`
}
