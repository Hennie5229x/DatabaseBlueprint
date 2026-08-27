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

type UserDefinedTableType struct {
	SchemaName        string
	TypeName          string
	IsMemoryOptimized bool
}

type UserDefinedType struct {
	SchemaName   string
	TypeName     string
	BaseTypeName string
	MaxLength    int
	Precision    int
	Scale        int
	IsNullable   bool
}

type UserDefinedTableTypeColumn struct {
	SchemaName            string
	TypeName              string
	ColumnID              int
	ColumnName            string
	DataTypeName          string
	MaxLength             int
	Precision             int
	Scale                 int
	CollationName         string
	IsNullable            bool
	IsIdentity            bool
	IsComputed            bool
	IdentitySeed          string
	IdentityIncrement     string
	ComputedDefinition    string
	IsPersisted           bool
	DefaultConstraintName string
	DefaultDefinition     string
}

type UserDefinedTableTypeKeyColumn struct {
	SchemaName         string
	TypeName           string
	ConstraintObjectID int
	ConstraintName     string
	ConstraintType     string
	IndexName          string
	IndexType          string
	ColumnName         string
	KeyOrdinal         int
	IsDescending       bool
	BucketCount        int
}

type UserDefinedTableTypeCheckConstraint struct {
	SchemaName         string
	TypeName           string
	ConstraintObjectID int
	ConstraintName     string
	ParentColumnID     int
	Definition         string
}

type UserDefinedTableTypeIndexColumn struct {
	SchemaName       string
	TypeName         string
	IndexID          int
	IndexName        string
	IndexType        string
	IsUnique         bool
	ColumnName       string
	KeyOrdinal       int
	IsDescending     bool
	IsIncluded       bool
	IncludeOrder     int
	HasFilter        bool
	FilterDefinition string
	BucketCount      int
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
	Order int    `json:"order"`
	Name  string `json:"name"`
	File  string `json:"file"`

	Type      string       `json:"-"`
	ObjectID  int          `json:"-"`
	Schema    string       `json:"-"`
	DependsOn []Dependency `json:"-"`
}

type Dependency struct {
	ObjectID int    `json:"-"`
	Schema   string `json:"-"`
	Name     string `json:"-"`
	Type     string `json:"-"`
	File     string `json:"-"`
}

type Synonyms struct {
	SchemaName     string
	SynonymName    string
	BaseObjectName string
}

type Sequences struct {
	SchemaName         string
	SequenceName       string
	DataTypeSchemaName string
	DataType           string
	Precision          int
	Scale              int
	StartValue         string
	IncrementBy        string
	MinValue           string
	MaxValue           string
	IsCycling          bool
	IsCached           bool
	CacheSize          *int
}

type Triggers struct {
	SchemaName          string
	TableSchemaName     string
	TableName           string
	TriggerName         string
	Definition          string
	IsInsteadOf         bool
	IsDisabled          bool
	IsNotForReplication bool
}

type Schemas struct {
	Name string
}

type DatabaseMetadata struct {
	Collation string `json:"collation"`
}
