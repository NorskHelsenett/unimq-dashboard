package httpsuite

import "context"

func GetClaimsFromContext(ctx context.Context) (map[string]interface{}, bool) {
	claims, ok := ctx.Value("claims").(map[string]interface{})
	return claims, ok
}
