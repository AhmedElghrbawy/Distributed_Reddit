//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var PostTags = newPostTagsTable("public", "post_tags", "")

type postTagsTable struct {
	postgres.Table

	// Columns
	TagName postgres.ColumnString
	PostID  postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type PostTagsTable struct {
	postTagsTable

	EXCLUDED postTagsTable
}

// AS creates new PostTagsTable with assigned alias
func (a PostTagsTable) AS(alias string) *PostTagsTable {
	return newPostTagsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new PostTagsTable with assigned schema name
func (a PostTagsTable) FromSchema(schemaName string) *PostTagsTable {
	return newPostTagsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new PostTagsTable with assigned table prefix
func (a PostTagsTable) WithPrefix(prefix string) *PostTagsTable {
	return newPostTagsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new PostTagsTable with assigned table suffix
func (a PostTagsTable) WithSuffix(suffix string) *PostTagsTable {
	return newPostTagsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newPostTagsTable(schemaName, tableName, alias string) *PostTagsTable {
	return &PostTagsTable{
		postTagsTable: newPostTagsTableImpl(schemaName, tableName, alias),
		EXCLUDED:      newPostTagsTableImpl("", "excluded", ""),
	}
}

func newPostTagsTableImpl(schemaName, tableName, alias string) postTagsTable {
	var (
		TagNameColumn  = postgres.StringColumn("tag_name")
		PostIDColumn   = postgres.StringColumn("post_id")
		allColumns     = postgres.ColumnList{TagNameColumn, PostIDColumn}
		mutableColumns = postgres.ColumnList{}
	)

	return postTagsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		TagName: TagNameColumn,
		PostID:  PostIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
