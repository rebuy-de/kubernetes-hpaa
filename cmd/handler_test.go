package cmd

import (
	"fmt"
	"testing"
	"time"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

func testCreateHPA() *v1.HorizontalPodAutoscaler {
	hpa := new(v1.HorizontalPodAutoscaler)
	hpa.Spec.MaxReplicas = 40
	min := int32(6)
	hpa.Spec.MinReplicas = &min
	t := v1meta.Date(
		2018, 02, 12,
		20, 56, 0, 0,
		time.FixedZone("UTC", 0))
	hpa.Status.LastScaleTime = &t

	return hpa
}

func TestHandleValidCases(t *testing.T) {
	cases := []struct {
		name            string
		offset          time.Duration
		lowerLimit      int32
		currentReplicas int32
		expect          int32
	}{
		{
			name:            "WithinColldown",
			offset:          3 * time.Minute,
			lowerLimit:      6,
			currentReplicas: 11,
			expect:          11,
		},
		{
			name:            "AfterCooldown",
			offset:          10 * time.Minute,
			lowerLimit:      6,
			currentReplicas: 11,
			expect:          10,
		},
		{
			name:            "NoActionLowerLimit",
			offset:          10 * time.Minute,
			lowerLimit:      11,
			currentReplicas: 11,
			expect:          11,
		},
		{
			name:            "EnforceLowerLimit",
			offset:          10 * time.Minute,
			lowerLimit:      12,
			currentReplicas: 11,
			expect:          12,
		},
		{
			name:            "ScaleUpWithinCooldown",
			offset:          2 * time.Minute,
			lowerLimit:      12,
			currentReplicas: 20,
			expect:          20,
		},
		{
			name:            "ScaleUpAfterCooldown",
			offset:          10 * time.Minute,
			lowerLimit:      12,
			currentReplicas: 20,
			expect:          19,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hpa := testCreateHPA()
			hpa.Status.CurrentReplicas = tc.currentReplicas
			hpa.ObjectMeta.Annotations = map[string]string{
				"rebuy.com/kubernetes-hpaa.lower-replica-limit": fmt.Sprint(tc.lowerLimit),
			}

			handler := NewHandler(5 * time.Minute)
			handler.clock = clock.NewFakeClock(hpa.Status.LastScaleTime.Add(tc.offset))

			hpa, err := handler.Run(hpa)
			if err != nil {
				t.Fatal(err)
			}

			have := *hpa.Spec.MinReplicas
			want := tc.expect
			if want != have {
				t.Fatalf("Wrong `spec.minReplicas`. Want: %d. Have: %d", want, have)
			}
		})
	}
}

func TestHandleMissingAnnotation(t *testing.T) {
	hpa := testCreateHPA()
	handler := NewHandler(5 * time.Minute)
	_, err := handler.Run(hpa)
	if err == nil {
		t.Fatal("Expected an error")
	}

	want := "no annotation with key rebuy.com/kubernetes-hpaa.lower-replica-limit found"
	have := err.Error()
	if want != have {
		t.Fatalf("Wrong error. Want `%s`. Have: `%s`.", want, have)
	}
}

func TestHandleInvalidAnnotation(t *testing.T) {
	hpa := testCreateHPA()
	hpa.ObjectMeta.Annotations = map[string]string{
		"rebuy.com/kubernetes-hpaa.lower-replica-limit": "foo",
	}

	handler := NewHandler(5 * time.Minute)
	_, err := handler.Run(hpa)
	if err == nil {
		t.Fatal("Expected an error")
	}

	want := `strconv.Atoi: parsing "foo": invalid syntax`
	have := err.Error()
	if want != have {
		t.Fatalf("Wrong error. Want `%s`. Have: `%s`.", want, have)
	}
}
