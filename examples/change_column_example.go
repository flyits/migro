package examples

import "github.com/migro/migro/pkg/schema"

// ExampleChangeColumnType demonstrates how to modify column types in migrations
func ExampleChangeColumnType() {
	// Example 1: Change a column from VARCHAR to TEXT
	table := schema.NewTable("users")
	table.IsAlter = true
	table.ChangeText("bio") // Change 'bio' column to TEXT type

	// Example 2: Change a column from INT to BIGINT
	table2 := schema.NewTable("posts")
	table2.IsAlter = true
	table2.ChangeBigInteger("view_count").Unsigned() // Change to BIGINT UNSIGNED

	// Example 3: Change VARCHAR length
	table3 := schema.NewTable("products")
	table3.IsAlter = true
	table3.ChangeString("name", 500) // Change VARCHAR(255) to VARCHAR(500)

	// Example 4: Change column type and add constraints
	table4 := schema.NewTable("orders")
	table4.IsAlter = true
	table4.ChangeDecimal("total_amount", 10, 2).Nullable().Default(0.00)

	// Example 5: Change multiple columns in one migration
	table5 := schema.NewTable("customers")
	table5.IsAlter = true
	table5.ChangeString("email", 320).Unique()
	table5.ChangeTimestamp("last_login").Nullable()
	table5.ChangeBoolean("is_active").Default(true)
}

// MigrationExample shows a complete migration with column changes
func MigrationExample() {
	// In your migration file:
	// Up method - modify columns
	upTable := schema.NewTable("articles")
	upTable.IsAlter = true
	upTable.ChangeText("content")                    // Change to TEXT for longer content
	upTable.ChangeBigInteger("author_id").Unsigned() // Change to BIGINT for larger IDs
	upTable.ChangeString("slug", 255).Unique()       // Ensure slug is unique

	// Down method - revert changes
	downTable := schema.NewTable("articles")
	downTable.IsAlter = true
	downTable.ChangeString("content", 1000)      // Revert to VARCHAR(1000)
	downTable.ChangeInteger("author_id")         // Revert to INT
	downTable.ChangeString("slug", 255)          // Remove unique constraint (handled separately)
}
