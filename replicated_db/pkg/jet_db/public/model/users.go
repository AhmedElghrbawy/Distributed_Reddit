//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type Users struct {
	Handle      string `sql:"primary_key"`
	DisplayName string
	Avatar      []byte
	Karma       int32
	CreatedAt   time.Time
}
