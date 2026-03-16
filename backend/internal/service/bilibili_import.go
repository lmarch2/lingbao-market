package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/lingbao-market/backend/internal/model"
)

const (
	bilibiliAPIBase = "https://api.bilibili.com"
	bilibiliWebBase = "https://www.bilibili.com"

	bilibiliUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

	bilibiliMaxPrice              = 999.0
	bilibiliMaxImportLimit        = 200
	bilibiliDefaultSearchPageSize = 20
	bilibiliMaxResponseBytes      = 4 << 20

	bilibiliMixinKeyTTL = 12 * time.Hour
)

var (
	bilibiliHTMLTagRe        = regexp.MustCompile(`<[^>]*>`)
	bilibiliWbiUnsafeCharsRe = regexp.MustCompile(`[!'()*]`)
	bilibiliPriceRangeRe     = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*[-~～]\s*(\d+(?:\.\d+)?)`)
	bilibiliNumberRe         = regexp.MustCompile(`\d+(?:\.\d+)?`)

	bilibiliCodeBracketRes = []*regexp.Regexp{
		regexp.MustCompile(`【([^】]{1,64})】`),
		regexp.MustCompile(`\[([^\]]{1,64})\]`),
		regexp.MustCompile(`（([^）]{1,64})）`),
		regexp.MustCompile(`\(([^)]{1,64})\)`),
		regexp.MustCompile(`《([^》]{1,64})》`),
	}
	bilibiliCodeKeywordRes = []*regexp.Regexp{
		regexp.MustCompile(`(?:兑换码|激活码|口令|暗号|代码)\s*[:：]?\s*([A-Za-z0-9]{3,12})`),
		regexp.MustCompile(`(?:码)\s*[:：]\s*([A-Za-z0-9]{3,12})`),
	}
	bilibiliPriceRes = []*regexp.Regexp{
		regexp.MustCompile(`[￥¥]\s*(\d+(?:\.\d+)?)`),
		regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:块|元|金)`),
		regexp.MustCompile(`(?:价格|价钱|售价|卖|出|收)\s*[:：]?\s*(\d+(?:\.\d+)?)`),
		regexp.MustCompile(`(\d+(?:\.\d+)?)\s*\+`),
	}

	bilibiliMixinKeyEncTab = []int{
		46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35,
		27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13,
		37, 48, 7, 16, 24, 55, 40, 61, 26, 17, 0, 1, 60, 51, 30, 4,
		22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11, 36, 20, 34, 44, 52,
	}
)

type BilibiliImportOptions struct {
	Keyword        string
	Limit          int
	MinPrice       float64
	SearchPages    int
	SearchPageSize int
	CommentPages   int
	Server         string
}

type BilibiliImporter struct {
	client *http.Client

	mu                sync.Mutex
	mixinKey          string
	mixinKeyFetchedAt time.Time
}

type BilibiliImportWarning struct {
	cause error
}

func (w *BilibiliImportWarning) Error() string {
	if w == nil || w.cause == nil {
		return ""
	}
	return w.cause.Error()
}

func (w *BilibiliImportWarning) Unwrap() error {
	if w == nil {
		return nil
	}
	return w.cause
}

func NewBilibiliImporter(extraCookie string) (*BilibiliImporter, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	importer := &BilibiliImporter{
		client: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
	}
	importer.addExtraCookies(strings.TrimSpace(extraCookie))
	return importer, nil
}

