package kit

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func Test_InOrderRangeMap(t *testing.T) {

	fmt.Println(time.Now())
	cases := []struct {
		name string
		m    map[int]string
		sort Sort
		want []int // 期望的键的顺序
	}{
		{
			name: "空map",
			m:    map[int]string{},
			sort: SortAsc,
			want: []int{},
		},
		{

			name: "已排序map",
			m: map[int]string{
				1: "一",
				2: "二",
				3: "三",
				4: "四",
				5: "五",
			},
			sort: SortAsc,
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name: "逆序map",
			m: map[int]string{
				5: "五",
				4: "四",
				3: "三",
				2: "二",
				1: "一",
			},
			sort: SortAsc,
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name: "随机顺序map",
			m: map[int]string{
				3: "三",
				1: "一",
				4: "四",
				2: "二",
				5: "五",
			},
			sort: SortDesc,
			want: []int{5, 4, 3, 2, 1},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			keys := []int{}
			InOrderRangeMap(c.m, func(v string, k int) {
				keys = append(keys, k)
			})
			if !reflect.DeepEqual(keys, c.want) {
				t.Errorf("InOrderRangeMap() 得到的键顺序 = %v, 期望 %v", keys, c.want)
			}
		})
	}
}
