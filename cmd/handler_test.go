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
	hpa.ObjectMeta.Annotations = map[string]string{
		"rebuy.com/kubernetes-hpaa.last-change": "2018-02-12T20:56:00+00:00",
	}

	return hpa
}

func TestHandleValidCases(t *testing.T) {
	now := v1meta.Date(
		2018, 02, 12,
		20, 56, 0, 0,
		time.FixedZone("UTC", 0))
	cases := []struct {
		name            string
		offset          time.Duration
		lowerLimit      int32
		currentReplicas int32
		expect          int32
	}{
		{
			name:            "WithinCooldown",
			offset:          3 * time.Minute,
			lowerLimit:      6,
			currentReplicas: 11,
			expect:          10,
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
			expect:          19,
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
			hpa.ObjectMeta.Annotations["rebuy.com/kubernetes-hpaa.lower-replica-limit"] = fmt.Sprint(tc.lowerLimit)
			hpa.ObjectMeta.Annotations["rebuy.com/kubernetes-hpaa.downscale-cooldown"] = "5m"
			v := HPAView(*hpa)

			handler := &Handler{
				clock: clock.NewFakeClock(now.Add(tc.offset)),
			}

			minReplicas, err := handler.Run(v)
			if err != nil {
				t.Fatal(err)
			}

			have := *minReplicas
			want := tc.expect
			if want != have {
				t.Fatalf("Wrong `spec.minReplicas`. Want: %d. Have: %d", want, have)
			}
		})
	}
}

func TestHandleNilDereference(t *testing.T) {
	hpa := new(v1.HorizontalPodAutoscaler)
	v := HPAView(*hpa)
	handler := new(Handler)
	_, _ = handler.Run(v)
}

func TestHandleMissingCooldownAnnotation(t *testing.T) {
	hpa := testCreateHPA()
	v := HPAView(*hpa)
	handler := new(Handler)
	minReplicas, err := handler.Run(v)
	if err != nil {
		t.Fatal(err)
	}
	if minReplicas != nil {
		t.Fatal("Unexpected result")
	}
}

func TestHandleMissingLimitAnnotation(t *testing.T) {
	hpa := testCreateHPA()
	hpa.ObjectMeta.Annotations = map[string]string{
		"rebuy.com/kubernetes-hpaa.downscale-cooldown": "5m",
	}
	v := HPAView(*hpa)

	handler := new(Handler)
	minReplicas, err := handler.Run(v)
	if err != nil {
		t.Fatal(err)
	}
	if minReplicas != nil {
		t.Fatal("Unexpected result")
	}
}

func TestHandleInvalidAnnotation(t *testing.T) {
	hpa := testCreateHPA()
	hpa.ObjectMeta.Annotations = map[string]string{
		"rebuy.com/kubernetes-hpaa.lower-replica-limit": "foo",
		"rebuy.com/kubernetes-hpaa.downscale-cooldown":  "5m",
	}
	v := HPAView(*hpa)

	handler := new(Handler)
	minReplicas, err := handler.Run(v)
	if err != nil {
		t.Fatal(err)
	}
	if minReplicas != nil {
		t.Fatal("Unexpected result")
	}
}
