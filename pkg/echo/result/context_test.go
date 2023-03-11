package result

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyContext(t *testing.T) {
	result := FromContext(context.Background())
	assert.Equal(t, "", result.Runtime.GoVersion)
}

func TestContext(t *testing.T) {
	r1 := Result{
		Runtime: GetRuntimeResult(),
	}

	assert.NotEmpty(t, r1.Runtime.GoVersion)
	assert.NotEmpty(t, r1.Runtime.GoArch)
	assert.NotEmpty(t, r1.Runtime.GoOS)

	ctx := WithResult(context.Background(), r1)
	r2 := FromContext(ctx)

	assert.Equal(t, r1.Runtime.GoVersion, r2.Runtime.GoVersion)
	assert.Equal(t, r1.Runtime.GoArch, r2.Runtime.GoArch)
	assert.Equal(t, r1.Runtime.GoOS, r2.Runtime.GoOS)
}
