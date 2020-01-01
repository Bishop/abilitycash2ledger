package xml_schema

func (t *txCategory) Flatten() (string, string) {
	name := ""
	next := t.Category

	for next != nil {
		name = name + ":" + next.Name
		next = next.Category
	}

	return t.Classifier, name[1:]
}

func (t txCategories) Map() map[string]string {
	result := make(map[string]string)

	for _, c := range t {
		classifier, name := c.Flatten()
		result[classifier] = name
	}

	return result
}
