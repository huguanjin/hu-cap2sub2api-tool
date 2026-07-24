package converter

import (
	"errors"
	"fmt"
	"strings"
)

// Sub2ApiConversionResult 对应 JS 版各 convertSub2Api*Account 的返回值。
type Sub2ApiConversionResult struct {
	SourceName     string
	SourceType     string
	ProviderLabel  string
	Email          string
	PlanType       string
	ExpiresAt      string
	EntryLabel     string
	Document       anyMap
	OutputFileName string
}

// Sub2ApiBatchResult 汇总 ConvertSub2ApiDocument 的输出。
type Sub2ApiBatchResult struct {
	Converted []*Sub2ApiConversionResult
	Skipped   []SkippedEntry
}

// NormalizeSub2ApiPlatform 对应 JS 版 normalizeSub2ApiPlatform：归一化 platform 别名。
func NormalizeSub2ApiPlatform(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "openai", "codex":
		return "codex"
	case "anthropic", "claude":
		return "claude"
	case "antigravity":
		return "antigravity"
	case "gemini":
		return "gemini"
	default:
		return ""
	}
}

func sub2ApiCredentials(account anyMap) (anyMap, error) {
	credentials, ok := getMap(account, "credentials")
	if !ok {
		return nil, errors.New("缺少 account.credentials 对象")
	}
	return credentials, nil
}

func sub2ApiExtra(account anyMap) anyMap {
	extra, ok := getMap(account, "extra")
	if !ok {
		return anyMap{}
	}
	return extra
}

// BuildSub2ApiEntryLabel 对应 JS 版 buildSub2ApiEntryLabel。
func BuildSub2ApiEntryLabel(accountRaw interface{}, index int) string {
	fallback := fmt.Sprintf("accounts[%d]", index)
	account, ok := asMap(accountRaw)
	if !ok {
		return fallback
	}

	credentials, _ := getMap(account, "credentials")
	extra := sub2ApiExtra(account)

	name, _ := getString(account, "name")
	extraEmail, _ := getString(extra, "email")
	credEmail, _ := getString(credentials, "email")
	credEmailAddr, _ := getString(credentials, "email_address")

	label, _ := firstNonEmpty(name, extraEmail, credEmail, credEmailAddr, fallback)
	return label
}

// ExtractSub2ApiAccounts 对应 JS 版 extractSub2ApiAccounts。
func ExtractSub2ApiAccounts(document interface{}) ([]interface{}, error) {
	if arr, ok := document.([]interface{}); ok {
		return arr, nil
	}

	if doc, ok := asMap(document); ok {
		if accountsRaw, ok := getSlice(doc, "accounts"); ok {
			return accountsRaw, nil
		}
		if platform, ok := getString(doc, "platform"); ok && strings.TrimSpace(platform) != "" {
			if _, credOK := getMap(doc, "credentials"); credOK {
				return []interface{}{doc}, nil
			}
		}
	}

	return nil, errors.New("不是有效的 sub2api 配置，缺少 accounts 数组")
}

// convertSub2ApiOpenAIAccount 对应 JS 版 convertSub2ApiOpenAIAccount。
func convertSub2ApiOpenAIAccount(account anyMap, opts ConvertOptions) (*Sub2ApiConversionResult, error) {
	credentials, err := sub2ApiCredentials(account)
	if err != nil {
		return nil, err
	}
	extra := sub2ApiExtra(account)
	now := opts.now()

	credAccessToken, _ := getString(credentials, "access_token")
	accessToken, accessOK := firstNonEmpty(credAccessToken)
	if !accessOK {
		return nil, errors.New("credentials.access_token 为空")
	}

	credRefreshToken, _ := getString(credentials, "refresh_token")
	refreshToken, _ := firstNonEmpty(credRefreshToken)

	credIDToken, _ := getString(credentials, "id_token")
	idToken, idOK := firstNonEmpty(credIDToken)
	if !idOK {
		return nil, errors.New("credentials.id_token 为空，无法生成 Codex CPA 文件")
	}

	accessPayload, _ := ParseJWTPayload(accessToken)
	credExpiresAt, _ := normalizeFlexibleTimestamp(credentials["expires_at"])
	expFromJWT := ""
	if accessPayload != nil {
		expFromJWT, _ = timestampFromUnixSeconds(accessPayload["exp"])
	}
	expFromNowPlus, _ := timestampFromNowPlusSeconds(credentials["expires_in"], now)
	expiresAt, _ := firstNonEmpty(credExpiresAt, expFromJWT, expFromNowPlus)

	extraEmail, _ := getString(extra, "email")
	credEmail, _ := getString(credentials, "email")
	email, _ := firstNonEmpty(extraEmail, credEmail)

	credPlanType, _ := getString(credentials, "plan_type")
	planType, _ := firstNonEmpty(credPlanType)

	accountName, _ := getString(account, "name")
	entryLabel, _ := firstNonEmpty(accountName, email)

	credAccountID, _ := getString(credentials, "chatgpt_account_id")
	accountID, _ := firstNonEmpty(credAccountID)

	lastRefresh, _ := normalizeFlexibleTimestamp(extra["last_refresh"])

	doc := stripMap(anyMap{
		"type":         "codex",
		"access_token": accessToken,
		"id_token":     idToken,
		"account_id":   accountID,
		"email":        email,
		"expired":      expiresAt,
		"last_refresh": lastRefresh,
		"plan_type":    planType,
	})
	// refresh_token 始终存在（即使为空字符串），对应 JS `refreshToken ?? ""` 在 strip 之后追加。
	doc["refresh_token"] = refreshToken

	return &Sub2ApiConversionResult{
		SourceType:     "codex",
		ProviderLabel:  "Codex / OpenAI",
		Email:          email,
		PlanType:       planType,
		ExpiresAt:      expiresAt,
		EntryLabel:     entryLabel,
		Document:       doc,
		OutputFileName: BuildCPAOutputFileName(accountName, email, "codex"),
	}, nil
}

