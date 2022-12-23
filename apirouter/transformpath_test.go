package apirouter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransformPathParamsWithColon(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedPath string
	}{
		{
			name:         "only /",
			path:         "/",
			expectedPath: "/",
		},
		{
			name:         "without params",
			path:         "/foo",
			expectedPath: "/foo",
		},
		{
			name:         "without params ending with /",
			path:         "/foo/",
			expectedPath: "/foo/",
		},
		{
			name:         "with params",
			path:         "/foo/:par1",
			expectedPath: "/foo/{par1}",
		},
		{
			name:         "with params ending with /",
			path:         "/foo/:par1/",
			expectedPath: "/foo/{par1}/",
		},
		{
			name:         "with multiple params",
			path:         "/:par1/:par2/:par3",
			expectedPath: "/{par1}/{par2}/{par3}",
		},
		{
			name:         "with multiple params ending with /",
			path:         "/:par1/:par2/:par3/",
			expectedPath: "/{par1}/{par2}/{par3}/",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			actual := TransformPathParamsWithColon(test.path)

			require.Equal(t, test.expectedPath, actual)
		})
	}
}
