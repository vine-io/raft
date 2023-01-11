package pb

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/vine-io/apimachinery/runtime"
)

func TestEpoch_Marshal_Unmarshal(t *testing.T) {
	m1 := make(map[string][]runtime.Object)
	m2 := make(map[string][]runtime.Object)

	product := &Product{}
	m1[FromProduct(product).TableName()] = []runtime.Object{
		&Product{
			Id:      "1",
			Date:    1,
			Company: "c1",
		},
		&Product{
			Id:      "2",
			Date:    2,
			Company: "c2",
		},
	}

	data, err := json.Marshal(m1)
	if err != nil {
		t.Fatalf("marshal %v", err)
	}

	err = json.Unmarshal(data, &m2)
	if err != nil {
		t.Fatalf("unmarshal %v", err)
	}

	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("no matched: \t%v\n\t%v", m1, m2)
	}
}
