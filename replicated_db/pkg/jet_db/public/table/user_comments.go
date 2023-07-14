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

var UserComments = newUserCommentsTable("public", "user_comments", "")

type userCommentsTable struct {
	postgres.Table

	// Columns
	ID              postgres.ColumnString
	Content         postgres.ColumnString
	Image           postgres.ColumnString
	NumberOfVotes   postgres.ColumnInteger
	OwnerHandle     postgres.ColumnString
	ParentCommentID postgres.ColumnString
	PostID          postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type UserCommentsTable struct {
	userCommentsTable

	EXCLUDED userCommentsTable
}

// AS creates new UserCommentsTable with assigned alias
func (a UserCommentsTable) AS(alias string) *UserCommentsTable {
	return newUserCommentsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new UserCommentsTable with assigned schema name
func (a UserCommentsTable) FromSchema(schemaName string) *UserCommentsTable {
	return newUserCommentsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new UserCommentsTable with assigned table prefix
func (a UserCommentsTable) WithPrefix(prefix string) *UserCommentsTable {
	return newUserCommentsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new UserCommentsTable with assigned table suffix
func (a UserCommentsTable) WithSuffix(suffix string) *UserCommentsTable {
	return newUserCommentsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newUserCommentsTable(schemaName, tableName, alias string) *UserCommentsTable {
	return &UserCommentsTable{
		userCommentsTable: newUserCommentsTableImpl(schemaName, tableName, alias),
		EXCLUDED:          newUserCommentsTableImpl("", "excluded", ""),
	}
}

func newUserCommentsTableImpl(schemaName, tableName, alias string) userCommentsTable {
	var (
		IDColumn              = postgres.StringColumn("id")
		ContentColumn         = postgres.StringColumn("content")
		ImageColumn           = postgres.StringColumn("image")
		NumberOfVotesColumn   = postgres.IntegerColumn("number_of_votes")
		OwnerHandleColumn     = postgres.StringColumn("owner_handle")
		ParentCommentIDColumn = postgres.StringColumn("parent_comment_id")
		PostIDColumn          = postgres.StringColumn("post_id")
		allColumns            = postgres.ColumnList{IDColumn, ContentColumn, ImageColumn, NumberOfVotesColumn, OwnerHandleColumn, ParentCommentIDColumn, PostIDColumn}
		mutableColumns        = postgres.ColumnList{ContentColumn, ImageColumn, NumberOfVotesColumn, OwnerHandleColumn, ParentCommentIDColumn, PostIDColumn}
	)

	return userCommentsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:              IDColumn,
		Content:         ContentColumn,
		Image:           ImageColumn,
		NumberOfVotes:   NumberOfVotesColumn,
		OwnerHandle:     OwnerHandleColumn,
		ParentCommentID: ParentCommentIDColumn,
		PostID:          PostIDColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
