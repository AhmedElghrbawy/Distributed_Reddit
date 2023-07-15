package main

type Executer interface {
	Execute(rdb *rdbServer) (interface{}, error) 
}