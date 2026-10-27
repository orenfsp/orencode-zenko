package municipality

type Municipality struct {
	ID   int64
	Name string
}

type CreateMunicipality struct {
	Name string
}
