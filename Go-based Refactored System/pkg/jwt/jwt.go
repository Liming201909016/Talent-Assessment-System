package jwt

import (
	"errors"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Create 创建 HS512 JWT，claim 由调用方构造；secret 与 Java token.secret 字节级一致。
// Java 参考：Jwts.builder().setClaims(claims).signWith(SignatureAlgorithm.HS512, secret).compact()
// Java 这里未设置 expiration / issuer / subject，因此我们也不加任何额外 Registered Claims，
// 使新旧双向 token 完全等价。
func Create(secret string, claims map[string]any) (string, error) {
	mc := jwtv5.MapClaims{}
	for k, v := range claims {
		mc[k] = v
	}
	t := jwtv5.NewWithClaims(jwtv5.SigningMethodHS512, mc)
	return t.SignedString([]byte(secret))
}

// Parse 解析 token 并返回 claims；仅接受 HS512。
func Parse(secret, token string) (map[string]any, error) {
	parsed, err := jwtv5.Parse(token, func(t *jwtv5.Token) (any, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwtv5.WithValidMethods([]string{"HS512"}))
	if err != nil {
		return nil, err
	}
	mc, ok := parsed.Claims.(jwtv5.MapClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid claims")
	}
	out := make(map[string]any, len(mc))
	for k, v := range mc {
		out[k] = v
	}
	return out, nil
}
