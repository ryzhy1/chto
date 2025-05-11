package converter

import (
	"boton-back/internal/domain/dto"
	"boton-back/internal/domain/models"
)

func ToCreateTenderDTO(tender *models.Tender) dto.TenderDTO {
	return dto.TenderDTO{
		Name:            tender.Name,
		Description:     tender.Description,
		ServiceType:     tender.ServiceType,
		Status:          tender.Status,
		OrganizationID:  tender.OrganizationID,
		CreatorUsername: tender.CreatorUsername,
	}
}
