package cmd

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	autoscaling "k8s.io/api/autoscaling/v1"
)

const (
	AnnotationLastChange        = "rebuy.com/kubernetes-hpaa.last-change"
	AnnotationDownscaleCooldown = "rebuy.com/kubernetes-hpaa.downscale-cooldown"
	AnnotationLowerReplicaLimit = "rebuy.com/kubernetes-hpaa.lower-replica-limit"
)

// HPAView provides a specialized view of an HPA, so the actual algorithm more
// understandable.
type HPAView autoscaling.HorizontalPodAutoscaler

func (v HPAView) Logger() *log.Entry {
	return log.WithFields(log.Fields{
		".metadata.name":          v.Name(),
		".metadata.namespace":     v.Namespace(),
		".spec.minReplicas":       v.MinReplicas(),
		".spec.maxReplicas":       v.Spec.MaxReplicas,
		".status.lastScaleTime":   v.Status.LastScaleTime,
		".status.desiredReplicas": v.Status.DesiredReplicas,
		"#lastChange":             v.LastChange(),
		"#lowerReplicaLimit":      v.LowerReplicaLimit(),
		"#downscaleCooldown":      v.DownscaleCooldown(),
	})
}

func (v HPAView) Name() string {
	return v.ObjectMeta.Name
}

func (v HPAView) Namespace() string {
	return v.ObjectMeta.Namespace
}

func (v HPAView) LastChange() time.Time {
	s, ok := v.ObjectMeta.Annotations[AnnotationLastChange]
	if !ok {
		return time.Unix(0, 0)
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		log.Errorf("failed to convert `%s` to time in last-change", s)
		return time.Unix(0, 0)
	}

	return t
}

func (v HPAView) LowerReplicaLimit() *int32 {
	s, ok := v.ObjectMeta.Annotations[AnnotationLowerReplicaLimit]
	if !ok {
		return nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		log.Errorf("failed to convert `%s` to int in lower-replica-limit", s)
		return nil
	}

	i32 := int32(i)
	return &i32
}

func (v HPAView) DownscaleCooldown() *time.Duration {
	s, ok := v.ObjectMeta.Annotations[AnnotationDownscaleCooldown]
	if !ok {
		return nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		log.Errorf("failed to convert `%s` to duration in downscale-cooldown", s)
		return nil
	}

	return &d
}

func (v HPAView) MinReplicas() int32 {
	if v.Spec.MinReplicas == nil {
		return 0
	}

	return *v.Spec.MinReplicas
}

func (v HPAView) DesiredReplicas() int32 {
	return v.Status.DesiredReplicas
}
