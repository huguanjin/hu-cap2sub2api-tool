package converter

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// parsedRecord 是各平台 CPA 记录解析结果的内部表示，对应 JS 版
// parseOpenAIRecord/parseClaudeRecord/parseAntigravityRecord/parseGeminiRecord 的返回值。
type parsedRecord struct {
	providerLabel string
	platform      string
	accountType   string
	email         string
	planType      string
	expiresAt     string
	credentials   anyMap
	extra         anyMap
}

// CPAConversionResult 对应 JS 版 convertCPARecord 的返回值。
type CPAConversionResult struct {
	SourceName     string
	SourceType     string
	ProviderLabel  string
	Email          string
	PlanType       string
	ExpiresAt      string
	Account        anyMap
	Document       anyMap
	OutputFileName string
}

func getOpenAIAuthSection(payload anyMap) anyMap {
	m, _ := getMap(payload, "https://api.openai.com/auth")
	return m
}

func getOpenAIProfileSection(payload anyMap) anyMap {
	m, _ := getMap(payload, "https://api.openai.com/profile")
	return m
}

// deriveOrganizationID 对应 JS 版 deriveOrganizationId。
func deriveOrganizationID(idAuth, accessAuth anyMap) string {
	for _, source := range []anyMap{idAuth, accessAuth} {
		if source == nil {
			continue
		}
		orgsRaw, ok := getSlice(source, "organizations")
		if !ok {
			continue
		}
		firstID := ""
		for _, orgRaw := range orgsRaw {
			org, ok := asMap(orgRaw)
			if !ok {
				continue
			}
			id, idOK := getString(org, "id")
			if !idOK || id == "" {
				continue
			}
			if firstID == "" {
				firstID = id
			}
			if isDefault, _ := getBool(org, "is_default"); isDefault {
				return id
			}
		}
		if firstID != "" {
			return firstID
		}
	}
	return ""
}

// buildCommonExtra 对应 JS 版 buildCommonExtra。
func buildCommonExtra(record anyMap, email string) anyMap {
	lastRefresh, _ := normalizeFlexibleTimestamp(record["last_refresh"])
	emailKey, _ := toEmailKey(email)
	return stripMap(anyMap{
		"email":        email,
		"email_key":    emailKey,
		"last_refresh": lastRefresh,
	})
}

// parseOpenAIRecord 对应 JS 版 parseOpenAIRecord（codex / OpenAI）。
func parseOpenAIRecord(record anyMap, opts ConvertOptions) (parsedRecord, error) {
	accessToken, _ := getString(record, "access_token")
	if strings.TrimSpace(accessToken) == "" {
		return parsedRecord{}, errors.New("缺少 access_token")
	}
	idToken, _ := getString(record, "id_token")
	if strings.TrimSpace(idToken) == "" {
		return parsedRecord{}, errors.New("缺少 id_token")
	}

	accessPayload, accessErr := ParseJWTPayload(accessToken)
	if accessErr != nil || accessPayload == nil {
		return parsedRecord{}, errors.New("access_token 不是有效 JWT")
	}
	idPayload, idErr := ParseJWTPayload(idToken)
	if idErr != nil || idPayload == nil {
		return parsedRecord{}, errors.New("id_token 不是有效 JWT")
	}

	accessAuth := getOpenAIAuthSection(accessPayload)
	idAuth := getOpenAIAuthSection(idPayload)
	accessProfile := getOpenAIProfileSection(accessPayload)
	now := opts.now()

	recordEmail, _ := getString(record, "email")
	accessProfileEmail, _ := getString(accessProfile, "email")
	accessPayloadEmail, _ := getString(accessPayload, "email")
	idPayloadEmail, _ := getString(idPayload, "email")
	email, _ := firstNonEmpty(recordEmail, accessProfileEmail, accessPayloadEmail, idPayloadEmail)

	expiredNorm, _ := normalizeFlexibleTimestamp(record["expired"])
	expFromJWT, _ := timestampFromUnixSeconds(accessPayload["exp"])
	expiresAt, _ := firstNonEmpty(expiredNorm, expFromJWT)

	recordPlanType, _ := getString(record, "plan_type")
	accessAuthPlanType, _ := getString(accessAuth, "chatgpt_plan_type")
	idAuthPlanType, _ := getString(idAuth, "chatgpt_plan_type")
	planType, _ := firstNonEmpty(recordPlanType, accessAuthPlanType, idAuthPlanType)

	recordAccountID, _ := getString(record, "account_id")
	accessAuthAccountID, _ := getString(accessAuth, "chatgpt_account_id")
	idAuthAccountID, _ := getString(idAuth, "chatgpt_account_id")
	chatgptAccountID, _ := firstNonEmpty(recordAccountID, accessAuthAccountID, idAuthAccountID)

	accessAuthUserID, _ := getString(accessAuth, "chatgpt_user_id")
	idAuthUserID, _ := getString(idAuth, "chatgpt_user_id")
	accessAuthUserIDAlt, _ := getString(accessAuth, "user_id")
	idAuthUserIDAlt, _ := getString(idAuth, "user_id")
	chatgptUserID, _ := firstNonEmpty(accessAuthUserID, idAuthUserID, accessAuthUserIDAlt, idAuthUserIDAlt)

	orgID := deriveOrganizationID(idAuth, accessAuth)

	credentials := anyMap{
		"access_token":       accessToken,
		"chatgpt_account_id": chatgptAccountID,
		"chatgpt_user_id":    chatgptUserID,
		"email":              email,
		"expires_at":         expiresAt,
		"id_token":           idToken,
		"organization_id":    orgID,
		"plan_type":          planType,
		"refresh_token":      record["refresh_token"],
	}
	if expiresIn, ok := getExpiresIn(expiresAt, now); ok {
		credentials["expires_in"] = expiresIn
	}

	return parsedRecord{
		providerLabel: "Codex / OpenAI",
		platform:      "openai",
		accountType:   "oauth",
		email:         email,
		planType:      planType,
		expiresAt:     expiresAt,
		credentials:   stripMap(credentials),
		extra:         buildCommonExtra(record, email),
	}, nil
}

