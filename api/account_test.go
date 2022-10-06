package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "db.sqlc.dev/app/db/mock"
	db "db.sqlc.dev/app/db/sqlc"
	"db.sqlc.dev/app/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

// check if response body(stored in recorder.Body) match the account
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	// call ioutil.ReadAll() to read all data from the response body and store it in a data variable.
	data, err := ioutil.ReadAll(body)
	// require no errors to be returned.
	require.NoError(t, err)

	var gotAccount db.Account
	// call json.Unmarshal to unmarshal the data to the gotAccount object
	err = json.Unmarshal(data, &gotAccount)
	// require no return err and gotAccount to be same as account
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

// TestGetAccount cover 100% code of getAccount method in ./api/account.go file
func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	// use an anonymous class to store the test data
	testCases := []struct {
		name      string
		accountID int64
		// buildStubs: build the stub that suits the purpose of each test case
		buildStubs func(store *mockdb.MockStore)
		// checkResponse: check the output of the API
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		// TEST CASE 1.: happy case
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				// build stub for GetAccount method by calling store.EXPECT().GetAccount()
				// GetAccount(query method, called by getAccount API method) requires two arguments: ctx, id
				// -> use gomock.Any() for 1st, gomock.Eq(account.ID) for 2nd
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					// specify we expect GetAccount method to be called exactly 1 time
					Times(1).
					// Return() function should match with the return values of the GetAccount function as defined in the Querier interface
					Return(account, nil)
			},
			// Check HTTP status code: require reponse(stroed in recorder.Code) to be http.StatusOK
			// Check response body: response body(stored recorder.Body) is a bytes.Buffer pointer
			// Check if account info stored in recorder.Body matches with generated account
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		// TEST CASE 2.: NOT FOUND case
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					// ask the stub to return an empty Account{} object together with a sql.ErrNoRows error
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		// TEST CASE 3.: InternalError case
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		// TEST CASE 4.: Invalid ID case
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					// GetAccount method should be executed when ID is invalid
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		// call t.Run to run each case as a separate sub-test of this unit test
		t.Run(tc.name, func(t *testing.T) {
			// initialize a controller using gomock.NewController
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// create a new mock store using mockdb.NewMockStore() generated function
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// Test HTTP server and send GetAccount request using built stub
			server := newTestServer(t, store)
			// use the recording feature of the httptest package to record the response of the API request
			recorder := httptest.NewRecorder()
			// declare the url path of the API to call
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			// test http GET method to declared url using http.NewRequest, require no return error
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// send API request through the server router and record its response in the recorder
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
