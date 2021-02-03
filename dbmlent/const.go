package dbmlent

const entTemplate = `package schema

%v

// %v holds the schema definition for the %v entity.
type %v struct {
	ent.Schema
}

%v

// Fields of the %v.
func (%v) Fields() []ent.Field {
	return %v
}

// Edges of the %v.
func (%v) Edges() []ent.Edge {
	return %v
}
`

const edgeTemplateTo = `
		edge.To("%v", %v.Type).
			StorageKey(edge.Column("%v"))%v,
`

const edgeTemplateFrom = `
		edge.From("%v", %v.Type).
			Ref("%v").
			Unique(),
`
