package main

type Config struct {
	Tables []*TableConfig
}

type TableConfig struct {
	Name string
}