// parseClaudeRecord 对应 JS 版 parseClaudeRecord。
func parseClaudeRecord(record anyMap) (parsedRecord, error) {
	accessToken, _ := getString(record, "access_token")
	if strings.TrimSpace(accessToken) == "" {
		return parsedRecord{}, errors.New("缺少 access_token")
	}

	recordEmail, _ := getString(record, "email")
	email, _ := firstNonEmpty(recordEmail)
	expiresAt, _ := normalizeFlexibleTimestamp(record["expired"])
	expiresAtUnixSecStr, _ := normalizeUnixSecondsString(expiresAt)

	idTokenRaw, _ := getString(record, "id_token")
	idToken, _ := firstNonEmpty(idTokenRaw)
	refreshTokenRaw, _ := getString(record, "refresh_token")
	refreshToken, _ := firstNonEmpty(refreshTokenRaw)

	credentials := stripMap(anyMap{
		"access_token":  accessToken,
		"email_address": email,
		"expires_at":    expiresAtUnixSecStr,
		"id_token":      idToken,
		"refresh_token": refreshToken,
	})

	return parsedRecord{
		providerLabel: "Claude",
		platform:      "anthropic",
		accountType:   "oauth",
		email:         email,
		expiresAt:     expiresAt,
		credentials:   credentials,
		extra:         buildCommonExtra(record, email),
	}, nil
}

// parseAntigravityRecord 对应 JS 版 parseAntigravityRecord。
func parseAntigravityRecord(record anyMap) (parsedRecord, error) {
	accessToken, _ := getString(record, "access_token")
	if strings.TrimSpace(accessToken) == "" {
		return parsedRecord{}, errors.New("缺少 access_token")
	}

	derivedExpiresAt := ""
	if explicit, ok := normalizeFlexibleTimestamp(record["expired"]); ok {
		derivedExpiresAt = explicit
	} else if timestamp, tsOK := record["timestamp"].(float64); tsOK {
		if expiresIn, eiOK := record["expires_in"].(float64); eiOK {
			ms := timestamp + expiresIn*1000
			derivedExpiresAt = formatISO(time.UnixMilli(int64(ms)))
		}
	}

	recordEmail, _ := getString(record, "email")
	email, _ := firstNonEmpty(recordEmail)
	recordPlanType, _ := getString(record, "plan_type")
	planType, _ := firstNonEmpty(recordPlanType)

	expiresAtUnixSecStr, _ := normalizeUnixSecondsString(derivedExpiresAt)
	recordProjectID, _ := getString(record, "project_id")
	projectID, _ := firstNonEmpty(recordProjectID)
	recordRefreshToken, _ := getString(record, "refresh_token")
	refreshToken, _ := firstNonEmpty(recordRefreshToken)
	recordTokenType, _ := getString(record, "token_type")
	tokenType, _ := firstNonEmpty(recordTokenType)

	credentials := anyMap{
		"access_token":  accessToken,
		"email":         email,
		"expires_at":    expiresAtUnixSecStr,
		"project_id":    projectID,
		"refresh_token": refreshToken,
		"token_type":    tokenType,
		"plan_type":     planType,
	}
	if expiresIn, ok := record["expires_in"].(float64); ok {
		credentials["expires_in"] = expiresIn
	}

	return parsedRecord{
		providerLabel: "Antigravity",
		platform:      "antigravity",
		accountType:   "oauth",
		email:         email,
		planType:      planType,
		expiresAt:     derivedExpiresAt,
		credentials:   stripMap(credentials),
		extra:         buildCommonExtra(record, email),
	}, nil
}

