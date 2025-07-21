package user_info

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

type userInfoService interface {
	DetermineUserGitHubEmail(ctx context.Context, accessToken string) (string, error)
	GetUserAppRole(ctx context.Context, accessToken string) (string, error)
}
type aggregator struct {
	service userInfoService
}

func NewAggregator(service userInfoService) *aggregator {
	return &aggregator{service: service}
}

func (a *aggregator) AggregateUserInfo(ctx context.Context, accessToken string) (string, string, error) {
	log.C(ctx).Info("aggregating user info in user_info_aggregator")

	emailChan := make(chan utils.ChannelResult[string], 1)
	roleChan := make(chan utils.ChannelResult[string], 1)

	go func() {
		userEmail, err := a.service.DetermineUserGitHubEmail(ctx, accessToken)
		emailChan <- utils.ChannelResult[string]{
			Result: userEmail,
			Err:    err,
		}
	}()

	go func() {
		userRole, err := a.service.GetUserAppRole(ctx, accessToken)
		roleChan <- utils.ChannelResult[string]{
			Result: userRole,
			Err:    err,
		}
	}()

	email, role, err := utils.OrchestrateGoRoutines(ctx, emailChan, roleChan)
	if err != nil {
		log.C(ctx).Errorf("failed to get tokens, error %s when trying to orchestrate goroutines", err.Error())
		return "", "", err
	}

	return email, role, nil
}