// convertSub2ApiClaudeAccount 对应 JS 版 convertSub2ApiClaudeAccount。
func convertSub2ApiClaudeAccount(account anyMap, opts ConvertOptions) (*Sub2ApiConversionResult, error) {
	credentials, err := sub2ApiCredentials(account)
	if err != nil {
		return nil, err
	}
	extra := sub2ApiExtra(account)
	now := opts.now()

	credAccessToken, _ := getString(credentials, "access_token")
	accessToken, accessOK := firstNonEmpty(credAccessToken)
	if !accessOK {
		return nil, errors.New("credentials.access_token 为空")
	}

	extraEmail, _ := getString(extra, "email")
	credEmailAddr, _ := getString(credentials, "email_address")
	email, _ := firstNonEmpty(extraEmail, credEmailAddr)

	credExpiresAt, _ := normalizeFlexibleTimestamp(credentials["expires_at"])
	expFromNowPlus, _ := timestampFromNowPlusSeconds(credentials["expires_in"], now)
	expiresAt, _ := firstNonEmpty(credExpiresAt, expFromNowPlus)

	accountName, _ := getString(account, "name")
	entryLabel, _ := firstNonEmpty(accountName, email)

	credIDToken, _ := getString(credentials, "id_token")
	idToken, _ := firstNonEmpty(credIDToken)
	credRefreshToken, _ := getString(credentials, "refresh_token")
	refreshToken, _ := firstNonEmpty(credRefreshToken)
	lastRefresh, _ := normalizeFlexibleTimestamp(extra["last_refresh"])

	doc := stripMap(anyMap{
		"type":          "claude",
		"access_token":  accessToken,
		"email":         email,
		"expired":       expiresAt,
		"id_token":      idToken,
		"last_refresh":  lastRefresh,
		"refresh_token": refreshToken,
	})

	return &Sub2ApiConversionResult{
		SourceType:     "claude",
		ProviderLabel:  "Claude",
		Email:          email,
		ExpiresAt:      expiresAt,
		EntryLabel:     entryLabel,
		Document:       doc,
		OutputFileName: BuildCPAOutputFileName(accountName, email, "claude"),
	}, nil
}

// convertSub2ApiAntigravityAccount 对应 JS 版 convertSub2ApiAntigravityAccount。
func convertSub2ApiAntigravityAccount(account anyMap, opts ConvertOptions) (*Sub2ApiConversionResult, error) {
	credentials, err := sub2ApiCredentials(account)
	if err != nil {
		return nil, err
	}
	extra := sub2ApiExtra(account)
	now := opts.now()

	credAccessToken, _ := getString(credentials, "access_token")
	accessToken, accessOK := firstNonEmpty(credAccessToken)
	if !accessOK {
		return nil, errors.New("credentials.access_token 为空")
	}

	extraEmail, _ := getString(extra, "email")
	credEmail, _ := getString(credentials, "email")
	email, _ := firstNonEmpty(extraEmail, credEmail)

	credPlanType, _ := getString(credentials, "plan_type")
	planType, _ := firstNonEmpty(credPlanType)

	credExpiresAt, _ := normalizeFlexibleTimestamp(credentials["expires_at"])
	expFromNowPlus, _ := timestampFromNowPlusSeconds(credentials["expires_in"], now)
	expiresAt, _ := firstNonEmpty(credExpiresAt, expFromNowPlus)

	accountName, _ := getString(account, "name")
	entryLabel, _ := firstNonEmpty(accountName, email)

	lastRefresh, _ := normalizeFlexibleTimestamp(extra["last_refresh"])
	credProjectID, _ := getString(credentials, "project_id")
	projectID, _ := firstNonEmpty(credProjectID)
	credRefreshToken, _ := getString(credentials, "refresh_token")
	refreshToken, _ := firstNonEmpty(credRefreshToken)
	credTokenType, _ := getString(credentials, "token_type")
	tokenType, _ := firstNonEmpty(credTokenType)

	doc := anyMap{
		"type":          "antigravity",
		"access_token":  accessToken,
		"email":         email,
		"expired":       expiresAt,
		"last_refresh":  lastRefresh,
		"plan_type":     planType,
		"project_id":    projectID,
		"refresh_token": refreshToken,
		"token_type":    tokenType,
	}
	if expiresIn, ok := toFiniteNumber(credentials["expires_in"]); ok {
		doc["expires_in"] = expiresIn
	}

	return &Sub2ApiConversionResult{
		SourceType:     "antigravity",
		ProviderLabel:  "Antigravity",
		Email:          email,
		PlanType:       planType,
		ExpiresAt:      expiresAt,
		EntryLabel:     entryLabel,
		Document:       stripMap(doc),
		OutputFileName: BuildCPAOutputFileName(accountName, email, "antigravity"),
	}, nil
}

