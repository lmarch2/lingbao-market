package api

import (
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lingbao-market/backend/internal/model"
	"github.com/lingbao-market/backend/internal/service"
)

type Handler struct {
	svc      *service.PriceService
	authSvc  *service.AuthService
	adminSvc *service.AdminService
	jwtKey   []byte
}

func NewHandler(svc *service.PriceService, authSvc *service.AuthService, adminSvc *service.AdminService, jwtKey string) *Handler {
	return &Handler{
		svc:      svc,
		authSvc:  authSvc,
		adminSvc: adminSvc,
		jwtKey:   []byte(jwtKey),
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Public
	api.Get("/feed", h.GetFeed)
	api.Get("/auth/captcha", h.GetCaptcha)
	api.Post("/auth/register", h.Register)
	api.Post("/auth/login", h.Login)
	api.Post("/feedback", h.SubmitFeedback)

	// Public (login not required)
	api.Post("/submit", h.SubmitPrice)
	// api.Post("/submit", h.authMiddleware, h.SubmitPrice) // Keep for reuse

	// Admin
	admin := api.Group("/admin", h.authMiddleware, h.adminMiddleware)
	admin.Get("/users", h.ListUsers)
	admin.Post("/users", h.CreateUser)
	admin.Patch("/users/:username/ban", h.SetUserBan)
	admin.Delete("/users/:username", h.DeleteUser)
	admin.Delete("/prices/:code", h.DeletePriceByCode)
	admin.Get("/feedback", h.ListFeedback)
	admin.Post("/feedback/:id/resolve", h.ResolveFeedback)
	admin.Get("/logs", h.ListLogs)
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req model.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	req.Username = strings.TrimSpace(req.Username)
	req.CaptchaID = strings.TrimSpace(req.CaptchaID)
	req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)
	if len(req.Username) < 3 || len(req.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "username min 3 chars, password min 6 chars"})
	}
	if ok, err := h.authSvc.VerifyCaptcha(c.Context(), req.CaptchaID, req.CaptchaCode); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "captcha verification failed"})
	} else if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "invalid captcha"})
	}

	user, err := h.authSvc.Register(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	_ = h.authSvc.DeleteCaptcha(c.Context(), req.CaptchaID)

	return c.Status(201).JSON(fiber.Map{"id": user.ID, "username": user.Username})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req model.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	req.Username = strings.TrimSpace(req.Username)
	req.CaptchaID = strings.TrimSpace(req.CaptchaID)
	req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)
	if ok, err := h.authSvc.VerifyCaptcha(c.Context(), req.CaptchaID, req.CaptchaCode); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "captcha verification failed"})
	} else if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "invalid captcha"})
	}

	resp, err := h.authSvc.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}
	_ = h.authSvc.DeleteCaptcha(c.Context(), req.CaptchaID)

	return c.JSON(resp)
}

func (h *Handler) GetCaptcha(c *fiber.Ctx) error {
	id, code, err := h.authSvc.CreateCaptcha(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create captcha"})
	}
	return c.JSON(fiber.Map{"captchaId": id, "code": code})
}

func (h *Handler) GetFeed(c *fiber.Ctx) error {
	sortBy := c.Query("sort", "time") // Default to time
	prices, err := h.svc.GetLatestFeed(c.Context(), 50, sortBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch feed"})
	}
	return c.JSON(prices)
}

func (h *Handler) SubmitPrice(c *fiber.Ctx) error {
	// User ID from middleware (optional usage)
	// claims := c.Locals("user").(jwt.MapClaims)
	// userID := claims["sub"].(string)

	var req model.SubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	req.Code = strings.TrimSpace(req.Code)
	req.Code = strings.Join(strings.Fields(req.Code), "")
	req.Code = strings.ToUpper(req.Code)

	if req.Code == "" || req.Price <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid data"})
	}

	// Validate Code Format (letters/digits incl. Chinese, 3-12 chars)
	runeCount := utf8.RuneCountInString(req.Code)
	if runeCount < 3 || runeCount > 12 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid code format"})
	}
	for _, r := range req.Code {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return c.Status(400).JSON(fiber.Map{"error": "invalid code format"})
	}

	item := model.PriceItem{
		Code:   req.Code,
		Price:  req.Price,
		Server: req.Server,
	}

	if err := h.svc.AddPrice(c.Context(), item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to submit"})
	}

	return c.Status(201).JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.authSvc.ListUsers(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list users"})
	}

	resp := make([]model.UserPublic, 0, len(users))
	for _, user := range users {
		resp = append(resp, model.UserPublic{
			ID:       user.ID,
			Username: user.Username,
			IsAdmin:  user.IsAdmin,
			Banned:   user.Banned,
		})
	}
	return c.JSON(resp)
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	payload.Username = strings.TrimSpace(payload.Username)
	if len(payload.Username) < 3 || len(payload.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "username min 3 chars, password min 6 chars"})
	}

	user, err := h.authSvc.CreateUser(c.Context(), payload.Username, payload.Password, payload.IsAdmin)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	resolver := h.actorFromCtx(c)
	_ = h.adminSvc.AppendLog(c.Context(), model.AdminLogEntry{
		Type:    "user_created",
		Message: "admin created user",
		Actor:   resolver,
		Metadata: map[string]string{
			"username": payload.Username,
			"isAdmin":  boolToString(payload.IsAdmin),
		},
	})

	return c.Status(201).JSON(model.UserPublic{
		ID:       user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		Banned:   user.Banned,
	})
}

