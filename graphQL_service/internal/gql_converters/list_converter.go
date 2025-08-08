package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
)

type listConverter struct{}

func NewListConverter() *listConverter {
	return &listConverter{}
}

func (*listConverter) ToGQL(list *models.List) *gql.List {
	return &gql.List{
		ID:          list.Id,
		Name:        list.Name,
		Description: list.Description,
		CreatedAt:   list.CreatedAt,
		LastUpdated: list.LastUpdated,
	}
}

func (l *listConverter) ToListPageGQL(listPage *models.ListPage) *gql.ListPage {
	if listPage == nil || len(listPage.Data) == 0 {
		return &gql.ListPage{
			Data:       make([]*gql.List, 0),
			PageInfo:   nil,
			TotalCount: 0,
		}
	}

	lists := listPage.Data
	gqlLists := make([]*gql.List, len(lists))

	for index, list := range lists {
		gqlList := l.ToGQL(list)
		gqlLists[index] = gqlList
	}

	return &gql.ListPage{
		Data: gqlLists,
		PageInfo: &gql.PageInfo{
			HasPrevPage: listPage.PageInfo.HasPrevPage,
			HasNextPage: listPage.PageInfo.HasNextPage,
			StartCursor: listPage.PageInfo.StartCursor,
			EndCursor:   listPage.PageInfo.EndCursor,
		},
		TotalCount: int32(listPage.TotalCount),
	}
}

func (*listConverter) ToModel(list *gql.List) *models.List {
	return &models.List{
		Id:          list.ID,
		Name:        list.Name,
		CreatedAt:   list.CreatedAt,
		LastUpdated: list.LastUpdated,
		Owner:       list.Owner.ID,
	}
}

func (*listConverter) CreateListInputGQLToHandlerModel(input gql.CreateListInput) *handler_models.CreateList {
	return &handler_models.CreateList{
		Name:        input.Name,
		Description: input.Description,
	}
}

func (*listConverter) UpdateListInputGQLToHandlerModel(input gql.UpdateListInput) *handler_models.UpdateList {
	return &handler_models.UpdateList{
		Name:        input.Name,
		Description: input.Description,
	}
}

func (*listConverter) FromGQLModelToDeleteListPayload(list *gql.List, success bool) *gql.DeleteListPayload {
	return &gql.DeleteListPayload{
		Success:     success,
		ID:          list.ID,
		Name:        &list.Name,
		Description: &list.Description,
		CreatedAt:   &list.CreatedAt,
		LastUpdated: &list.LastUpdated,
	}
}

func (l *listConverter) ManyFromGQLModelToDeleteListPayload(lists []*gql.List, success bool) []*gql.DeleteListPayload {
	if lists == nil {
		return nil
	}

	payloads := make([]*gql.DeleteListPayload, 0, len(lists))
	for _, list := range lists {
		payload := l.FromGQLModelToDeleteListPayload(list, success)
		payloads = append(payloads, payload)
	}

	return payloads
}
