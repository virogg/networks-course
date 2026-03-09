package domain

import domainerr "github.com/virogg/networks-course/service/internal/domain/errors"

type Product struct {
	ID          int64
	Name        string
	Description string
	IconPath    string
}

func NewProduct(name, description string) *Product {
	return &Product{
		Name:        name,
		Description: description,
	}
}

func (p *Product) Validate() error {
	if p.Name == "" {
		return domainerr.ErrInvalidProductName
	}
	return nil
}
