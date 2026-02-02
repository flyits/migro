package schema

// Table represents a database table definition with fluent API
type Table struct {
	Name           string
	Columns        []*Column
	Indexes        []*Index
	ForeignKeys    []*ForeignKey
	PrimaryKey     []string
	Engine         string // MySQL specific
	Charset        string // MySQL specific
	Collation      string // MySQL specific
	IfNotExists    bool
	IsAlter        bool          // true if this is an ALTER TABLE operation
	DropColumns    []string      // columns to drop in ALTER TABLE
	DropIndexes    []string      // indexes to drop in ALTER TABLE
	DropForeignKeys []string     // foreign keys to drop in ALTER TABLE
	RenameColumns  map[string]string // old name -> new name
}

// NewTable creates a new table definition
func NewTable(name string) *Table {
	return &Table{
		Name:          name,
		Columns:       make([]*Column, 0),
		Indexes:       make([]*Index, 0),
		ForeignKeys:   make([]*ForeignKey, 0),
		RenameColumns: make(map[string]string),
	}
}

// addColumn is a helper to add a column and return it for chaining
func (t *Table) addColumn(name string, colType ColumnType) *Column {
	col := &Column{
		Name: name,
		Type: colType,
	}
	t.Columns = append(t.Columns, col)
	return col
}

// ID adds an auto-incrementing big integer primary key column named "id"
func (t *Table) ID() *Column {
	col := t.addColumn("id", TypeBigInteger)
	col.IsAutoIncrement = true
	col.IsUnsigned = true
	col.IsPrimary = true
	return col
}

// String adds a VARCHAR column
func (t *Table) String(name string, length int) *Column {
	col := t.addColumn(name, TypeString)
	col.Length = length
	return col
}

// Text adds a TEXT column
func (t *Table) Text(name string) *Column {
	return t.addColumn(name, TypeText)
}

// Integer adds an INT column
func (t *Table) Integer(name string) *Column {
	return t.addColumn(name, TypeInteger)
}

// BigInteger adds a BIGINT column
func (t *Table) BigInteger(name string) *Column {
	return t.addColumn(name, TypeBigInteger)
}

// SmallInteger adds a SMALLINT column
func (t *Table) SmallInteger(name string) *Column {
	return t.addColumn(name, TypeSmallInteger)
}

// TinyInteger adds a TINYINT column
func (t *Table) TinyInteger(name string) *Column {
	return t.addColumn(name, TypeTinyInteger)
}

// Float adds a FLOAT column
func (t *Table) Float(name string) *Column {
	return t.addColumn(name, TypeFloat)
}

// Double adds a DOUBLE column
func (t *Table) Double(name string) *Column {
	return t.addColumn(name, TypeDouble)
}

// Decimal adds a DECIMAL column with precision and scale
func (t *Table) Decimal(name string, precision, scale int) *Column {
	col := t.addColumn(name, TypeDecimal)
	col.Precision = precision
	col.Scale = scale
	return col
}

// Boolean adds a BOOLEAN column
func (t *Table) Boolean(name string) *Column {
	return t.addColumn(name, TypeBoolean)
}

// Date adds a DATE column
func (t *Table) Date(name string) *Column {
	return t.addColumn(name, TypeDate)
}

// DateTime adds a DATETIME column
func (t *Table) DateTime(name string) *Column {
	return t.addColumn(name, TypeDateTime)
}

// Timestamp adds a TIMESTAMP column
func (t *Table) Timestamp(name string) *Column {
	return t.addColumn(name, TypeTimestamp)
}

// Time adds a TIME column
func (t *Table) Time(name string) *Column {
	return t.addColumn(name, TypeTime)
}

// JSON adds a JSON column
func (t *Table) JSON(name string) *Column {
	return t.addColumn(name, TypeJSON)
}

// Binary adds a BINARY/BLOB column
func (t *Table) Binary(name string) *Column {
	return t.addColumn(name, TypeBinary)
}

// UUID adds a UUID column (CHAR(36) for MySQL, UUID for PostgreSQL)
func (t *Table) UUID(name string) *Column {
	return t.addColumn(name, TypeUUID)
}

// Timestamps adds created_at and updated_at timestamp columns
func (t *Table) Timestamps() {
	t.Timestamp("created_at").Nullable()
	t.Timestamp("updated_at").Nullable()
}

// SoftDeletes adds a deleted_at timestamp column for soft deletes
func (t *Table) SoftDeletes() {
	t.Timestamp("deleted_at").Nullable()
}

// Index adds an index on the specified columns
func (t *Table) Index(columns ...string) *Index {
	idx := NewIndex(columns...)
	t.Indexes = append(t.Indexes, idx)
	return idx
}

// Unique adds a unique index on the specified columns
func (t *Table) Unique(columns ...string) *Index {
	idx := NewIndex(columns...).Unique()
	t.Indexes = append(t.Indexes, idx)
	return idx
}

// Primary sets the primary key columns
func (t *Table) Primary(columns ...string) *Index {
	idx := NewIndex(columns...).Primary()
	t.PrimaryKey = columns
	t.Indexes = append(t.Indexes, idx)
	return idx
}

// Foreign adds a foreign key constraint
func (t *Table) Foreign(column string) *ForeignKey {
	fk := NewForeignKey(column)
	t.ForeignKeys = append(t.ForeignKeys, fk)
	return fk
}

// DropColumn marks a column for deletion (ALTER TABLE)
func (t *Table) DropColumn(name string) {
	t.DropColumns = append(t.DropColumns, name)
}

// DropIndex marks an index for deletion (ALTER TABLE)
func (t *Table) DropIndex(name string) {
	t.DropIndexes = append(t.DropIndexes, name)
}

// DropForeign marks a foreign key for deletion (ALTER TABLE)
func (t *Table) DropForeign(name string) {
	t.DropForeignKeys = append(t.DropForeignKeys, name)
}

// RenameColumn renames a column (ALTER TABLE)
func (t *Table) RenameColumn(from, to string) {
	t.RenameColumns[from] = to
}

// SetEngine sets the storage engine (MySQL only)
func (t *Table) SetEngine(engine string) *Table {
	t.Engine = engine
	return t
}

// SetCharset sets the character set (MySQL only)
func (t *Table) SetCharset(charset string) *Table {
	t.Charset = charset
	return t
}

// SetCollation sets the collation (MySQL only)
func (t *Table) SetCollation(collation string) *Table {
	t.Collation = collation
	return t
}
