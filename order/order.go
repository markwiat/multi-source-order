package order

import (
	"errors"
	"sort"
)

func GetSortedElements(initial Element, constraint Constraint, noNextChecker NoNextErrorChecker, sources []Container) ([]SortedItem, error) {
	if err := sanitize(initial, constraint); err != nil {
		return nil, err
	}

	result := make([]SortedItem, 0)
	holder, err := initStateHolder(initial, sources, noNextChecker)
	if err != nil {
		return nil, err
	}

	for !constraint.isResultFull(len(result)) {
		last := pickOne(holder)
		if last == nil || !constraint.accept(last.Element) {
			break
		}
		if err := generateOne(holder, last.Element, sources, noNextChecker, last.index); err != nil {
			return nil, err
		}
		result = append(result, last.SortedItem)
	}

	return result, nil
}

func sanitize(initial Element, constraint Constraint) error {
	if initial == nil {
		return errors.New("initial element cannot be nil")
	}

	if !constraint.accept(initial) {
		return errors.New("highest element from constraint must not be lower than initial one")
	}

	return nil
}

func pickOne(holder stateHolder) *calcElement {
	valids, invalids := splitByValidness(holder)
	if len(valids) == 0 {
		return nil
	}
	lowest := pickLowest(valids)
	deleteInvalid(holder, invalids)

	return &lowest
}

func pickLowest(valids []calcElement) calcElement {
	lowest := valids[0]
	for i := 1; i < len(valids); i++ {
		e := valids[i]
		if e.Element.Before(lowest.Element) {
			lowest = e
		}
	}

	return lowest
}

func initStateHolder(initial Element, sources []Container, noNextChecker NoNextErrorChecker) (stateHolder, error) {
	holder := make(map[int]SortedItem, len(sources))
	for i, _ := range sources {
		if err := generateOne(holder, initial, sources, noNextChecker, i); err != nil {
			return holder, err
		}
	}

	return holder, nil
}

func generateOne(holder stateHolder, element Element, sources []Container, noNextChecker NoNextErrorChecker, index int) error {
	next, err := sources[index].NextAfter(element)
	if err != nil {
		if noNextChecker(err) {
			next = nil
		} else {
			return err
		}
	}
	se := SortedItem{
		ContainerId: sources[index].Id(),
		Element:     next,
	}

	holder[index] = se

	return nil
}

func splitByValidness(holder stateHolder) (valids, invalids []calcElement) {
	indexs := sortedIndexes(holder)
	for _, index := range indexs {
		v := holder[index]
		ce := calcElement{SortedItem: v, index: index}
		if v.Element == nil {
			invalids = append(invalids, ce)
		} else {
			valids = append(valids, ce)
		}
	}

	return
}

func sortedIndexes(holder stateHolder) []int {
	result := make([]int, 0, len(holder))
	for k, _ := range holder {
		result = append(result, k)
	}

	sort.Ints(result)

	return result
}

func deleteInvalid(holder stateHolder, invalids []calcElement) {
	for _, ce := range invalids {
		delete(holder, ce.index)
	}
}

type stateHolder map[int]SortedItem

type calcElement struct {
	SortedItem
	index int
}
