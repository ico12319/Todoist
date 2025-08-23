package resource_identifier

type ResourceIdentifier interface {
	GetResourceIdentifier() string
	SetResourceIdentifier(resourceIdentifier string)
}

type GenericResourceIdentifier struct {
	resourceIdentifier string
}

func (g *GenericResourceIdentifier) GetResourceIdentifier() string {
	return g.resourceIdentifier
}

func (g *GenericResourceIdentifier) SetResourceIdentifier(resourceIdentifier string) {
	g.resourceIdentifier = resourceIdentifier
}
