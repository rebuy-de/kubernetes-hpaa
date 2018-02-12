package cmd

import "k8s.io/api/autoscaling/v1"

func Handle(hpa *v1.HorizontalPodAutoscaler) (*v1.HorizontalPodAutoscaler, error) {
	return hpa, nil
}