// convertSub2ApiGeminiAccount 对应 JS 版 convertSub2ApiGeminiAccount。
func convertSub2ApiGeminiAccount(account anyMap) (*Sub2ApiConversionResult, error) {
	credentials, err := sub2ApiCredentials(account)
	if err != nil {
		return nil, err
	}
	extra := sub2ApiExtra(account)

	credAccessToken, _ := getString(credentials, "access_token")
	accessToken, accessOK := firstNonEmpty(credAccessToken)
	if !accessOK {
		return nil, errors.New("credentials.access_token 为空")
	}

	extraEmail, _ := getString(extra, "email")
	email, _ := firstNonEmpty(extraEmail)
	expiresAt, _ := normalizeFlexibleTimestamp(credentials["expires_at"])

	accountName, _ := getString(account, "name")
	entryLabel, _ := firstNonEmpty(accountName, email)

	lastRefresh, _ := normalizeFlexibleTimestamp(extra["last_refresh"])
	credProjectID, _ := getString(credentials, "project_id")
	projectID, _ := firstNonEmpty(credProjectID)
	credRefreshToken, _ := getString(credentials, "refresh_token")
	refreshToken, _ := firstNonEmpty(credRefreshToken)
	credScope, _ := getString(credentials, "scope")
	scope, _ := firstNonEmpty(credScope)
	credTokenType, _ := getString(credentials, "token_type")
	tokenType, _ := firstNonEmpty(credTokenType)

	doc := anyMap{
		"type":         "gemini",
		"email":        email,
		"last_refresh": lastRefresh,
		"project_id":   projectID,
		"token": anyMap{
			"access_token":  accessToken,
			"expiry":        expiresAt,
			"refresh_token": refreshToken,
			"scope":         scope,
			"token_type":    tokenType,
		},
	}
	if checkedVal, ok := extra["checked"].(bool); ok {
		doc["checked"] = checkedVal
	}
	if autoVal, ok := extra["auto"].(bool); ok {
		doc["auto"] = autoVal
	}

	return &Sub2ApiConversionResult{
		SourceType:     "gemini",
		ProviderLabel:  "Gemini",
		Email:          email,
		ExpiresAt:      expiresAt,
		EntryLabel:     entryLabel,
		Document:       stripMap(doc),
		OutputFileName: BuildCPAOutputFileName(accountName, email, "gemini"),
	}, nil
}

// ConvertSub2ApiAccount 对应 JS 版 convertSub2ApiAccount：按 platform 分发到具体转换函数。
func ConvertSub2ApiAccount(accountRaw interface{}, opts ConvertOptions) (*Sub2ApiConversionResult, error) {
	account, ok := asMap(accountRaw)
	if !ok {
		return nil, errors.New("account 不是对象")
	}

	accountType, _ := getString(account, "type")
	normalizedType := strings.ToLower(strings.TrimSpace(accountType))
	if normalizedType != "" && normalizedType != "oauth" {
		return nil, fmt.Errorf("暂不支持 type=%s 的 sub2api 账号", accountType)
	}

	platformRaw, _ := getString(account, "platform")
	sourceType := NormalizeSub2ApiPlatform(platformRaw)

	switch sourceType {
	case "codex":
		return convertSub2ApiOpenAIAccount(account, opts)
	case "claude":
		return convertSub2ApiClaudeAccount(account, opts)
	case "antigravity":
		return convertSub2ApiAntigravityAccount(account, opts)
	case "gemini":
		return convertSub2ApiGeminiAccount(account)
	default:
		return nil, fmt.Errorf("暂不支持 platform=%s 的 sub2api 账号", platformRaw)
	}
}

// ConvertSub2ApiDocument 对应 JS 版 convertSub2ApiDocument：批量转换并收集跳过项。
func ConvertSub2ApiDocument(document interface{}, opts ConvertOptions) (*Sub2ApiBatchResult, error) {
	accounts, err := ExtractSub2ApiAccounts(document)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, errors.New("sub2api 配置中的 accounts 为空")
	}

	result := &Sub2ApiBatchResult{}
	for index, accountRaw := range accounts {
		entryLabel := BuildSub2ApiEntryLabel(accountRaw, index)

		converted, convErr := ConvertSub2ApiAccount(accountRaw, opts)
		if convErr != nil {
			result.Skipped = append(result.Skipped, SkippedEntry{
				SourceName: opts.SourceName,
				EntryLabel: entryLabel,
				Reason:     convErr.Error(),
			})
			continue
		}

		converted.SourceName = opts.SourceName
		converted.EntryLabel = entryLabel
		result.Converted = append(result.Converted, converted)
	}

	return result, nil
}
