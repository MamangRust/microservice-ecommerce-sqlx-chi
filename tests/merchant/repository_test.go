package merchant_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/repository"
	user_repo "github.com/MamangRust/microservice-ecommerce-grpc-user/repository"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type MerchantRepositoryTestSuite struct {
	tests.BaseTestSuite
	repo       repository.MerchantCommandRepository
	queryRepo  repository.MerchantQueryRepository
	userRepo   user_repo.UserCommandRepository
	userID     int
	merchantID int
}

func (s *MerchantRepositoryTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	s.SetupUserService()
	s.repo = repository.NewMerchantCommandRepository(s.SQLxDB())
	s.queryRepo = repository.NewMerchantQueryRepository(s.SQLxDB())
	s.userRepo = user_repo.NewUserCommandRepository(s.SQLxDB())

	// 1. Seed dependencies
	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
}

func (s *MerchantRepositoryTestSuite) Test1_CreateMerchant() {
	req := &requests.CreateMerchantRequest{
		Name:         "Test Merchant",
		UserID:       s.userID,
		Description:  "Detailed description of the merchant.",
		Address:      "Merchant Street No. 1",
		ContactEmail: "merchant@example.com",
		ContactPhone: "08123456789",
		Status:       "active",
	}

	merchant, err := s.repo.Create(context.Background(), req)
	s.NoError(err)
	s.NotNil(merchant)
	s.Equal(req.Name, merchant.Name)
	s.Equal(int32(s.userID), merchant.UserID)
	s.merchantID = int(merchant.MerchantID)
}

func (s *MerchantRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	found, err := s.queryRepo.FindByID(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.merchantID, int(found.MerchantID))
}

func (s *MerchantRepositoryTestSuite) Test3_UpdateMerchant() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	updateReq := &requests.UpdateMerchantRequest{
		MerchantID: &s.merchantID,
		Name:       "Updated Merchant",
		UserID:     s.userID,
		Status:     "active",
	}

	updated, err := s.repo.Update(ctx, updateReq)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(updateReq.Name, updated.Name)
	s.Equal(updateReq.Status, updated.Status)
}

func (s *MerchantRepositoryTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	_, err := s.repo.Trash(ctx, s.merchantID)
	s.NoError(err)

	_, err = s.repo.Restore(ctx, s.merchantID)
	s.NoError(err)
}

func (s *MerchantRepositoryTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	success, err := s.repo.DeletePermanent(ctx, s.merchantID)
	s.NoError(err)
	s.True(success)
}

func TestMerchantRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantRepositoryTestSuite))
}
