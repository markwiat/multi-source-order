package order

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

type testElement int

func (te testElement) Before(other Element) bool {
	return te < other.(testElement)
}

type testContainer struct {
	nums        []int
	containerId string
}

func (tc testContainer) ContainerId() any {
	return tc.containerId
}

func (tc testContainer) NextAfter(e Element) (Element, error) {
	base := int(e.(testElement))
	var next *int
	for i, v := range tc.nums {
		if v > base {
			if next == nil || v < *next {
				next = &tc.nums[i]
			}
		}
	}

	if next == nil {
		return nil, nil
	}

	return testElement(*next), nil
}

type resultItem struct {
	containerId string
	num         int
}

func toResultItem(se SortedItem) resultItem {
	return resultItem{
		containerId: se.ContainerId.(string),
		num:         int(se.Element.(testElement)),
	}
}

func toResultItems(origs []SortedItem) []resultItem {
	result := make([]resultItem, len(origs))
	for i, v := range origs {
		result[i] = toResultItem(v)
	}

	return result
}

var noConstraint Constraint = CreateConstraint()
var sizeConstraint Constraint = CreateConstraint(WithSizeLimit(5))
var highLimitConstraint Constraint = CreateConstraint(WithHighestElemnt(testElement(10)))
var container1 testContainer = testContainer{containerId: "c1", nums: []int{1, 2, 3, 7, 9}}
var container2 testContainer = testContainer{containerId: "c2", nums: []int{3, 10, 12}}
var container3 testContainer = testContainer{containerId: "c3", nums: []int{3, 12, 15}}
var containers = []Container{container1, container2, container3}
var fullExpected = []resultItem{
	{
		containerId: "c1",
		num:         1,
	},
	{
		containerId: "c1",
		num:         2,
	},
	{
		containerId: "c1",
		num:         3,
	},
	{
		containerId: "c2",
		num:         3,
	},
	{
		containerId: "c3",
		num:         3,
	},
	{
		containerId: "c1",
		num:         7,
	},
	{
		containerId: "c1",
		num:         9,
	},
	{
		containerId: "c2",
		num:         10,
	},
	{
		containerId: "c2",
		num:         12,
	},
	{
		containerId: "c3",
		num:         12,
	},
	{
		containerId: "c3",
		num:         15,
	},
}

func TestGetSortedElements(t *testing.T) {
	type args struct {
		initial    Element
		constraint Constraint
		sources    []Container
	}
	tests := []struct {
		name    string
		args    args
		want    []resultItem
		hasNext bool
		wantErr bool
	}{
		{
			name: "no constraint",
			args: args{
				initial:    testElement(0),
				constraint: noConstraint,
				sources:    containers,
			},
			want:    fullExpected,
			wantErr: false,
			hasNext: false,
		},
		{
			name: "not from begin",
			args: args{
				initial:    testElement(1),
				constraint: noConstraint,
				sources:    containers,
			},
			want:    fullExpected[1:],
			wantErr: false,
			hasNext: false,
		},
		{
			name: "size constraint",
			args: args{
				initial:    testElement(0),
				constraint: sizeConstraint,
				sources:    containers,
			},
			want:    fullExpected[0:5],
			wantErr: false,
			hasNext: true,
		},
		{
			name: "high limit constraint",
			args: args{
				initial:    testElement(0),
				constraint: highLimitConstraint,
				sources:    containers,
			},
			want:    fullExpected[0:8],
			wantErr: false,
			hasNext: false,
		},
		{
			name: "too high initial",
			args: args{
				initial:    testElement(20),
				constraint: highLimitConstraint,
				sources:    containers,
			},
			want:    nil,
			wantErr: true,
			hasNext: false,
		},
		{
			name: "no initial",
			args: args{
				initial:    nil,
				constraint: noConstraint,
				sources:    containers,
			},
			want:    nil,
			wantErr: true,
			hasNext: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, hasNext, err := GetSortedElements(tt.args.initial, tt.args.constraint, tt.args.sources)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSortedElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				assert.IsEqual(got, nil)
				return
			}
			gotMapped := toResultItems(got)
			assert.Equal(t, gotMapped, tt.want)
			assert.Equal(t, hasNext, tt.hasNext)
		})
	}
}
