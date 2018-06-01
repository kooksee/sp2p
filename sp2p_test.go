package sp2p

import (
	"testing"
	"sort"
	"fmt"
)

func TestSort(t *testing.T) {
	a := []int{1, 4, 2, 6, 9, 3, 2, 4, 6, 89, 3, 2, 3, 5}
	sort.Ints(a)

	fmt.Println(a)
}
