package cmd

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	autoscaling "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	clock  clock.Clock
	Client kubernetes.Interface
}

func (h *Handler) now() time.Time {
	if h.clock == nil {
		h.clock = new(clock.RealClock)
	}

	return h.clock.Now()
}

func (h *Handler) OnAdd(obj interface{}) {
	h.Handle(obj)
}

func (h *Handler) OnUpdate(oldObj, newObj interface{}) {
	h.Handle(newObj)
}

func (h *Handler) OnDelete(obj interface{}) {}

func (h *Handler) Handle(obj interface{}) {
	hpa, ok := obj.(*autoscaling.HorizontalPodAutoscaler)
	if !ok {
		log.WithFields(log.Fields{
			"Type": fmt.Sprintf("%T", obj),
		}).Error("got unexpected object")
		return
	}

	v := HPAView(*hpa)
	logger := v.Logger()
	logger.Debug("received new HPA")

	minReplicas, err := h.Run(v)
	if err != nil {
		logger.Warn(err)
		return
	}

	if minReplicas == nil {
		logger.Debug("no changes")
		return
	}

	hpa.Spec.MinReplicas = minReplicas
	hpa.ObjectMeta.Annotations[AnnotationLastChange] = time.Now().Format(time.RFC3339)

	_, err = h.Client.
		AutoscalingV1().
		HorizontalPodAutoscalers(hpa.ObjectMeta.Namespace).
		Update(hpa)
	if err != nil {
		logger.Error(err)
	}
}

func (h *Handler) Run(v HPAView) (*int32, error) {
	logger := v.Logger()
	now := h.now()

	lrl := v.LowerReplicaLimit()
	if lrl == nil {
		logger.Debug("ignoring HPA, because of missing annotations")
		return nil, nil
	}

	oldMinReplicas := v.MinReplicas()
	newMinReplicas := max(v.DesiredReplicas()-1, *lrl)

	if newMinReplicas == oldMinReplicas {
		logger.Debug("(nothingtodohere)")
		return nil, nil
	}

	if newMinReplicas > oldMinReplicas {
		logger.Infof("scaling up .spec.minReplicas from %d to %d",
			v.MinReplicas(), newMinReplicas)
		return &newMinReplicas, nil
	}

	deadline := v.LastChange().Add(*v.DownscaleCooldown())
	if now.Before(deadline) {
		logger.Infof("next scale down is in %v (%v)", deadline.Sub(now), deadline)
		return nil, nil
	}

	logger.Infof("scaling down .spec.minReplicas from %d to %d",
		v.MinReplicas(), newMinReplicas)

	return &newMinReplicas, nil
}
