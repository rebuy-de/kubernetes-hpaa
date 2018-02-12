package cmd

import (
	"fmt"
	"strconv"
	"time"

	"k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

type Handler struct {
	clock clock.Clock

	DownscaleCooldown time.Duration
}

func NewHandler(cd time.Duration) *Handler {
	return &Handler{
		clock: new(clock.RealClock),

		DownscaleCooldown: cd,
	}
}

func (h *Handler) Run(hpa *v1.HorizontalPodAutoscaler) (*v1.HorizontalPodAutoscaler, error) {
	now := h.clock.Now()
	last := hpa.Status.LastScaleTime
	deadline := last.Add(h.DownscaleCooldown)

	currReplicas := hpa.Status.CurrentReplicas
	lowerLimit, err := getAnnotationInt(hpa, "lower-replica-limit")
	if err != nil {
		return nil, err
	}

	newMinReplicas := currReplicas

	if now.After(deadline) {
		newMinReplicas--
	}

	newMinReplicas = max(newMinReplicas, lowerLimit)
	hpa.Spec.MinReplicas = &newMinReplicas

	return hpa, nil
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func getAnnotationInt(hpa *v1.HorizontalPodAutoscaler, name string) (int32, error) {
	tpl := "rebuy.com/kubernetes-hpaa.%s"
	key := fmt.Sprintf(tpl, name)

	s, ok := hpa.ObjectMeta.Annotations[key]
	if !ok {
		return 0, fmt.Errorf("no annotation with key %s found", key)
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return int32(i), nil
}