func (i *BilibiliImporter) ImportHighPriceCodes(ctx context.Context, svc *PriceService, opts BilibiliImportOptions) (int, error) {
	if i == nil {
		return 0, errors.New("bilibili importer is nil")
	}
	if svc == nil {
		return 0, errors.New("price service is nil")
	}

	keyword := strings.TrimSpace(opts.Keyword)
	if keyword == "" {
		return 0, errors.New("keyword is empty")
	}

	limit := opts.Limit
	if limit <= 0 {
		return 0, errors.New("limit must be positive")
	}
	if limit > bilibiliMaxImportLimit {
		limit = bilibiliMaxImportLimit
	}

	searchPages := opts.SearchPages
	if searchPages <= 0 {
		searchPages = 1
	}

	searchPageSize := opts.SearchPageSize
	if searchPageSize <= 0 {
		searchPageSize = bilibiliDefaultSearchPageSize
	}
	if searchPageSize > 50 {
		searchPageSize = 50
	}

	commentPages := opts.CommentPages
	if commentPages <= 0 {
		commentPages = 1
	}
	if commentPages > 5 {
		commentPages = 5
	}

	minPrice := opts.MinPrice
	if minPrice <= 0 {
		minPrice = 1
	}

	server := strings.TrimSpace(opts.Server)

	if err := i.ensureWebCookies(ctx); err != nil {
		return 0, err
	}

	codeToPrice := make(map[string]float64)
	var firstWarn error

	for page := 1; page <= searchPages; page++ {
		videos, err := i.searchVideos(ctx, keyword, page, searchPageSize)
		if err != nil {
			return 0, err
		}

		for _, v := range videos {
			if ctx.Err() != nil {
				return 0, ctx.Err()
			}

			baseText := strings.TrimSpace(stripHTML(v.Title) + " " + stripHTML(v.Description))
			fallbackPrice, _ := extractPriceFromText(baseText)

			if fallbackPrice > 0 {
				if code := extractCodeFromText(baseText); code != "" {
					codeToPrice[code] = maxFloat(codeToPrice[code], fallbackPrice)
				}
			}

			referer := bilibiliWebBase + "/"
			if strings.TrimSpace(v.BVID) != "" {
				referer = bilibiliWebBase + "/video/" + strings.TrimSpace(v.BVID)
			}
			messages, err := i.fetchReplyMessages(ctx, v.AID, commentPages, referer)
			if err != nil {
				if firstWarn == nil {
					firstWarn = err
				}
				continue
			}

			for _, msg := range messages {
				if ctx.Err() != nil {
					return 0, ctx.Err()
				}

				price, ok := extractPriceFromText(msg)
				if !ok && fallbackPrice > 0 {
					price = fallbackPrice
					ok = true
				}
				if !ok || price < minPrice {
					continue
				}

				code := extractCodeFromText(msg)
				if code == "" {
					continue
				}

				codeToPrice[code] = maxFloat(codeToPrice[code], price)
			}
		}
	}

	type candidate struct {
		Code  string
		Price float64
	}
	candidates := make([]candidate, 0, len(codeToPrice))
	for code, price := range codeToPrice {
		if price < minPrice {
			continue
		}
		candidates = append(candidates, candidate{Code: code, Price: price})
	}
	sort.Slice(candidates, func(a, b int) bool {
		if candidates[a].Price == candidates[b].Price {
			return candidates[a].Code < candidates[b].Code
		}
		return candidates[a].Price > candidates[b].Price
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	imported := 0
	for _, c := range candidates {
		if ctx.Err() != nil {
			return imported, ctx.Err()
		}
		item := model.PriceItem{
			Code:   c.Code,
			Price:  c.Price,
			Server: server,
		}
		if err := svc.AddPrice(ctx, item); err != nil {
			return imported, err
		}
		imported++
	}

	if firstWarn != nil {
		return imported, &BilibiliImportWarning{cause: firstWarn}
	}

	return imported, nil
}

type bilibiliSearchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
	Data    struct {
		Result []bilibiliVideo `json:"result"`
	} `json:"data"`
}

type bilibiliVideo struct {
	AID         int64  `json:"aid"`
	BVID        string `json:"bvid"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (i *BilibiliImporter) searchVideos(ctx context.Context, keyword string, page, pageSize int) ([]bilibiliVideo, error) {
	params := url.Values{}
	params.Set("search_type", "video")
	params.Set("keyword", keyword)
	params.Set("page", strconv.Itoa(page))
	params.Set("page_size", strconv.Itoa(pageSize))

	var resp bilibiliSearchResponse
	if err := i.getSignedJSON(ctx, "/x/web-interface/wbi/search/type", params, "https://search.bilibili.com/", &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		msg := strings.TrimSpace(resp.Message)
		if msg == "" {
			msg = strings.TrimSpace(resp.Msg)
		}
		return nil, fmt.Errorf("bilibili search failed: code=%d msg=%q", resp.Code, msg)
	}
	return resp.Data.Result, nil
}

type bilibiliNavResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	} `json:"data"`
}

func (i *BilibiliImporter) getMixinKey(ctx context.Context) (string, error) {
	i.mu.Lock()
	cachedKey := i.mixinKey
	cachedAt := i.mixinKeyFetchedAt
	i.mu.Unlock()

	if cachedKey != "" && time.Since(cachedAt) < bilibiliMixinKeyTTL {
		return cachedKey, nil
	}

	var nav bilibiliNavResponse
	if err := i.getJSON(ctx, bilibiliAPIBase+"/x/web-interface/nav", bilibiliWebBase+"/", &nav); err != nil {
		return "", err
	}
	imgURL := strings.TrimSpace(nav.Data.WbiImg.ImgURL)
	subURL := strings.TrimSpace(nav.Data.WbiImg.SubURL)
	if nav.Code != 0 && (imgURL == "" || subURL == "") {
		return "", fmt.Errorf("bilibili nav failed: code=%d msg=%q", nav.Code, strings.TrimSpace(nav.Message))
	}
	if imgURL == "" || subURL == "" {
		return "", fmt.Errorf("bilibili nav missing wbi images: code=%d msg=%q", nav.Code, strings.TrimSpace(nav.Message))
	}

	imgKey, err := extractWbiKeyFromURL(imgURL)
	if err != nil {
		return "", fmt.Errorf("invalid wbi img url: %w", err)
	}
	subKey, err := extractWbiKeyFromURL(subURL)
	if err != nil {
		return "", fmt.Errorf("invalid wbi sub url: %w", err)
	}

	mixinKey, err := buildMixinKey(imgKey, subKey)
	if err != nil {
		return "", err
	}

	i.mu.Lock()
	i.mixinKey = mixinKey
	i.mixinKeyFetchedAt = time.Now()
	i.mu.Unlock()

	return mixinKey, nil
}

func buildMixinKey(imgKey, subKey string) (string, error) {
	orig := imgKey + subKey
	if len(orig) < 64 {
		return "", errors.New("wbi key too short")
	}

	var b strings.Builder
	b.Grow(64)
	for _, idx := range bilibiliMixinKeyEncTab {
		if idx < 0 || idx >= len(orig) {
			return "", errors.New("invalid mixin index")
		}
		b.WriteByte(orig[idx])
	}

	mixed := b.String()
	if len(mixed) < 32 {
		return "", errors.New("mixin key too short")
	}
	return mixed[:32], nil
}

func extractWbiKeyFromURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	base := strings.TrimSpace(path.Base(parsed.Path))
	if base == "" || base == "." || base == "/" {
		return "", errors.New("missing filename")
	}
	base = strings.TrimSuffix(base, path.Ext(base))
	if base == "" {
		return "", errors.New("missing key")
	}
	return base, nil
}

type bilibiliReplyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Cursor struct {
			IsEnd           bool `json:"is_end"`
			PaginationReply struct {
				NextOffset string `json:"next_offset"`
			} `json:"pagination_reply"`
		} `json:"cursor"`
		Replies    []bilibiliReply `json:"replies"`
		TopReplies []bilibiliReply `json:"top_replies"`
		Hots       []bilibiliReply `json:"hots"`
	} `json:"data"`
}

type bilibiliReply struct {
	Content struct {
		Message string `json:"message"`
	} `json:"content"`
	Replies []bilibiliReply `json:"replies"`
}

func (i *BilibiliImporter) fetchReplyMessages(ctx context.Context, oid int64, pages int, referer string) ([]string, error) {
	offset := ""
	var messages []string

	for page := 0; page < pages; page++ {
		if ctx.Err() != nil {
			return messages, ctx.Err()
		}

		params := url.Values{}
		params.Set("oid", strconv.FormatInt(oid, 10))
		params.Set("type", "1")
		params.Set("mode", "3")
		params.Set("plat", "1")
		params.Set("web_location", "1315875")

		paginationStr, err := json.Marshal(map[string]string{"offset": offset})
		if err != nil {
			return messages, err
		}
		params.Set("pagination_str", string(paginationStr))

		var resp bilibiliReplyResponse
		if err := i.getSignedJSON(ctx, "/x/v2/reply/wbi/main", params, referer, &resp); err != nil {
			return messages, err
		}
		if resp.Code != 0 {
			return messages, fmt.Errorf("bilibili reply failed: code=%d msg=%q", resp.Code, strings.TrimSpace(resp.Message))
		}
		if resp.Data == nil {
			return messages, nil
		}

		appendReplyMessages(&messages, resp.Data.Hots)
		appendReplyMessages(&messages, resp.Data.TopReplies)
		appendReplyMessages(&messages, resp.Data.Replies)

		next := strings.TrimSpace(resp.Data.Cursor.PaginationReply.NextOffset)
		if resp.Data.Cursor.IsEnd || next == "" || next == offset {
			return messages, nil
		}
		offset = next
	}

	return messages, nil
}

func appendReplyMessages(dst *[]string, replies []bilibiliReply) {
	for _, r := range replies {
		msg := strings.TrimSpace(r.Content.Message)
		if msg != "" {
			*dst = append(*dst, msg)
		}
		if len(r.Replies) > 0 {
			appendReplyMessages(dst, r.Replies)
		}
	}
}

func (i *BilibiliImporter) addExtraCookies(raw string) {
	if raw == "" || i.client == nil || i.client.Jar == nil {
		return
	}
	webURL, err := url.Parse(bilibiliWebBase)
	if err != nil {
		return
	}

	var cookies []*http.Cookie
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if name == "" || value == "" {
			continue
		}
		cookies = append(cookies, &http.Cookie{
			Name:  name,
			Value: value,
			Path:  "/",
		})
	}

	if len(cookies) > 0 {
		i.client.Jar.SetCookies(webURL, cookies)
	}
}

func (i *BilibiliImporter) ensureWebCookies(ctx context.Context) error {
	if i.client == nil {
		return errors.New("http client is nil")
	}
	if i.client.Jar == nil {
		return nil
	}

	webURL, err := url.Parse(bilibiliWebBase)
	if err != nil {
		return err
	}
	for _, c := range i.client.Jar.Cookies(webURL) {
		if strings.EqualFold(c.Name, "buvid3") {
			return nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bilibiliWebBase+"/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", bilibiliUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512<<10))

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("bilibili home http status %d", resp.StatusCode)
	}
	return nil
}

func (i *BilibiliImporter) getJSON(ctx context.Context, fullURL string, referer string, out any) error {
	if i.client == nil {
		return errors.New("http client is nil")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}

	ref := strings.TrimSpace(referer)
	if ref == "" {
		ref = bilibiliWebBase + "/"
	}

	req.Header.Set("User-Agent", bilibiliUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", ref)
	req.Header.Set("Origin", bilibiliWebBase)

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, bilibiliMaxResponseBytes))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bilibili http status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode bilibili response: %w", err)
	}
	return nil
}

func (i *BilibiliImporter) getSignedJSON(ctx context.Context, apiPath string, params url.Values, referer string, out any) error {
	mixinKey, err := i.getMixinKey(ctx)
	if err != nil {
		return err
	}
	signed := signWbi(params, mixinKey, time.Now())
	fullURL := bilibiliAPIBase + apiPath + "?" + encodeWbiQuery(signed)
	return i.getJSON(ctx, fullURL, referer, out)
}

func signWbi(params url.Values, mixinKey string, now time.Time) url.Values {
	signed := url.Values{}
	for key, values := range params {
		if len(values) == 0 {
			continue
		}
		signed.Set(key, sanitizeWbiValue(values[0]))
	}
	signed.Set("wts", strconv.FormatInt(now.Unix(), 10))

	keys := make([]string, 0, len(signed))
	for key := range signed {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for idx, key := range keys {
		if idx > 0 {
			b.WriteByte('&')
		}
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(encodeWbiComponent(signed.Get(key)))
	}

	sum := md5.Sum([]byte(b.String() + mixinKey))
	signed.Set("w_rid", fmt.Sprintf("%x", sum))
	return signed
}

func sanitizeWbiValue(value string) string {
	return bilibiliWbiUnsafeCharsRe.ReplaceAllString(value, "")
}

func encodeWbiQuery(params url.Values) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for idx, key := range keys {
		if idx > 0 {
			b.WriteByte('&')
		}
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(encodeWbiComponent(params.Get(key)))
	}
	return b.String()
}

func encodeWbiComponent(value string) string {
	escaped := url.QueryEscape(value)
	return strings.ReplaceAll(escaped, "+", "%20")
}

func extractCodeFromText(text string) string {
	normalized := strings.TrimSpace(html.UnescapeString(stripHTML(text)))
	if normalized == "" {
		return ""
	}

	for _, re := range bilibiliCodeBracketRes {
		match := re.FindStringSubmatch(normalized)
		if len(match) < 2 {
			continue
		}
		code := normalizeCode(match[1])
		if isValidImportedCode(code) {
			return code
		}
	}

	for _, re := range bilibiliCodeKeywordRes {
		match := re.FindStringSubmatch(normalized)
		if len(match) < 2 {
			continue
		}
		code := normalizeCode(match[1])
		if isValidImportedCode(code) {
			return code
		}
	}

	tokens := scanLetterDigitTokens(normalized)
	for i := len(tokens) - 1; i >= 0; i-- {
		code := normalizeCode(tokens[i])
		if isValidImportedCode(code) {
			return code
		}
	}

	return ""
}

func scanLetterDigitTokens(text string) []string {
	var tokens []string
	var b strings.Builder

	flush := func() {
		if b.Len() == 0 {
			return
		}
		tokens = append(tokens, b.String())
		b.Reset()
	}

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	return tokens
}

func normalizeCode(raw string) string {
	compact := strings.TrimSpace(raw)
	compact = strings.Join(strings.Fields(compact), "")
	return strings.ToUpper(compact)
}

func isValidImportedCode(code string) bool {
	if code == "" {
		return false
	}
	runeCount := utf8.RuneCountInString(code)
	if runeCount < 3 || runeCount > 12 {
		return false
	}

	hasASCIILetter := false
	for _, r := range code {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			hasASCIILetter = true
		}
		if r >= 'a' && r <= 'z' {
			hasASCIILetter = true
		}
	}

	return hasASCIILetter
}

func extractPriceFromText(text string) (float64, bool) {
	normalized := strings.TrimSpace(html.UnescapeString(stripHTML(text)))
	if normalized == "" {
		return 0, false
	}
	normalized = strings.NewReplacer(",", " ", "，", " ").Replace(normalized)

	if match := bilibiliPriceRangeRe.FindStringSubmatch(normalized); len(match) >= 3 {
		left, okL := parseAndNormalizePrice(match[1])
		right, okR := parseAndNormalizePrice(match[2])
		if okL && okR {
			if right >= left {
				return right, true
			}
			return left, true
		}
		if okL {
			return left, true
		}
		if okR {
			return right, true
		}
	}

	for _, re := range bilibiliPriceRes {
		match := re.FindStringSubmatch(normalized)
		if len(match) < 2 {
			continue
		}
		value, ok := parseAndNormalizePrice(match[1])
		if ok {
			return value, true
		}
	}

	numbers := bilibiliNumberRe.FindAllString(normalized, -1)
	if len(numbers) == 1 {
		value, ok := parseAndNormalizePrice(numbers[0])
		if ok {
			return value, true
		}
	}

	return 0, false
}

func parseAndNormalizePrice(raw string) (float64, bool) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}

	value = math.Floor(value)
	if value < 1 || value > bilibiliMaxPrice {
		return 0, false
	}
	return value, true
}

func stripHTML(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	return strings.TrimSpace(bilibiliHTMLTagRe.ReplaceAllString(text, ""))
}

func maxFloat(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}
