package xml_schema

func (t *txCategory) Flatten() (string, string) {
	name := t.Name
	next := t.Category

	for next != nil {
		name = name + "\\" + next.Name
		next = next.Category
	}

	return t.Classifier, name
}

func (t txCategories) List() []string {
	result := make([]string, 0)

	for _, c := range t {
		_, name := c.Flatten()
		result = append(result, name)
	}

	return result
}
