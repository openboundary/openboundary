package schema

// UsecaseSchema validates usecase component specs.
type UsecaseSchema struct{}

// Kind returns the component kind.
func (s *UsecaseSchema) Kind() Kind {
	return KindUsecase
}

// Validate validates the usecase spec.
func (s *UsecaseSchema) Validate(spec map[string]interface{}) error {
	// TODO: Implement validation
	// Required fields: input, output, steps
	// Optional fields: dependencies
	return nil
}
