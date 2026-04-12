package pick_test

import (
	"testing"

	"charm.land/huh/v2"
	"github.com/gechr/primer/pick"
	"github.com/stretchr/testify/require"
)

func TestMultiSelectEmptyItemsFastPath(t *testing.T) {
	t.Parallel()

	var theme huh.Theme

	values, err := pick.MultiSelect[int]("title", nil, theme, 3, false)

	require.NoError(t, err)
	require.Nil(t, values)
}

func TestErrCanceledIsSet(t *testing.T) {
	t.Parallel()

	require.EqualError(t, pick.ErrCanceled, "canceled")
}
