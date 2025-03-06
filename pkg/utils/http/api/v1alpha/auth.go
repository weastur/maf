package v1alpha

import (
	"crypto/sha256"
	"crypto/subtle"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/rs/zerolog/log"
)

var (
	apiKey          = "root" // pragma: allowlist secret
	unprotectedURLs = []*regexp.Regexp{
		regexp.MustCompile(`.*\/version$`),
	}
)

func authFilter(c *fiber.Ctx) bool {
	url := strings.ToLower(c.OriginalURL())

	for _, pattern := range unprotectedURLs {
		if pattern.MatchString(url) {
			log.Trace().Str("url", url).Msg("URL is unprotected")

			return true
		}
	}

	return false
}

func errorHandler(c *fiber.Ctx, err error) error {
	return WrapResponse(c, StatusError, nil, err)
}

func validateAPIKey(_ *fiber.Ctx, key string) (bool, error) {
	hashedAPIKey := sha256.Sum256([]byte(apiKey))
	hashedKey := sha256.Sum256([]byte(key))

	if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
		return true, nil
	}

	log.Warn().Msg("API key is missing or malformed")

	return false, keyauth.ErrMissingOrMalformedAPIKey
}

func AuthMiddleware() fiber.Handler {
	return keyauth.New(keyauth.Config{
		ErrorHandler: errorHandler,
		Next:         authFilter,
		Validator:    validateAPIKey,
		KeyLookup:    "header:X-Auth-Token",
		AuthScheme:   "",
		ContextKey:   "token",
	})
}
