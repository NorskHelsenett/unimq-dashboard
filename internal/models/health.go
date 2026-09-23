package models

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

type HealthStatus struct {
	Database Status
	Dex      Status
	RabbitMQ Status
}

func NewHealthStatus() *HealthStatus {
	return &HealthStatus{
		Database: StatusUnhealthy,
		Dex:      StatusUnhealthy,
		RabbitMQ: StatusUnhealthy,
	}
}

func (hs *HealthStatus) IsHealthy() bool {
	return hs.Database == StatusHealthy && hs.Dex == StatusHealthy && hs.RabbitMQ == StatusHealthy
}
