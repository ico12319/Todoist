package middlewares

/*
func TestActionPermissionMiddleware_ServeHTTP(t *testing.T) {
	tests := []struct {
		testName          string
		contextUser       models.User
		mockConfig        func() *mocks.IService
		shouldCallNext    bool
		shouldEncodeError bool
		expectedHttpCode  int
		receivedFromNext  *userRoleKey
		responseBody      string
	}{
		{
			testName:    "ListId received from context is invalid and middleware encodes error and httpStatusBadRequest",
			contextUser: initUser(adminEmail, adminRole),
			mockConfig: func() *mocks.IService {
				mockService := &mocks.IService{}
				mockService.EXPECT().GetListRecord(listId).Return(nil, fmt.Errorf("list with id %s does not exist", listId)).Once()
				return mockService
			},
			shouldCallNext:    false,
			shouldEncodeError: true,
			expectedHttpCode:  http.StatusBadRequest,
			responseBody:      `{"error":"list with id list1 does not exist"}` + "\n",
		},

		{
			"ListId received from context is valid but the user from the context is not an admin or owner and the list is not shared with him" +
				"so the middleware encodes httpStatusUnauthorized",
			initUser(testEmail, readerRole),
			func() *mocks.IService {
				mockService := &mocks.IService{}
				mockService.EXPECT().GetListRecord(listId).Return(initList(notOwnerEmail, writerRole, models.User{"", ""}), nil).Once()
				return mockService
			},
			false,
			true,
			http.StatusUnauthorized,
			nil,
			`{"error":"unauthorized user"}` + "\n",
		},

		{
			"ListId received from context is valid and the user from the context is an admin" +
				"so the middleware encodes httpStatusOK and calls next",
			initUser(adminEmail, adminRole),
			func() *mocks.IService {
				mockService := &mocks.IService{}
				mockService.EXPECT().GetListRecord(listId).Return(initList(notOwnerEmail, writerRole, models.User{"", ""}), nil).Once()
				return mockService
			},
			true,
			false,
			http.StatusOK,
			&userRoleKey{
				role:    adminRole,
				isOwner: false,
			},
			"",
		},

		{
			"ListId received from context is valid and the user from the context is the owner of the list" +
				"so the middleware encodes httpStatusOK and calls next",
			initUser(ownerEmail, writerRole),
			func() *mocks.IService {
				mockService := &mocks.IService{}
				mockService.EXPECT().GetListRecord(listId).Return(initList(ownerEmail, writerRole, models.User{"", ""}), nil).Once()
				return mockService
			},
			true,
			false,
			http.StatusOK,
			&userRoleKey{
				role:    "writer",
				isOwner: true,
			},
			"",
		},

		{
			"ListId received from context is valid and the list is shared with the user from the context" +
				"so the middleware encodes httpStatusOK and calls next",
			initUser(sharedWithMeEmail, readerRole),
			func() *mocks.IService {
				mockService := &mocks.IService{}
				mockService.EXPECT().GetListRecord(listId).Return(initList(ownerEmail, writerRole, initUser(sharedWithMeEmail, readerRole)), nil).Once()
				return mockService
			},
			true,
			false,
			http.StatusOK,
			&userRoleKey{
				role:    readerRole,
				isOwner: false,
			},
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockService := new(mocks.IService)
			if readyMock := test.mockConfig(); readyMock != nil {
				mockService = readyMock
			}
			defer mockService.AssertExpectations(t)

			isNextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				isNextCalled = true
				userInfo, ok := r.Context().Value(UserRoleKey).(userRoleKey)
				require.True(t, ok)
				require.Equal(t, *test.receivedFromNext, userInfo)
			})

			rr := httptest.NewRecorder()
			ctx := context.Background()
			ctx = context.WithValue(ctx, ListId, listId)
			ctx = context.WithValue(ctx, UserKey, test.contextUser)
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

			middleware := newActionPermissionMiddleware(next, mockService)
			middleware.ServeHTTP(rr, req)

			if test.shouldEncodeError {
				require.Equal(t, test.responseBody, rr.Body.String())
			}
			require.Equal(t, test.shouldCallNext, isNextCalled)
			require.Equal(t, test.expectedHttpCode, rr.Code)
		})
	}
}
*/
