package schema

// ColumnType represents the type of a database column
type ColumnType int

const (
	TypeString ColumnType = iota
	TypeText
	TypeInteger
	TypeBigInteger
	TypeSmallInteger
	TypeTinyInteger
	TypeFloat
	TypeDouble
	TypeDecimal
	TypeBoolean
	TypeDate
	TypeDateTime
	TypeTimestamp
	TypeTime
	TypeJSON
	TypeBinary
	TypeUUID
)

// Column represents a database column definition
type Column struct {
	Name            string
	Type            ColumnType
	Length          int
	Precision       int
	Scale           int
	IsNullable      bool
	DefaultValue    interface{}
	IsAutoIncrement bool
	IsUnsigned      bool
	IsPrimary       bool
	IsUnique        bool
	ColumnComment   string
	After           string // for MySQL ALTER TABLE
	Change          bool   // indicates column modification
}

// Nullable sets the column as nullable
func (c *Column) Nullable() *Column {
	c.IsNullable = true
	return c
}

// Default sets the default value for the column
func (c *Column) Default(value interface{}) *Column {
	c.DefaultValue = value
	return c
}

// Unsigned sets the column as unsigned (for numeric types)
func (c *Column) Unsigned() *Column {
	c.IsUnsigned = true
	return c
}

// AutoIncrement sets the column as auto-incrementing
func (c *Column) AutoIncrement() *Column {
	c.IsAutoIncrement = true
	return c
}

// Primary sets the column as primary key
func (c *Column) Primary() *Column {
	c.IsPrimary = true
	return c
}

// Unique sets the column as unique
func (c *Column) Unique() *Column {
	c.IsUnique = true
	return c
}

// Comment sets a comment for the column
func (c *Column) Comment(text string) *Column {
	c.ColumnComment = text
	return c
}

// PlaceAfter places this column after another column (MySQL only)
func (c *Column) PlaceAfter(column string) *Column {
	c.After = column
	return c
}