// parseGeminiRecord 对应 JS 版 parseGeminiRecord。
func parseGeminiRecord(record anyMap) (parsedRecord, error) {
	rawToken, ok := getMap(record, "token")
	if !ok {
		return parsedRecord{}, errors.New("缺少 token 对象")
	}

	tokenAccessToken, _ := getString(rawToken, "access_token")
	tokenAccessTokenCamel, _ := getString(rawToken, "accessToken")
	accessToken, accessOK := firstNonEmpty(tokenAccessToken, tokenAccessTokenCamel)
	if !accessOK {
		return parsedRecord{}, errors.New("token 中缺少 access_token")
	}

	expiryA, _ := normalizeFlexibleTimestamp(rawToken["expiry"])
	expiryB, _ := normalizeFlexibleTimestamp(rawToken["expires_at"])
	expiryC, _ := normalizeFlexibleTimestamp(rawToken["expiration"])
	expiryD, _ := timestampFromUnixSeconds(rawToken["expires_in_abs"])
	expiresAt, _ := firstNonEmpty(expiryA, expiryB, expiryC, expiryD)

	recordProjectID, _ := getString(record, "project_id")
	projectID, projectOK := firstNonEmpty(recordProjectID)
	oauthType := ""
	if projectOK {
		oauthType = "code_assist"
	}

	recordEmail, _ := getString(record, "email")
	email, _ := firstNonEmpty(recordEmail)

	expiresAtUnixSecStr, _ := normalizeUnixSecondsString(expiresAt)

	tokenRefresh, _ := getString(rawToken, "refresh_token")
	tokenRefreshCamel, _ := getString(rawToken, "refreshToken")
	refreshToken, _ := firstNonEmpty(tokenRefresh, tokenRefreshCamel)

	scope, _ := joinScopes(firstDefined(rawToken["scope"], rawToken["scopes"]))

	tokenType1, _ := getString(rawToken, "token_type")
	tokenType2, _ := getString(rawToken, "tokenType")
	tokenType, _ := firstNonEmpty(tokenType1, tokenType2)

	credentials := stripMap(anyMap{
		"access_token":  accessToken,
		"expires_at":    expiresAtUnixSecStr,
		"oauth_type":    oauthType,
		"project_id":    projectID,
		"refresh_token": refreshToken,
		"scope":         scope,
		"token_type":    tokenType,
	})

	extra := anyMap{}
	for k, v := range buildCommonExtra(record, email) {
		extra[k] = v
	}
	if autoVal, ok := record["auto"].(bool); ok {
		extra["auto"] = autoVal
	}
	if checkedVal, ok := record["checked"].(bool); ok {
		extra["checked"] = checkedVal
	}

	return parsedRecord{
		providerLabel: "Gemini",
		platform:      "gemini",
		accountType:   "oauth",
		email:         email,
		expiresAt:     expiresAt,
		credentials:   credentials,
		extra:         stripMap(extra),
	}, nil
}

// ConvertCPARecord 对应 JS 版 convertCPARecord：将单条 CPA 记录转换为 sub2api 账号文档。
func ConvertCPARecord(record anyMap, opts ConvertOptions) (*CPAConversionResult, error) {
	if record == nil {
		return nil, errors.New("文件不是 JSON 对象")
	}

	rawType, _ := getString(record, "type")
	sourceType := strings.ToLower(strings.TrimSpace(rawType))
	exportedAt, _ := normalizeTimestamp(opts.now())

	dispatchType := sourceType
	if dispatchType == "" {
		dispatchType = "codex"
	}

	var parsed parsedRecord
	var err error
	switch dispatchType {
	case "codex":
		parsed, err = parseOpenAIRecord(record, opts)
	case "claude":
		parsed, err = parseClaudeRecord(record)
	case "antigravity":
		parsed, err = parseAntigravityRecord(record)
	case "gemini":
		parsed, err = parseGeminiRecord(record)
	default:
		return nil, fmt.Errorf("暂不支持 type=%s 的 CPA 文件", rawType)
	}
	if err != nil {
		return nil, err
	}

	if len(parsed.credentials) == 0 {
		return nil, errors.New("没有可导出的认证字段")
	}

	accountName, _ := firstNonEmpty(parsed.email, opts.SourceName, "converted-account")
	account := stripMap(anyMap{
		"name":        accountName,
		"platform":    parsed.platform,
		"type":        parsed.accountType,
		"concurrency": float64(10),
		"priority":    float64(1),
		"credentials": parsed.credentials,
		"extra":       parsed.extra,
	})

	document := anyMap{
		"exported_at": exportedAt,
		"proxies":     []interface{}{},
		"accounts":    []interface{}{account},
	}

	resultSourceType := sourceType
	if resultSourceType == "" {
		resultSourceType = "codex"
	}

	return &CPAConversionResult{
		SourceName:     opts.SourceName,
		SourceType:     resultSourceType,
		ProviderLabel:  parsed.providerLabel,
		Email:          parsed.email,
		PlanType:       parsed.planType,
		ExpiresAt:      parsed.expiresAt,
		Account:        account,
		Document:       document,
		OutputFileName: BuildSub2ApiOutputFileName(opts.SourceName, parsed.email),
	}, nil
}
