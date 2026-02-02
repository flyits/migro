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
}

// Ensure Table implements Blueprint
var _ Blueprint = (*Table)(nil)
