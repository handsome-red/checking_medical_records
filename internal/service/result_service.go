package service

// import (
// 	"med_book/internal/model"
// 	"med_book/internal/repository"

// 	"github.com/google/uuid"
// )

// type ResultService struct {
// 	userRepo    repository.UserRepositoryInterface
// 	sessionRepo repository.SessionRepositoryInterface
// }

// func NewResultService(
// 	userRepo repository.UserRepositoryInterface,
// 	sessionRepo repository.SessionRepositoryInterface,
// ) *ResultService {
// 	return &ResultService{
// 		userRepo:    userRepo,
// 		sessionRepo: sessionRepo,
// 	}
// }

// func (r *ResultService) GetByUserID(id uuid.UUID) ([]*model.Session, error) {
// 	result, err := r.sessionRepo.FindByID(id)
// }