func (h *Handler) SetUserBan(c *fiber.Ctx) error {
	username := strings.TrimSpace(c.Params("username"))
	if username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing username"})
	}
	var payload struct {
		Banned bool `json:"banned"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := h.authSvc.SetBanned(c.Context(), username, payload.Banned)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	resolver := h.actorFromCtx(c)
	_ = h.adminSvc.AppendLog(c.Context(), model.AdminLogEntry{
		Type:    "user_ban_changed",
		Message: "admin changed user ban state",
		Actor:   resolver,
		Metadata: map[string]string{
			"username": username,
			"banned":   boolToString(payload.Banned),
		},
	})

	return c.JSON(model.UserPublic{
		ID:       user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		Banned:   user.Banned,
	})
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	username := strings.TrimSpace(c.Params("username"))
	if username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing username"})
	}
	if err := h.authSvc.DeleteUser(c.Context(), username); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete user"})
	}
	resolver := h.actorFromCtx(c)
	_ = h.adminSvc.AppendLog(c.Context(), model.AdminLogEntry{
		Type:    "user_deleted",
		Message: "admin deleted user",
		Actor:   resolver,
		Metadata: map[string]string{
			"username": username,
		},
	})
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *Handler) DeletePriceByCode(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing code"})
	}
	if decoded, err := url.PathUnescape(code); err == nil {
		code = decoded
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing code"})
	}

	removedTime, removedPrice, err := h.svc.DeletePricesByCode(c.Context(), strings.ToUpper(code))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete code"})
	}
	resolver := h.actorFromCtx(c)
	_ = h.adminSvc.AppendLog(c.Context(), model.AdminLogEntry{
		Type:    "price_deleted",
		Message: "admin deleted code from market feed",
		Actor:   resolver,
		Metadata: map[string]string{
			"code":         strings.ToUpper(code),
			"removedTime":  int64ToString(removedTime),
			"removedPrice": int64ToString(removedPrice),
		},
	})

	return c.JSON(fiber.Map{
		"status":        "ok",
		"removed_time":  removedTime,
		"removed_price": removedPrice,
	})
}

func (h *Handler) SubmitFeedback(c *fiber.Ctx) error {
	var req model.FeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	req.Code = strings.TrimSpace(strings.ToUpper(req.Code))
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Code == "" || req.Reason == "" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid feedback data"})
	}
	if len(req.Reason) > 300 {
		return c.Status(400).JSON(fiber.Map{"error": "reason too long"})
	}

	reporter := "guest"
	if actor := h.actorFromOptionalAuth(c); actor != "" {
		reporter = actor
	}

	feedback, err := h.adminSvc.AddFeedback(c.Context(), req.Code, req.Reason, reporter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to submit feedback"})
	}

	return c.Status(201).JSON(feedback)
}

func (h *Handler) ListFeedback(c *fiber.Ctx) error {
	includeResolved := c.QueryBool("includeResolved", true)
	feedback, err := h.adminSvc.ListFeedback(c.Context(), 200, includeResolved)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list feedback"})
	}
	return c.JSON(feedback)
}

func (h *Handler) ResolveFeedback(c *fiber.Ctx) error {
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing feedback id"})
	}

	var req model.ResolveFeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action != "keep" && action != "delete" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid action"})
	}

	feedback, err := h.adminSvc.GetFeedback(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "feedback not found"})
	}

	removedTime := int64(0)
	removedPrice := int64(0)
	if action == "delete" {
		removedTime, removedPrice, err = h.svc.DeletePricesByCode(c.Context(), strings.ToUpper(feedback.Code))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to delete code"})
		}
	}

	resolver := h.actorFromCtx(c)
	resolved, err := h.adminSvc.ResolveFeedback(c.Context(), id, resolver, action, removedTime, removedPrice)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resolved)
}

func (h *Handler) ListLogs(c *fiber.Ctx) error {
	logs, err := h.adminSvc.ListLogs(c.Context(), 200)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list logs"})
	}
	return c.JSON(logs)
}

func (h *Handler) actorFromCtx(c *fiber.Ctx) string {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return "system"
	}
	if username, ok := claims["username"].(string); ok && strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	return "system"
}

func boolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func (h *Handler) actorFromOptionalAuth(c *fiber.Ctx) string {
	authHeader := strings.TrimSpace(c.Get("Authorization"))
	if authHeader == "" {
		return ""
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenString == "" {
		return ""
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.jwtKey, nil
	})
	if err != nil || !token.Valid {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	if username, ok := claims["username"].(string); ok && strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	if name, ok := claims["name"].(string); ok && strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}

	return ""
}

func (h *Handler) authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "missing authorization header"})
	}

	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.jwtKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token claims"})
	}

	username := ""
	if val, ok := claims["username"].(string); ok {
		username = val
	} else if val, ok := claims["name"].(string); ok {
		username = val
	}
	if username != "" {
		banned, err := h.authSvc.IsBanned(c.Context(), username)
		if err == nil && banned {
			return c.Status(403).JSON(fiber.Map{"error": "account banned"})
		}
	}

	c.Locals("user", claims)
	return c.Next()
}

func (h *Handler) adminMiddleware(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(403).JSON(fiber.Map{"error": "access denied"})
	}
	isAdmin, _ := claims["admin"].(bool)
	if !isAdmin {
		// Some JWT libs marshal bools as float64
		if v, ok := claims["admin"].(float64); ok {
			isAdmin = v == 1
		}
	}
	if !isAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "admin required"})
	}
	return c.Next()
}
