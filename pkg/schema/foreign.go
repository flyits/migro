package schema

// ForeignKeyAction represents the action to take on foreign key events
type ForeignKeyAction string

const (
	ActionCascade  ForeignKeyAction = "CASCADE"
	ActionRestrict ForeignKeyAction = "RESTRICT"
	ActionSetNull  ForeignKeyAction = "SET NULL"
	ActionNoAction ForeignKeyAction = "NO ACTION"
)

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Name            string
	Columns         []string
	ReferenceTable  string
	ReferenceColumn string
	OnDelete        ForeignKeyAction
	OnUpdate        ForeignKeyAction
}

// NewForeignKey creates a new foreign key for the given column
func NewForeignKey(column string) *ForeignKey {
	return &ForeignKey{
		Columns:  []string{column},
		OnDelete: ActionRestrict,
		OnUpdate: ActionRestrict,
	}
}

// Named sets the name of the foreign key constraint
func (f *ForeignKey) Named(name string) *ForeignKey {
	f.Name = name
	return f
}

// References sets the referenced table and column
func (f *ForeignKey) References(table, column string) *ForeignKey {
	f.ReferenceTable = table
	f.ReferenceColumn = column
	return f
}

// OnDeleteCascade sets the ON DELETE action to CASCADE
func (f *ForeignKey) OnDeleteCascade() *ForeignKey {
	f.OnDelete = ActionCascade
	return f
}

// OnDeleteSetNull sets the ON DELETE action to SET NULL
func (f *ForeignKey) OnDeleteSetNull() *ForeignKey {
	f.OnDelete = ActionSetNull
	return f
}

// OnDeleteRestrict sets the ON DELETE action to RESTRICT
func (f *ForeignKey) OnDeleteRestrict() *ForeignKey {
	f.OnDelete = ActionRestrict
	return f
}

// OnUpdateCascade sets the ON UPDATE action to CASCADE
func (f *ForeignKey) OnUpdateCascade() *ForeignKey {
	f.OnUpdate = ActionCascade
	return f
}

// OnUpdateSetNull sets the ON UPDATE action to SET NULL
func (f *ForeignKey) OnUpdateSetNull() *ForeignKey {
	f.OnUpdate = ActionSetNull
	return f
}

// OnUpdateRestrict sets the ON UPDATE action to RESTRICT
func (f *ForeignKey) OnUpdateRestrict() *ForeignKey {
	f.OnUpdate = ActionRestrict
	return f
}
