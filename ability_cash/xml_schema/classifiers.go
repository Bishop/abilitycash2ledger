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

func (t txCategories) Map() map[string]string {
	result := make(map[string]string)

	for _, c := range t {
		classifier, name := c.Flatten()
		result[classifier] = name
	}

	return result
}

func (c *Classifier) Categories() <-chan string {
	result := make(chan string)

	go func() {
		for _, category := range c.Income {
			category.iterateCategories("", result)
		}
		for _, category := range c.Expense {
			category.iterateCategories("", result)
		}
		for _, category := range c.Single {
			category.iterateCategories("", result)
		}
		close(result)
	}()

	return result
}

func (c *txCategoryTI) iterateCategories(prefix string, ch chan<- string) {
	fullName := prefix + "\\" + c.Name

	ch <- fullName[1:]

	if c.Categories != nil {
		for _, category := range *c.Categories {
			category.iterateCategories(fullName, ch)
		}
	}
}
