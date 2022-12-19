package model

type Cache struct {
	Status  int
	Headers map[string][]string
	Body    []byte
}
