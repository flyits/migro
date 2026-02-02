package schema

// Blueprint is the interface for table schema operations
type Blueprint interface {
	// Column operations
	ID() *Column
	String(name string, length int) *Column
	Text(name string) *Column
	Integer(name string) *Column
	BigInteger(name string) *Column
	SmallInteger(name string) *Column
	TinyInteger(name string) *Column
	Float(name string) *Column
	Double(name string) *Column
	Decimal(name string, precision, scale int) *Column
	Boolean(name string) *Column
	Date(name string) *Column
	DateTime(name string) *Column
	Timestamp(name string) *Column
	Time(name string) *Column
	JSON(name string) *Column
	Binary(name string) *Column
	UUID(name string) *Column
	Timestamps()
	SoftDeletes()

	// Index operations
	Index(columns ...string) *Index
	Unique(columns ...string) *Index
	Primary(columns ...string) *Index

	// Foreign key operations
	Foreign(column string) *ForeignKey

	// ALTER TABLE operations
	DropColumn(name string)
	DropIndex(name string)
	DropForeign(name string)
	RenameColumn(from, to string)

	// Column modification operations
	ChangeColumn(name string, colType ColumnType) *Column
	ChangeString(name string, length int) *Column
	ChangeText(name string) *Column
	ChangeInteger(name string) *Column
	ChangeBigInteger(name string) *Column
	ChangeSmallInteger(name string) *Column
	ChangeTinyInteger(name string) *Column
	ChangeFloat(name string) *Column
	ChangeDouble(name string) *Column
	ChangeDecimal(name string, precision, scale int) *Column
	ChangeBoolean(name string) *Column
	ChangeDate(name string) *Column
	ChangeDateTime(name string) *Column
	ChangeTimestamp(name string) *Column
	ChangeTime(name string) *Column
	ChangeJSON(name string) *Column
	ChangeBinary(name string) *Column
	ChangeUUID(name string) *Column
}

// Ensure Table implements Blueprint
var _ Blueprint = (*Table)(nil)
