package graphitepickletest

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRule_String(t *testing.T) {
	tests := []struct {
		name string
		rule *Rule
		s    string
	}{
		{
			name: "optional",
			rule: &Rule{
				Path: "a.b.c",
				Exprs: []*Expr{
					{Op: LessThan, Value: 3.0},
				},
			},
			s: "~a.b.c:<3",
		},
		{
			name: "required",
			rule: &Rule{
				Required: true,
				Path:     "a.b.c",
				Exprs: []*Expr{
					{Op: LessThan, Value: 3.0},
				},
			},
			s: "a.b.c:<3",
		},
		{
			name: "operators",
			rule: &Rule{
				Required: true,
				Path:     "a.b.c",
				Exprs: []*Expr{
					{Op: LessThan, Value: 3.0},
					{Op: LessEqual, Value: 2.15},
					{Op: GreaterThan, Value: 0.0},
					{Op: GreaterEqual, Value: -3.0},
				},
			},
			s: "a.b.c:<3,<=2.15,>0,>=-3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.rule.String()
			if s != tt.s {
				t.Errorf("%+v = %q; want %q", tt.rule, s, tt.s)
			}
		})
	}
}

func TestMatchEmpty(t *testing.T) {
	r := Match([]*Rule{}, []*Metric{})
	if len(r) != 0 {
		t.Errorf("should match all rules if empty rules and metrics; but found invalid data: %v", r)
	}
}

func TestMatch_path(t *testing.T) {
	tests := []struct {
		name    string
		rules   []*Rule
		metrics []*Metric
		want    []*InvalidData
	}{
		{
			name: "simple path/success",
			rules: []*Rule{
				{
					Required: true,
					Path:     "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.metric1.value", Value: 3.0},
			},
			want: nil,
		},
		{
			name: "simple path/failure",
			rules: []*Rule{
				{
					Required: true,
					Path:     "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.metric1.value1", Value: 3.0},
				{Path: "custom.metric2.value", Value: 3.0},
				{Path: "custom.metric1", Value: 3.0},
			},
			want: []*InvalidData{
				{
					Rule: &Rule{
						Required: true,
						Path:     "custom.metric1.value",
						Exprs: []*Expr{
							{Op: LessEqual, Value: 3.0},
						},
					},
				},
				{
					Metric: &Metric{Path: "custom.metric1.value1", Value: 3.0},
				},
				{
					Metric: &Metric{Path: "custom.metric2.value", Value: 3.0},
				},
				{
					Metric: &Metric{Path: "custom.metric1", Value: 3.0},
				},
			},
		},
		{
			name: "optional path/matched",
			rules: []*Rule{
				{
					Path: "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: []*Metric{
				{Path: "custom.metric1.value", Value: 3.0},
			},
			want: []*InvalidData{},
		},
		{
			name: "optional path/passed",
			rules: []*Rule{
				{
					Path: "custom.metric1.value",
					Exprs: []*Expr{
						{Op: LessEqual, Value: 3.0},
					},
				},
			},
			metrics: nil,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Match(tt.rules, tt.metrics)
			checkResults(t, "only result", a, tt.want)
			checkResults(t, "only expected", tt.want, a)
		})
	}
}

func checkResults(t *testing.T, name string, a1, a2 []*InvalidData) {
	t.Helper()
	if diffs := diffInvalidData(a1, a2); len(diffs) > 0 {
		t.Run(name, func(t *testing.T) {
			for _, d := range diffs {
				t.Errorf("unexpected %v", d)
			}
		})
	}
}

// diffInvalidData returns a slice in a part of the a1 that is *NOT* contained in a2.
func diffInvalidData(a1, a2 []*InvalidData) []*InvalidData {
	var x []*InvalidData
	for _, v1 := range a1 {
		var found bool
		for _, v2 := range a2 {
			if reflect.DeepEqual(v1, v2) {
				found = true
			}
		}
		if !found {
			x = append(x, v1)
		}
	}
	return x
}

func (p *InvalidData) String() string {
	s := ""
	if p.Rule != nil {
		s += fmt.Sprintf("rule={%v}", p.Rule)
	}
	if p.Metric != nil {
		if s != "" {
			s += " "
		}
		s += fmt.Sprintf("value={%s=%g}", p.Metric.Path, p.Metric.Value)
	}
	return s
}