package sql_query_decorators

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBaseQuery_DetermineCorrectSqlQuery(t *testing.T) {
	t.Run("returning correct query", func(t *testing.T) {
		base := newBaseQuery(baseQueryString)
		returnedQuery := base.DetermineCorrectSqlQuery(context.Background())
		require.Equal(t, baseQueryString, returnedQuery)
	})
}
