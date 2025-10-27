package models

type NodeInfo struct {
	Role     string
	Name     string
	CPUs     int64
	Memory   int64
	Platform string // Eg: Architecture - OS
}
