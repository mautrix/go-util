package emojirunes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.mau.fi/util/emojirunes"
)

func TestIsOnlyEmojis(t *testing.T) {
	assert.True(t, emojirunes.IsOnlyEmojis("🤔"))
	assert.True(t, emojirunes.IsOnlyEmojis("👨‍👩‍👧‍👦"))
}

func TestIsOnlyEmojis_Keycaps(t *testing.T) {
	assert.True(t, emojirunes.IsOnlyEmojis("#️⃣*️⃣1️⃣2️⃣3️⃣4️⃣5️⃣6️⃣7️⃣8️⃣9️⃣0️⃣"))
}
