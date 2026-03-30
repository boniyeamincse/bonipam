package service

import (
	"boni-pam/internal/domain"
	"strings"
	"time"
)

type ContextResolverService struct{}

func NewContextResolverService() *ContextResolverService {
	return &ContextResolverService{}
}

func (s *ContextResolverService) Resolve(req domain.PolicyEvaluationRequest) map[string]interface{} {
	resolved := make(map[string]interface{}, len(req.Attributes)+5)
	for k, v := range req.Attributes {
		resolved[k] = v
	}

	if sourceIP := strings.TrimSpace(req.SourceIP); sourceIP != "" {
		resolved["source_ip"] = sourceIP
	}
	if deviceID := strings.TrimSpace(req.DeviceID); deviceID != "" {
		resolved["device_id"] = deviceID
	}
	if trust := strings.TrimSpace(req.DeviceTrust); trust != "" {
		resolved["device_trust"] = strings.ToLower(trust)
	}
	if req.RiskScore != nil {
		resolved["risk_score"] = *req.RiskScore
	}

	now := time.Now().UTC()
	if requestTime := strings.TrimSpace(req.RequestTime); requestTime != "" {
		if parsed, err := time.Parse(time.RFC3339, requestTime); err == nil {
			now = parsed.UTC()
		}
	}

	resolved["time"] = now.Format("15:04")
	resolved["timestamp"] = now.Format(time.RFC3339)
	resolved["day_of_week"] = strings.ToLower(now.Weekday().String())

	return resolved
}
