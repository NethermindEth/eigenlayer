package hardwarechecker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleQueryNodeExporter() {
	address := "http://demo.robustperception.io:9090"
	cpuQuery := "count(count(node_cpu_seconds_total) by (cpu))"

	val, err := QueryNodeExporter(address, cpuQuery)
	if err != nil {
		panic(err)
	}
	fmt.Println("Number of Cores:", val)
	// Output: Number of Cores: 1
}

func TestHardwareMetrics_Meets(t *testing.T) {
	type fields struct {
		CPU       float64
		RAM       float64
		DiskSpace float64
	}
	type args struct {
		hm HardwareMetrics
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "metrics meet the requirements",
			fields: fields{
				CPU:       4,
				RAM:       8,
				DiskSpace: 100,
			},
			args: args{
				hm: HardwareMetrics{
					CPU:       2,
					RAM:       4,
					DiskSpace: 50,
				},
			},
			want: false,
		},
		{
			name: "metrics do not meet the requirements",
			fields: fields{
				CPU:       4,
				RAM:       8,
				DiskSpace: 100,
			},
			args: args{
				hm: HardwareMetrics{
					CPU:       8,
					RAM:       16,
					DiskSpace: 200,
				},
			},
			want: true,
		},
		{
			name: "metrics wtih same requirements",
			fields: fields{
				CPU:       4,
				RAM:       8,
				DiskSpace: 100,
			},
			args: args{
				hm: HardwareMetrics{
					CPU:       4,
					RAM:       8,
					DiskSpace: 100,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HardwareMetrics{
				CPU:       tt.fields.CPU,
				RAM:       tt.fields.RAM,
				DiskSpace: tt.fields.DiskSpace,
			}
			got := h.Meets(tt.args.hm)
			// t.Errorf("HardwareMetrics.Meets() = %v, want %v", got, tt.want)
			assert.Equal(t, tt.want, got)

		})
	}
}
