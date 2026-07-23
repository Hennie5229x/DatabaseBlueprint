package sqlserver

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
	ForeignKeyObjectID int
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
