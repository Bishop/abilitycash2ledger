package xml_schema

func (t *txCategory) Flatten() (string, string) {
	name := t.Name
	next := t.Category

	for next != nil {
		name = name + ":" + next.Name
		next = next.Category
	}

	return t.Classifier, name
}

func (t txCategories) Map() map[string]string {
	result := make(map[string]string)

	for _, c := range t {
		classifier, name := c.Flatten()
		result[classifier] = name
	}

	return result
}
