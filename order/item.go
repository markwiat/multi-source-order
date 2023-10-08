package order

type Element interface {
	Before(other Element) bool
}

type Container interface {
	NextAfter(e Element) (Element, error)
	Id() any
}

type SortedItem struct {
	ContainerId any
	Element     Element
}

type NoNextErrorChecker func(error) bool

const noSizeLimit = -1

type Constraint struct {
	sizeLimit int
	highest   Element
}

type updator func(*Constraint)

func WithSizeLimit(sizeLimit uint) updator {
	return func(c *Constraint) {
		c.sizeLimit = int(sizeLimit)
	}
}

func WithHighestElemnt(highest Element) updator {
	return func(c *Constraint) {
		c.highest = highest
	}
}

func CreateConstraint(updators ...updator) Constraint {
	constraint := Constraint{
		sizeLimit: noSizeLimit,
	}
	for _, u := range updators {
		u(&constraint)
	}

	return constraint
}

func (c Constraint) hasSizeLimit() bool {
	return c.sizeLimit != noSizeLimit
}

func (c Constraint) isResultFull(resultSize int) bool {
	return c.hasSizeLimit() && resultSize >= c.sizeLimit
}

func (c Constraint) accept(e Element) bool {
	return c.highest == nil || !c.highest.Before(e)
}
